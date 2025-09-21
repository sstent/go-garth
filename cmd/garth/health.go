package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/rodaine/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go-garth/internal/data" // Import the data package
	types "go-garth/internal/models/types"
	"go-garth/pkg/garmin"
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

	vo2maxCmd = &cobra.Command{
		Use:   "vo2max",
		Short: "Get VO2 Max data",
		Long:  `Fetch VO2 Max data for a specified date range.`,
		RunE:  runVO2Max,
	}

	hrZonesCmd = &cobra.Command{
		Use:   "hr-zones",
		Short: "Get Heart Rate Zones data",
		Long:  `Fetch Heart Rate Zones data.`,
		RunE:  runHRZones,
	}

	trainingStatusCmd = &cobra.Command{
		Use:   "training-status",
		Short: "Get Training Status data",
		Long:  `Fetch Training Status data.`,
		RunE:  runTrainingStatus,
	}

	trainingLoadCmd = &cobra.Command{
		Use:   "training-load",
		Short: "Get Training Load data",
		Long:  `Fetch Training Load data.`,
		RunE:  runTrainingLoad,
	}

	fitnessAgeCmd = &cobra.Command{
		Use:   "fitness-age",
		Short: "Get Fitness Age data",
		Long:  `Fetch Fitness Age data.`,
		RunE:  runFitnessAge,
	}

	wellnessCmd = &cobra.Command{
		Use:   "wellness",
		Short: "Get comprehensive wellness data",
		Long:  `Fetch comprehensive wellness data including body composition and resting heart rate trends.`,
		RunE:  runWellness,
	}

	healthDateFrom  string
	healthDateTo    string
	healthDays      int
	healthWeek      bool
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

	healthCmd.AddCommand(vo2maxCmd)
	vo2maxCmd.Flags().StringVar(&healthDateFrom, "from", "", "Start date for data fetching (YYYY-MM-DD)")
	vo2maxCmd.Flags().StringVar(&healthDateTo, "to", "", "End date for data fetching (YYYY-MM-DD)")
	vo2maxCmd.Flags().StringVar(&healthAggregate, "aggregate", "", "Aggregate data by (day, week, month, year)")

	healthCmd.AddCommand(hrZonesCmd)

	healthCmd.AddCommand(trainingStatusCmd)
	trainingStatusCmd.Flags().StringVar(&healthDateFrom, "from", "", "Date for data fetching (YYYY-MM-DD, defaults to today)")

	healthCmd.AddCommand(trainingLoadCmd)
	trainingLoadCmd.Flags().StringVar(&healthDateFrom, "from", "", "Date for data fetching (YYYY-MM-DD, defaults to today)")

	healthCmd.AddCommand(fitnessAgeCmd)

	healthCmd.AddCommand(wellnessCmd)
	wellnessCmd.Flags().StringVar(&healthDateFrom, "from", "", "Start date for data fetching (YYYY-MM-DD)")
	wellnessCmd.Flags().StringVar(&healthDateTo, "to", "", "End date for data fetching (YYYY-MM-DD)")
	wellnessCmd.Flags().StringVar(&healthAggregate, "aggregate", "", "Aggregate data by (day, week, month, year)")
}

