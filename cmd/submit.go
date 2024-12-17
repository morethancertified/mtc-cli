package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/morethancertified/mtc-cli/internal/mtcapi"
	"github.com/morethancertified/mtc-cli/internal/types"
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

		fmt.Println("\nCommands to run")
		fmt.Println("------")
		for _, command := range lesson.CliCommands {
			fmt.Println(command)
		}

		printTasksTable(lesson.Tasks)

		input := confirmation.New("Ready to run commands?", confirmation.Yes)
		ready, err := input.RunPrompt()
		if err != nil {
			fmt.Println("Error getting confirmation:", err)
			return
		}
		if !ready {
			fmt.Println("Aborting...")
			return
		}

		// widgets.RunProgressBar()

		cliCommandResults := []types.CLICommandResult{}
		for _, command := range lesson.CliCommands {
			cliCommandResult := types.CLICommandResult{
				Command: command,
			}

			cmd := exec.Command("sh", "-c", "LANG=en_US.UTF-8 "+command)

			b, err := cmd.Output()
			fmt.Printf("\nRan command: %s\n", cmd.String())
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

		fmt.Println("\nLesson submitted!")
		printTasksTable(lesson.Tasks)
	},
}

func init() {
	rootCmd.AddCommand(submitCmd)
	submitCmd.Flags().BoolP("reset", "r", false, "Reset the lesson tasks")
}

func printTasksTable(tasks []types.Task) {
	fmt.Println("\nTASKS:")
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Title", "Status"})
	for _, task := range tasks {
		t.AppendRow(table.Row{task.Title, task.Status})
	}
	t.SetStyle(table.StyleRounded)
	fmt.Println(t.Render())
}
