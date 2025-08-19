package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/morethancertified/mtc-cli/internal/tui"
	"github.com/morethancertified/mtc-cli/internal/types"
	"github.com/spf13/cobra"
	"github.com/morethancertified/mtc-cli/internal/mtcapi"
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
				fmt.Println("No API base URL configured. Run 'sprintctl submit' first.")
				return
			}
			
			apiClient := mtcapi.New(config.ApiBaseURL)
			lesson, err = apiClient.GetLesson(lessonToken)
			if err != nil {
				fmt.Printf("Error getting lesson status: %s\n", err)
				return
			}
			
			fmt.Printf("Fresh status for lesson %s:\n", lessonToken)
		} else {
			// Use cached data (existing behavior)
			localConfigFile := filepath.Join(wd, ".sprint.json")

			// Check if config file exists
			if _, err := os.Stat(localConfigFile); os.IsNotExist(err) {
				fmt.Println("No cached grading report found.")
				fmt.Println("Run 'sprintctl submit <lesson-token>' to generate a report.")
				fmt.Println("Or run 'sprintctl status <lesson-token>' to get fresh status.")
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
			fmt.Println("Cached lesson status:")
		}

		// Launch TUI for interactive grading report
		// Note: status command doesn't support resubmit, so we pass empty token and ignore resubmit result
		_, err = tui.RunGradingReport(lesson.Tasks, "")
		if err != nil {
			// Fallback to simple text output if TUI fails
			fmt.Printf("Error launching TUI: %s\n", err)
			fmt.Println("\nFalling back to simple output:")
			printTasksTable(lesson.Tasks)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
