package garmin

import (
	"time"
)

// SleepData represents sleep summary data
type SleepData struct {
	Date                 time.Time `json:"calendarDate"`
	SleepScore           int       `json:"sleepScore"`
	TotalSleepSeconds    int       `json:"totalSleepSeconds"`
	DeepSleepSeconds     int       `json:"deepSleepSeconds"`
	LightSleepSeconds    int       `json:"lightSleepSeconds"`
	RemSleepSeconds      int       `json:"remSleepSeconds"`
	AwakeSleepSeconds    int       `json:"awakeSleepSeconds"`
	// Add more fields as needed
}

// HrvData represents Heart Rate Variability data
type HrvData struct {
	Date                 time.Time `json:"calendarDate"`
	HrvValue             float64   `json:"hrvValue"`
	// Add more fields as needed
}

// StressData represents stress level data
type StressData struct {
	Date                 time.Time `json:"calendarDate"`
	StressLevel          int       `json:"stressLevel"`
	RestStressLevel      int       `json:"restStressLevel"`
	// Add more fields as needed
}

// BodyBatteryData represents Body Battery data
type BodyBatteryData struct {
	Date                 time.Time `json:"calendarDate"`
	BatteryLevel         int       `json:"batteryLevel"`
	Charge               int       `json:"charge"`
	Drain                int       `json:"drain"`
	// Add more fields as needed
}