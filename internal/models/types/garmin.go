package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	// Default location for conversions (set to UTC by default)
	defaultLocation *time.Location
)

func init() {
	var err error
	defaultLocation, err = time.LoadLocation("UTC")
	if err != nil {
		panic(err)
	}
}

// ParseTimestamp converts a millisecond timestamp to time.Time in default location
func ParseTimestamp(ts int) time.Time {
	return time.Unix(0, int64(ts)*int64(time.Millisecond)).In(defaultLocation)
}

// parseAggregationKey is a helper function to parse aggregation key back to a time.Time object
func ParseAggregationKey(key, aggregate string) time.Time {
	switch aggregate {
	case "day":
		t, _ := time.Parse("2006-01-02", key)
		return t
	case "week":
		year, _ := strconv.Atoi(key[:4])
		week, _ := strconv.Atoi(key[6:])
		t := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		// Find the first Monday of the year
		for t.Weekday() != time.Monday {
			t = t.AddDate(0, 0, 1)
		}
		// Add weeks
		return t.AddDate(0, 0, (week-1)*7)
	case "month":
		t, _ := time.Parse("2006-01", key)
		return t
	case "year":
		t, _ := time.Parse("2006", key)
		return t
	}
	return time.Time{}
}

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
	Date          time.Time `json:"calendarDate"`
	VO2MaxRunning *float64  `json:"vo2MaxRunning"`
	VO2MaxCycling *float64  `json:"vo2MaxCycling"`
	UserProfilePK int       `json:"userProfilePk"`
}

// Add these new structs
type VO2MaxEntry struct {
	Value        float64   `json:"value"`
	ActivityType string    `json:"activityType"` // "running" or "cycling"
	Date         time.Time `json:"date"`
	Source       string    `json:"source"` // "user_settings", "activity", etc.
}

type VO2Max struct {
	Value        float64   `json:"vo2Max"`
	FitnessLevel string    `json:"fitnessLevel"`
	UpdatedDate  time.Time `json:"date"`
}

// VO2MaxProfile represents the current VO2 max profile from user settings
type VO2MaxProfile struct {
	UserProfilePK int          `json:"userProfilePk"`
	LastUpdated   time.Time    `json:"lastUpdated"`
	Running       *VO2MaxEntry `json:"running,omitempty"`
	Cycling       *VO2MaxEntry `json:"cycling,omitempty"`
}

// SleepLevel represents different sleep stages
type SleepLevel struct {
	StartGMT      time.Time `json:"startGmt"`
	EndGMT        time.Time `json:"endGmt"`
	ActivityLevel float64   `json:"activityLevel"`
	SleepLevel    string    `json:"sleepLevel"` // "deep", "light", "rem", "awake"
}

// SleepMovement represents movement during sleep
type SleepMovement struct {
	StartGMT      time.Time `json:"startGmt"`
	EndGMT        time.Time `json:"endGmt"`
	ActivityLevel float64   `json:"activityLevel"`
}

// SleepScore represents detailed sleep scoring
type SleepScore struct {
	Overall          int                 `json:"overall"`
	Composition      SleepScoreBreakdown `json:"composition"`
	Revitalization   SleepScoreBreakdown `json:"revitalization"`
	Duration         SleepScoreBreakdown `json:"duration"`
	DeepPercentage   float64             `json:"deepPercentage"`
	LightPercentage  float64             `json:"lightPercentage"`
	RemPercentage    float64             `json:"remPercentage"`
	RestfulnessValue float64             `json:"restfulnessValue"`
}

type SleepScoreBreakdown struct {
	QualifierKey   string  `json:"qualifierKey"`
	OptimalStart   float64 `json:"optimalStart"`
	OptimalEnd     float64 `json:"optimalEnd"`
	Value          float64 `json:"value"`
	IdealStartSecs *int    `json:"idealStartInSeconds"`
	IdealEndSecs   *int    `json:"idealEndInSeconds"`
}

