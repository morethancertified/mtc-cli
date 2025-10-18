package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/morethancertified/sprintctl/internal/mtcapi"
	"github.com/morethancertified/sprintctl/internal/styles"
	"github.com/morethancertified/sprintctl/internal/tui"
	"github.com/morethancertified/sprintctl/internal/types"
	"github.com/morethancertified/sprintctl/internal/widgets"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var submitCmd = &cobra.Command{
	Use:     "submit <lesson-token>",
	Short:   "Submit a lesson for grading",
	Args:    cobra.ExactArgs(1),
	Example: "sprintctl submit cm4ppz694200blze51ts1234",
	Run: func(cmd *cobra.Command, args []string) {
		// Check for a local project config file and create one if it doesn't exist.
		wd, err := os.Getwd()
		cobra.CheckErr(err)
		localConfigFile := filepath.Join(wd, ".sprint.json")

		if _, err := os.Stat(localConfigFile); os.IsNotExist(err) {
			fmt.Println(styles.InfoStyle.Render(" FIRST TIME SETUP "))
			fmt.Println(styles.BoxStyle.Render("Please select the platform this lab is for:"))

			// Create a map of display names to API URLs
			platformMap := map[string]string{
				"New Learning Platform (https://labs.morethancertified.com/api/v1)": "https://labs.morethancertified.com/api/v1",
				"Legacy Video Platform (https://app.morethancertified.com/api/v1)":  "https://app.morethancertified.com/api/v1",
				"CloudSprints (https://cloudsprints.com/api/v1)": "https://cloudsprints.com/api/v1",
			}

			// Create simple string choices for clean display
			platformOptions := []string{
				"New Learning Platform (https://labs.morethancertified.com/api/v1)",
				"Legacy Video Platform (https://app.morethancertified.com/api/v1)",
				"CloudSprints (https://cloudsprints.com/api/v1)",
			}

			sp := selection.New("Choose the platform:", platformOptions)
			choice, err := sp.RunPrompt()
			cobra.CheckErr(err)

			// Look up the URL for the selected platform
			selectedURL := platformMap[choice]

			// Create the config map and save it to .sprint.json
			config := map[string]interface{}{"api_base_url": selectedURL}
			file, err := json.MarshalIndent(config, "", "  ")
			cobra.CheckErr(err)

			err = os.WriteFile(localConfigFile, file, 0644)
			cobra.CheckErr(err)

			// Set the value for the current run and merge in the new config
			viper.Set("api_base_url", selectedURL)
			viper.MergeInConfig() // Re-read to ensure it's loaded for this session
			fmt.Println(styles.SuccessStyle.Render(" CONFIGURATION SAVED "))
			fmt.Println(styles.BoxStyle.Render(fmt.Sprintf("Configuration saved to: %s", localConfigFile)))
			fmt.Println(styles.Separator(60))
		}

		lessonToken := args[0]
		reset, _ := cmd.Flags().GetBool("reset")
		apiClient := mtcapi.New(viper.GetString("api_base_url"))
		lesson, err := apiClient.GetLesson(lessonToken)
		if err != nil {
			fmt.Println("Error getting lesson:", err)
			return
		}

		if reset {
			lesson, err = apiClient.ResetLesson(lessonToken)
			if err != nil {
				fmt.Println(styles.ErrorStyle.Render(" ERROR "), "resetting lesson:", err)
				return
			}
			fmt.Println(styles.SuccessStyle.Render(" LESSON RESET "))
			printTasksTable(lesson.Tasks)
			return
		}

		printTasksTable(lesson.Tasks)
		
		// Run the submission flow
		runSubmissionFlow(lessonToken, apiClient)
	},
}

func init() {
	rootCmd.AddCommand(submitCmd)
	submitCmd.Flags().BoolP("reset", "r", false, "Reset the lesson tasks")
}

