package garmin

import (
	"garmin-connect/internal/data"
	"time"
)

// BodyBatteryData represents Body Battery data.
type BodyBatteryData = data.DailyBodyBatteryStress

// SleepData represents sleep data.
type SleepData = data.DailySleepDTO

// HRVData represents HRV data.
type HRVData = data.HRVData

// WeightData represents weight data.
type WeightData = data.WeightData

// GetBodyBattery retrieves Body Battery data for a given date.
func (c *Client) GetBodyBattery(date time.Time) (*BodyBatteryData, error) {
	bb := &data.DailyBodyBatteryStress{}
	result, err := bb.Get(date, c.Client)
	if err != nil {
		return nil, err
	}
	return result.(*BodyBatteryData), nil
}

// GetSleep retrieves sleep data for a given date.
func (c *Client) GetSleep(date time.Time) (*SleepData, error) {
	sleep := &data.DailySleepDTO{}
	result, err := sleep.Get(date, c.Client)
	if err != nil {
		return nil, err
	}
	return result.(*SleepData), nil
}

// GetHRV retrieves HRV data for a given date.
func (c *Client) GetHRV(date time.Time) (*HRVData, error) {
	hrv := &data.HRVData{}
	result, err := hrv.Get(date, c.Client)
	if err != nil {
		return nil, err
	}
	return result.(*HRVData), nil
}

// GetWeight retrieves weight data for a given date.
func (c *Client) GetWeight(date time.Time) (*WeightData, error) {
	weight := &data.WeightData{}
	result, err := weight.Get(date, c.Client)
	if err != nil {
		return nil, err
	}
	return result.(*WeightData), nil
}
