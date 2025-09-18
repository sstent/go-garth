package types

import "time"

// GarminTime represents Garmin's timestamp format with custom JSON parsing
type GarminTime struct {
	time.Time
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
