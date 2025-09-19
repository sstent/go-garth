package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"go-garth/internal/api/client"
	"go-garth/internal/utils"
)

// HRVSummary represents Heart Rate Variability summary data
type HRVSummary struct {
	UserProfilePK     int       `json:"userProfilePk"`
	CalendarDate      time.Time `json:"calendarDate"`
	StartTimestampGMT time.Time `json:"startTimestampGmt"`
	EndTimestampGMT   time.Time `json:"endTimestampGmt"`
	WeeklyAvg         float64   `json:"weeklyAvg"`
	LastNightAvg      float64   `json:"lastNightAvg"`
	Baseline          float64   `json:"baseline"`
}

// HRVReading represents an individual HRV reading
type HRVReading struct {
	Timestamp     int     `json:"timestamp"`
	StressLevel   int     `json:"stressLevel"`
	HeartRate     int     `json:"heartRate"`
	RRInterval    int     `json:"rrInterval"`
	Status        string  `json:"status"`
	SignalQuality float64 `json:"signalQuality"`
}

// TimestampAsTime converts the reading timestamp to time.Time using timeutils
func (r *HRVReading) TimestampAsTime() time.Time {
	return utils.ParseTimestamp(r.Timestamp)
}

// RRSeconds converts the RR interval to seconds
func (r *HRVReading) RRSeconds() float64 {
	return float64(r.RRInterval) / 1000.0
}

// HRVData represents complete HRV data
type HRVData struct {
	UserProfilePK int          `json:"userProfilePk"`
	HRVSummary    HRVSummary   `json:"hrvSummary"`
	HRVReadings   []HRVReading `json:"hrvReadings"`
	BaseData
}

// Validate ensures HRVSummary fields meet requirements
func (h *HRVSummary) Validate() error {
	if h.WeeklyAvg < 0 {
		return errors.New("WeeklyAvg must be non-negative")
	}
	if h.LastNightAvg < 0 {
		return errors.New("LastNightAvg must be non-negative")
	}
	if h.Baseline < 0 {
		return errors.New("Baseline must be non-negative")
	}
	if h.CalendarDate.IsZero() {
		return errors.New("CalendarDate must be set")
	}
	if h.StartTimestampGMT.IsZero() || h.EndTimestampGMT.IsZero() {
		return errors.New("Timestamps must be set")
	}
	if h.EndTimestampGMT.Before(h.StartTimestampGMT) {
		return errors.New("EndTimestampGMT must be after StartTimestampGMT")
	}
	return nil
}

// Validate ensures HRVReading fields meet requirements
func (r *HRVReading) Validate() error {
	if r.StressLevel < 0 || r.StressLevel > 100 {
		return fmt.Errorf("StressLevel must be between 0-100, got %d", r.StressLevel)
	}
	if r.HeartRate <= 0 {
		return fmt.Errorf("HeartRate must be positive, got %d", r.HeartRate)
	}
	if r.RRInterval <= 0 {
		return fmt.Errorf("RRInterval must be positive, got %d", r.RRInterval)
	}
	if r.SignalQuality < 0 || r.SignalQuality > 1 {
		return fmt.Errorf("SignalQuality must be between 0-1, got %f", r.SignalQuality)
	}
	return nil
}

// Validate ensures HRVData meets all requirements
func (h *HRVData) Validate() error {
	if h.UserProfilePK <= 0 {
		return errors.New("UserProfilePK must be positive")
	}
	if err := h.HRVSummary.Validate(); err != nil {
		return fmt.Errorf("HRVSummary validation failed: %w", err)
	}
	for i, reading := range h.HRVReadings {
		if err := reading.Validate(); err != nil {
			return fmt.Errorf("HRVReading[%d] validation failed: %w", i, err)
		}
	}
	return nil
}

// DailyVariability calculates the average RR interval for the day
func (h *HRVData) DailyVariability() float64 {
	if len(h.HRVReadings) == 0 {
		return 0
	}
	var total float64
	for _, r := range h.HRVReadings {
		total += r.RRSeconds()
	}
	return total / float64(len(h.HRVReadings))
}

// MinHRVReading returns the reading with the lowest RR interval
func (h *HRVData) MinHRVReading() HRVReading {
	if len(h.HRVReadings) == 0 {
		return HRVReading{}
	}
	min := h.HRVReadings[0]
	for _, r := range h.HRVReadings {
		if r.RRInterval < min.RRInterval {
			min = r
		}
	}
	return min
}

// MaxHRVReading returns the reading with the highest RR interval
func (h *HRVData) MaxHRVReading() HRVReading {
	if len(h.HRVReadings) == 0 {
		return HRVReading{}
	}
	max := h.HRVReadings[0]
	for _, r := range h.HRVReadings {
		if r.RRInterval > max.RRInterval {
			max = r
		}
	}
	return max
}

// ParseHRVReadings converts values array to structured readings
func ParseHRVReadings(valuesArray [][]any) []HRVReading {
	readings := make([]HRVReading, 0, len(valuesArray))
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

		readings = append(readings, HRVReading{
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

// Get implements the Data interface for HRVData
func (h *HRVData) Get(day time.Time, client *client.Client) (any, error) {
	dateStr := day.Format("2006-01-02")
	path := fmt.Sprintf("/wellness-service/wellness/dailyHrvData/%s?date=%s",
		client.Username, dateStr)

	data, err := client.ConnectAPI(path, "GET", nil, nil)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	var hrvData HRVData
	if err := json.Unmarshal(data, &hrvData); err != nil {
		return nil, err
	}

	if err := hrvData.Validate(); err != nil {
		return nil, fmt.Errorf("HRV data validation failed: %w", err)
	}

	return hrvData, nil
}

// List implements the Data interface for concurrent fetching
func (h *HRVData) List(end time.Time, days int, client *client.Client, maxWorkers int) ([]any, error) {
	// Implementation to be added
	return []any{}, nil
}
