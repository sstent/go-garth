package data

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	shared "go-garth/shared/interfaces"
	types "go-garth/internal/models/types"
)

// ParseBodyBatteryReadings converts body battery values array to structured readings
func ParseBodyBatteryReadings(valuesArray [][]any) []types.BodyBatteryReading {
	readings := make([]types.BodyBatteryReading, 0)
	for _, values := range valuesArray {
		if len(values) < 4 {
			continue
		}

		timestamp, ok1 := values[0].(float64)
		status, ok2 := values[1].(string)
		level, ok3 := values[2].(float64)
		version, ok4 := values[3].(float64)

		if !ok1 || !ok2 || !ok3 || !ok4 {
			continue
		}

		readings = append(readings, types.BodyBatteryReading{
			Timestamp: int(timestamp),
			Status:    status,
			Level:     int(level),
			Version:   version,
		})
	}
	sort.Slice(readings, func(i, j int) bool {
		return readings[i].Timestamp < readings[j].Timestamp
	})
	return readings
}

func (d *types.DetailedBodyBatteryData) Get(day time.Time, c shared.APIClient) (interface{}, error) {
	dateStr := day.Format("2006-01-02")

	// Get main Body Battery data
	path1 := fmt.Sprintf("/wellness-service/wellness/dailyStress/%s", dateStr)
	data1, err := c.ConnectAPI(path1, "GET", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Body Battery stress data: %w", err)
	}

	// Get Body Battery events
	path2 := fmt.Sprintf("/wellness-service/wellness/bodyBattery/%s", dateStr)
	data2, err := c.ConnectAPI(path2, "GET", nil, nil)
	if err != nil {
		// Events might not be available, continue without them
		data2 = []byte("[]")
	}

	var result types.DetailedBodyBatteryData
	if len(data1) > 0 {
		if err := json.Unmarshal(data1, &result); err != nil {
			return nil, fmt.Errorf("failed to parse Body Battery data: %w", err)
		}
	}

	var events []types.BodyBatteryEvent
	if len(data2) > 0 {
		if err := json.Unmarshal(data2, &events); err == nil {
			result.Events = events
		}
	}

	return &result, nil
}

// GetCurrentLevel returns the most recent Body Battery level
func (d *types.DetailedBodyBatteryData) GetCurrentLevel() int {
	if len(d.BodyBatteryValuesArray) == 0 {
		return 0
	}

	readings := ParseBodyBatteryReadings(d.BodyBatteryValuesArray)
	if len(readings) == 0 {
		return 0
	}

	return readings[len(readings)-1].Level
}

// GetDayChange returns the Body Battery change for the day
func (d *types.DetailedBodyBatteryData) GetDayChange() int {
	readings := ParseBodyBatteryReadings(d.BodyBatteryValuesArray)
	if len(readings) < 2 {
		return 0
	}

	return readings[len(readings)-1].Level - readings[0].Level
}
