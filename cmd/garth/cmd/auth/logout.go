package auth

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var LogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and remove saved session",
	Long:  `Remove the saved session file and clear authentication tokens`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionPath, _ := cmd.Flags().GetString("session")
		if sessionPath == "" {
			sessionPath = "session.json"
		}

		// Check if session file exists
		if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
			fmt.Printf("No session file found at %s\n", sessionPath)
			return nil
		}

		// Remove session file
		if err := os.Remove(sessionPath); err != nil {
			return fmt.Errorf("failed to remove session file: %w", err)
		}

		fmt.Printf("Logged out successfully. Session file removed: %s\n", sessionPath)
		return nil
	},
}

func init() {
	LogoutCmd.Flags().StringP("session", "s", "session.json", "Session file path to remove")
}
