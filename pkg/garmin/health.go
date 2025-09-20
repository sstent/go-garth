package garmin

import (
	"encoding/json"
	"fmt"
	"time"

	internalClient "go-garth/internal/api/client"
	"go-garth/internal/models/types"
)

func (c *Client) GetDailyHRVData(date time.Time) (*types.DailyHRVData, error) {
	return getDailyHRVData(date, c.Client)
}

func getDailyHRVData(day time.Time, client *internalClient.Client) (*types.DailyHRVData, error) {
	dateStr := day.Format("2006-01-02")
	path := fmt.Sprintf("/wellness-service/wellness/dailyHrvData/%s?date=%s",
		client.Username, dateStr)

	data, err := client.ConnectAPI(path, "GET", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get HRV data: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var response struct {
		HRVSummary  types.DailyHRVData `json:"hrvSummary"`
		HRVReadings []types.HRVReading `json:"hrvReadings"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse HRV response: %w", err)
	}

	// Combine summary and readings
	response.HRVSummary.HRVReadings = response.HRVReadings
	return &response.HRVSummary, nil
}

func (c *Client) GetDetailedSleepData(date time.Time) (*types.DetailedSleepData, error) {
	return getDetailedSleepData(date, c.Client)
}

func getDetailedSleepData(day time.Time, client *internalClient.Client) (*types.DetailedSleepData, error) {
	dateStr := day.Format("2006-01-02")
	path := fmt.Sprintf("/wellness-service/wellness/dailySleepData/%s?date=%s&nonSleepBufferMinutes=60",
		client.Username, dateStr)

	data, err := client.ConnectAPI(path, "GET", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get detailed sleep data: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var response struct {
		DailySleepDTO                       *types.DetailedSleepData `json:"dailySleepDTO"`
		SleepMovement                       []types.SleepMovement    `json:"sleepMovement"`
		RemSleepData                        bool                     `json:"remSleepData"`
		SleepLevels                         []types.SleepLevel       `json:"sleepLevels"`
		SleepRestlessMoments                []interface{}            `json:"sleepRestlessMoments"`
		RestlessMomentsCount                int                      `json:"restlessMomentsCount"`
		WellnessSpO2SleepSummaryDTO         interface{}              `json:"wellnessSpO2SleepSummaryDTO"`
		WellnessEpochSPO2DataDTOList        []interface{}            `json:"wellnessEpochSPO2DataDTOList"`
		WellnessEpochRespirationDataDTOList []interface{}            `json:"wellnessEpochRespirationDataDTOList"`
		SleepStress                         interface{}              `json:"sleepStress"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse detailed sleep response: %w", err)
	}

	if response.DailySleepDTO == nil {
		return nil, nil
	}

	// Populate additional data
	response.DailySleepDTO.SleepMovement = response.SleepMovement
	response.DailySleepDTO.SleepLevels = response.SleepLevels

	return response.DailySleepDTO, nil
}
