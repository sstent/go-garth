package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"garmin-connect/pkg/garmin"
)

var (
	healthCmd = &cobra.Command{
		Use:   "health",
		Short: "Manage Garmin Connect health data",
		Long:  `Provides commands to fetch various health metrics like sleep, HRV, stress, and body battery.`,
	}

	sleepCmd = &cobra.Command{
		Use:   "sleep",
		Short: "Get sleep data",
		Long:  `Fetch sleep data for a specified date range.`,
		RunE:  runSleep,
	}

	hrvCmd = &cobra.Command{
		Use:   "hrv",
		Short: "Get HRV data",
		Long:  `Fetch Heart Rate Variability (HRV) data.`,
		RunE:  runHrv,
	}

	stressCmd = &cobra.Command{
		Use:   "stress",
		Short: "Get stress data",
		Long:  `Fetch stress data.`,
		RunE:  runStress,
	}

	bodyBatteryCmd = &cobra.Command{
		Use:   "bodybattery",
		Short: "Get Body Battery data",
		Long:  `Fetch Body Battery data.`,
		RunE:  runBodyBattery,
	}

	// Flags for health commands
	healthDateFrom string
	healthDateTo   string
	healthDays     int
	healthWeek     bool
	healthYesterday bool
)

func init() {
	rootCmd.AddCommand(healthCmd)

	healthCmd.AddCommand(sleepCmd)
	sleepCmd.Flags().StringVar(&healthDateFrom, "from", "", "Start date for data fetching (YYYY-MM-DD)")
	sleepCmd.Flags().StringVar(&healthDateTo, "to", "", "End date for data fetching (YYYY-MM-DD)")

	healthCmd.AddCommand(hrvCmd)
	hrvCmd.Flags().IntVar(&healthDays, "days", 0, "Number of past days to fetch data for")

	healthCmd.AddCommand(stressCmd)
	stressCmd.Flags().BoolVar(&healthWeek, "week", false, "Fetch data for the current week")

	healthCmd.AddCommand(bodyBatteryCmd)
	bodyBatteryCmd.Flags().BoolVar(&healthYesterday, "yesterday", false, "Fetch data for yesterday")
}

func runSleep(cmd *cobra.Command, args []string) error {
	garminClient, err := garmin.NewClient("www.garmin.com") // TODO: Domain should be configurable
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	sessionFile := "garmin_session.json" // TODO: Make session file configurable
	if err := garminClient.LoadSession(sessionFile); err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	var startDate, endDate time.Time

	if healthDateFrom != "" {
		startDate, err = time.Parse("2006-01-02", healthDateFrom)
		if err != nil {
			return fmt.Errorf("invalid date format for --from: %w", err)
		}
	} else {
		startDate = time.Now().AddDate(0, 0, -7) // Default to last 7 days
	}

	if healthDateTo != "" {
		endDate, err = time.Parse("2006-01-02", healthDateTo)
		if err != nil {
			return fmt.Errorf("invalid date format for --to: %w", err)
		}
	} else {
		endDate = time.Now() // Default to today
	}

	sleepData, err := garminClient.GetSleepData(startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to get sleep data: %w", err)
	}

	if len(sleepData) == 0 {
		fmt.Println("No sleep data found.")
		return nil
	}

	fmt.Println("Sleep Data:")
	for _, data := range sleepData {
		fmt.Printf("- Date: %s, Score: %d, Total Sleep: %s\n",
			data.Date.Format("2006-01-02"), data.SleepScore, (time.Duration(data.TotalSleepSeconds) * time.Second).String())
	}

	return nil
}

func runHrv(cmd *cobra.Command, args []string) error {
	garminClient, err := garmin.NewClient("www.garmin.com") // TODO: Domain should be configurable
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	sessionFile := "garmin_session.json" // TODO: Make session file configurable
	if err := garminClient.LoadSession(sessionFile); err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	days := healthDays
	if days == 0 {
		days = 7 // Default to 7 days if not specified
	}

	hrvData, err := garminClient.GetHrvData(days)
	if err != nil {
		return fmt.Errorf("failed to get HRV data: %w", err)
	}

	if len(hrvData) == 0 {
		fmt.Println("No HRV data found.")
		return nil
	}

	fmt.Println("HRV Data:")
	for _, data := range hrvData {
		fmt.Printf("- Date: %s, HRV: %.2f\n", data.Date.Format("2006-01-02"), data.HrvValue)
	}

	return nil
}

func runStress(cmd *cobra.Command, args []string) error {
	garminClient, err := garmin.NewClient("www.garmin.com") // TODO: Domain should be configurable
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	sessionFile := "garmin_session.json" // TODO: Make session file configurable
	if err := garminClient.LoadSession(sessionFile); err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	var startDate, endDate time.Time
	if healthWeek {
		now := time.Now()
		weekday := now.Weekday()
		// Calculate the start of the current week (Sunday)
		startDate = now.AddDate(0, 0, -int(weekday))
		endDate = startDate.AddDate(0, 0, 6) // End of the current week (Saturday)
	} else {
		// Default to today if no specific range or week is given
		startDate = time.Now()
		endDate = time.Now()
	}

	stressData, err := garminClient.GetStressData(startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to get stress data: %w", err)
	}

	if len(stressData) == 0 {
		fmt.Println("No stress data found.")
		return nil
	}

	fmt.Println("Stress Data:")
	for _, data := range stressData {
		fmt.Printf("- Date: %s, Stress Level: %d, Rest Stress Level: %d\n",
			data.Date.Format("2006-01-02"), data.StressLevel, data.RestStressLevel)
	}

	return nil
}

func runBodyBattery(cmd *cobra.Command, args []string) error {
	garminClient, err := garmin.NewClient("www.garmin.com") // TODO: Domain should be configurable
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	sessionFile := "garmin_session.json" // TODO: Make session file configurable
	if err := garminClient.LoadSession(sessionFile); err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	var startDate, endDate time.Time
	if healthYesterday {
		startDate = time.Now().AddDate(0, 0, -1)
		endDate = startDate
	} else {
		// Default to today if no specific range or yesterday is given
		startDate = time.Now()
		endDate = time.Now()
	}

	bodyBatteryData, err := garminClient.GetBodyBatteryData(startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to get Body Battery data: %w", err)
	}

	if len(bodyBatteryData) == 0 {
		fmt.Println("No Body Battery data found.")
		return nil
	}

	fmt.Println("Body Battery Data:")
	for _, data := range bodyBatteryData {
		fmt.Printf("- Date: %s, Level: %d, Charge: %d, Drain: %d\n",
			data.Date.Format("2006-01-02"), data.BatteryLevel, data.Charge, data.Drain)
	}

	return nil
}
