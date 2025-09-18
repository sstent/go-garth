package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"garmin-connect/pkg/garmin"
	"garmin-connect/internal/auth/credentials"
)

func main() {
	// Parse command line flags
	var outputTokens = flag.Bool("tokens", false, "Output OAuth tokens in JSON format")
	var dataType = flag.String("data", "", "Data type to fetch (bodybattery, sleep, hrv, weight)")
	var statsType = flag.String("stats", "", "Stats type to fetch (steps, stress, hydration, intensity, sleep, hrv)")

	var dateStr = flag.String("date", "", "Date in YYYY-MM-DD format (default: yesterday)")
	var days = flag.Int("days", 1, "Number of days to fetch")
	var outputFile = flag.String("output", "", "Output file for JSON results")
	flag.Parse()

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

	// If tokens flag is set, output tokens and exit
	if *outputTokens {
		outputTokensJSON(garminClient)
		return
	}

	// Handle data requests
	if *dataType != "" {
		handleDataRequest(garminClient, *dataType, *dateStr, *days, *outputFile)
		return
	}

	// Handle stats requests
	if *statsType != "" {
		handleStatsRequest(garminClient, *statsType, *dateStr, *days, *outputFile)
		return
	}

	// Default: show recent activities
	activities, err := garminClient.GetActivities(5)
	if err != nil {
		log.Fatalf("Failed to get activities: %v", err)
	}
	displayActivities(activities)
}

func outputTokensJSON(c *garmin.Client) {
	tokens := struct {
		OAuth1 *garmin.OAuth1Token `json:"oauth1"`
		OAuth2 *garmin.OAuth2Token `json:"oauth2"`
	}{
		OAuth1: c.OAuth1Token(),
		OAuth2: c.OAuth2Token(),
	}

	jsonBytes, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal tokens: %v", err)
	}
	fmt.Println(string(jsonBytes))
}

func handleDataRequest(c *garmin.Client, dataType, dateStr string, days int, outputFile string) {
	endDate := time.Now().AddDate(0, 0, -1) // default to yesterday
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			log.Fatalf("Invalid date format: %v", err)
		}
		endDate = parsedDate
	}

	var result interface{}
	var err error

	switch dataType {
	case "bodybattery":
		result, err = c.GetBodyBattery(endDate)
	case "sleep":
		result, err = c.GetSleep(endDate)
	case "hrv":
		result, err = c.GetHRV(endDate)
	case "weight":
		result, err = c.GetWeight(endDate)
	default:
		log.Fatalf("Unknown data type: %s", dataType)
	}

	if err != nil {
		log.Fatalf("Failed to get %s data: %v", dataType, err)
	}

	outputResult(result, outputFile)
}

func handleStatsRequest(c *garmin.Client, statsType, dateStr string, days int, outputFile string) {
	endDate := time.Now().AddDate(0, 0, -1) // default to yesterday
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			log.Fatalf("Invalid date format: %v", err)
		}
		endDate = parsedDate
	}

	var stats garmin.Stats
	switch statsType {
	case "steps":
		stats = garmin.NewDailySteps()
	case "stress":
		stats = garmin.NewDailyStress()
	case "hydration":
		stats = garmin.NewDailyHydration()
	case "intensity":
		stats = garmin.NewDailyIntensityMinutes()
	case "sleep":
		stats = garmin.NewDailySleep()
	case "hrv":
		stats = garmin.NewDailyHRV()
	default:
		log.Fatalf("Unknown stats type: %s", statsType)
	}

	result, err := stats.List(endDate, days, c.Client)
	if err != nil {
		log.Fatalf("Failed to get %s stats: %v", statsType, err)
	}

	outputResult(result, outputFile)
}

func outputResult(data interface{}, outputFile string) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal result: %v", err)
	}

	if outputFile != "" {
		if err := os.WriteFile(outputFile, jsonBytes, 0644); err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
		fmt.Printf("Results saved to %s\n", outputFile)
	} else {
		fmt.Println(string(jsonBytes))
	}
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