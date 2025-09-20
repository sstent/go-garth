# High Priority Endpoints Implementation Guide

## Overview
This guide covers implementing the most commonly requested Garmin Connect API endpoints that are currently missing from your codebase. We'll focus on the high-priority endpoints that provide detailed health and fitness data.

## 1. Detailed Sleep Data Implementation

### Files to Create/Modify

#### A. Create `internal/data/sleep_detailed.go`
```go
package data

import (
	"encoding/json"
	"fmt"
	"time"

	"go-garth/internal/api/client"
	"go-garth/internal/types"
)

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
	Overall           int     `json:"overall"`
	Composition       SleepScoreBreakdown `json:"composition"`
	Revitalization    SleepScoreBreakdown `json:"revitalization"`
	Duration          SleepScoreBreakdown `json:"duration"`
	DeepPercentage    float64 `json:"deepPercentage"`
	LightPercentage   float64 `json:"lightPercentage"`
	RemPercentage     float64 `json:"remPercentage"`
	RestfulnessValue  float64 `json:"restfulnessValue"`
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
	UserProfilePK          int             `json:"userProfilePk"`
	CalendarDate           time.Time       `json:"calendarDate"`
	SleepStartTimestampGMT time.Time       `json:"sleepStartTimestampGmt"`
	SleepEndTimestampGMT   time.Time       `json:"sleepEndTimestampGmt"`
	SleepStartTimestampLocal time.Time     `json:"sleepStartTimestampLocal"`
	SleepEndTimestampLocal   time.Time     `json:"sleepEndTimestampLocal"`
	UnmeasurableSleepSeconds int           `json:"unmeasurableSleepSeconds"`
	DeepSleepSeconds         int           `json:"deepSleepSeconds"`
	LightSleepSeconds        int           `json:"lightSleepSeconds"`
	RemSleepSeconds          int           `json:"remSleepSeconds"`
	AwakeSleepSeconds        int           `json:"awakeSleepSeconds"`
	DeviceRemCapable         bool          `json:"deviceRemCapable"`
	SleepLevels              []SleepLevel  `json:"sleepLevels"`
	SleepMovement            []SleepMovement `json:"sleepMovement"`
	SleepScores              *SleepScore   `json:"sleepScores"`
	AverageSpO2Value         *float64      `json:"averageSpO2Value"`
	LowestSpO2Value          *int          `json:"lowestSpO2Value"`
	HighestSpO2Value         *int          `json:"highestSpO2Value"`
	AverageRespirationValue  *float64      `json:"averageRespirationValue"`
	LowestRespirationValue   *float64      `json:"lowestRespirationValue"`
	HighestRespirationValue  *float64      `json:"highestRespirationValue"`
	AvgSleepStress           *float64      `json:"avgSleepStress"`
	BaseData
}

// NewDetailedSleepData creates a new DetailedSleepData instance
func NewDetailedSleepData() *DetailedSleepData {
	sleep := &DetailedSleepData{}
	sleep.GetFunc = sleep.get
	return sleep
}

func (d *DetailedSleepData) get(day time.Time, client *client.Client) (interface{}, error) {
	dateStr := day.Format("2006-01-02")
	path := fmt.Sprintf("/wellness-service/wellness/dailySleepData/%s?date=%s&nonSleepBufferMinutes=60",
		client.Username, dateStr)

	data, err := client.ConnectAPI(path, "GET", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get detailed sleep data: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var response struct {
		DailySleepDTO            *DetailedSleepData `json:"dailySleepDTO"`
		SleepMovement            []SleepMovement    `json:"sleepMovement"`
		RemSleepData            bool               `json:"remSleepData"`
		SleepLevels             []SleepLevel       `json:"sleepLevels"`
		SleepRestlessMoments    []interface{}      `json:"sleepRestlessMoments"`
		RestlessMomentsCount    int                `json:"restlessMomentsCount"`
		WellnessSpO2SleepSummaryDTO interface{}   `json:"wellnessSpO2SleepSummaryDTO"`
		WellnessEpochSPO2DataDTOList []interface{} `json:"wellnessEpochSPO2DataDTOList"`
		WellnessEpochRespirationDataDTOList []interface{} `json:"wellnessEpochRespirationDataDTOList"`
		SleepStress             interface{}        `json:"sleepStress"`
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

	return response.DailySleepDTO, nil
}

// GetSleepEfficiency calculates sleep efficiency percentage
func (d *DetailedSleepData) GetSleepEfficiency() float64 {
	totalTime := d.SleepEndTimestampGMT.Sub(d.SleepStartTimestampGMT).Seconds()
	sleepTime := float64(d.DeepSleepSeconds + d.LightSleepSeconds + d.RemSleepSeconds)
	if totalTime == 0 {
		return 0
	}
	return (sleepTime / totalTime) * 100
}

// GetTotalSleepTime returns total sleep time in hours
func (d *DetailedSleepData) GetTotalSleepTime() float64 {
	totalSeconds := d.DeepSleepSeconds + d.LightSleepSeconds + d.RemSleepSeconds
	return float64(totalSeconds) / 3600.0
}
```

