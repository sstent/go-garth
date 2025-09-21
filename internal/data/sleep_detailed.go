package data

import (
	"encoding/json"
	"fmt"
	"time"

	types "go-garth/internal/models/types"
	shared "go-garth/shared/interfaces"
)

// DetailedSleepDataWithMethods embeds types.DetailedSleepData and adds methods
type DetailedSleepDataWithMethods struct {
	types.DetailedSleepData
}

func (d *DetailedSleepDataWithMethods) Get(day time.Time, c shared.APIClient) (interface{}, error) {
	dateStr := day.Format("2006-01-02")
	path := fmt.Sprintf("/wellness-service/wellness/dailySleepData/%s?date=%s&nonSleepBufferMinutes=60",
		c.GetUsername(), dateStr)

	data, err := c.ConnectAPI(path, "GET", nil, nil)
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

	return &DetailedSleepDataWithMethods{DetailedSleepData: *response.DailySleepDTO}, nil
}

// GetSleepEfficiency calculates sleep efficiency percentage
func (d *DetailedSleepDataWithMethods) GetSleepEfficiency() float64 {
	totalTime := d.SleepEndTimestampGMT.Sub(d.SleepStartTimestampGMT).Seconds()
	sleepTime := float64(d.DeepSleepSeconds + d.LightSleepSeconds + d.RemSleepSeconds)
	if totalTime == 0 {
		return 0
	}
	return (sleepTime / totalTime) * 100
}

// GetTotalSleepTime returns total sleep time in hours
func (d *DetailedSleepDataWithMethods) GetTotalSleepTime() float64 {
	totalSeconds := d.DeepSleepSeconds + d.LightSleepSeconds + d.RemSleepSeconds
	return float64(totalSeconds) / 3600.0
}
