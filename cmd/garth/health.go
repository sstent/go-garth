package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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
	healthAggregate string
)

func init() {
	rootCmd.AddCommand(healthCmd)

	healthCmd.AddCommand(sleepCmd)
	sleepCmd.Flags().StringVar(&healthDateFrom, "from", "", "Start date for data fetching (YYYY-MM-DD)")
	sleepCmd.Flags().StringVar(&healthDateTo, "to", "", "End date for data fetching (YYYY-MM-DD)")
	sleepCmd.Flags().StringVar(&healthAggregate, "aggregate", "", "Aggregate data by (day, week, month, year)")

	healthCmd.AddCommand(hrvCmd)
	hrvCmd.Flags().IntVar(&healthDays, "days", 0, "Number of past days to fetch data for")
	hrvCmd.Flags().StringVar(&healthAggregate, "aggregate", "", "Aggregate data by (day, week, month, year)")

	healthCmd.AddCommand(stressCmd)
	stressCmd.Flags().BoolVar(&healthWeek, "week", false, "Fetch data for the current week")
	stressCmd.Flags().StringVar(&healthAggregate, "aggregate", "", "Aggregate data by (day, week, month, year)")

	healthCmd.AddCommand(bodyBatteryCmd)
	bodyBatteryCmd.Flags().BoolVar(&healthYesterday, "yesterday", false, "Fetch data for yesterday")
	bodyBatteryCmd.Flags().StringVar(&healthAggregate, "aggregate", "", "Aggregate data by (day, week, month, year)")
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

	// Apply aggregation if requested
	if healthAggregate != "" {
		aggregatedSleep := make(map[string]struct {
			TotalSleepSeconds int
			SleepScore        int
			Count             int
		})

		for _, data := range sleepData {
			key := ""
			switch healthAggregate {
			case "day":
				key = data.Date.Format("2006-01-02")
			case "week":
				year, week := data.Date.ISOWeek()
				key = fmt.Sprintf("%d-W%02d", year, week)
			case "month":
				key = data.Date.Format("2006-01")
			case "year":
				key = data.Date.Format("2006")
			default:
				return fmt.Errorf("unsupported aggregation period: %s", healthAggregate)
			}

			entry := aggregatedSleep[key]
			entry.TotalSleepSeconds += data.TotalSleepSeconds
			entry.SleepScore += data.SleepScore
			entry.Count++
			aggregatedSleep[key] = entry
		}

		// Convert aggregated data back to a slice for output
		sleepData = []garmin.SleepData{}
		for key, entry := range aggregatedSleep {
			sleepData = append(sleepData, garmin.SleepData{
				Date:              parseAggregationKey(key, healthAggregate), // Helper to parse key back to date
				TotalSleepSeconds: entry.TotalSleepSeconds / entry.Count,
				SleepScore:        entry.SleepScore / entry.Count,
			})
		}
	}

	outputFormat := viper.GetString("output")

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(sleepData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal sleep data to JSON: %w", err)
		}
		fmt.Println(string(data))
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		defer writer.Flush()

		writer.Write([]string{"Date", "SleepScore", "TotalSleepSeconds", "DeepSleepSeconds", "LightSleepSeconds", "RemSleepSeconds", "AwakeSleepSeconds"})
		for _, data := range sleepData {
			writer.Write([]string{
				data.Date.Format("2006-01-02"),
				fmt.Sprintf("%d", data.SleepScore),
				fmt.Sprintf("%d", data.TotalSleepSeconds),
				fmt.Sprintf("%d", data.DeepSleepSeconds),
				fmt.Sprintf("%d", data.LightSleepSeconds),
				fmt.Sprintf("%d", data.RemSleepSeconds),
				fmt.Sprintf("%d", data.AwakeSleepSeconds),
			})
		}
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Date", "Score", "Total Sleep", "Deep", "Light", "REM", "Awake"})
		for _, data := range sleepData {
			table.Append([]string{
				data.Date.Format("2006-01-02"),
				fmt.Sprintf("%d", data.SleepScore),
				(time.Duration(data.TotalSleepSeconds) * time.Second).String(),
				(time.Duration(data.DeepSleepSeconds) * time.Second).String(),
				(time.Duration(data.LightSleepSeconds) * time.Second).String(),
				(time.Duration(data.RemSleepSeconds) * time.Second).String(),
				(time.Duration(data.AwakeSleepSeconds) * time.Second).String(),
			})
		}
		table.Render()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}

