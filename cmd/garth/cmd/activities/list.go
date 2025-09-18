package activities

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/sstent/go-garth/garth/types"
)

var (
	outputFormat string
	startDate    string
	endDate      string
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List activities from Garmin Connect",
	Long:  `List activities with filtering by date range and multiple output formats`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient("")
		if err != nil {
			return fmt.Errorf("client creation failed: %w", err)
		}

		start, end, err := getDateFilter()
		if err != nil {
			return fmt.Errorf("invalid date filter: %w", err)
		}
		activities, err := c.GetActivities(start, end)
		if err != nil {
			return fmt.Errorf("failed to get activities: %w", err)
		}

		switch outputFormat {
		case "table":
			renderTable(activities)
		case "json":
			if err := renderJSON(activities); err != nil {
				return err
			}
		case "csv":
			if err := renderCSV(activities); err != nil {
				return err
			}
		default:
			return fmt.Errorf("invalid output format: %s", outputFormat)
		}

		return nil
	},
}

func init() {
	ListCmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table|json|csv)")
	ListCmd.Flags().StringVar(&startDate, "start", "", "Start date (YYYY-MM-DD)")
	ListCmd.Flags().StringVar(&endDate, "end", "", "End date (YYYY-MM-DD)")
}

func getDateFilter() (time.Time, time.Time, error) {
	var start, end time.Time
	var err error

	if startDate != "" {
		start, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid start date format: %w", err)
		}
	}
	if endDate != "" {
		end, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid end date format: %w", err)
		}
	}
	return start, end, nil
}

func renderTable(activities []types.Activity) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header("ID", "Name", "Type", "Date", "Distance (km)", "Duration")

	for _, a := range activities {
		table.Append(
			fmt.Sprint(a.ActivityID),
			a.ActivityName,
			a.ActivityType.TypeKey,
			a.StartTimeLocal,
			fmt.Sprintf("%.2f", a.Distance/1000),
			time.Duration(a.Duration*float64(time.Second)).String(),
		)
	}

	table.Render()
}

func renderJSON(activities []types.Activity) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(activities)
}

func renderCSV(activities []types.Activity) error {
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	if err := w.Write([]string{"ID", "Name", "Type", "Date", "Distance (km)", "Duration"}); err != nil {
		return err
	}

	for _, a := range activities {
		record := []string{
			fmt.Sprint(a.ActivityID),
			a.ActivityName,
			a.ActivityType.TypeKey,
			a.StartTimeLocal,
			fmt.Sprintf("%.2f", a.Distance/1000),
			time.Duration(a.Duration * float64(time.Second)).String(),
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}
	return nil
}