// DetailedSleepData represents comprehensive sleep data
type DetailedSleepData struct {
	UserProfilePK            int             `json:"userProfilePk"`
	CalendarDate             time.Time       `json:"calendarDate"`
	SleepStartTimestampGMT   time.Time       `json:"sleepStartTimestampGmt"`
	SleepEndTimestampGMT     time.Time       `json:"sleepEndTimestampGmt"`
	SleepStartTimestampLocal time.Time       `json:"sleepStartTimestampLocal"`
	SleepEndTimestampLocal   time.Time       `json:"sleepEndTimestampLocal"`
	UnmeasurableSleepSeconds int             `json:"unmeasurableSleepSeconds"`
	DeepSleepSeconds         int             `json:"deepSleepSeconds"`
	LightSleepSeconds        int             `json:"lightSleepSeconds"`
	RemSleepSeconds          int             `json:"remSleepSeconds"`
	AwakeSleepSeconds        int             `json:"awakeSleepSeconds"`
	DeviceRemCapable         bool            `json:"deviceRemCapable"`
	SleepLevels              []SleepLevel    `json:"sleepLevels"`
	SleepMovement            []SleepMovement `json:"sleepMovement"`
	SleepScores              *SleepScore     `json:"sleepScores"`
	AverageSpO2Value         *float64        `json:"averageSpO2Value"`
	LowestSpO2Value          *int            `json:"lowestSpO2Value"`
	HighestSpO2Value         *int            `json:"highestSpO2Value"`
	AverageRespirationValue  *float64        `json:"averageRespirationValue"`
	LowestRespirationValue   *float64        `json:"lowestRespirationValue"`
	HighestRespirationValue  *float64        `json:"highestRespirationValue"`
	AvgSleepStress           *float64        `json:"avgSleepStress"`
}

// HRVBaseline represents HRV baseline data
type HRVBaseline struct {
	LowUpper      int     `json:"lowUpper"`
	BalancedLow   int     `json:"balancedLow"`
	BalancedUpper int     `json:"balancedUpper"`
	MarkerValue   float64 `json:"markerValue"`
}

// DailyHRVData represents comprehensive daily HRV data
type DailyHRVData struct {
	UserProfilePK          int          `json:"userProfilePk"`
	CalendarDate           time.Time    `json:"calendarDate"`
	WeeklyAvg              *float64     `json:"weeklyAvg"`
	LastNightAvg           *float64     `json:"lastNightAvg"`
	LastNight5MinHigh      *float64     `json:"lastNight5MinHigh"`
	Baseline               HRVBaseline  `json:"baseline"`
	Status                 string       `json:"status"`
	FeedbackPhrase         string       `json:"feedbackPhrase"`
	CreateTimeStamp        time.Time    `json:"createTimeStamp"`
	HRVReadings            []HRVReading `json:"hrvReadings"`
	StartTimestampGMT      time.Time    `json:"startTimestampGmt"`
	EndTimestampGMT        time.Time    `json:"endTimestampGmt"`
	StartTimestampLocal    time.Time    `json:"startTimestampLocal"`
	EndTimestampLocal      time.Time    `json:"endTimestampLocal"`
	SleepStartTimestampGMT time.Time    `json:"sleepStartTimestampGmt"`
	SleepEndTimestampGMT   time.Time    `json:"sleepEndTimestampGmt"`
}

// BodyBatteryEvent represents events that impact Body Battery
type BodyBatteryEvent struct {
	EventType              string    `json:"eventType"` // "sleep", "activity", "stress"
	EventStartTimeGMT      time.Time `json:"eventStartTimeGmt"`
	TimezoneOffset         int       `json:"timezoneOffset"`
	DurationInMilliseconds int       `json:"durationInMilliseconds"`
	BodyBatteryImpact      int       `json:"bodyBatteryImpact"`
	FeedbackType           string    `json:"feedbackType"`
	ShortFeedback          string    `json:"shortFeedback"`
}

// DetailedBodyBatteryData represents comprehensive Body Battery data
type DetailedBodyBatteryData struct {
	UserProfilePK          int                `json:"userProfilePk"`
	CalendarDate           time.Time          `json:"calendarDate"`
	StartTimestampGMT      time.Time          `json:"startTimestampGmt"`
	EndTimestampGMT        time.Time          `json:"endTimestampGmt"`
	StartTimestampLocal    time.Time          `json:"startTimestampLocal"`
	EndTimestampLocal      time.Time          `json:"endTimestampLocal"`
	MaxStressLevel         int                `json:"maxStressLevel"`
	AvgStressLevel         int                `json:"avgStressLevel"`
	BodyBatteryValuesArray [][]interface{}    `json:"bodyBatteryValuesArray"`
	StressValuesArray      [][]int            `json:"stressValuesArray"`
	Events                 []BodyBatteryEvent `json:"bodyBatteryEvents"`
}

