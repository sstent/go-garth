package auth

import (
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/sstent/go-garth/garth/client"
)

var (
	email    string
	password string
)

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Garmin Connect",
	Long:  `Authenticate using Garmin Connect credentials and save session tokens`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if password == "" {
			fmt.Print("Enter password: ")
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return fmt.Errorf("password input failed: %w", err)
			}
			password = string(bytePassword)
			fmt.Println()
		}

		c, err := client.NewClient("")
		if err != nil {
			return fmt.Errorf("client creation failed: %w", err)
		}

		if err := c.Login(email, password); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		if err := c.SaveSession("session.json"); err != nil {
			return fmt.Errorf("session save failed: %w", err)
		}

		fmt.Println("Logged in successfully. Session saved to session.json")
		return nil
	},
}

func init() {
	LoginCmd.Flags().StringVarP(&email, "email", "e", "", "Garmin login email")
	LoginCmd.Flags().StringVarP(&password, "password", "p", "", "Garmin login password (optional, will prompt if empty)")

	// Mark required flags
	LoginCmd.MarkFlagRequired("email")
}