#### B. Add methods to `internal/api/client/client.go`
```go
// GetDetailedSleepData retrieves comprehensive sleep data for a date
func (c *Client) GetDetailedSleepData(date time.Time) (*types.DetailedSleepData, error) {
	sleepData := data.NewDetailedSleepData()
	result, err := sleepData.Get(date, c)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	detailedSleep, ok := result.(*types.DetailedSleepData)
	if !ok {
		return nil, fmt.Errorf("unexpected sleep data type")
	}

	return detailedSleep, nil
}
```

## 2. Heart Rate Variability (HRV) Implementation

#### A. Update `internal/data/hrv.go` (extend existing)
Add these methods to your existing HRV implementation:

```go
// HRVStatus represents HRV status and baseline
type HRVStatus struct {
	Status           string  `json:"status"` // "BALANCED", "UNBALANCED", "POOR"
	FeedbackPhrase   string  `json:"feedbackPhrase"`
	BaselineLowUpper int     `json:"baselineLowUpper"`
	BalancedLow      int     `json:"balancedLow"`
	BalancedUpper    int     `json:"balancedUpper"`
	MarkerValue      float64 `json:"markerValue"`
}

// DailyHRVData represents comprehensive daily HRV data
type DailyHRVData struct {
	UserProfilePK           int          `json:"userProfilePk"`
	CalendarDate            time.Time    `json:"calendarDate"`
	WeeklyAvg               *float64     `json:"weeklyAvg"`
	LastNightAvg            *float64     `json:"lastNightAvg"`
	LastNight5MinHigh       *float64     `json:"lastNight5MinHigh"`
	Baseline                HRVBaseline  `json:"baseline"`
	Status                  string       `json:"status"`
	FeedbackPhrase          string       `json:"feedbackPhrase"`
	CreateTimeStamp         time.Time    `json:"createTimeStamp"`
	HRVReadings             []HRVReading `json:"hrvReadings"`
	StartTimestampGMT       time.Time    `json:"startTimestampGmt"`
	EndTimestampGMT         time.Time    `json:"endTimestampGmt"`
	StartTimestampLocal     time.Time    `json:"startTimestampLocal"`
	EndTimestampLocal       time.Time    `json:"endTimestampLocal"`
	SleepStartTimestampGMT  time.Time    `json:"sleepStartTimestampGmt"`
	SleepEndTimestampGMT    time.Time    `json:"sleepEndTimestampGmt"`
	BaseData
}

type HRVBaseline struct {
	LowUpper      int     `json:"lowUpper"`
	BalancedLow   int     `json:"balancedLow"`
	BalancedUpper int     `json:"balancedUpper"`
	MarkerValue   float64 `json:"markerValue"`
}

// Update the existing get method in hrv.go
func (h *DailyHRVData) get(day time.Time, client *client.Client) (interface{}, error) {
	dateStr := day.Format("2006-01-02")
	path := fmt.Sprintf("/wellness-service/wellness/dailyHrvData/%s?date=%s",
		client.Username, dateStr)

	data, err := client.ConnectAPI(path, "GET", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get HRV data: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var response struct {
		HRVSummary  DailyHRVData `json:"hrvSummary"`
		HRVReadings []HRVReading `json:"hrvReadings"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse HRV response: %w", err)
	}

	// Combine summary and readings
	response.HRVSummary.HRVReadings = response.HRVReadings
	return &response.HRVSummary, nil
}
```

## 3. Body Battery Detailed Implementation

#### A. Update `internal/data/body_battery.go`
Add these structures and methods:

```go
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
	UserProfilePK          int                    `json:"userProfilePk"`
	CalendarDate           time.Time              `json:"calendarDate"`
	StartTimestampGMT      time.Time              `json:"startTimestampGmt"`
	EndTimestampGMT        time.Time              `json:"endTimestampGmt"`
	StartTimestampLocal    time.Time              `json:"startTimestampLocal"`
	EndTimestampLocal      time.Time              `json:"endTimestampLocal"`
	MaxStressLevel         int                    `json:"maxStressLevel"`
	AvgStressLevel         int                    `json:"avgStressLevel"`
	BodyBatteryValuesArray [][]interface{}        `json:"bodyBatteryValuesArray"`
	StressValuesArray      [][]int                `json:"stressValuesArray"`
	Events                 []BodyBatteryEvent     `json:"bodyBatteryEvents"`
	BaseData
}

