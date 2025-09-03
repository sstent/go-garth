package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sstent/go-garth"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Note: Using system environment variables (no .env file found)")
	}

	// Get credentials from environment
	username := os.Getenv("GARMIN_USERNAME")
	password := os.Getenv("GARMIN_PASSWORD")
	if username == "" || password == "" {
		log.Fatal("GARMIN_USERNAME or GARMIN_PASSWORD not set in environment")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create token storage and authenticator
	storage := garth.NewMemoryStorage()
	auth := garth.NewAuthenticator(garth.ClientOptions{
		Storage:  storage,
		TokenURL: "https://connectapi.garmin.com/oauth-service/oauth/token",
		Timeout:  30 * time.Second,
	})

	// Authenticate
	token, err := auth.Login(ctx, username, password, "")
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	log.Printf("Authenticated successfully! Token expires at: %s", token.Expiry.Format(time.RFC3339))

	// Create HTTP client with authentication transport
	httpClient := &http.Client{
		Transport: garth.NewAuthTransport(auth.(*garth.GarthAuthenticator), storage, nil),
	}

	// Create API client
	apiClient := garth.NewAPIClient("https://connectapi.garmin.com", httpClient)

	// Create activity service
	activityService := garth.NewActivityService(apiClient)

	// List last 20 activities
	activities, err := activityService.List(ctx, garth.ActivityListOptions{
		Limit: 20,
	})
	if err != nil {
		log.Fatalf("Failed to get activities: %v", err)
	}

	// Print activities
	fmt.Println("\nLast 20 Activities:")
	fmt.Println("=======================================")
	for i, activity := range activities {
		fmt.Printf("%d. %s [%s] - %s\n", i+1,
			activity.Name,
			activity.Type,
			activity.StartTime.Format("2006-01-02 15:04"))
		fmt.Printf("   Distance: %.2f km, Duration: %.0f min\n\n",
			activity.Distance/1000,
			activity.Duration/60)
	}

	fmt.Println("Example completed successfully!")
}
