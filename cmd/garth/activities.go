package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"garmin-connect/pkg/garmin"
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
		Long:  `Download activity data in various formats (e.g., GPX, TCX).`,
		Args:  cobra.ExactArgs(1),
		RunE:  runDownloadActivity,
	}

	searchActivitiesCmd = &cobra.Command{
		Use:   "search",
		Short: "Search activities",
		Long:  `Search Garmin Connect activities by a query string.`,
		RunE:  runSearchActivities,
	}

	// Flags for listActivitiesCmd
	activityLimit      int
	activityOffset     int
	activityType       string
	activityDateFrom   string
	activityDateTo     string

	// Flags for downloadActivitiesCmd
	downloadFormat string
	outputDir      string
	downloadOriginal bool
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
		Limit: activityLimit,
		Offset: activityOffset,
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

	fmt.Println("Activities:")
	for _, activity := range activities {
		fmt.Printf("- ID: %d, Name: %s, Type: %s, Date: %s, Distance: %.2f km, Duration: %.0f s\n",
			activity.ActivityID, activity.ActivityName, activity.ActivityType,
			activity.Starttime.Format("2006-01-02 15:04:05"), activity.Distance/1000, activity.Duration)
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
	fmt.Printf("  Type: %s\n", activityDetail.ActivityType)
	fmt.Printf("  Date: %s\n", activityDetail.Starttime.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Distance: %.2f km\n", activityDetail.Distance/1000)
	fmt.Printf("  Duration: %.0f s\n", activityDetail.Duration)
	fmt.Printf("  Description: %s\n", activityDetail.Description)

	return nil
}

func runDownloadActivity(cmd *cobra.Command, args []string) error {
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

	if downloadFormat == "csv" {
		activityDetail, err := garminClient.GetActivity(activityID)
		if err != nil {
			return fmt.Errorf("failed to get activity details for CSV export: %w", err)
		}

		filename := fmt.Sprintf("%d.csv", activityID)
		outputPath := filename
		if outputDir != "" {
			outputPath = filepath.Join(outputDir, filename)
		}

		file, err := os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create CSV file: %w", err)
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
			activityDetail.ActivityType,
			activityDetail.Starttime.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%.2f", activityDetail.Distance/1000),
			fmt.Sprintf("%.0f", activityDetail.Duration),
			activityDetail.Description,
		})

		fmt.Printf("Activity %d summary exported to %s\n", activityID, outputPath)
		return nil
	}

	opts := garmin.DownloadOptions{
		Format:    downloadFormat,
		OutputDir: outputDir,
		Original:  downloadOriginal,
	}

	fmt.Printf("Downloading activity %d in %s format to %s...\n", activityID, downloadFormat, outputDir)
	if err := garminClient.DownloadActivity(activityID, opts); err != nil {
		return fmt.Errorf("failed to download activity: %w", err)
	}

	fmt.Printf("Activity %d downloaded successfully.\n", activityID)
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
			activity.ActivityID, activity.ActivityName, activity.ActivityType,
			activity.Starttime.Format("2006-01-02"))
	}

	return nil
}
