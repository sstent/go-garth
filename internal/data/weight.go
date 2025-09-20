package data

import (
	"encoding/json"
	"fmt"
	"time"

	shared "go-garth/shared/interfaces"
	types "go-garth/internal/models/types"
)

// Validate checks if weight data contains valid values
func (w *types.WeightData) Validate() error {
	if w.Weight <= 0 {
		return fmt.Errorf("invalid weight value")
	}
	if w.BMI < 10 || w.BMI > 50 {
		return fmt.Errorf("BMI out of valid range")
	}
	return nil
}

// Get implements the Data interface for WeightData
func (w *types.WeightData) Get(day time.Time, c shared.APIClient) (any, error) {
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
		WeightList []types.WeightData `json:"weightList"`
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
func (w *types.WeightData) List(end time.Time, days int, c shared.APIClient, maxWorkers int) ([]any, error) {
	results, errs := w.BaseData.List(end, days, c, maxWorkers)
	if len(errs) > 0 {
		// Return first error for now
		return results, errs[0]
	}
	return results, nil
}
