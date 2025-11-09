package cmd

import (
	"github.com/spf13/cobra"
)

// gradeCmd is an alias for submitCmd
var gradeCmd = &cobra.Command{
	Use:     "grade [lesson-token]",
	Short:   "Grade a lesson (alias for submit)",
	Args:    cobra.MaximumNArgs(1),
	Example: "mtc grade\nmtc grade cm4ppz694200blze51ts1234",
	Run:     submitCmd.Run,
}

func init() {
	rootCmd.AddCommand(gradeCmd)
	gradeCmd.Flags().BoolP("reset", "r", false, "Reset the lesson tasks")
}
