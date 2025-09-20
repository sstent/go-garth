package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	types "go-garth/internal/models/types"
	"go-garth/pkg/garmin"
)

var (
	statsYear      bool
	statsAggregate string
	statsFrom      string
)

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
		distanceData = []types.DistanceData{}
		for key, entry := range aggregatedDistance {
			distanceData = append(distanceData, types.DistanceData{
				Date:     types.ParseAggregationKey(key, statsAggregate), // Helper to parse key back to date
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
		tbl := table.New("Date", "Distance (km)")
		for _, data := range distanceData {
			tbl.AddRow(
				data.Date.Format("2006-01-02"),
				fmt.Sprintf("%.2f", data.Distance/1000),
			)
		}
		tbl.Print()
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
		caloriesData = []types.CaloriesData{}
		for key, entry := range aggregatedCalories {
			caloriesData = append(caloriesData, types.CaloriesData{
				Date:     types.ParseAggregationKey(key, statsAggregate), // Helper to parse key back to date
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
		tbl := table.New("Date", "Calories")
		for _, data := range caloriesData {
			tbl.AddRow(
				data.Date.Format("2006-01-02"),
				fmt.Sprintf("%d", data.Calories),
			)
		}
		tbl.Print()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}
