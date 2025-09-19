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

// VO2MaxData represents VO2 max data
type VO2MaxData struct {
	Date       time.Time `json:"calendarDate"`
	VO2MaxRunning float64   `json:"vo2MaxRunning"`
	VO2MaxCycling float64   `json:"vo2MaxCycling"`
	// Add more fields as needed
}

// HeartRateZones represents heart rate zone data
type HeartRateZones struct {
	RestingHR        int       `json:"resting_hr"`
	MaxHR            int       `json:"max_hr"`
	LactateThreshold int       `json:"lactate_threshold"`
	Zones            []HRZone  `json:"zones"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// HRZone represents a single heart rate zone
type HRZone struct {
	Zone   int    `json:"zone"`
	MinBPM int    `json:"min_bpm"`
	MaxBPM int    `json:"max_bpm"`
	Name   string `json:"name"`
}

// WellnessData represents additional wellness metrics
type WellnessData struct {
	Date         time.Time `json:"calendarDate"`
	RestingHR    *int      `json:"resting_hr"`
	Weight       *float64  `json:"weight"`
	BodyFat      *float64  `json:"body_fat"`
	BMI          *float64  `json:"bmi"`
	BodyWater    *float64  `json:"body_water"`
	BoneMass     *float64  `json:"bone_mass"`
	MuscleMass   *float64  `json:"muscle_mass"`
	// Add more fields as needed
}