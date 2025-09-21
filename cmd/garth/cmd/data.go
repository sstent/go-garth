package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"go-garth/internal/auth/credentials"
	"go-garth/pkg/garmin"

	"github.com/spf13/cobra"
)

var (
	dataDateStr    string
	dataDays       int
	dataOutputFile string
)

var dataCmd = &cobra.Command{
	Use:   "data [type]",
	Short: "Fetch various data types from Garmin Connect",
	Long:  `Fetch data such as bodybattery, sleep, HRV, and weight from Garmin Connect.`,
	Args:  cobra.ExactArgs(1), // Expects one argument: the data type
	Run: func(cmd *cobra.Command, args []string) {
		dataType := args[0]

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
		if dataDateStr != "" {
			parsedDate, err := time.Parse("2006-01-02", dataDateStr)
			if err != nil {
				log.Fatalf("Invalid date format: %v", err)
			}
			endDate = parsedDate
		}

		var result interface{}

		switch dataType {
		case "bodybattery":
			result, err = garminClient.GetBodyBatteryData(endDate)
		case "sleep":
			result, err = garminClient.GetSleepData(endDate)
		case "hrv":
			result, err = garminClient.GetHrvData(endDate)
		// case "weight":
		// 	result, err = garminClient.GetWeight(endDate)
		default:
			log.Fatalf("Unknown data type: %s", dataType)
		}

		if err != nil {
			log.Fatalf("Failed to get %s data: %v", dataType, err)
		}

		outputResult(result, dataOutputFile)
	},
}

func init() {
	rootCmd.AddCommand(dataCmd)

	dataCmd.Flags().StringVar(&dataDateStr, "date", "", "Date in YYYY-MM-DD format (default: yesterday)")
	dataCmd.Flags().StringVar(&dataOutputFile, "output", "", "Output file for JSON results")
	// dataCmd.Flags().IntVar(&dataDays, "days", 1, "Number of days to fetch") // Not used for single day data types
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
