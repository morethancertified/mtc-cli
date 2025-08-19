package cmd

import (
	"fmt"
	"strings"

	"github.com/erikgeiser/promptkit/textinput"
	"github.com/morethancertified/mtc-cli/internal/auth"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login [email]",
	Short: "Authenticate with CloudSprints using email OTP",
	Long: `Authenticate with CloudSprints using a one-time password sent to your email.
You can provide your email as an argument or enter it when prompted.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var email string
		
		// Get email from args or prompt
		if len(args) > 0 {
			email = args[0]
		} else {
			input := textinput.New("Enter your email:")
			input.Placeholder = "user@example.com"
			
			var err error
			email, err = input.RunPrompt()
			if err != nil {
				fmt.Printf("Error getting email: %v\n", err)
				return
			}
		}
		
		// Validate email format (basic check)
		if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
			fmt.Println("Please enter a valid email address")
			return
		}
		
		// Check if already authenticated
		if auth.IsAuthenticated() {
			fmt.Println("You are already authenticated!")
			fmt.Println("Run 'sprintctl logout' to sign out first.")
			return
		}
		
		// Request OTP
		fmt.Printf("Sending OTP to %s...\n", email)
		err := auth.LoginWithOTP(email)
		if err != nil {
			fmt.Printf("Error sending OTP: %v\n", err)
			return
		}
		
		fmt.Println("✅ OTP sent! Check your email.")
		
		// Prompt for OTP code
		codeInput := textinput.New("Enter the 6-digit code from your email:")
		codeInput.Placeholder = "123456"
		
		code, err := codeInput.RunPrompt()
		if err != nil {
			fmt.Printf("Error getting code: %v\n", err)
			return
		}
		
		// Verify OTP
		fmt.Println("Verifying code...")
		err = auth.VerifyOTP(email, code)
		if err != nil {
			fmt.Printf("❌ Authentication failed: %v\n", err)
			return
		}
		
		fmt.Println("✅ Successfully authenticated!")
		fmt.Println("You can now use sprintctl to submit lessons and access your data.")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
