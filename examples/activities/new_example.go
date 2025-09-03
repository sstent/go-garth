package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sstent/go-garth"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Note: Using system environment variables (no .env file found)")
	}

	// Get credentials
	username := os.Getenv("GARMIN_USERNAME")
	password := os.Getenv("GARMIN_PASSWORD")
	mfaToken := os.Getenv("GARMIN_MFA_TOKEN")
	if username == "" || password == "" {
		log.Fatal("GARMIN_USERNAME or GARMIN_PASSWORD not set in environment")
	}

	// Create client
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := garth.NewGarminClient(ctx, username, password, mfaToken)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Get profile
	profile, err := client.Profile.Get(ctx)
	if err != nil {
		log.Fatalf("Failed to get profile: %v", err)
	}
	log.Printf("User Profile: %s %s (%s)", profile.FirstName, profile.LastName, profile.UserID)

	// List activities
	activities, err := client.Activities.List(ctx, garth.ActivityListOptions{Limit: 5})
	if err != nil {
		log.Fatalf("Failed to get activities: %v", err)
	}

	log.Println("Recent Activities:")
	for _, activity := range activities {
		fmt.Printf("- %s: %s (%s)\n",
			activity.StartTime.Format("2006-01-02"),
			activity.Name,
			activity.Type)
	}
}
