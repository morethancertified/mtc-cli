package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/morethancertified/sprintctl/internal/mtcapi"
	"github.com/morethancertified/sprintctl/internal/styles"
	"github.com/morethancertified/sprintctl/internal/tui"
	"github.com/morethancertified/sprintctl/internal/types"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [lesson-token]",
	Short: "Show the status of a lesson (uses cached data if no token provided)",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		wd, err := os.Getwd()
		cobra.CheckErr(err)

		var lesson types.Lesson

		if len(args) > 0 {
			// Fresh API call with provided token
			lessonToken := args[0]
			
			// Read API base URL from config
			localConfigFile := filepath.Join(wd, ".sprint.json")
			var config struct {
				ApiBaseURL string `json:"api_base_url"`
			}
			
			if configData, err := os.ReadFile(localConfigFile); err == nil {
				json.Unmarshal(configData, &config)
			}
			
			if config.ApiBaseURL == "" {
				fmt.Println(styles.ErrorStyle.Render(" CONFIGURATION ERROR "))
				fmt.Println(styles.BoxStyle.Render("No API base URL configured.\nRun 'sprintctl submit' first to set up your configuration."))
				return
			}
			
			apiClient := mtcapi.New(config.ApiBaseURL)
			lesson, err = apiClient.GetLesson(lessonToken)
			if err != nil {
				fmt.Println(styles.ErrorStyle.Render(" API ERROR "), err)
				return
			}
			
			fmt.Println(styles.InfoStyle.Render(" FRESH STATUS "))
			fmt.Println(styles.BoxStyle.Render(fmt.Sprintf("Lesson Token: %s", lessonToken)))
		} else {
			// Use cached data (existing behavior)
			localConfigFile := filepath.Join(wd, ".sprint.json")

			// Check if config file exists
			if _, err := os.Stat(localConfigFile); os.IsNotExist(err) {
				fmt.Println(styles.WarningStyle.Render(" NO CACHED DATA "))
				fmt.Println(styles.BoxStyle.Render("No cached grading report found.\n\nOptions:\n• Run 'sprintctl submit <lesson-token>' to generate a report\n• Run 'sprintctl status <lesson-token>' to get fresh status"))
				return
			}

			// Read the config file
			configData, err := os.ReadFile(localConfigFile)
			if err != nil {
				fmt.Printf("Error reading config file: %s\n", err)
				return
			}

			var config struct {
				ApiBaseURL string       `json:"api_base_url"`
				LastLesson types.Lesson `json:"last_lesson"`
			}

			err = json.Unmarshal(configData, &config)
			if err != nil {
				fmt.Printf("Error parsing config file: %s\n", err)
				return
			}

			if len(config.LastLesson.Tasks) == 0 {
				fmt.Println("No lesson data found in cache.")
				fmt.Println("Run 'sprintctl submit <lesson-token>' to generate a report.")
				return
			}

			lesson = config.LastLesson
			fmt.Println(styles.InfoStyle.Render(" CACHED STATUS "))
			fmt.Println(styles.BoxStyle.Render("Using cached lesson data from last submission"))
		}

		// Launch TUI for interactive grading report
		// Note: status command doesn't support resubmit, so we pass empty token and ignore resubmit result
		_, err = tui.RunGradingReport(lesson.Tasks, "")
		if err != nil {
			// Fallback to simple text output if TUI fails
			fmt.Println(styles.WarningStyle.Render(" TUI ERROR "), err)
			fmt.Println(styles.SectionHeaderStyle.Render("📋 FALLBACK OUTPUT"))
			printTasksTable(lesson.Tasks)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
