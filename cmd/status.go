package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/morethancertified/mtc-cli/internal/tui"
	"github.com/morethancertified/mtc-cli/internal/types"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of the last graded lesson",
	Run: func(cmd *cobra.Command, args []string) {
		wd, err := os.Getwd()
		cobra.CheckErr(err)

		localConfigFile := filepath.Join(wd, ".sprint.json")

		// Check if config file exists
		if _, err := os.Stat(localConfigFile); os.IsNotExist(err) {
			fmt.Println("No cached grading report found.")
			fmt.Println("Run 'sprintctl submit <lesson-token>' to generate a report.")
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

		// Check if we have cached lesson data
		if config.LastLesson.ID == "" {
			fmt.Println("No cached grading report found.")
			fmt.Println("Run 'sprintctl submit <lesson-token>' to generate a report.")
			return
		}

		// Launch TUI for interactive grading report
		err = tui.RunGradingReport(config.LastLesson.Tasks)
		if err != nil {
			// Fallback to simple text output if TUI fails
			fmt.Printf("Error launching TUI: %s\n", err)
			fmt.Println("\nFalling back to simple output:")
			printTasksTable(config.LastLesson.Tasks)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
