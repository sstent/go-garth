package auth

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/sstent/go-garth/garth/client"
)

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long:  `Check if you are currently authenticated with Garmin Connect`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionPath, _ := cmd.Flags().GetString("session")
		if sessionPath == "" {
			sessionPath = "session.json"
		}

		// Check if session file exists
		if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
			fmt.Printf("❌ Not authenticated\n")
			fmt.Printf("Session file not found: %s\n", sessionPath)
			fmt.Printf("Run 'garth auth login' to authenticate\n")
			return nil
		}

		// Try to create client and load session
		c, err := client.NewClient("")
		if err != nil {
			return fmt.Errorf("client creation failed: %w", err)
		}

		if err := c.LoadSession(sessionPath); err != nil {
			fmt.Printf("❌ Session file exists but is invalid\n")
			fmt.Printf("Error: %v\n", err)
			fmt.Printf("Run 'garth auth login' to re-authenticate\n")
			return nil
		}

		// Try to make a simple authenticated request to verify session
		if _, err := c.GetUserProfile(); err != nil {
			fmt.Printf("❌ Session exists but authentication failed\n")
			fmt.Printf("Error: %v\n", err)
			fmt.Printf("Run 'garth auth login' to re-authenticate\n")
			return nil
		}

		fmt.Printf("✅ Authenticated\n")
		fmt.Printf("Session file: %s\n", sessionPath)
		return nil
	},
}

func init() {
	StatusCmd.Flags().StringP("session", "s", "session.json", "Session file path to check")
}
