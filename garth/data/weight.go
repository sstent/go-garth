package data

import (
	"encoding/json"
	"fmt"
	"time"

	"garmin-connect/garth/client"
)

// WeightData represents weight measurement data
type WeightData struct {
	UserProfilePK     int       `json:"userProfilePk"`
	CalendarDate      time.Time `json:"calendarDate"`
	Timestamp         time.Time `json:"timestamp"`
	Weight            float64   `json:"weight"` // in kilograms
	BMI               float64   `json:"bmi"`
	BodyFatPercentage float64   `json:"bodyFatPercentage"`
	BoneMass          float64   `json:"boneMass"`   // in kg
	MuscleMass        float64   `json:"muscleMass"` // in kg
	Hydration         float64   `json:"hydration"`  // in kg
	BaseData
}

// Validate checks if weight data contains valid values
func (w *WeightData) Validate() error {
	if w.Weight <= 0 {
		return fmt.Errorf("invalid weight value")
	}
	if w.BMI < 10 || w.BMI > 50 {
		return fmt.Errorf("BMI out of valid range")
	}
	return nil
}

// Get implements the Data interface for WeightData
func (w *WeightData) Get(day time.Time, client *client.Client) (any, error) {
	startDate := day.Format("2006-01-02")
	endDate := day.Format("2006-01-02")
	path := fmt.Sprintf("/weight-service/weight/dateRange?startDate=%s&endDate=%s",
		startDate, endDate)

	data, err := client.ConnectAPI(path, "GET", nil, nil)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	var response struct {
		WeightList []WeightData `json:"weightList"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	if len(response.WeightList) == 0 {
		return nil, nil
	}

	weightData := response.WeightList[0]
	// Convert grams to kilograms
	weightData.Weight = weightData.Weight / 1000
	weightData.BoneMass = weightData.BoneMass / 1000
	weightData.MuscleMass = weightData.MuscleMass / 1000
	weightData.Hydration = weightData.Hydration / 1000

	return weightData, nil
}

// List implements the Data interface for concurrent fetching
func (w *WeightData) List(end time.Time, days int, client *client.Client, maxWorkers int) ([]any, error) {
	results, errs := w.BaseData.List(end, days, client, maxWorkers)
	if len(errs) > 0 {
		// Return first error for now
		return results, errs[0]
	}
	return results, nil
}
