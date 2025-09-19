package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"go-garth/internal/auth/credentials"
	"go-garth/pkg/garmin"

	"github.com/spf13/cobra"
)

var tokensCmd = &cobra.Command{
	Use:   "tokens",
	Short: "Output OAuth tokens in JSON format",
	Long:  `Output the OAuth1 and OAuth2 tokens in JSON format after a successful login.`, 
	Run: func(cmd *cobra.Command, args []string) {
		// Load credentials from .env file
		_, _, domain, err := credentials.LoadEnvCredentials()
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
			log.Fatalf("No existing session found. Please run 'garth login' first.")
		}

		tokens := struct {
			OAuth1 *garmin.OAuth1Token `json:"oauth1"`
			OAuth2 *garmin.OAuth2Token `json:"oauth2"`
		}{
			OAuth1: garminClient.OAuth1Token(),
			OAuth2: garminClient.OAuth2Token(),
		}

		jsonBytes, err := json.MarshalIndent(tokens, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal tokens: %v", err)
		}
		fmt.Println(string(jsonBytes))
	},
}

func init() {
	rootCmd.AddCommand(tokensCmd)
}
