package data

import (
	"encoding/json"
	"fmt"
	"time"

	types "go-garth/internal/models/types"
	shared "go-garth/shared/interfaces"
)

// TrainingStatusWithMethods embeds types.TrainingStatus and adds methods
type TrainingStatusWithMethods struct {
	types.TrainingStatus
}

func (t *TrainingStatusWithMethods) Get(day time.Time, c shared.APIClient) (interface{}, error) {
	dateStr := day.Format("2006-01-02")
	path := fmt.Sprintf("/metrics-service/metrics/trainingStatus/%s", dateStr)

	data, err := c.ConnectAPI(path, "GET", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get training status: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var result types.TrainingStatus
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse training status: %w", err)
	}

	return &TrainingStatusWithMethods{TrainingStatus: result}, nil
}

// TrainingLoadWithMethods embeds types.TrainingLoad and adds methods
type TrainingLoadWithMethods struct {
	types.TrainingLoad
}

func (t *TrainingLoadWithMethods) Get(day time.Time, c shared.APIClient) (interface{}, error) {
	dateStr := day.Format("2006-01-02")
	endDate := day.AddDate(0, 0, 6).Format("2006-01-02") // Get week of data
	path := fmt.Sprintf("/metrics-service/metrics/trainingLoad/%s/%s", dateStr, endDate)

	data, err := c.ConnectAPI(path, "GET", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get training load: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var results []types.TrainingLoad
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("failed to parse training load: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &TrainingLoadWithMethods{TrainingLoad: results[0]}, nil
}
