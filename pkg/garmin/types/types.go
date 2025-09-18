package types

// DailyBodyBatteryStress represents Garmin's daily body battery and stress data
type DailyBodyBatteryStress struct {
	CalendarDate string  `json:"calendarDate"`
	BodyBattery  int     `json:"bodyBattery"`
	Stress       int     `json:"stress"`
	RestStress   float64 `json:"restStress"`
}

// Activity represents a Garmin activity
type Activity struct {
	ActivityID   int64   `json:"activityId"`
	StartTime    string  `json:"startTime"`
	ActivityType string  `json:"activityType"`
	Duration     float64 `json:"duration"`
	Distance     float64 `json:"distance"`
}

// OAuth2Token represents the authentication tokens from Garmin's OAuth flow
type OAuth2Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}
