package main

import (
	"fmt"
	"log"
	"time"

	"garmin-connect/garth/client"
	"garmin-connect/garth/credentials"
	"garmin-connect/garth/types"
)

func main() {
	// Load credentials from .env file
	email, password, domain, err := credentials.LoadEnvCredentials()
	if err != nil {
		log.Fatalf("Failed to load credentials: %v", err)
	}

	// Create client
	garminClient, err := client.NewClient(domain)
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

	// Test getting activities
	activities, err := garminClient.GetActivities(5)
	if err != nil {
		log.Fatalf("Failed to get activities: %v", err)
	}

	// Display activities
	displayActivities(activities)
}

func displayActivities(activities []types.Activity) {
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
