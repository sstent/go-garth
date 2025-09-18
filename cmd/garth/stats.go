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
	statsAggregate string
)

func init() {
	rootCmd.AddCommand(statsCmd)

	statsCmd.AddCommand(stepsCmd)
	stepsCmd.Flags().BoolVar(&statsMonth, "month", false, "Fetch data for the current month")
	stepsCmd.Flags().StringVar(&statsAggregate, "aggregate", "", "Aggregate data by (day, week, month, year)")

	statsCmd.AddCommand(distanceCmd)
	distanceCmd.Flags().BoolVar(&statsYear, "year", false, "Fetch data for the current year")
	distanceCmd.Flags().StringVar(&statsAggregate, "aggregate", "", "Aggregate data by (day, week, month, year)")

	statsCmd.AddCommand(caloriesCmd)
	caloriesCmd.Flags().StringVar(&statsFrom, "from", "", "Start date for data fetching (YYYY-MM-DD)")
	caloriesCmd.Flags().StringVar(&statsAggregate, "aggregate", "", "Aggregate data by (day, week, month, year)")
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

	// Apply aggregation if requested
	if statsAggregate != "" {
		aggregatedSteps := make(map[string]struct {
			Steps int
			Count int
		})

		for _, data := range stepsData {
			key := ""
			switch statsAggregate {
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
				return fmt.Errorf("unsupported aggregation period: %s", statsAggregate)
			}

			entry := aggregatedSteps[key]
			entry.Steps += data.Steps
			entry.Count++
			aggregatedSteps[key] = entry
		}

		// Convert aggregated data back to a slice for output
		stepsData = []garmin.StepsData{}
		for key, entry := range aggregatedSteps {
			stepsData = append(stepsData, garmin.StepsData{
				Date:  parseAggregationKey(key, statsAggregate), // Helper to parse key back to date
				Steps: entry.Steps / entry.Count,
			})
		}
	}

	outputFormat := viper.GetString("output")

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(stepsData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal steps data to JSON: %w", err)
		}
		fmt.Println(string(data))
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		defer writer.Flush()

		writer.Write([]string{"Date", "Steps"})
		for _, data := range stepsData {
			writer.Write([]string{
				data.Date.Format("2006-01-02"),
				fmt.Sprintf("%d", data.Steps),
			})
		}
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Date", "Steps"})
		for _, data := range stepsData {
			table.Append([]string{
				data.Date.Format("2006-01-02"),
				fmt.Sprintf("%d", data.Steps),
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

	// Apply aggregation if requested
	if statsAggregate != "" {
		aggregatedDistance := make(map[string]struct {
			Distance float64
			Count    int
		})

		for _, data := range distanceData {
			key := ""
			switch statsAggregate {
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
				return fmt.Errorf("unsupported aggregation period: %s", statsAggregate)
			}

			entry := aggregatedDistance[key]
			entry.Distance += data.Distance
			entry.Count++
			aggregatedDistance[key] = entry
		}

		// Convert aggregated data back to a slice for output
		distanceData = []garmin.DistanceData{}
		for key, entry := range aggregatedDistance {
			distanceData = append(distanceData, garmin.DistanceData{
				Date:     parseAggregationKey(key, statsAggregate), // Helper to parse key back to date
				Distance: entry.Distance / float64(entry.Count),
			})
		}
	}

	outputFormat := viper.GetString("output")

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(distanceData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal distance data to JSON: %w", err)
		}
		fmt.Println(string(data))
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		defer writer.Flush()

		writer.Write([]string{"Date", "Distance(km)"})
		for _, data := range distanceData {
			writer.Write([]string{
				data.Date.Format("2006-01-02"),
				fmt.Sprintf("%.2f", data.Distance/1000),
			})
		}
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Date", "Distance (km)"})
		for _, data := range distanceData {
			table.Append([]string{
				data.Date.Format("2006-01-02"),
				fmt.Sprintf("%.2f", data.Distance/1000),
			})
		}
		table.Render()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
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

	// Apply aggregation if requested
	if statsAggregate != "" {
		aggregatedCalories := make(map[string]struct {
			Calories int
			Count    int
		})

		for _, data := range caloriesData {
			key := ""
			switch statsAggregate {
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
				return fmt.Errorf("unsupported aggregation period: %s", statsAggregate)
			}

			entry := aggregatedCalories[key]
			entry.Calories += data.Calories
			entry.Count++
			aggregatedCalories[key] = entry
		}

		// Convert aggregated data back to a slice for output
		caloriesData = []garmin.CaloriesData{}
		for key, entry := range aggregatedCalories {
			caloriesData = append(caloriesData, garmin.CaloriesData{
				Date:     parseAggregationKey(key, statsAggregate), // Helper to parse key back to date
				Calories: entry.Calories / entry.Count,
			})
		}
	}

	outputFormat := viper.GetString("output")

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(caloriesData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal calories data to JSON: %w", err)
		}
		fmt.Println(string(data))
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		defer writer.Flush()

		writer.Write([]string{"Date", "Calories"})
		for _, data := range caloriesData {
			writer.Write([]string{
				data.Date.Format("2006-01-02"),
				fmt.Sprintf("%d", data.Calories),
			})
		}
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Date", "Calories"})
		for _, data := range caloriesData {
			table.Append([]string{
				data.Date.Format("2006-01-02"),
				fmt.Sprintf("%d", data.Calories),
			})
		}
		table.Render()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}
