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
	"github.com/morethancertified/mtc-cli/internal/mtcapi"
	"github.com/morethancertified/mtc-cli/internal/tui"
	"github.com/morethancertified/mtc-cli/internal/types"
	"github.com/morethancertified/mtc-cli/internal/widgets"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var submitCmd = &cobra.Command{
	Use:     "submit <lesson-token>",
	Short:   "Submit a lesson for grading",
	Args:    cobra.ExactArgs(1),
	Example: "mtc submit cm4ppz694200blze51ts1234",
	Run: func(cmd *cobra.Command, args []string) {
		// Check for a local project config file and create one if it doesn't exist.
		wd, err := os.Getwd()
		cobra.CheckErr(err)
		localConfigFile := filepath.Join(wd, ".sprint.json")

		if _, err := os.Stat(localConfigFile); os.IsNotExist(err) {
			fmt.Println("First time submitting for this project.")
			fmt.Println("Please select the platform this lab is for:")

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

			// Create the config map and save it to .mtc.json
			config := map[string]interface{}{"api_base_url": selectedURL}
			file, err := json.MarshalIndent(config, "", "  ")
			cobra.CheckErr(err)

			err = os.WriteFile(localConfigFile, file, 0644)
			cobra.CheckErr(err)

			// Set the value for the current run and merge in the new config
			viper.Set("api_base_url", selectedURL)
			viper.MergeInConfig() // Re-read to ensure it's loaded for this session
			fmt.Println("Configuration saved to", localConfigFile)
			fmt.Println("------------------------------------------------------------------")
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
				fmt.Println("Error resetting lesson:", err)
				return
			}
			fmt.Println("\nLesson reset!")
			printTasksTable(lesson.Tasks)
			return
		}

		printTasksTable(lesson.Tasks)
		fmt.Println("\nWe will now run the following command(s) to validate your lesson:")
		fmt.Println("------------------------------------------------------------------")
		for _, command := range lesson.CliCommands {
			fmt.Println(command)
		}
		fmt.Println("------------------------------------------------------------------")

		input := confirmation.New("Continue?", confirmation.Yes)
		ready, err := input.RunPrompt()
		if err != nil {
			fmt.Println("Error getting confirmation:", err)
			return
		}
		if !ready {
			fmt.Println("Aborting...")
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
			fmt.Println("Error submitting lesson:", err)
			return
		}

		fmt.Println("\nGrading complete!")

		// Debug: Print what we got from the API
		fmt.Printf("DEBUG - Received %d tasks from API\n", len(lesson.Tasks))
		for i, task := range lesson.Tasks {
			fmt.Printf("Task %d: %s (Status: %s, AI Explanation length: %d)\n", 
				i+1, task.Title, task.Status, len(task.AiExplanation))
		}

		// Cache lesson data for status command
		err = cacheLessonData(lesson, localConfigFile)
		if err != nil {
			fmt.Printf("Warning: Could not cache lesson data: %s\n", err)
		}

		// Launch TUI for interactive grading report
		err = tui.RunGradingReport(lesson.Tasks)
		if err != nil {
			// Fallback to table view if TUI fails
			printTasksTable(lesson.Tasks)
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(submitCmd)
	submitCmd.Flags().BoolP("reset", "r", false, "Reset the lesson tasks")
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
	fmt.Println("\nTASK STATUS:")
	fmt.Println("------------")
	for _, task := range tasks {
		var status string
		switch task.Status {
		case "COMPLETED":
			status = "✅"
		case "FAILED":
			status = "❌"
		default:
			status = "⏳"
		}
		fmt.Printf("%s %s\n", status, task.Title)
	}

	// t := table.New()
	// t.SetHeaders("Task", "Status")
	// for _, task := range tasks {
	// 	t.AddRow(task.Title, task.Status)
	// }
	// t.SetStyle(table.StyleLight)
	// fmt.Println(t.Render())
}
