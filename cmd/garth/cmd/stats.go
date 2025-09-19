package cmd

import (
	"log"
	"time"

	"go-garth/internal/auth/credentials"
	"go-garth/pkg/garmin"

	"github.com/spf13/cobra"
)

var (
	statsDateStr string
	statsDays int
	statsOutputFile string
)

var statsCmd = &cobra.Command{
	Use:   "stats [type]",
	Short: "Fetch various stats types from Garmin Connect",
	Long:  `Fetch stats such as steps, stress, hydration, intensity, sleep, and HRV from Garmin Connect.`, 
	Args:  cobra.ExactArgs(1), // Expects one argument: the stats type
	Run: func(cmd *cobra.Command, args []string) {
		statsType := args[0]

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

		endDate := time.Now().AddDate(0, 0, -1) // default to yesterday
		if statsDateStr != "" {
			parsedDate, err := time.Parse("2006-01-02", statsDateStr)
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

		result, err := stats.List(endDate, statsDays, garminClient.Client)
		if err != nil {
			log.Fatalf("Failed to get %s stats: %v", statsType, err)
		}

		outputResult(result, statsOutputFile)
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)

	statsCmd.Flags().StringVar(&statsDateStr, "date", "", "Date in YYYY-MM-DD format (default: yesterday)")
	statsCmd.Flags().IntVar(&statsDays, "days", 1, "Number of days to fetch")
	statsCmd.Flags().StringVar(&statsOutputFile, "output", "", "Output file for JSON results")
}
