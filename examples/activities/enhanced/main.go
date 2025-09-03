package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sstent/go-garth"
)

func main() {
	// Load environment variables from .env file (robust path handling)
	var envPath string
	if _, err := os.Stat("../../../.env"); err == nil {
		envPath = "../../../.env"
	} else if _, err := os.Stat("../../.env"); err == nil {
		envPath = "../../.env"
	} else {
		envPath = ".env"
	}

	if err := godotenv.Load(envPath); err != nil {
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

	// Example 1: List recent activities
	fmt.Println("\n=== RECENT ACTIVITIES ===")
	recentActivities, err := activityService.List(ctx, garth.ActivityListOptions{
		Limit: 10,
	})
	if err != nil {
		log.Printf("Failed to get recent activities: %v", err)
	} else {
		printActivities(recentActivities, "Recent Activities")
	}

	// Example 2: List activities from last 30 days
	fmt.Println("\n=== ACTIVITIES FROM LAST 30 DAYS ===")
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	recentMonthActivities, err := activityService.List(ctx, garth.ActivityListOptions{
		StartDate: thirtyDaysAgo,
		EndDate:   time.Now(),
		Limit:     20,
	})
	if err != nil {
		log.Printf("Failed to get activities from last 30 days: %v", err)
	} else {
		printActivities(recentMonthActivities, "Last 30 Days")
	}

	// Example 3: List running activities only
	fmt.Println("\n=== RUNNING ACTIVITIES ===")
	runningActivities, err := activityService.List(ctx, garth.ActivityListOptions{
		ActivityType: "running",
		Limit:        15,
	})
	if err != nil {
		log.Printf("Failed to get running activities: %v", err)
	} else {
		printActivities(runningActivities, "Running Activities")
	}

	// Example 4: Search activities by name
	fmt.Println("\n=== ACTIVITIES CONTAINING 'MORNING' ===")
	morningActivities, err := activityService.List(ctx, garth.ActivityListOptions{
		NameContains: "morning",
		Limit:        10,
	})
	if err != nil {
		log.Printf("Failed to search activities: %v", err)
	} else {
		printActivities(morningActivities, "Activities with 'morning' in name")
	}

	// Example 5: Get detailed information for the first activity
	if len(recentActivities) > 0 {
		firstActivity := recentActivities[0]
		fmt.Printf("\n=== DETAILS FOR ACTIVITY: %s ===\n", firstActivity.Name)

		details, err := activityService.Get(ctx, firstActivity.ActivityID)
		if err != nil {
			log.Printf("Failed to get activity details: %v", err)
		} else {
			printActivityDetails(details)
		}
	}

	fmt.Println("\nEnhanced activity listing example completed successfully!")
}

func printActivities(activities []garth.Activity, title string) {
	if len(activities) == 0 {
		fmt.Printf("No %s found\n", title)
		return
	}

	fmt.Printf("\n%s (%d total):\n", title, len(activities))
	fmt.Println(strings.Repeat("=", 50))

	for i, activity := range activities {
		fmt.Printf("%d. %s [%s]\n", i+1, activity.Name, activity.Type)
		fmt.Printf("   Date: %s\n", activity.StartTime.Format("2006-01-02 15:04"))
		fmt.Printf("   Distance: %.2f km\n", activity.Distance/1000)
		fmt.Printf("   Duration: %s\n", formatDuration(activity.Duration))
		fmt.Printf("   Calories: %d\n", activity.Calories)
		fmt.Println()
	}
}

func printActivityDetails(details *garth.ActivityDetails) {
	fmt.Printf("Activity ID: %d\n", details.ActivityID)
	fmt.Printf("Name: %s\n", details.Name)
	fmt.Printf("Type: %s\n", details.Type)
	fmt.Printf("Description: %s\n", details.Description)
	fmt.Printf("Start Time: %s\n", details.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Distance: %.2f km\n", details.Distance/1000)
	fmt.Printf("Duration: %s\n", formatDuration(details.Duration))
	fmt.Printf("Calories: %d\n", details.Calories)

	if details.ElevationGain > 0 {
		fmt.Printf("Elevation Gain: %.0f m\n", details.ElevationGain)
	}
	if details.ElevationLoss > 0 {
		fmt.Printf("Elevation Loss: %.0f m\n", details.ElevationLoss)
	}
	if details.MaxHeartRate > 0 {
		fmt.Printf("Max Heart Rate: %d bpm\n", details.MaxHeartRate)
	}
	if details.AvgHeartRate > 0 {
		fmt.Printf("Avg Heart Rate: %d bpm\n", details.AvgHeartRate)
	}
	if details.MaxSpeed > 0 {
		fmt.Printf("Max Speed: %.1f km/h\n", details.MaxSpeed*3.6)
	}
	if details.AvgSpeed > 0 {
		fmt.Printf("Avg Speed: %.1f km/h\n", details.AvgSpeed*3.6)
	}
	if details.Steps > 0 {
		fmt.Printf("Steps: %d\n", details.Steps)
	}
}

func formatDuration(seconds float64) string {
	duration := time.Duration(seconds * float64(time.Second))
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
