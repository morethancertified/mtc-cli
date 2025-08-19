package cmd

import (
	"fmt"

	"github.com/morethancertified/mtc-cli/internal/auth"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Sign out of CloudSprints",
	Long:  "Remove stored authentication credentials and sign out of CloudSprints.",
	Run: func(cmd *cobra.Command, args []string) {
		if !auth.IsAuthenticated() {
			fmt.Println("You are not currently authenticated.")
			return
		}
		
		err := auth.Logout()
		if err != nil {
			fmt.Printf("Error signing out: %v\n", err)
			return
		}
		
		fmt.Println("✅ Successfully signed out!")
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