func NewDetailedBodyBatteryData() *DetailedBodyBatteryData {
	bb := &DetailedBodyBatteryData{}
	bb.GetFunc = bb.get
	return bb
}

func (d *DetailedBodyBatteryData) get(day time.Time, client *client.Client) (interface{}, error) {
	dateStr := day.Format("2006-01-02")
	
	// Get main Body Battery data
	path1 := fmt.Sprintf("/wellness-service/wellness/dailyStress/%s", dateStr)
	data1, err := client.ConnectAPI(path1, "GET", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Body Battery stress data: %w", err)
	}

	// Get Body Battery events
	path2 := fmt.Sprintf("/wellness-service/wellness/bodyBattery/%s", dateStr)
	data2, err := client.ConnectAPI(path2, "GET", nil, nil)
	if err != nil {
		// Events might not be available, continue without them
		data2 = []byte("[]")
	}

	var result DetailedBodyBatteryData
	if len(data1) > 0 {
		if err := json.Unmarshal(data1, &result); err != nil {
			return nil, fmt.Errorf("failed to parse Body Battery data: %w", err)
		}
	}

	var events []BodyBatteryEvent
	if len(data2) > 0 {
		if err := json.Unmarshal(data2, &events); err == nil {
			result.Events = events
		}
	}

	return &result, nil
}

// GetCurrentLevel returns the most recent Body Battery level
func (d *DetailedBodyBatteryData) GetCurrentLevel() int {
	if len(d.BodyBatteryValuesArray) == 0 {
		return 0
	}
	
	readings := ParseBodyBatteryReadings(d.BodyBatteryValuesArray)
	if len(readings) == 0 {
		return 0
	}
	
	return readings[len(readings)-1].Level
}

// GetDayChange returns the Body Battery change for the day
func (d *DetailedBodyBatteryData) GetDayChange() int {
	readings := ParseBodyBatteryReadings(d.BodyBatteryValuesArray)
	if len(readings) < 2 {
		return 0
	}
	
	return readings[len(readings)-1].Level - readings[0].Level
}
```

## 4. Training Status & Load Implementation

#### A. Create `internal/data/training.go`
```go
package data

import (
	"encoding/json"
	"fmt"
	"time"

	"go-garth/internal/api/client"
)

// TrainingStatus represents current training status
type TrainingStatus struct {
	CalendarDate          time.Time `json:"calendarDate"`
	TrainingStatusKey     string    `json:"trainingStatusKey"` // "DETRAINING", "RECOVERY", "MAINTAINING", "PRODUCTIVE", "PEAKING", "OVERREACHING", "UNPRODUCTIVE", "NONE"
	TrainingStatusTypeKey string    `json:"trainingStatusTypeKey"`
	TrainingStatusValue   int       `json:"trainingStatusValue"`
	LoadRatio             float64   `json:"loadRatio"`
	BaseData
}

// TrainingLoad represents training load data
type TrainingLoad struct {
	CalendarDate              time.Time `json:"calendarDate"`
	AcuteTrainingLoad         float64   `json:"acuteTrainingLoad"`
	ChronicTrainingLoad       float64   `json:"chronicTrainingLoad"`
	TrainingLoadRatio         float64   `json:"trainingLoadRatio"`
	TrainingEffectAerobic     float64   `json:"trainingEffectAerobic"`
	TrainingEffectAnaerobic   float64   `json:"trainingEffectAnaerobic"`
	BaseData
}

// FitnessAge represents fitness age calculation
type FitnessAge struct {
	FitnessAge       int       `json:"fitnessAge"`
	ChronologicalAge int       `json:"chronologicalAge"`
	VO2MaxRunning    float64   `json:"vo2MaxRunning"`
	LastUpdated      time.Time `json:"lastUpdated"`
}

func NewTrainingStatus() *TrainingStatus {
	ts := &TrainingStatus{}
	ts.GetFunc = ts.get
	return ts
}

func (t *TrainingStatus) get(day time.Time, client *client.Client) (interface{}, error) {
	dateStr := day.Format("2006-01-02")
	path := fmt.Sprintf("/metrics-service/metrics/trainingStatus/%s", dateStr)

	data, err := client.ConnectAPI(path, "GET", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get training status: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var result TrainingStatus
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse training status: %w", err)
	}

	return &result, nil
}

func NewTrainingLoad() *TrainingLoad {
	tl := &TrainingLoad{}
	tl.GetFunc = tl.get
	return tl
}

