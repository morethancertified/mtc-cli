package cmd

import (
	"fmt"

	"github.com/morethancertified/mtc-cli/internal/auth"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
	Long:  "Commands for managing authentication with CloudSprints.",
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	Run: func(cmd *cobra.Command, args []string) {
		if auth.IsAuthenticated() {
			fmt.Println("✅ Authenticated")
			fmt.Println("You are signed in to CloudSprints.")
		} else {
			fmt.Println("❌ Not authenticated")
			fmt.Println("Run 'sprintctl login' to sign in.")
		}
	},
}

var authTokenCmd = &cobra.Command{
	Use:   "token <access_token>",
	Short: "Manually set authentication token (beta)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		token := args[0]
		
		if len(token) < 10 {
			fmt.Println("❌ Invalid token format")
			return
		}
		
		err := auth.StoreToken(token)
		if err != nil {
			fmt.Printf("❌ Error storing token: %v\n", err)
			return
		}
		
		fmt.Println("✅ Token stored successfully!")
		fmt.Println("You can now use sprintctl commands.")
	},
}

func init() {
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authTokenCmd)
	rootCmd.AddCommand(authCmd)
}