// Helper function to parse aggregation key back to a time.Time object
func parseAggregationKey(key, aggregate string) time.Time {
	switch aggregate {
	case "day":
		t, _ := time.Parse("2006-01-02", key)
		return t
	case "week":
		year, _ := strconv.Atoi(key[:4])
		week, _ := strconv.Atoi(key[6:])
		t := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		// Find the first Monday of the year
		for t.Weekday() != time.Monday {
			t = t.AddDate(0, 0, 1)
		}
		// Add weeks
		return t.AddDate(0, 0, (week-1)*7)
	case "month":
		t, _ := time.Parse("2006-01", key)
		return t
	case "year":
		t, _ := time.Parse("2006", key)
		return t
	}
	return time.Time{}
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

	// Apply aggregation if requested
	if healthAggregate != "" {
		aggregatedHrv := make(map[string]struct {
			HrvValue float64
			Count    int
		})

		for _, data := range hrvData {
			key := ""
			switch healthAggregate {
			case "day":
				key = data.Date.Format("2006-01-02")
			case "week":
				year, week := data.Date.ISOWeek()
				key = fmt.Sprintf("%d-W%02d", year, week)
			case "month":
				key = data.Date.Format("2006-01")
			case "year":
				key = data.Date.Format("2006")
			default:
				return fmt.Errorf("unsupported aggregation period: %s", healthAggregate)
			}

			entry := aggregatedHrv[key]
			entry.HrvValue += data.HrvValue
			entry.Count++
			aggregatedHrv[key] = entry
		}

		// Convert aggregated data back to a slice for output
		hrvData = []garmin.HrvData{}
		for key, entry := range aggregatedHrv {
			hrvData = append(hrvData, garmin.HrvData{
				Date:     parseAggregationKey(key, healthAggregate), // Helper to parse key back to date
				HrvValue: entry.HrvValue / float64(entry.Count),
			})
		}
	}

	outputFormat := viper.GetString("output")

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(hrvData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal HRV data to JSON: %w", err)
		}
		fmt.Println(string(data))
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		defer writer.Flush()

		writer.Write([]string{"Date", "HRV Value"})
		for _, data := range hrvData {
			writer.Write([]string{
				data.Date.Format("2006-01-02"),
				fmt.Sprintf("%.2f", data.HrvValue),
			})
		}
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Date", "HRV Value"})
		for _, data := range hrvData {
			table.Append([]string{
				data.Date.Format("2006-01-02"),
				fmt.Sprintf("%.2f", data.HrvValue),
			})
		}
		table.Render()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
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

	// Apply aggregation if requested
	if healthAggregate != "" {
		aggregatedStress := make(map[string]struct {
			StressLevel     int
			RestStressLevel int
			Count           int
		})

		for _, data := range stressData {
			key := ""
			switch healthAggregate {
			case "day":
				key = data.Date.Format("2006-01-02")
			case "week":
				year, week := data.Date.ISOWeek()
				key = fmt.Sprintf("%d-W%02d", year, week)
			case "month":
				key = data.Date.Format("2006-01")
			case "year":
				key = data.Date.Format("2006")
			default:
				return fmt.Errorf("unsupported aggregation period: %s", healthAggregate)
			}

			entry := aggregatedStress[key]
			entry.StressLevel += data.StressLevel
			entry.RestStressLevel += data.RestStressLevel
			entry.Count++
			aggregatedStress[key] = entry
		}

		// Convert aggregated data back to a slice for output
		stressData = []garmin.StressData{}
		for key, entry := range aggregatedStress {
			stressData = append(stressData, garmin.StressData{
				Date:            parseAggregationKey(key, healthAggregate), // Helper to parse key back to date
				StressLevel:     entry.StressLevel / entry.Count,
				RestStressLevel: entry.RestStressLevel / entry.Count,
			})
		}
	}

	outputFormat := viper.GetString("output")

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(stressData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal stress data to JSON: %w", err)
		}
		fmt.Println(string(data))
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		defer writer.Flush()

		writer.Write([]string{"Date", "StressLevel", "RestStressLevel"})
		for _, data := range stressData {
			writer.Write([]string{
				data.Date.Format("2006-01-02"),
				fmt.Sprintf("%d", data.StressLevel),
				fmt.Sprintf("%d", data.RestStressLevel),
			})
		}
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Date", "Stress Level", "Rest Stress Level"})
		for _, data := range stressData {
			table.Append([]string{
				data.Date.Format("2006-01-02"),
				fmt.Sprintf("%d", data.StressLevel),
				fmt.Sprintf("%d", data.RestStressLevel),
			})
		}
		table.Render()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
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

	// Apply aggregation if requested
	if healthAggregate != "" {
		aggregatedBodyBattery := make(map[string]struct {
			BatteryLevel int
			Charge       int
			Drain        int
			Count        int
		})

		for _, data := range bodyBatteryData {
			key := ""
			switch healthAggregate {
			case "day":
				key = data.Date.Format("2006-01-02")
			case "week":
				year, week := data.Date.ISOWeek()
				key = fmt.Sprintf("%d-W%02d", year, week)
			case "month":
				key = data.Date.Format("2006-01")
			case "year":
				key = data.Date.Format("2006")
			default:
				return fmt.Errorf("unsupported aggregation period: %s", healthAggregate)
			}

			entry := aggregatedBodyBattery[key]
			entry.BatteryLevel += data.BatteryLevel
			entry.Charge += data.Charge
			entry.Drain += data.Drain
			entry.Count++
			aggregatedBodyBattery[key] = entry
		}

		// Convert aggregated data back to a slice for output
		bodyBatteryData = []garmin.BodyBatteryData{}
		for key, entry := range aggregatedBodyBattery {
			bodyBatteryData = append(bodyBatteryData, garmin.BodyBatteryData{
				Date:         parseAggregationKey(key, healthAggregate), // Helper to parse key back to date
				BatteryLevel: entry.BatteryLevel / entry.Count,
				Charge:       entry.Charge / entry.Count,
				Drain:        entry.Drain / entry.Count,
			})
		}
	}

	outputFormat := viper.GetString("output")

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(bodyBatteryData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal Body Battery data to JSON: %w", err)
		}
		fmt.Println(string(data))
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		defer writer.Flush()

		writer.Write([]string{"Date", "BatteryLevel", "Charge", "Drain"})
		for _, data := range bodyBatteryData {
			writer.Write([]string{
				data.Date.Format("2006-01-02"),
				fmt.Sprintf("%d", data.BatteryLevel),
				fmt.Sprintf("%d", data.Charge),
				fmt.Sprintf("%d", data.Drain),
			})
		}
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Date", "Level", "Charge", "Drain"})
		for _, data := range bodyBatteryData {
			table.Append([]string{
				data.Date.Format("2006-01-02"),
				fmt.Sprintf("%d", data.BatteryLevel),
				fmt.Sprintf("%d", data.Charge),
				fmt.Sprintf("%d", data.Drain),
			})
		}
		table.Render()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}