func runSleep(cmd *cobra.Command, args []string) error {
	garminClient, err := garmin.NewClient(viper.GetString("domain"))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := garminClient.LoadSession(viper.GetString("session_file")); err != nil {
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

	var allSleepData []*data.DetailedSleepDataWithMethods
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		// Create a new instance of DetailedSleepDataWithMethods for each day
		sleepDataFetcher := &data.DetailedSleepDataWithMethods{}
		sleepData, err := sleepDataFetcher.Get(d, garminClient.InternalClient())
		if err != nil {
			return fmt.Errorf("failed to get sleep data for %s: %w", d.Format("2006-01-02"), err)
		}
		if sleepData != nil {
			// Type assert the result back to DetailedSleepDataWithMethods
			if sdm, ok := sleepData.(*data.DetailedSleepDataWithMethods); ok {
				allSleepData = append(allSleepData, sdm)
			} else {
				return fmt.Errorf("unexpected type returned for sleep data: %T", sleepData)
			}
		}
	}

	if len(allSleepData) == 0 {
		fmt.Println("No sleep data found.")
		return nil
	}

	outputFormat := viper.GetString("output.format")

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(allSleepData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal sleep data to JSON: %w", err)
		}
		fmt.Println(string(data))
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		defer writer.Flush()

		writer.Write([]string{"Date", "SleepScore", "TotalSleep", "Deep", "Light", "REM", "Awake", "AvgSpO2", "LowestSpO2", "AvgRespiration"})
		for _, data := range allSleepData {
			writer.Write([]string{
				data.CalendarDate.Format("2006-01-02"),
				fmt.Sprintf("%d", data.SleepScores.Overall),
				(time.Duration(data.DeepSleepSeconds+data.LightSleepSeconds+data.RemSleepSeconds) * time.Second).String(),
				(time.Duration(data.DeepSleepSeconds) * time.Second).String(),
				(time.Duration(data.LightSleepSeconds) * time.Second).String(),
				(time.Duration(data.RemSleepSeconds) * time.Second).String(),
				(time.Duration(data.AwakeSleepSeconds) * time.Second).String(),
				func() string {
					if data.AverageSpO2Value != nil {
						return fmt.Sprintf("%.2f", *data.AverageSpO2Value)
					}
					return "N/A"
				}(),
				func() string {
					if data.LowestSpO2Value != nil {
						return fmt.Sprintf("%d", *data.LowestSpO2Value)
					}
					return "N/A"
				}(),
				func() string {
					if data.AverageRespirationValue != nil {
						return fmt.Sprintf("%.2f", *data.AverageRespirationValue)
					}
					return "N/A"
				}(),
			})
		}
	case "table":
		tbl := table.New("Date", "Score", "Total Sleep", "Deep", "Light", "REM", "Awake", "Avg SpO2", "Lowest SpO2", "Avg Resp")
		for _, data := range allSleepData {
			tbl.AddRow(
				data.CalendarDate.Format("2006-01-02"),
				fmt.Sprintf("%d", data.SleepScores.Overall),
				(time.Duration(data.DeepSleepSeconds+data.LightSleepSeconds+data.RemSleepSeconds) * time.Second).String(),
				(time.Duration(data.DeepSleepSeconds) * time.Second).String(),
				(time.Duration(data.LightSleepSeconds) * time.Second).String(),
				(time.Duration(data.RemSleepSeconds) * time.Second).String(),
				(time.Duration(data.AwakeSleepSeconds) * time.Second).String(),
				func() string {
					if data.AverageSpO2Value != nil {
						return fmt.Sprintf("%.2f", *data.AverageSpO2Value)
					}
					return "N/A"
				}(),
				func() string {
					if data.LowestSpO2Value != nil {
						return fmt.Sprintf("%d", *data.LowestSpO2Value)
					}
					return "N/A"
				}(),
				func() string {
					if data.AverageRespirationValue != nil {
						return fmt.Sprintf("%.2f", *data.AverageRespirationValue)
					}
					return "N/A"
				}(),
			)
		}
		tbl.Print()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}

func runHrv(cmd *cobra.Command, args []string) error {
	garminClient, err := garmin.NewClient(viper.GetString("domain"))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := garminClient.LoadSession(viper.GetString("session_file")); err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	days := healthDays
	if days == 0 {
		days = 7 // Default to 7 days if not specified
	}

	var allHrvData []*data.DailyHRVDataWithMethods
	for d := time.Now().AddDate(0, 0, -days+1); !d.After(time.Now()); d = d.AddDate(0, 0, 1) {
		hrvDataFetcher := &data.DailyHRVDataWithMethods{}
		hrvData, err := hrvDataFetcher.Get(d, garminClient.InternalClient())
		if err != nil {
			return fmt.Errorf("failed to get HRV data for %s: %w", d.Format("2006-01-02"), err)
		}
		if hrvData != nil {
			if hdm, ok := hrvData.(*data.DailyHRVDataWithMethods); ok {
				allHrvData = append(allHrvData, hdm)
			} else {
				return fmt.Errorf("unexpected type returned for HRV data: %T", hrvData)
			}
		}
	}

	if len(allHrvData) == 0 {
		fmt.Println("No HRV data found.")
		return nil
	}

	outputFormat := viper.GetString("output.format")

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(allHrvData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal HRV data to JSON: %w", err)
		}
		fmt.Println(string(data))
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		defer writer.Flush()

		writer.Write([]string{"Date", "WeeklyAvg", "LastNightAvg", "Status", "Feedback"})
		for _, data := range allHrvData {
			writer.Write([]string{
				data.CalendarDate.Format("2006-01-02"),
				func() string {
					if data.WeeklyAvg != nil {
						return fmt.Sprintf("%.2f", *data.WeeklyAvg)
					}
					return "N/A"
				}(),
				func() string {
					if data.LastNightAvg != nil {
						return fmt.Sprintf("%.2f", *data.LastNightAvg)
					}
					return "N/A"
				}(),
				data.Status,
				data.FeedbackPhrase,
			})
		}
	case "table":
		tbl := table.New("Date", "Weekly Avg", "Last Night Avg", "Status", "Feedback")
		for _, data := range allHrvData {
			tbl.AddRow(
				data.CalendarDate.Format("2006-01-02"),
				func() string {
					if data.WeeklyAvg != nil {
						return fmt.Sprintf("%.2f", *data.WeeklyAvg)
					}
					return "N/A"
				}(),
				func() string {
					if data.LastNightAvg != nil {
						return fmt.Sprintf("%.2f", *data.LastNightAvg)
					}
					return "N/A"
				}(),
				data.Status,
				data.FeedbackPhrase,
			)
		}
		tbl.Print()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}

