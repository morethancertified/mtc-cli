package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/morethancertified/mtc-cli/internal/mtcapi"
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

		printTasksTable(lesson.Tasks)
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(submitCmd)
	submitCmd.Flags().BoolP("reset", "r", false, "Reset the lesson tasks")
}

func printTasksTable(tasks []types.Task) {
	fmt.Println("\nTASK STATUS:")
	fmt.Println("------------")
	for _, task := range tasks {
		status := "⚪"
		if task.Status == "COMPLETED" {
			status = "✅"
		} else if task.Status == "FAILED" {
			status = "❌"
		}

		fmt.Printf("%s %s\n", status, task.Title)
	}
	// t := table.NewWriter()
	// t.AppendHeader(table.Row{"Title", "Status"})
	// for _, task := range tasks {
	// 	t.AppendRow(table.Row{task.Title, task.Status})
	// }
	// t.SetStyle(table.StyleLight)
	// fmt.Println(t.Render())
}
