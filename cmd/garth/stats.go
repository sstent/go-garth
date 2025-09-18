package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"garmin-connect/pkg/garmin"
)

var (
	statsCmd = &cobra.Command{
		Use:   "stats",
		Short: "Manage Garmin Connect statistics",
		Long:  `Provides commands to fetch various statistics like steps, distance, and calories.`,
	}

	stepsCmd = &cobra.Command{
		Use:   "steps",
		Short: "Get steps statistics",
		Long:  `Fetch steps statistics for a specified period.`,
		RunE:  runSteps,
	}

	distanceCmd = &cobra.Command{
		Use:   "distance",
		Short: "Get distance statistics",
		Long:  `Fetch distance statistics for a specified period.`,
		RunE:  runDistance,
	}

	caloriesCmd = &cobra.Command{
		Use:   "calories",
		Short: "Get calories statistics",
		Long:  `Fetch calories statistics for a specified period.`,
		RunE:  runCalories,
	}

	// Flags for stats commands
	statsMonth bool
	statsYear  bool
	statsFrom  string
)

func init() {
	rootCmd.AddCommand(statsCmd)

	statsCmd.AddCommand(stepsCmd)
	stepsCmd.Flags().BoolVar(&statsMonth, "month", false, "Fetch data for the current month")

	statsCmd.AddCommand(distanceCmd)
	distanceCmd.Flags().BoolVar(&statsYear, "year", false, "Fetch data for the current year")

	statsCmd.AddCommand(caloriesCmd)
	caloriesCmd.Flags().StringVar(&statsFrom, "from", "", "Start date for data fetching (YYYY-MM-DD)")
}

func runSteps(cmd *cobra.Command, args []string) error {
	garminClient, err := garmin.NewClient("www.garmin.com") // TODO: Domain should be configurable
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	sessionFile := "garmin_session.json" // TODO: Make session file configurable
	if err := garminClient.LoadSession(sessionFile); err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	var startDate, endDate time.Time
	if statsMonth {
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, -1) // Last day of the month
	} else {
		// Default to today if no specific range or month is given
		startDate = time.Now()
		endDate = time.Now()
	}

	stepsData, err := garminClient.GetStepsData(startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to get steps data: %w", err)
	}

	if len(stepsData) == 0 {
		fmt.Println("No steps data found.")
		return nil
	}

	fmt.Println("Steps Data:")
	for _, data := range stepsData {
		fmt.Printf("- Date: %s, Steps: %d\n", data.Date.Format("2006-01-02"), data.Steps)
	}

	return nil
}

func runDistance(cmd *cobra.Command, args []string) error {
	garminClient, err := garmin.NewClient("www.garmin.com") // TODO: Domain should be configurable
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	sessionFile := "garmin_session.json" // TODO: Make session file configurable
	if err := garminClient.LoadSession(sessionFile); err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	var startDate, endDate time.Time
	if statsYear {
		now := time.Now()
		startDate = time.Date(now.Year(), time.January, 1, 0, 0, 0, 0, now.Location())
		endDate = time.Date(now.Year(), time.December, 31, 0, 0, 0, 0, now.Location()) // Last day of the year
	} else {
		// Default to today if no specific range or year is given
		startDate = time.Now()
		endDate = time.Now()
	}

	distanceData, err := garminClient.GetDistanceData(startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to get distance data: %w", err)
	}

	if len(distanceData) == 0 {
		fmt.Println("No distance data found.")
		return nil
	}

	fmt.Println("Distance Data:")
	for _, data := range distanceData {
		fmt.Printf("- Date: %s, Distance: %.2f km\n", data.Date.Format("2006-01-02"), data.Distance/1000)
	}

	return nil
}

func runCalories(cmd *cobra.Command, args []string) error {
	garminClient, err := garmin.NewClient("www.garmin.com") // TODO: Domain should be configurable
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	sessionFile := "garmin_session.json" // TODO: Make session file configurable
	if err := garminClient.LoadSession(sessionFile); err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	var startDate, endDate time.Time
	if statsFrom != "" {
		startDate, err = time.Parse("2006-01-02", statsFrom)
		if err != nil {
			return fmt.Errorf("invalid date format for --from: %w", err)
		}
		endDate = time.Now() // Default end date to today if only from is provided
	} else {
		// Default to today if no specific range is given
		startDate = time.Now()
		endDate = time.Now()
	}

	caloriesData, err := garminClient.GetCaloriesData(startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to get calories data: %w", err)
	}

	if len(caloriesData) == 0 {
		fmt.Println("No calories data found.")
		return nil
	}

	fmt.Println("Calories Data:")
	for _, data := range caloriesData {
		fmt.Printf("- Date: %s, Calories: %d\n", data.Date.Format("2006-01-02"), data.Calories)
	}

	return nil
}
