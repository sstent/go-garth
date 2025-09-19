package types

import (
	"fmt"
	"strings"
	"time"
)

// GarminTime represents Garmin's timestamp format with custom JSON parsing
type GarminTime struct {
	time.Time
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// It parses Garmin's specific timestamp format.
func (gt *GarminTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), `"`)
	if s == "null" {
		return nil
	}

	// Try parsing with milliseconds (e.g., "2018-09-01T00:13:25.000")
	// Garmin sometimes returns .0 for milliseconds, which Go's time.Parse handles as .000
	// The 'Z' in the layout indicates a UTC time without a specific offset, which is often how these are interpreted.
	// If the input string does not contain 'Z', it will be parsed as local time.
	// For consistency, we'll assume UTC if no timezone is specified.
	layouts := []string{
		"2006-01-02T15:04:05.0", // Example: 2018-09-01T00:13:25.0
		"2006-01-02T15:04:05",   // Example: 2018-09-01T00:13:25
		"2006-01-02",            // Example: 2018-09-01
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			gt.Time = t
			return nil
		}
	}

	return fmt.Errorf("cannot parse %q into a GarminTime", s)
}

// SessionData represents saved session information
type SessionData struct {
	Domain    string `json:"domain"`
	Username  string `json:"username"`
	AuthToken string `json:"auth_token"`
}

// ActivityType represents the type of activity
type ActivityType struct {
	TypeID       int    `json:"typeId"`
	TypeKey      string `json:"typeKey"`
	ParentTypeID *int   `json:"parentTypeId,omitempty"`
}

// EventType represents the event type of an activity
type EventType struct {
	TypeID  int    `json:"typeId"`
	TypeKey string `json:"typeKey"`
}

// Activity represents a Garmin Connect activity
type Activity struct {
	ActivityID      int64        `json:"activityId"`
	ActivityName    string       `json:"activityName"`
	Description     string       `json:"description"`
	StartTimeLocal  GarminTime   `json:"startTimeLocal"`
	StartTimeGMT    GarminTime   `json:"startTimeGMT"`
	ActivityType    ActivityType `json:"activityType"`
	EventType       EventType    `json:"eventType"`
	Distance        float64      `json:"distance"`
	Duration        float64      `json:"duration"`
	ElapsedDuration float64      `json:"elapsedDuration"`
	MovingDuration  float64      `json:"movingDuration"`
	ElevationGain   float64      `json:"elevationGain"`
	ElevationLoss   float64      `json:"elevationLoss"`
	AverageSpeed    float64      `json:"averageSpeed"`
	MaxSpeed        float64      `json:"maxSpeed"`
	Calories        float64      `json:"calories"`
	AverageHR       float64      `json:"averageHR"`
	MaxHR           float64      `json:"maxHR"`
}

// UserProfile represents a Garmin user profile
type UserProfile struct {
	UserName        string     `json:"userName"`
	DisplayName     string     `json:"displayName"`
	LevelUpdateDate GarminTime `json:"levelUpdateDate"`
	// Add other fields as needed from API response
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

// StepsData represents steps statistics
type StepsData struct {
	Date  time.Time `json:"calendarDate"`
	Steps int       `json:"steps"`
}

// DistanceData represents distance statistics
type DistanceData struct {
	Date     time.Time `json:"calendarDate"`
	Distance float64   `json:"distance"` // in meters
}

// CaloriesData represents calories statistics
type CaloriesData struct {
	Date     time.Time `json:"calendarDate"`
	Calories int       `json:"activeCalories"`
}