package data

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	garth "github.com/sstent/go-garth/pkg/garth/types"
	shared "github.com/sstent/go-garth/shared/interfaces"
)

// BodyBatteryReading represents a single body battery data point
type BodyBatteryReading struct {
	Timestamp int     `json:"timestamp"`
	Status    string  `json:"status"`
	Level     int     `json:"level"`
	Version   float64 `json:"version"`
}

// ParseBodyBatteryReadings converts body battery values array to structured readings
// Accepts mixed numeric types (int, int64, float64, json.Number) for robustness.
func ParseBodyBatteryReadings(valuesArray [][]any) []BodyBatteryReading {
	readings := make([]BodyBatteryReading, 0, len(valuesArray))

	toInt := func(v any) (int, bool) {
		switch t := v.(type) {
		case int:
			return t, true
		case int32:
			return int(t), true
		case int64:
			return int(t), true
		case float32:
			return int(t), true
		case float64:
			return int(t), true
		case json.Number:
			i, err := t.Int64()
			if err == nil {
				return int(i), true
			}
			f, err := t.Float64()
			if err == nil {
				return int(f), true
			}
			return 0, false
		default:
			return 0, false
		}
	}

	toFloat64 := func(v any) (float64, bool) {
		switch t := v.(type) {
		case float32:
			return float64(t), true
		case float64:
			return t, true
		case int:
			return float64(t), true
		case int32:
			return float64(t), true
		case int64:
			return float64(t), true
		case json.Number:
			f, err := t.Float64()
			if err == nil {
				return f, true
			}
			return 0, false
		default:
			return 0, false
		}
	}

	for _, values := range valuesArray {
		if len(values) < 4 {
			continue
		}

		ts, ok1 := toInt(values[0])
		status, ok2 := values[1].(string)
		lvl, ok3 := toInt(values[2])
		ver, ok4 := toFloat64(values[3])

		if !ok1 || !ok2 || !ok3 || !ok4 {
			continue
		}

		readings = append(readings, BodyBatteryReading{
			Timestamp: ts,
			Status:    status,
			Level:     lvl,
			Version:   ver,
		})
	}

	sort.Slice(readings, func(i, j int) bool {
		return readings[i].Timestamp < readings[j].Timestamp
	})

	return readings
}

// BodyBatteryDataWithMethods embeds garth.DetailedBodyBatteryData and adds methods
type BodyBatteryDataWithMethods struct {
	garth.DetailedBodyBatteryData
}

func (d *BodyBatteryDataWithMethods) Get(day time.Time, c shared.APIClient) (interface{}, error) {
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

	var result garth.DetailedBodyBatteryData
	if len(data1) > 0 {
		if err := json.Unmarshal(data1, &result); err != nil {
			return nil, fmt.Errorf("failed to parse Body Battery data: %w", err)
		}
	}

	var events []garth.BodyBatteryEvent
	if len(data2) > 0 {
		if err := json.Unmarshal(data2, &events); err == nil {
			result.Events = events
		}
	}

	return &BodyBatteryDataWithMethods{DetailedBodyBatteryData: result}, nil
}

// GetCurrentLevel returns the most recent Body Battery level
func (d *BodyBatteryDataWithMethods) GetCurrentLevel() int {
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
func (d *BodyBatteryDataWithMethods) GetDayChange() int {
	readings := ParseBodyBatteryReadings(d.BodyBatteryValuesArray)
	if len(readings) < 2 {
		return 0
	}

	return readings[len(readings)-1].Level - readings[0].Level
}

// Added for test compatibility and public API alignment
// DailyBodyBatteryStress wraps garth.DetailedBodyBatteryData and provides a Get method compatible with existing tests.
// See [type DailyBodyBatteryStress](internal/data/body_battery.go:0) and [func (*DailyBodyBatteryStress).Get](internal/data/body_battery.go:0)
type DailyBodyBatteryStress struct {
	garth.DetailedBodyBatteryData
}

// Get retrieves Body Battery daily stress data and associated events for a given day.
// Mirrors logic in BodyBatteryDataWithMethods.Get to maintain a consistent behavior.
// Returns (*DailyBodyBatteryStress, nil) on success, (nil, nil) when no data available.
func (d *DailyBodyBatteryStress) Get(day time.Time, c shared.APIClient) (interface{}, error) {
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

	var result garth.DetailedBodyBatteryData
	if len(data1) > 0 {
		if err := json.Unmarshal(data1, &result); err != nil {
			return nil, fmt.Errorf("failed to parse Body Battery data: %w", err)
		}
	}

	var events []garth.BodyBatteryEvent
	if len(data2) > 0 {
		if err := json.Unmarshal(data2, &events); err == nil {
			result.Events = events
		}
	}

	return &DailyBodyBatteryStress{DetailedBodyBatteryData: result}, nil
}
