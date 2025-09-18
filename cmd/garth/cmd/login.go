package cmd

import (
	"fmt"
	"log"

	"garmin-connect/internal/auth/credentials"
	"garmin-connect/pkg/garmin"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Garmin Connect",
	Long:  `Login to Garmin Connect using credentials from .env file and save the session.`, 
	Run: func(cmd *cobra.Command, args []string) {
		// Load credentials from .env file
		email, password, domain, err := credentials.LoadEnvCredentials()
		if err != nil {
			log.Fatalf("Failed to load credentials: %v", err)
		}

		// Create client
		garminClient, err := garmin.NewClient(domain)
		if err != nil {
			log.Fatalf("Failed to create client: %v", err)
		}

		// Try to load existing session first
		sessionFile := "garmin_session.json"
		if err := garminClient.LoadSession(sessionFile); err != nil {
			fmt.Println("No existing session found, logging in with credentials from .env...")

			if err := garminClient.Login(email, password); err != nil {
				log.Fatalf("Login failed: %v", err)
			}

			// Save session for future use
			if err := garminClient.SaveSession(sessionFile); err != nil {
				fmt.Printf("Failed to save session: %v\n", err)
			}
		} else {
			fmt.Println("Loaded existing session")
		}

		fmt.Println("Login successful!")
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
