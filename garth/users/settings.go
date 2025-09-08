package users

import (
	"time"

	"garmin-connect/garth/client"
)

type PowerFormat struct {
	FormatID      int     `json:"formatId"`
	FormatKey     string  `json:"formatKey"`
	MinFraction   int     `json:"minFraction"`
	MaxFraction   int     `json:"maxFraction"`
	GroupingUsed  bool    `json:"groupingUsed"`
	DisplayFormat *string `json:"displayFormat"`
}

type FirstDayOfWeek struct {
	DayID              int    `json:"dayId"`
	DayName            string `json:"dayName"`
	SortOrder          int    `json:"sortOrder"`
	IsPossibleFirstDay bool   `json:"isPossibleFirstDay"`
}

type WeatherLocation struct {
	UseFixedLocation *bool    `json:"useFixedLocation"`
	Latitude         *float64 `json:"latitude"`
	Longitude        *float64 `json:"longitude"`
	LocationName     *string  `json:"locationName"`
	ISOCountryCode   *string  `json:"isoCountryCode"`
	PostalCode       *string  `json:"postalCode"`
}

type UserData struct {
	Gender                         string                   `json:"gender"`
	Weight                         float64                  `json:"weight"`
	Height                         float64                  `json:"height"`
	TimeFormat                     string                   `json:"timeFormat"`
	BirthDate                      time.Time                `json:"birthDate"`
	MeasurementSystem              string                   `json:"measurementSystem"`
	ActivityLevel                  *string                  `json:"activityLevel"`
	Handedness                     string                   `json:"handedness"`
	PowerFormat                    PowerFormat              `json:"powerFormat"`
	HeartRateFormat                PowerFormat              `json:"heartRateFormat"`
	FirstDayOfWeek                 FirstDayOfWeek           `json:"firstDayOfWeek"`
	VO2MaxRunning                  *float64                 `json:"vo2MaxRunning"`
	VO2MaxCycling                  *float64                 `json:"vo2MaxCycling"`
	LactateThresholdSpeed          *float64                 `json:"lactateThresholdSpeed"`
	LactateThresholdHeartRate      *float64                 `json:"lactateThresholdHeartRate"`
	DiveNumber                     *int                     `json:"diveNumber"`
	IntensityMinutesCalcMethod     string                   `json:"intensityMinutesCalcMethod"`
	ModerateIntensityMinutesHRZone int                      `json:"moderateIntensityMinutesHrZone"`
	VigorousIntensityMinutesHRZone int                      `json:"vigorousIntensityMinutesHrZone"`
	HydrationMeasurementUnit       string                   `json:"hydrationMeasurementUnit"`
	HydrationContainers            []map[string]interface{} `json:"hydrationContainers"`
	HydrationAutoGoalEnabled       bool                     `json:"hydrationAutoGoalEnabled"`
	FirstbeatMaxStressScore        *float64                 `json:"firstbeatMaxStressScore"`
	FirstbeatCyclingLTTimestamp    *int64                   `json:"firstbeatCyclingLtTimestamp"`
	FirstbeatRunningLTTimestamp    *int64                   `json:"firstbeatRunningLtTimestamp"`
	ThresholdHeartRateAutoDetected bool                     `json:"thresholdHeartRateAutoDetected"`
	FTPAutoDetected                *bool                    `json:"ftpAutoDetected"`
	TrainingStatusPausedDate       *string                  `json:"trainingStatusPausedDate"`
	WeatherLocation                *WeatherLocation         `json:"weatherLocation"`
	GolfDistanceUnit               *string                  `json:"golfDistanceUnit"`
	GolfElevationUnit              *string                  `json:"golfElevationUnit"`
	GolfSpeedUnit                  *string                  `json:"golfSpeedUnit"`
	ExternalBottomTime             *float64                 `json:"externalBottomTime"`
}

type UserSleep struct {
	SleepTime        int  `json:"sleepTime"`
	DefaultSleepTime bool `json:"defaultSleepTime"`
	WakeTime         int  `json:"wakeTime"`
	DefaultWakeTime  bool `json:"defaultWakeTime"`
}

type UserSleepWindow struct {
	SleepWindowFrequency              string `json:"sleepWindowFrequency"`
	StartSleepTimeSecondsFromMidnight int    `json:"startSleepTimeSecondsFromMidnight"`
	EndSleepTimeSecondsFromMidnight   int    `json:"endSleepTimeSecondsFromMidnight"`
}

type UserSettings struct {
	ID               int               `json:"id"`
	UserData         UserData          `json:"userData"`
	UserSleep        UserSleep         `json:"userSleep"`
	ConnectDate      *string           `json:"connectDate"`
	SourceType       *string           `json:"sourceType"`
	UserSleepWindows []UserSleepWindow `json:"userSleepWindows,omitempty"`
}

func GetSettings(c *client.Client) (*UserSettings, error) {
	// Implementation will be added in client.go
	return nil, nil
}
