package data

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	garth "github.com/sstent/go-garth/pkg/garth/types"
	shared "github.com/sstent/go-garth/shared/interfaces"
)

// DailyHRVDataWithMethods embeds garth.DailyHRVData and adds methods
type DailyHRVDataWithMethods struct {
	garth.DailyHRVData
}

// Get implements the Data interface for DailyHRVData
func (h *DailyHRVDataWithMethods) Get(day time.Time, c shared.APIClient) (interface{}, error) {
	dateStr := day.Format("2006-01-02")
	path := fmt.Sprintf("/wellness-service/wellness/dailyHrvData/%s?date=%s",
		c.GetUsername(), dateStr)

	data, err := c.ConnectAPI(path, "GET", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get HRV data: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var response struct {
		HRVSummary  garth.DailyHRVData `json:"hrvSummary"`
		HRVReadings []garth.HRVReading `json:"hrvReadings"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse HRV response: %w", err)
	}

	// Combine summary and readings
	response.HRVSummary.HRVReadings = response.HRVReadings
	return &DailyHRVDataWithMethods{DailyHRVData: response.HRVSummary}, nil
}

// ParseHRVReadings converts body battery values array to structured readings
func ParseHRVReadings(valuesArray [][]any) []garth.HRVReading {
	readings := make([]garth.HRVReading, 0, len(valuesArray))
	for _, values := range valuesArray {
		if len(values) < 6 {
			continue
		}

		// Extract values with type assertions
		timestamp, _ := values[0].(int)
		stressLevel, _ := values[1].(int)
		heartRate, _ := values[2].(int)
		rrInterval, _ := values[3].(int)
		status, _ := values[4].(string)
		signalQuality, _ := values[5].(float64)

		readings = append(readings, garth.HRVReading{
			Timestamp:     timestamp,
			StressLevel:   stressLevel,
			HeartRate:     heartRate,
			RRInterval:    rrInterval,
			Status:        status,
			SignalQuality: signalQuality,
		})
	}
	sort.Slice(readings, func(i, j int) bool {
		return readings[i].Timestamp < readings[j].Timestamp
	})
	return readings
}
