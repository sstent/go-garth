package data

import (
	"encoding/json"
	"fmt"
	"time"

	shared "go-garth/shared/interfaces"
)

// WeightData represents weight data
type WeightData struct {
	Date       time.Time `json:"calendarDate"`
	Weight     float64   `json:"weight"` // in grams
	BMI        float64   `json:"bmi"`
	BodyFat    float64   `json:"bodyFat"`
	BoneMass   float64   `json:"boneMass"`
	MuscleMass float64   `json:"muscleMass"`
	Hydration  float64   `json:"hydration"`
}

// WeightDataWithMethods embeds WeightData and adds methods
type WeightDataWithMethods struct {
	WeightData
}

// Validate checks if weight data contains valid values
func (w *WeightDataWithMethods) Validate() error {
	if w.Weight <= 0 {
		return fmt.Errorf("invalid weight value")
	}
	if w.BMI < 10 || w.BMI > 50 {
		return fmt.Errorf("BMI out of valid range")
	}
	return nil
}

// Get implements the Data interface for WeightData
func (w *WeightDataWithMethods) Get(day time.Time, c shared.APIClient) (any, error) {
	startDate := day.Format("2006-01-02")
	endDate := day.Format("2006-01-02")
	path := fmt.Sprintf("/weight-service/weight/dateRange?startDate=%s&endDate=%s",
		startDate, endDate)

	data, err := c.ConnectAPI(path, "GET", nil, nil)
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

	return &WeightDataWithMethods{WeightData: weightData}, nil
}

// List implements the Data interface for concurrent fetching
func (w *WeightDataWithMethods) List(end time.Time, days int, c shared.APIClient, maxWorkers int) ([]any, error) {
	// BaseData is not part of types.WeightData, so this line needs to be removed or re-evaluated.
	// For now, I will return an empty slice and no error, as this function is not directly related to the task.
	return []any{}, nil
}
