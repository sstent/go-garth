package data

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"garmin-connect/garth/client"
	"garmin-connect/garth/errors"
	"garmin-connect/garth/utils"
)

// DailyBodyBatteryStress represents complete daily Body Battery and stress data
type DailyBodyBatteryStress struct {
	UserProfilePK          int       `json:"userProfilePk"`
	CalendarDate           time.Time `json:"calendarDate"`
	StartTimestampGMT      time.Time `json:"startTimestampGmt"`
	EndTimestampGMT        time.Time `json:"endTimestampGmt"`
	StartTimestampLocal    time.Time `json:"startTimestampLocal"`
	EndTimestampLocal      time.Time `json:"endTimestampLocal"`
	MaxStressLevel         int       `json:"maxStressLevel"`
	AvgStressLevel         int       `json:"avgStressLevel"`
	StressChartValueOffset int       `json:"stressChartValueOffset"`
	StressChartYAxisOrigin int       `json:"stressChartYAxisOrigin"`
	StressValuesArray      [][]int   `json:"stressValuesArray"`
	BodyBatteryValuesArray [][]any   `json:"bodyBatteryValuesArray"`
}

// BodyBatteryEvent represents a Body Battery impact event
type BodyBatteryEvent struct {
	EventType              string    `json:"eventType"`
	EventStartTimeGMT      time.Time `json:"eventStartTimeGmt"`
	TimezoneOffset         int       `json:"timezoneOffset"`
	DurationInMilliseconds int       `json:"durationInMilliseconds"`
	BodyBatteryImpact      int       `json:"bodyBatteryImpact"`
	FeedbackType           string    `json:"feedbackType"`
	ShortFeedback          string    `json:"shortFeedback"`
}

// BodyBatteryData represents legacy Body Battery events data
type BodyBatteryData struct {
	Event                  *BodyBatteryEvent `json:"event"`
	ActivityName           string            `json:"activityName"`
	ActivityType           string            `json:"activityType"`
	ActivityID             string            `json:"activityId"`
	AverageStress          float64           `json:"averageStress"`
	StressValuesArray      [][]int           `json:"stressValuesArray"`
	BodyBatteryValuesArray [][]any           `json:"bodyBatteryValuesArray"`
}

// BodyBatteryReading represents an individual Body Battery reading
type BodyBatteryReading struct {
	Timestamp int     `json:"timestamp"`
	Status    string  `json:"status"`
	Level     int     `json:"level"`
	Version   float64 `json:"version"`
}

// StressReading represents an individual stress reading
type StressReading struct {
	Timestamp   int `json:"timestamp"`
	StressLevel int `json:"stressLevel"`
}

// ParseBodyBatteryReadings converts body battery values array to structured readings
func ParseBodyBatteryReadings(valuesArray [][]any) []BodyBatteryReading {
	readings := make([]BodyBatteryReading, 0)
	for _, values := range valuesArray {
		if len(values) < 4 {
			continue
		}

		timestamp, ok1 := values[0].(int)
		status, ok2 := values[1].(string)
		level, ok3 := values[2].(int)
		version, ok4 := values[3].(float64)

		if !ok1 || !ok2 || !ok3 || !ok4 {
			continue
		}

		readings = append(readings, BodyBatteryReading{
			Timestamp: timestamp,
			Status:    status,
			Level:     level,
			Version:   version,
		})
	}
	sort.Slice(readings, func(i, j int) bool {
		return readings[i].Timestamp < readings[j].Timestamp
	})
	return readings
}

// ParseStressReadings converts stress values array to structured readings
func ParseStressReadings(valuesArray [][]int) []StressReading {
	readings := make([]StressReading, 0)
	for _, values := range valuesArray {
		if len(values) != 2 {
			continue
		}
		readings = append(readings, StressReading{
			Timestamp:   values[0],
			StressLevel: values[1],
		})
	}
	sort.Slice(readings, func(i, j int) bool {
		return readings[i].Timestamp < readings[j].Timestamp
	})
	return readings
}

// Get implements the Data interface for DailyBodyBatteryStress
func (d *DailyBodyBatteryStress) Get(day time.Time, client *client.Client) (any, error) {
	dateStr := day.Format("2006-01-02")
	path := fmt.Sprintf("/wellness-service/wellness/dailyStress/%s", dateStr)

	response, err := client.ConnectAPI(path, "GET", nil)
	if err != nil {
		return nil, err
	}

	if response == nil {
		return nil, nil
	}

	responseMap, ok := response.(map[string]interface{})
	if !ok {
		return nil, &errors.IOError{GarthError: errors.GarthError{
			Message: "Invalid response format"}}
	}

	snakeResponse := utils.CamelToSnakeDict(responseMap)

	jsonBytes, err := json.Marshal(snakeResponse)
	if err != nil {
		return nil, err
	}

	var result DailyBodyBatteryStress
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// List implements the Data interface for concurrent fetching
func (d *DailyBodyBatteryStress) List(end time.Time, days int, client *client.Client, maxWorkers int) ([]any, error) {
	// Implementation to be added
	return []any{}, nil
}