func runStress(cmd *cobra.Command, args []string) error {
	garminClient, err := garmin.NewClient(viper.GetString("domain"))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := garminClient.LoadSession(viper.GetString("session_file")); err != nil {
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
		stressData = []types.StressData{}
		for key, entry := range aggregatedStress {
			stressData = append(stressData, types.StressData{
				Date:            types.ParseAggregationKey(key, healthAggregate),
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
		tbl := table.New("Date", "Stress Level", "Rest Stress Level")
		for _, data := range stressData {
			tbl.AddRow(
				data.Date.Format("2006-01-02"),
				fmt.Sprintf("%d", data.StressLevel),
				fmt.Sprintf("%d", data.RestStressLevel),
			)
		}
		tbl.Print()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}

func runBodyBattery(cmd *cobra.Command, args []string) error {
	garminClient, err := garmin.NewClient(viper.GetString("domain"))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := garminClient.LoadSession(viper.GetString("session_file")); err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	var targetDate time.Time
	if healthYesterday {
		targetDate = time.Now().AddDate(0, 0, -1)
	} else {
		targetDate = time.Now()
	}

	bodyBatteryDataFetcher := &data.BodyBatteryDataWithMethods{}
	result, err := bodyBatteryDataFetcher.Get(targetDate, garminClient.InternalClient())
	if err != nil {
		return fmt.Errorf("failed to get Body Battery data: %w", err)
	}
	bodyBatteryData, ok := result.(*data.BodyBatteryDataWithMethods)
	if !ok {
		return fmt.Errorf("unexpected type for Body Battery data: %T", result)
	}

	if bodyBatteryData == nil {
		fmt.Println("No Body Battery data found.")
		return nil
	}

	outputFormat := viper.GetString("output.format")

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

		writer.Write([]string{"Date", "CurrentLevel", "DayChange", "MaxStressLevel", "AvgStressLevel"})
		writer.Write([]string{
			bodyBatteryData.CalendarDate.Format("2006-01-02"),
			fmt.Sprintf("%d", bodyBatteryData.GetCurrentLevel()),
			fmt.Sprintf("%d", bodyBatteryData.GetDayChange()),
			fmt.Sprintf("%d", bodyBatteryData.MaxStressLevel),
			fmt.Sprintf("%d", bodyBatteryData.AvgStressLevel),
		})
	case "table":
		tbl := table.New("Date", "Current Level", "Day Change", "Max Stress", "Avg Stress")
		tbl.AddRow(
			bodyBatteryData.CalendarDate.Format("2006-01-02"),
			fmt.Sprintf("%d", bodyBatteryData.GetCurrentLevel()),
			fmt.Sprintf("%d", bodyBatteryData.GetDayChange()),
			fmt.Sprintf("%d", bodyBatteryData.MaxStressLevel),
			fmt.Sprintf("%d", bodyBatteryData.AvgStressLevel),
		)
		tbl.Print()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}

func runVO2Max(cmd *cobra.Command, args []string) error {
	client, err := garmin.NewClient(viper.GetString("domain"))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.LoadSession(viper.GetString("session_file")); err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	profile, err := client.InternalClient().GetCurrentVO2Max()
	if err != nil {
		return fmt.Errorf("failed to get VO2 Max data: %w", err)
	}

	if profile.Running == nil && profile.Cycling == nil {
		fmt.Println("No VO2 Max data found.")
		return nil
	}

	outputFormat := viper.GetString("output.format")

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(profile, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal VO2 Max data to JSON: %w", err)
		}
		fmt.Println(string(data))
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		defer writer.Flush()

		writer.Write([]string{"Type", "Value", "Date", "Source"})
		if profile.Running != nil {
			writer.Write([]string{
				profile.Running.ActivityType,
				fmt.Sprintf("%.2f", profile.Running.Value),
				profile.Running.Date.Format("2006-01-02"),
				profile.Running.Source,
			})
		}
		if profile.Cycling != nil {
			writer.Write([]string{
				profile.Cycling.ActivityType,
				fmt.Sprintf("%.2f", profile.Cycling.Value),
				profile.Cycling.Date.Format("2006-01-02"),
				profile.Cycling.Source,
			})
		}
	case "table":
		tbl := table.New("Type", "Value", "Date", "Source")

		if profile.Running != nil {
			tbl.AddRow(
				profile.Running.ActivityType,
				fmt.Sprintf("%.2f", profile.Running.Value),
				profile.Running.Date.Format("2006-01-02"),
				profile.Running.Source,
			)
		}
		if profile.Cycling != nil {
			tbl.AddRow(
				profile.Cycling.ActivityType,
				fmt.Sprintf("%.2f", profile.Cycling.Value),
				fmt.Sprintf("%.2f", profile.Cycling.Value),
				profile.Cycling.Date.Format("2006-01-02"),
				profile.Cycling.Source,
			)
		}
		tbl.Print()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}

func runHRZones(cmd *cobra.Command, args []string) error {
	garminClient, err := garmin.NewClient(viper.GetString("domain"))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := garminClient.LoadSession(viper.GetString("session_file")); err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	hrZonesData, err := garminClient.GetHeartRateZones()
	if err != nil {
		return fmt.Errorf("failed to get Heart Rate Zones data: %w", err)
	}

	if hrZonesData == nil {
		fmt.Println("No Heart Rate Zones data found.")
		return nil
	}

	outputFormat := viper.GetString("output")

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(hrZonesData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal Heart Rate Zones data to JSON: %w", err)
		}
		fmt.Println(string(data))
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		defer writer.Flush()

		writer.Write([]string{"Zone", "MinBPM", "MaxBPM", "Name"})
		for _, zone := range hrZonesData.Zones {
			writer.Write([]string{
				strconv.Itoa(zone.Zone),
				strconv.Itoa(zone.MinBPM),
				strconv.Itoa(zone.MaxBPM),
				zone.Name,
			})
		}
	case "table":
		tbl := table.New("Resting HR", "Max HR", "Lactate Threshold", "Updated At")
		tbl.AddRow(
			strconv.Itoa(hrZonesData.RestingHR),
			strconv.Itoa(hrZonesData.MaxHR),
			strconv.Itoa(hrZonesData.LactateThreshold),
			hrZonesData.UpdatedAt.Format("2006-01-02 15:04:05"),
		)
		tbl.Print()

		fmt.Println()

		zonesTable := table.New("Zone", "Min BPM", "Max BPM", "Name")
		for _, zone := range hrZonesData.Zones {
			zonesTable.AddRow(
				strconv.Itoa(zone.Zone),
				strconv.Itoa(zone.MinBPM),
				strconv.Itoa(zone.MaxBPM),
				zone.Name,
			)
		}
		zonesTable.Print()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}

func runWellness(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("not implemented")
}

func runTrainingStatus(cmd *cobra.Command, args []string) error {
	garminClient, err := garmin.NewClient(viper.GetString("domain"))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := garminClient.LoadSession(viper.GetString("session_file")); err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	var targetDate time.Time
	if healthDateFrom != "" {
		targetDate, err = time.Parse("2006-01-02", healthDateFrom)
		if err != nil {
			return fmt.Errorf("invalid date format for --from: %w", err)
		}
	} else {
		targetDate = time.Now()
	}

	trainingStatusFetcher := &data.TrainingStatusWithMethods{}
	trainingStatus, err := trainingStatusFetcher.Get(targetDate, garminClient.InternalClient())
	if err != nil {
		return fmt.Errorf("failed to get training status: %w", err)
	}

	if trainingStatus == nil {
		fmt.Println("No training status data found.")
		return nil
	}

	tsm, ok := trainingStatus.(*data.TrainingStatusWithMethods)
	if !ok {
		return fmt.Errorf("unexpected type returned for training status: %T", trainingStatus)
	}

	outputFormat := viper.GetString("output.format")

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(tsm, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal training status to JSON: %w", err)
		}
		fmt.Println(string(data))
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		defer writer.Flush()

		writer.Write([]string{"Date", "Status", "LoadRatio"})
		writer.Write([]string{
			tsm.CalendarDate.Format("2006-01-02"),
			tsm.TrainingStatusKey,
			fmt.Sprintf("%.2f", tsm.LoadRatio),
		})
	case "table":
		tbl := table.New("Date", "Status", "Load Ratio")
		tbl.AddRow(
			tsm.CalendarDate.Format("2006-01-02"),
			tsm.TrainingStatusKey,
			fmt.Sprintf("%.2f", tsm.LoadRatio),
		)
		tbl.Print()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}

func runTrainingLoad(cmd *cobra.Command, args []string) error {
	garminClient, err := garmin.NewClient(viper.GetString("domain"))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := garminClient.LoadSession(viper.GetString("session_file")); err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	var targetDate time.Time
	if healthDateFrom != "" {
		targetDate, err = time.Parse("2006-01-02", healthDateFrom)
		if err != nil {
			return fmt.Errorf("invalid date format for --from: %w", err)
		}
	} else {
		targetDate = time.Now()
	}

	trainingLoadFetcher := &data.TrainingLoadWithMethods{}
	trainingLoad, err := trainingLoadFetcher.Get(targetDate, garminClient.InternalClient())
	if err != nil {
		return fmt.Errorf("failed to get training load: %w", err)
	}

	if trainingLoad == nil {
		fmt.Println("No training load data found.")
		return nil
	}

	tlm, ok := trainingLoad.(*data.TrainingLoadWithMethods)
	if !ok {
		return fmt.Errorf("unexpected type returned for training load: %T", trainingLoad)
	}

	outputFormat := viper.GetString("output.format")

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(tlm, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal training load to JSON: %w", err)
		}
		fmt.Println(string(data))
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		defer writer.Flush()

		writer.Write([]string{"Date", "AcuteLoad", "ChronicLoad", "LoadRatio"})
		writer.Write([]string{
			tlm.CalendarDate.Format("2006-01-02"),
			fmt.Sprintf("%.2f", tlm.AcuteTrainingLoad),
			fmt.Sprintf("%.2f", tlm.ChronicTrainingLoad),
			fmt.Sprintf("%.2f", tlm.TrainingLoadRatio),
		})
	case "table":
		tbl := table.New("Date", "Acute Load", "Chronic Load", "Load Ratio")
		tbl.AddRow(
			tlm.CalendarDate.Format("2006-01-02"),
			fmt.Sprintf("%.2f", tlm.AcuteTrainingLoad),
			fmt.Sprintf("%.2f", tlm.ChronicTrainingLoad),
			fmt.Sprintf("%.2f", tlm.TrainingLoadRatio),
		)
		tbl.Print()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}

func runFitnessAge(cmd *cobra.Command, args []string) error {
	garminClient, err := garmin.NewClient(viper.GetString("domain"))
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := garminClient.LoadSession(viper.GetString("session_file")); err != nil {
		return fmt.Errorf("not logged in: %w", err)
	}

	fitnessAge, err := garminClient.GetFitnessAge()
	if err != nil {
		return fmt.Errorf("failed to get fitness age: %w", err)
	}

	if fitnessAge == nil {
		fmt.Println("No fitness age data found.")
		return nil
	}

	outputFormat := viper.GetString("output.format")

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(fitnessAge, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal fitness age to JSON: %w", err)
		}
		fmt.Println(string(data))
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		defer writer.Flush()

		writer.Write([]string{"FitnessAge", "ChronologicalAge", "VO2MaxRunning", "LastUpdated"})
		writer.Write([]string{
			fmt.Sprintf("%d", fitnessAge.FitnessAge),
			fmt.Sprintf("%d", fitnessAge.ChronologicalAge),
			fmt.Sprintf("%.2f", fitnessAge.VO2MaxRunning),
			fitnessAge.LastUpdated.Format("2006-01-02"),
		})
	case "table":
		tbl := table.New("Fitness Age", "Chronological Age", "VO2 Max Running", "Last Updated")
		tbl.AddRow(
			fmt.Sprintf("%d", fitnessAge.FitnessAge),
			fmt.Sprintf("%d", fitnessAge.ChronologicalAge),
			fmt.Sprintf("%.2f", fitnessAge.VO2MaxRunning),
			fitnessAge.LastUpdated.Format("2006-01-02"),
		)
		tbl.Print()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}
