package data

import (
	"encoding/json"
	"fmt"
	"time"

	"garmin-connect/garth/client"
	"garmin-connect/garth/errors"
	"garmin-connect/garth/utils"
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

	dailySleepDto, exists := snakeResponse["daily_sleep_dto"].(map[string]interface{})
	if !exists || dailySleepDto["id"] == nil {
		return nil, nil // No sleep data
	}

	jsonBytes, err := json.Marshal(snakeResponse)
	if err != nil {
		return nil, err
	}

	var result struct {
		DailySleepDTO *DailySleepDTO  `json:"daily_sleep_dto"`
		SleepMovement []SleepMovement `json:"sleep_movement"`
	}

	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// List implements the Data interface for concurrent fetching
func (d *DailySleepDTO) List(end time.Time, days int, client *client.Client, maxWorkers int) ([]any, error) {
	// Implementation to be added
	return []any{}, nil
}