// TrainingStatus represents current training status
type TrainingStatus struct {
	CalendarDate          time.Time `json:"calendarDate"`
	TrainingStatusKey     string    `json:"trainingStatusKey"` // "DETRAINING", "RECOVERY", "MAINTAINING", "PRODUCTIVE", "PEAKING", "OVERREACHING", "UNPRODUCTIVE", "NONE"
	TrainingStatusTypeKey string    `json:"trainingStatusTypeKey"`
	TrainingStatusValue   int       `json:"trainingStatusValue"`
	LoadRatio             float64   `json:"loadRatio"`
}

// TrainingLoad represents training load data
type TrainingLoad struct {
	CalendarDate            time.Time `json:"calendarDate"`
	AcuteTrainingLoad       float64   `json:"acuteTrainingLoad"`
	ChronicTrainingLoad     float64   `json:"chronicTrainingLoad"`
	TrainingLoadRatio       float64   `json:"trainingLoadRatio"`
	TrainingEffectAerobic   float64   `json:"trainingEffectAerobic"`
	TrainingEffectAnaerobic float64   `json:"trainingEffectAnaerobic"`
}

// FitnessAge represents fitness age calculation
type FitnessAge struct {
	FitnessAge       int       `json:"fitnessAge"`
	ChronologicalAge int       `json:"chronologicalAge"`
	VO2MaxRunning    float64   `json:"vo2MaxRunning"`
	LastUpdated      time.Time `json:"lastUpdated"`
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
	Date       time.Time `json:"calendarDate"`
	RestingHR  *int      `json:"resting_hr"`
	Weight     *float64  `json:"weight"`
	BodyFat    *float64  `json:"body_fat"`
	BMI        *float64  `json:"bmi"`
	BodyWater  *float64  `json:"body_water"`
	BoneMass   *float64  `json:"bone_mass"`
	MuscleMass *float64  `json:"muscle_mass"`
	// Add more fields as needed
}

// SleepData represents sleep summary data
type SleepData struct {
	Date              time.Time `json:"calendarDate"`
	SleepScore        int       `json:"sleepScore"`
	TotalSleepSeconds int       `json:"totalSleepSeconds"`
	DeepSleepSeconds  int       `json:"deepSleepSeconds"`
	LightSleepSeconds int       `json:"lightSleepSeconds"`
	RemSleepSeconds   int       `json:"remSleepSeconds"`
	AwakeSleepSeconds int       `json:"awakeSleepSeconds"`
	// Add more fields as needed
}

// HrvData represents Heart Rate Variability data
type HrvData struct {
	Date     time.Time `json:"calendarDate"`
	HrvValue float64   `json:"hrvValue"`
	// Add more fields as needed
}

// HRVStatus represents HRV status and baseline
type HRVStatus struct {
	Status           string  `json:"status"` // "BALANCED", "UNBALANCED", "POOR"
	FeedbackPhrase   string  `json:"feedbackPhrase"`
	BaselineLowUpper int     `json:"baselineLowUpper"`
	BalancedLow      int     `json:"balancedLow"`
	BalancedUpper    int     `json:"balancedUpper"`
	MarkerValue      float64 `json:"markerValue"`
}

// HRVReading represents an individual HRV reading
type HRVReading struct {
	Timestamp     int     `json:"timestamp"`
	StressLevel   int     `json:"stressLevel"`
	HeartRate     int     `json:"heartRate"`
	RRInterval    int     `json:"rrInterval"`
	Status        string  `json:"status"`
	SignalQuality float64 `json:"signalQuality"`
}

// TimestampAsTime converts the reading timestamp to time.Time using timeutils
func (r *HRVReading) TimestampAsTime() time.Time {
	return ParseTimestamp(r.Timestamp)
}

// RRSeconds converts the RR interval to seconds
func (r *HRVReading) RRSeconds() float64 {
	return float64(r.RRInterval) / 1000.0
}

// StressData represents stress level data
type StressData struct {
	Date            time.Time `json:"calendarDate"`
	StressLevel     int       `json:"stressLevel"`
	RestStressLevel int       `json:"restStressLevel"`
	// Add more fields as needed
}

// BodyBatteryData represents Body Battery data
type BodyBatteryData struct {
	Date         time.Time `json:"calendarDate"`
	BatteryLevel int       `json:"batteryLevel"`
	Charge       int       `json:"charge"`
	Drain        int       `json:"drain"`
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
