package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	types "go-garth/internal/types"
	"go-garth/internal/utils"
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

	// VO2 Max Command
	vo2maxCmd = &cobra.Command{
		Use:   "vo2max",
		Short: "Get VO2 Max data",
		Long:  `Fetch VO2 Max data for a specified date range.`,
		RunE:  runVO2Max,
	}

	healthCmd.AddCommand(vo2maxCmd)
	vo2maxCmd.Flags().StringVar(&healthDateFrom, "from", "", "Start date for data fetching (YYYY-MM-DD)")
	vo2maxCmd.Flags().StringVar(&healthDateTo, "to", "", "End date for data fetching (YYYY-MM-DD)")
	vo2maxCmd.Flags().StringVar(&healthAggregate, "aggregate", "", "Aggregate data by (day, week, month, year)")

	// Heart Rate Zones Command
	hrZonesCmd = &cobra.Command{
		Use:   "hr-zones",
		Short: "Get Heart Rate Zones data",
		Long:  `Fetch Heart Rate Zones data.`,
		RunE:  runHRZones,
	}

	healthCmd.AddCommand(hrZonesCmd)

	// Wellness Command
	wellnessCmd = &cobra.Command{
		Use:   "wellness",
		Short: "Get comprehensive wellness data",
		Long:  `Fetch comprehensive wellness data including body composition and resting heart rate trends.`,
		RunE:  runWellness,
	}

	healthCmd.AddCommand(wellnessCmd)
	wellnessCmd.Flags().StringVar(&healthDateFrom, "from", "", "Start date for data fetching (YYYY-MM-DD)")
	wellnessCmd.Flags().StringVar(&healthDateTo, "to", "", "End date for data fetching (YYYY-MM-DD)")
	wellnessCmd.Flags().StringVar(&healthAggregate, "aggregate", "", "Aggregate data by (day, week, month, year)")
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
		sleepData = []types.SleepData{}
		for key, entry := range aggregatedSleep {
			sleepData = append(sleepData, types.SleepData{
				Date:              utils.ParseAggregationKey(key, healthAggregate),
				TotalSleepSeconds: entry.TotalSleepSeconds / entry.Count,
				SleepScore:        entry.SleepScore / entry.Count,
				DeepSleepSeconds:  0,
				LightSleepSeconds: 0,
				RemSleepSeconds:   0,
				AwakeSleepSeconds: 0,
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
		table.Header([]string{"Date", "Score", "Total Sleep", "Deep", "Light", "REM", "Awake"})
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
		hrvData = []types.HrvData{}
		for key, entry := range aggregatedHrv {
			hrvData = append(hrvData, types.HrvData{
				Date:     utils.ParseAggregationKey(key, healthAggregate),
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
		table.Header([]string{"Date", "HRV Value"})
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
		stressData = []types.StressData{}
		for key, entry := range aggregatedStress {
			stressData = append(stressData, types.StressData{
				Date:            utils.ParseAggregationKey(key, healthAggregate),
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
		table.Header([]string{"Date", "Stress Level", "Rest Stress Level"})
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
		bodyBatteryData = []types.BodyBatteryData{}
		for key, entry := range aggregatedBodyBattery {
			bodyBatteryData = append(bodyBatteryData, types.BodyBatteryData{
				Date:         utils.ParseAggregationKey(key, healthAggregate),
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
		table.Header([]string{"Date", "Battery Level", "Charge", "Drain"})
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

func runVO2Max(cmd *cobra.Command, args []string) error {
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
		endDate = time.Now() // Default to today
		parsedDate, err := time.Parse("2006-01-02", healthDateTo)
		if err != nil {
			return fmt.Errorf("invalid date format for --to: %w", err)
		}
		endDate = parsedDate
	}

	vo2maxData, err := garminClient.GetVO2MaxData(startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to get VO2 Max data: %w", err)
	}

	if len(vo2maxData) == 0 {
		fmt.Println("No VO2 Max data found.")
		return nil
	}

	// Apply aggregation if requested
	if healthAggregate != "" {
		aggregatedVO2Max := make(map[string]struct {
			VO2MaxRunning float64
			VO2MaxCycling float64
			Count         int
		})

		for _, data := range vo2maxData {
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

			entry := aggregatedVO2Max[key]
			entry.VO2MaxRunning += data.VO2MaxRunning
			entry.VO2MaxCycling += data.VO2MaxCycling
			entry.Count++
			aggregatedVO2Max[key] = entry
		}

		// Convert aggregated data back to a slice for output
		vo2maxData = []types.VO2MaxData{}
		for key, entry := range aggregatedVO2Max {
			vo2maxData = append(vo2maxData, types.VO2MaxData{
				Date:          utils.ParseAggregationKey(key, healthAggregate),
				VO2MaxRunning: entry.VO2MaxRunning / float64(entry.Count),
				VO2MaxCycling: entry.VO2MaxCycling / float64(entry.Count),
			})
		}
	}

	outputFormat := viper.GetString("output")

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(vo2maxData, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal VO2 Max data to JSON: %w", err)
		}
		fmt.Println(string(data))
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		defer writer.Flush()

		writer.Write([]string{"Date", "VO2MaxRunning", "VO2MaxCycling"})
		for _, data := range vo2maxData {
			writer.Write([]string{
				data.Date.Format("2006-01-02"),
				fmt.Sprintf("%.2f", data.VO2MaxRunning),
				fmt.Sprintf("%.2f", data.VO2MaxCycling),
			})
		}
	case "table":
		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"Date", "VO2 Max Running", "VO2 Max Cycling"})
		for _, data := range vo2maxData {
			table.Append([]string{
				data.Date.Format("2006-01-02"),
				fmt.Sprintf("%.2f", data.VO2MaxRunning),
				fmt.Sprintf("%.2f", data.VO2MaxCycling),
			})
		}
		table.Render()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}

func runHRZones(cmd *cobra.Command, args []string) error {
	garminClient, err := garmin.NewClient("www.garmin.com") // TODO: Domain should be configurable
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	sessionFile := "garmin_session.json" // TODO: Make session file configurable
	if err := garminClient.LoadSession(sessionFile); err != nil {
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
		table := tablewriter.NewWriter(os.Stdout)
		table.Header([]string{"Resting HR", "Max HR", "Lactate Threshold", "Updated At"})
		table.Append([]string{
			strconv.Itoa(hrZonesData.RestingHR),
			strconv.Itoa(hrZonesData.MaxHR),
			strconv.Itoa(hrZonesData.LactateThreshold),
			hrZonesData.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
		table.Render()

		fmt.Println()

		zonesTable := tablewriter.NewWriter(os.Stdout)
		zonesTable.Header([]string{"Zone", "Min BPM", "Max BPM", "Name"})
		for _, zone := range hrZonesData.Zones {
			zonesTable.Append([]string{
				strconv.Itoa(zone.Zone),
				strconv.Itoa(zone.MinBPM),
				strconv.Itoa(zone.MaxBPM),
				zone.Name,
			})
		}
		zonesTable.Render()
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	return nil
}

var wellnessCmd *cobra.Command

func runWellness(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("not implemented")
}