func runSubmissionFlow(lessonToken string, apiClient *mtcapi.MtcApiClient) {
	// Get current working directory for config file
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting working directory:", err)
		return
	}
	localConfigFile := filepath.Join(wd, ".sprint.json")

	// Get fresh lesson data
	lesson, err := apiClient.GetLesson(lessonToken)
	if err != nil {
		fmt.Println("Error getting lesson:", err)
		return
	}

	fmt.Println(styles.SectionHeaderStyle.Render("🔍 VALIDATION COMMANDS"))
	fmt.Println(styles.BoxStyle.Render("The following commands will be executed to validate your lesson:"))
	
	commandsText := ""
	for _, command := range lesson.CliCommands {
		commandsText += styles.CommandStyle.Render(command) + "\n"
	}
	fmt.Println(styles.CodeBlockStyle.Render(commandsText))
	fmt.Println(styles.Separator(60))

	input := confirmation.New("Continue?", confirmation.Yes)
	ready, err := input.RunPrompt()
	if err != nil {
		fmt.Println("Error getting confirmation:", err)
		return
	}
	if !ready {
		fmt.Println(styles.WarningStyle.Render(" ABORTED "))
		return
	}

	widgets.RunProgressBar()

	cliCommandResults := []types.CLICommandResult{}
	for _, command := range lesson.CliCommands {
		cliCommandResult := types.CLICommandResult{
			Command: command,
		}

		cmd := exec.Command("sh", "-c", "LANG=en_US.UTF-8 "+command)

		b, err := cmd.Output()
		if ee, ok := err.(*exec.ExitError); ok {
			cliCommandResult.ExitCode = ee.ExitCode()
			cliCommandResult.Stderr = strings.TrimRight(string(ee.Stderr), "\n\t\r")
		} else if err != nil {
			cliCommandResult.ExitCode = -69
		} else {
			cliCommandResult.Stdout = strings.TrimRight(string(b), "\n\t\r")
		}

		cliCommandResults = append(cliCommandResults, cliCommandResult)
	}

	lesson, err = apiClient.SubmitLesson(lessonToken, cliCommandResults)
	if err != nil {
		fmt.Println(styles.ErrorStyle.Render(" SUBMISSION ERROR "), err)
		return
	}

	fmt.Println(styles.SuccessStyle.Render(" GRADING COMPLETE! "))

	// Show grading results summary
	completed := 0
	failed := 0
	for _, task := range lesson.Tasks {
		switch task.Status {
		case "COMPLETED":
			completed++
		case "FAILED":
			failed++
		}
	}
	
	summaryText := fmt.Sprintf("Tasks Completed: %d/%d", completed, len(lesson.Tasks))
	if failed > 0 {
		summaryText += fmt.Sprintf(" | Failed: %d", failed)
	}
	
	fmt.Println(styles.BoxStyle.Render(summaryText))
	fmt.Println(styles.ProgressBar(completed, len(lesson.Tasks), 40))

	// Cache lesson data for status command
	err = cacheLessonData(lesson, localConfigFile)
	if err != nil {
		fmt.Printf("Warning: Could not cache lesson data: %s\n", err)
	}

	// Launch TUI for interactive grading report
	shouldResubmit, err := tui.RunGradingReport(lesson.Tasks, lessonToken)
	if err != nil {
		// Fallback to table view if TUI fails
		fmt.Println(styles.WarningStyle.Render(" TUI ERROR "))
		printTasksTable(lesson.Tasks)
		fmt.Println()
	} else if shouldResubmit {
		// User wants to resubmit - run the submission flow again
		fmt.Println(styles.InfoStyle.Render(" RESUBMITTING LESSON "))
		runSubmissionFlow(lessonToken, apiClient)
	}
}

func cacheLessonData(lesson types.Lesson, configFile string) error {
	// Read existing config
	var config map[string]interface{}
	if data, err := os.ReadFile(configFile); err == nil {
		json.Unmarshal(data, &config)
	} else {
		config = make(map[string]interface{})
	}

	// Add lesson data to config
	config["last_lesson"] = lesson

	// Write back to file
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}

func printTasksTable(tasks []types.Task) {
	fmt.Println(styles.SectionHeaderStyle.Render("📋 TASK STATUS"))
	
	for _, task := range tasks {
		statusIcon := styles.StatusIcon(task.Status)
		statusStyle := styles.StatusText(task.Status)
		
		taskLine := fmt.Sprintf("%s %s %s", 
			statusIcon, 
			task.Title,
			statusStyle.Render(task.Status))
		
		fmt.Println(styles.ListItemStyle.Render(taskLine))
	}
	fmt.Println()
}
