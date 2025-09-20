package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/rodaine/table"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go-garth/pkg/garmin"
)

var (
	activitiesCmd = &cobra.Command{
		Use:   "activities",
		Short: "Manage Garmin Connect activities",
		Long:  `Provides commands to list, get details, search, and download Garmin Connect activities.`,
	}

	listActivitiesCmd = &cobra.Command{
		Use:   "list",
		Short: "List recent activities",
		Long:  `List recent Garmin Connect activities with optional filters.`,
		RunE:  runListActivities,
	}

	getActivitiesCmd = &cobra.Command{
		Use:   "get [activityID]",
		Short: "Get activity details",
		Long:  `Get detailed information for a specific Garmin Connect activity.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runGetActivity,
	}

	downloadActivitiesCmd = &cobra.Command{
		Use:   "download [activityID]",
		Short: "Download activity data",
		Args:  cobra.RangeArgs(0, 1), RunE: runDownloadActivity,
	}

	searchActivitiesCmd = &cobra.Command{
		Use:   "search",
		Short: "Search activities",
		Long:  `Search Garmin Connect activities by a query string.`,
		RunE:  runSearchActivities,
	}

	// Flags for listActivitiesCmd
	activityLimit    int
	activityOffset   int
	activityType     string
	activityDateFrom string
	activityDateTo   string

	// Flags for downloadActivitiesCmd
	downloadFormat   string
	outputDir        string
	downloadOriginal bool
	downloadAll      bool
)

func init() {
	rootCmd.AddCommand(activitiesCmd)

	activitiesCmd.AddCommand(listActivitiesCmd)
	listActivitiesCmd.Flags().IntVar(&activityLimit, "limit", 20, "Maximum number of activities to retrieve")
	listActivitiesCmd.Flags().IntVar(&activityOffset, "offset", 0, "Offset for activities list")
	listActivitiesCmd.Flags().StringVar(&activityType, "type", "", "Filter activities by type (e.g., running, cycling)")
	listActivitiesCmd.Flags().StringVar(&activityDateFrom, "from", "", "Start date for filtering activities (YYYY-MM-DD)")
	listActivitiesCmd.Flags().StringVar(&activityDateTo, "to", "", "End date for filtering activities (YYYY-MM-DD)")

	activitiesCmd.AddCommand(getActivitiesCmd)

	activitiesCmd.AddCommand(downloadActivitiesCmd)
	downloadActivitiesCmd.Flags().StringVar(&downloadFormat, "format", "gpx", "Download format (gpx, tcx, fit, csv)")
	downloadActivitiesCmd.Flags().StringVar(&outputDir, "output-dir", ".", "Output directory for downloaded files")
	downloadActivitiesCmd.Flags().BoolVar(&downloadOriginal, "original", false, "Download original uploaded file")

	downloadActivitiesCmd.Flags().BoolVar(&downloadAll, "all", false, "Download all activities matching filters")
	downloadActivitiesCmd.Flags().StringVar(&activityType, "type", "", "Filter activities by type (e.g., running, cycling)")
	downloadActivitiesCmd.Flags().StringVar(&activityDateFrom, "from", "", "Start date for filtering activities (YYYY-MM-DD)")
	downloadActivitiesCmd.Flags().StringVar(&activityDateTo, "to", "", "End date for filtering activities (YYYY-MM-DD)")

	activitiesCmd.AddCommand(searchActivitiesCmd)
	searchActivitiesCmd.Flags().StringP("query", "q", "", "Query string to search for activities")
}

func runListActivities(cmd *cobra.Command, args []string) error {
	garminClient, err := garmin.NewClient("www.garmin.com") // TODO: Domain should be configurable
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	sessionFile := "garmin_session.json" // TODO: Make session file configurable
	if err := garminClient.LoadSession(sessionFile); err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	opts := garmin.ActivityOptions{
		Limit:        activityLimit,
		Offset:       activityOffset,
		ActivityType: activityType,
	}

	if activityDateFrom != "" {
		opts.DateFrom, err = time.Parse("2006-01-02", activityDateFrom)
		if err != nil {
			return fmt.Errorf("invalid date format for --from: %w", err)
		}
	}

	if activityDateTo != "" {
		opts.DateTo, err = time.Parse("2006-01-02", activityDateTo)
		if err != nil {
			return fmt.Errorf("invalid date format for --to: %w", err)
		}
	}

	activities, err := garminClient.ListActivities(opts)
	if err != nil {
		return fmt.Errorf("failed to list activities: %w", err)
	}

	if len(activities) == 0 {
		fmt.Println("No activities found.")
		return nil
	}

	outputFormat := viper.GetString("output")

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(activities, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal activities to JSON: %w", err)
		}
		fmt.Println(string(data))
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		defer writer.Flush()

		writer.Write([]string{"ActivityID", "ActivityName", "ActivityType", "StartTime", "Distance(km)", "Duration(s)"})
		for _, activity := range activities {
			writer.Write([]string{
				fmt.Sprintf("%d", activity.ActivityID),
				activity.ActivityName,
				activity.ActivityType.TypeKey,
				activity.StartTimeLocal.Format("2006-01-02 15:04:05"),
				fmt.Sprintf("%.2f", activity.Distance/1000),
				fmt.Sprintf("%.0f", activity.Duration),
			})
		}
	case "table":
		tbl := table.New("ID", "Name", "Type", "Date", "Distance (km)", "Duration (s)")
		for _, activity := range activities {
			tbl.AddRow(
				fmt.Sprintf("%d", activity.ActivityID),
				activity.ActivityName,
				activity.ActivityType.TypeKey,
				activity.StartTimeLocal.Format("2006-01-02 15:04:05"),
				fmt.Sprintf("%.2f", activity.Distance/1000),
				fmt.Sprintf("%.0f", activity.Duration),
			)
		}
		tbl.Print()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}

func runGetActivity(cmd *cobra.Command, args []string) error {
	activityIDStr := args[0]
	activityID, err := strconv.Atoi(activityIDStr)
	if err != nil {
		return fmt.Errorf("invalid activity ID: %w", err)
	}

	garminClient, err := garmin.NewClient("www.garmin.com") // TODO: Domain should be configurable
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	sessionFile := "garmin_session.json" // TODO: Make session file configurable
	if err := garminClient.LoadSession(sessionFile); err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	activityDetail, err := garminClient.GetActivity(activityID)
	if err != nil {
		return fmt.Errorf("failed to get activity details: %w", err)
	}

	fmt.Printf("Activity Details (ID: %d):\n", activityDetail.ActivityID)
	fmt.Printf("  Name: %s\n", activityDetail.ActivityName)
	fmt.Printf("  Type: %s\n", activityDetail.ActivityType.TypeKey)
	fmt.Printf("  Date: %s\n", activityDetail.StartTimeLocal.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Distance: %.2f km\n", activityDetail.Distance/1000)
	fmt.Printf("  Duration: %.0f s\n", activityDetail.Duration)
	fmt.Printf("  Description: %s\n", activityDetail.Description)

	return nil
}

func runDownloadActivity(cmd *cobra.Command, args []string) error {
	var wg sync.WaitGroup
	const concurrencyLimit = 5 // Limit concurrent downloads
	sem := make(chan struct{}, concurrencyLimit)

	garminClient, err := garmin.NewClient("www.garmin.com") // TODO: Domain should be configurable
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	sessionFile := "garmin_session.json" // TODO: Make session file configurable
	if err := garminClient.LoadSession(sessionFile); err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	var activitiesToDownload []garmin.Activity

	if downloadAll || len(args) == 0 {
		opts := garmin.ActivityOptions{
			ActivityType: activityType,
		}

		if activityDateFrom != "" {
			opts.DateFrom, err = time.Parse("2006-01-02", activityDateFrom)
			if err != nil {
				return fmt.Errorf("invalid date format for --from: %w", err)
			}
		}

		if activityDateTo != "" {
			opts.DateTo, err = time.Parse("2006-01-02", activityDateTo)
			if err != nil {
				return fmt.Errorf("invalid date format for --to: %w", err)
			}
		}

		activitiesToDownload, err = garminClient.ListActivities(opts)
		if err != nil {
			return fmt.Errorf("failed to list activities for batch download: %w", err)
		}

		if len(activitiesToDownload) == 0 {
			fmt.Println("No activities found matching the filters for download.")
			return nil
		}
	} else if len(args) == 1 {
		activityIDStr := args[0]
		activityID, err := strconv.Atoi(activityIDStr)
		if err != nil {
			return fmt.Errorf("invalid activity ID: %w", err)
		}
		// For single download, we need to fetch the activity details to get its name and type
		activityDetail, err := garminClient.GetActivity(activityID)
		if err != nil {
			return fmt.Errorf("failed to get activity details for download: %w", err)
		}
		activitiesToDownload = []garmin.Activity{activityDetail.Activity}
	} else {
		return fmt.Errorf("invalid arguments: specify an activity ID or use --all with filters")
	}

	fmt.Printf("Starting download of %d activities...\n", len(activitiesToDownload))

	bar := progressbar.NewOptions(len(activitiesToDownload),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription("Downloading activities..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerPadding: " ",
			BarStart:      "[ ",
			BarEnd:        " ]",
		}),
	)

	for _, activity := range activitiesToDownload {
		wg.Add(1)
		sem <- struct{}{}
		go func(activity garmin.Activity) {
			defer wg.Done()
			defer func() { <-sem }()

			if downloadFormat == "csv" {
				activityDetail, err := garminClient.GetActivity(int(activity.ActivityID))
				if err != nil {
					fmt.Printf("Warning: Failed to get activity details for CSV export for activity %d: %v\n", activity.ActivityID, err)
					bar.Add(1)
					return
				}

				filename := fmt.Sprintf("%d.csv", activity.ActivityID)
				outputPath := filename
				if outputDir != "" {
					outputPath = filepath.Join(outputDir, filename)
				}

				file, err := os.Create(outputPath)
				if err != nil {
					fmt.Printf("Warning: Failed to create CSV file for activity %d: %v\n", activity.ActivityID, err)
					bar.Add(1)
					return
				}
				defer file.Close()

				writer := csv.NewWriter(file)
				defer writer.Flush()

				// Write header
				writer.Write([]string{"ActivityID", "ActivityName", "ActivityType", "StartTime", "Distance(km)", "Duration(s)", "Description"})

				// Write data
				writer.Write([]string{
					fmt.Sprintf("%d", activityDetail.ActivityID),
					activityDetail.ActivityName,
					activityDetail.ActivityType.TypeKey,
					activityDetail.StartTimeLocal.Format("2006-01-02 15:04:05"),
					fmt.Sprintf("%.2f", activityDetail.Distance/1000),
					fmt.Sprintf("%.0f", activityDetail.Duration),
					activityDetail.Description,
				})

				fmt.Printf("Activity %d summary exported to %s\n", activity.ActivityID, outputPath)
			} else {
				filename := fmt.Sprintf("%d.%s", activity.ActivityID, downloadFormat)
				if downloadOriginal {
					filename = fmt.Sprintf("%d_original.fit", activity.ActivityID) // Assuming original is .fit
				}
				outputPath := filepath.Join(outputDir, filename)

				// Check if file already exists
				if _, err := os.Stat(outputPath); err == nil {
					fmt.Printf("Skipping activity %d: file already exists at %s\n", activity.ActivityID, outputPath)
					bar.Add(1)
					return
				} else if !os.IsNotExist(err) {
					fmt.Printf("Warning: Failed to check existence of file %s for activity %d: %v\n", outputPath, activity.ActivityID, err)
					bar.Add(1)
					return
				}

				opts := garmin.DownloadOptions{
					Format:    downloadFormat,
					OutputDir: outputDir,
					Original:  downloadOriginal,
					Filename:  filename, // Pass filename to opts
				}

				fmt.Printf("Downloading activity %d in %s format to %s...\n", activity.ActivityID, downloadFormat, outputPath)
				if err := garminClient.DownloadActivity(int(activity.ActivityID), opts); err != nil {
					fmt.Printf("Warning: Failed to download activity %d: %v\n", activity.ActivityID, err)
					bar.Add(1)
					return
				}

				fmt.Printf("Activity %d downloaded successfully.\n", activity.ActivityID)
			}
			bar.Add(1)
		}(activity)
	}

	wg.Wait()
	bar.Finish()
	fmt.Println("All downloads finished.")

	return nil
}

func runSearchActivities(cmd *cobra.Command, args []string) error {
	query, err := cmd.Flags().GetString("query")
	if err != nil || query == "" {
		return fmt.Errorf("search query cannot be empty")
	}

	garminClient, err := garmin.NewClient("www.garmin.com") // TODO: Domain should be configurable
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	sessionFile := "garmin_session.json" // TODO: Make session file configurable
	if err := garminClient.LoadSession(sessionFile); err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	activities, err := garminClient.SearchActivities(query)
	if err != nil {
		return fmt.Errorf("failed to search activities: %w", err)
	}

	if len(activities) == 0 {
		fmt.Printf("No activities found for query '%s'.\n", query)
		return nil
	}

	fmt.Printf("Activities matching '%s':\n", query)
	for _, activity := range activities {
		fmt.Printf("- ID: %d, Name: %s, Type: %s, Date: %s\n",
			activity.ActivityID, activity.ActivityName, activity.ActivityType.TypeKey,
			activity.StartTimeLocal.Format("2006-01-02"))
	}

	return nil
}
