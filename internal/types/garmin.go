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
	StartTimeLocal  string       `json:"startTimeLocal"`
	StartTimeGMT    string       `json:"startTimeGMT"`
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
