package cmd

import (
	"fmt"
	"log"
	"time"

	"go-garth/internal/auth/credentials"
	"go-garth/pkg/garmin"

	"github.com/spf13/cobra"
)

var activitiesCmd = &cobra.Command{
	Use:   "activities",
	Short: "Display recent Garmin Connect activities",
	Long:  `Fetches and displays a list of recent activities from Garmin Connect.`,
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

		opts := garmin.ActivityOptions{
			Limit: 5,
		}
		activities, err := garminClient.ListActivities(opts)
		if err != nil {
			log.Fatalf("Failed to get activities: %v", err)
		}
		displayActivities(activities)
	},
}

func init() {
	rootCmd.AddCommand(activitiesCmd)
}

func displayActivities(activities []garmin.Activity) {
	fmt.Printf("\n=== Recent Activities ===\n")
	for i, activity := range activities {
		fmt.Printf("%d. %s\n", i+1, activity.ActivityName)
		fmt.Printf("   Type: %s\n", activity.ActivityType.TypeKey)
		fmt.Printf("   Date: %s\n", activity.StartTimeLocal)
		if activity.Distance > 0 {
			fmt.Printf("   Distance: %.2f km\n", activity.Distance/1000)
		}
		if activity.Duration > 0 {
			duration := time.Duration(activity.Duration) * time.Second
			fmt.Printf("   Duration: %v\n", duration.Round(time.Second))
		}
		fmt.Println()
	}
}
