package data

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sstent/go-garth/garth/client"
)

// SleepScores represents sleep scoring data
type SleepScores struct {
	TotalSleepSeconds int `json:"totalSleepSeconds"`
	SleepScores       []struct {
		StartTimeGMT time.Time `json:"startTimeGmt"`
		EndTimeGMT   time.Time `json:"endTimeGmt"`
		SleepScore   int       `json:"sleepScore"`
	} `json:"sleepScores"`
	SleepMovement []SleepMovement `json:"sleepMovement"`
	// Add other fields from Python implementation
}

// SleepMovement represents movement during sleep
type SleepMovement struct {
	StartGMT      time.Time `json:"startGmt"`
	EndGMT        time.Time `json:"endGmt"`
	ActivityLevel int       `json:"activityLevel"`
}

// DailySleepDTO represents daily sleep data
type DailySleepDTO struct {
	UserProfilePK          int         `json:"userProfilePk"`
	CalendarDate           time.Time   `json:"calendarDate"`
	SleepStartTimestampGMT time.Time   `json:"sleepStartTimestampGmt"`
	SleepEndTimestampGMT   time.Time   `json:"sleepEndTimestampGmt"`
	SleepScores            SleepScores `json:"sleepScores"`
	BaseData
}

// Get implements the Data interface for DailySleepDTO
func (d *DailySleepDTO) Get(day time.Time, client *client.Client) (any, error) {
	dateStr := day.Format("2006-01-02")
	path := fmt.Sprintf("/wellness-service/wellness/dailySleepData/%s?nonSleepBufferMinutes=60&date=%s",
		client.Username, dateStr)

	data, err := client.ConnectAPI(path, "GET", nil, nil)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, nil
	}

	var response struct {
		DailySleepDTO *DailySleepDTO  `json:"dailySleepDto"`
		SleepMovement []SleepMovement `json:"sleepMovement"`
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
func (d *DailySleepDTO) List(end time.Time, days int, client *client.Client, maxWorkers int) ([]any, error) {
	// Implementation to be added
	return []any{}, nil
}