func (t *TrainingLoad) get(day time.Time, client *client.Client) (interface{}, error) {
	dateStr := day.Format("2006-01-02")
	endDate := day.AddDate(0, 0, 6).Format("2006-01-02") // Get week of data
	path := fmt.Sprintf("/metrics-service/metrics/trainingLoad/%s/%s", dateStr, endDate)

	data, err := client.ConnectAPI(path, "GET", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get training load: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var results []TrainingLoad
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("failed to parse training load: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0], nil
}
```

## 5. Client Methods Integration

#### Add these methods to `internal/api/client/client.go`:

```go
// GetTrainingStatus retrieves current training status
func (c *Client) GetTrainingStatus(date time.Time) (*types.TrainingStatus, error) {
	trainingStatus := data.NewTrainingStatus()
	result, err := trainingStatus.Get(date, c)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	status, ok := result.(*types.TrainingStatus)
	if !ok {
		return nil, fmt.Errorf("unexpected training status type")
	}

	return status, nil
}

// GetTrainingLoad retrieves training load data
func (c *Client) GetTrainingLoad(date time.Time) (*types.TrainingLoad, error) {
	trainingLoad := data.NewTrainingLoad()
	result, err := trainingLoad.Get(date, c)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil
	}

	load, ok := result.(*types.TrainingLoad)
	if !ok {
		return nil, fmt.Errorf("unexpected training load type")
	}

	return load, nil
}

// GetFitnessAge retrieves fitness age calculation
func (c *Client) GetFitnessAge() (*types.FitnessAge, error) {
	path := "/fitness-service/fitness/fitnessAge"
	
	data, err := c.ConnectAPI(path, "GET", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get fitness age: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var fitnessAge types.FitnessAge
	if err := json.Unmarshal(data, &fitnessAge); err != nil {
		return nil, fmt.Errorf("failed to parse fitness age: %w", err)
	}

	fitnessAge.LastUpdated = time.Now()
	return &fitnessAge, nil
}
```

## Implementation Steps

### Phase 1: Sleep Data (Week 1)
1. Create `internal/data/sleep_detailed.go`
2. Update `internal/types/garmin.go` with sleep types
3. Add client methods
4. Create tests
5. Test with real data

### Phase 2: HRV Enhancement (Week 2)
1. Update existing `internal/data/hrv.go`
2. Add new HRV types to types file
3. Enhance client methods
4. Create comprehensive tests

### Phase 3: Body Battery Details (Week 3)
1. Update `internal/data/body_battery.go`
2. Add event tracking
3. Add convenience methods
4. Create tests

### Phase 4: Training Metrics (Week 4)
1. Create `internal/data/training.go`
2. Add training types
3. Implement client methods
4. Create tests and validation

## Testing Strategy

Create test files for each new data type:

```go
// Example test structure
func TestDetailedSleepData_Get(t *testing.T) {
    // Mock response from API
    mockResponse := `{
        "dailySleepDTO": {
            "userProfilePk": 12345,
            "calendarDate": "2023-06-15",
            "deepSleepSeconds": 7200,
            "lightSleepSeconds": 14400,
            "remSleepSeconds": 3600,
            "awakeSleepSeconds": 1800
        },
        "sleepMovement": [],
        "sleepLevels": []
    }`
    
    // Create mock client
    server := testutils.MockJSONResponse(http.StatusOK, mockResponse)
    defer server.Close()
    
    // Test implementation
    // ... test logic
}
```

## Error Handling Patterns

For each endpoint, implement consistent error handling:

```go
func (d *DataType) get(day time.Time, client *client.Client) (interface{}, error) {
    data, err := client.ConnectAPI(path, "GET", nil, nil)
    if err != nil {
        // Log the error but don't fail completely
        fmt.Printf("Warning: Failed to get %s data: %v\n", "datatype", err)
        return nil, nil // Return nil data, not error for missing data
    }
    
    if len(data) == 0 {
        return nil, nil // No data available
    }
    
    // Parse and validate
    var result DataType
    if err := json.Unmarshal(data, &result); err != nil {
        return nil, fmt.Errorf("failed to parse %s data: %w", "datatype", err)
    }
    
    return &result, nil
}
```

## Usage Examples

After implementation, users can access the data like this:

```go
// Get detailed sleep data
sleepData, err := client.GetDetailedSleepData(time.Now().AddDate(0, 0, -1))
if err != nil {
    log.Fatal(err)
}
if sleepData != nil {
    fmt.Printf("Sleep efficiency: %.1f%%\n", sleepData.GetSleepEfficiency())
    fmt.Printf("Total sleep: %.1f hours\n", sleepData.GetTotalSleepTime())
}

// Get training status
status, err := client.GetTrainingStatus(time.Now())
if err != nil {
    log.Fatal(err)
}
if status != nil {
    fmt.Printf("Training Status: %s\n", status.TrainingStatusKey)
    fmt.Printf("Load Ratio: %.2f\n", status.LoadRatio)
}
```

This implementation guide provides a comprehensive foundation for adding the most requested Garmin Connect API endpoints to your Go client.
