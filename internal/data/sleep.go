package data

import (
	"encoding/json"
	"fmt"
	"time"

	garth "github.com/sstent/go-garth/pkg/garth/types"
	shared "github.com/sstent/go-garth/shared/interfaces"
)

// DailySleepDTO represents daily sleep data
type DailySleepDTO struct {
	UserProfilePK          int              `json:"userProfilePk"`
	CalendarDate           time.Time        `json:"calendarDate"`
	SleepStartTimestampGMT time.Time        `json:"sleepStartTimestampGmt"`
	SleepEndTimestampGMT   time.Time        `json:"sleepEndTimestampGmt"`
	SleepScores            garth.SleepScore `json:"sleepScores"` // Using garth.SleepScore
	shared.BaseData
}

// Get implements the Data interface for DailySleepDTO
func (d *DailySleepDTO) Get(day time.Time, c shared.APIClient) (any, error) {
	dateStr := day.Format("2006-01-02")
	path := fmt.Sprintf("/wellness-service/wellness/dailySleepData/%s?nonSleepBufferMinutes=60&date=%s",
		c.GetUsername(), dateStr)

	data, err := c.ConnectAPI(path, "GET", nil, nil)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	var response struct {
		DailySleepDTO *DailySleepDTO        `json:"dailySleepDto"`
		SleepMovement []garth.SleepMovement `json:"sleepMovement"` // Using garth.SleepMovement
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	if response.DailySleepDTO == nil {
		return nil, nil
	}

	return response, nil
}

// List implements the Data interface for concurrent fetching
func (d *DailySleepDTO) List(end time.Time, days int, c shared.APIClient, maxWorkers int) ([]any, error) {
	// Implementation to be added
	return []any{}, nil
}
