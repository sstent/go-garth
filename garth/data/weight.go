package data

import (
	"errors"
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
	BaseData
}

// Validate checks if weight data contains valid values
func (w *WeightData) Validate() error {
	if w.Weight <= 0 {
		return errors.New("invalid weight value")
	}
	return nil
}

// Get implements the Data interface for WeightData
func (w *WeightData) Get(day time.Time, client *client.Client) (any, error) {
	// Implementation to be added
	return nil, nil
}

// List implements the Data interface for concurrent fetching
func (w *WeightData) List(end time.Time, days int, client *client.Client, maxWorkers int) ([]any, error) {
	// Implementation to be added
	return []any{}, nil
}
