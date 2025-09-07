# Garmin Connect API Endpoints and Go Structs

This document provides a comprehensive overview of all Garmin Connect API endpoints accessible through the Garth library, along with corresponding Go structs for JSON response handling.

## Base URLs
- **Connect API**: `https://connectapi.{domain}`
- **SSO**: `https://sso.{domain}`
- **Domain**: `garmin.com` (or `garmin.cn` for China)

## Authentication Endpoints

### OAuth1 Token Request
- **Endpoint**: `GET /oauth-service/oauth/preauthorized`
- **Query Parameters**: `ticket`, `login-url`, `accepts-mfa-tokens=true`
- **Purpose**: Get OAuth1 token after SSO login

```go
type OAuth1TokenResponse struct {
    OAuthToken       string `json:"oauth_token"`
    OAuthTokenSecret string `json:"oauth_token_secret"`
    MFAToken         string `json:"mfa_token,omitempty"`
}
```

### OAuth2 Token Exchange
- **Endpoint**: `POST /oauth-service/oauth/exchange/user/2.0`
- **Purpose**: Exchange OAuth1 token for OAuth2 access token

```go
type OAuth2TokenResponse struct {
    Scope                    string `json:"scope"`
    JTI                      string `json:"jti"`
    TokenType                string `json:"token_type"`
    AccessToken              string `json:"access_token"`
    RefreshToken             string `json:"refresh_token"`
    ExpiresIn                int    `json:"expires_in"`
    ExpiresAt                int    `json:"expires_at"`
    RefreshTokenExpiresIn    int    `json:"refresh_token_expires_in"`
    RefreshTokenExpiresAt    int    `json:"refresh_token_expires_at"`
}
```

## User Profile Endpoints

### Social Profile
- **Endpoint**: `GET /userprofile-service/socialProfile`
- **Purpose**: Get user's social profile information

```go
type UserProfile struct {
    ID                              int      `json:"id"`
    ProfileID                       int      `json:"profileId"`
    GarminGUID                      string   `json:"garminGuid"`
    DisplayName                     string   `json:"displayName"`
    FullName                        string   `json:"fullName"`
    UserName                        string   `json:"userName"`
    ProfileImageType                *string  `json:"profileImageType"`
    ProfileImageURLLarge            *string  `json:"profileImageUrlLarge"`
    ProfileImageURLMedium           *string  `json:"profileImageUrlMedium"`
    ProfileImageURLSmall            *string  `json:"profileImageUrlSmall"`
    Location                        *string  `json:"location"`
    FacebookURL                     *string  `json:"facebookUrl"`
    TwitterURL                      *string  `json:"twitterUrl"`
    PersonalWebsite                 *string  `json:"personalWebsite"`
    Motivation                      *string  `json:"motivation"`
    Bio                            *string  `json:"bio"`
    PrimaryActivity                 *string  `json:"primaryActivity"`
    FavoriteActivityTypes           []string `json:"favoriteActivityTypes"`
    RunningTrainingSpeed            float64  `json:"runningTrainingSpeed"`
    CyclingTrainingSpeed            float64  `json:"cyclingTrainingSpeed"`
    FavoriteCyclingActivityTypes    []string `json:"favoriteCyclingActivityTypes"`
    CyclingClassification           *string  `json:"cyclingClassification"`
    CyclingMaxAvgPower              float64  `json:"cyclingMaxAvgPower"`
    SwimmingTrainingSpeed           float64  `json:"swimmingTrainingSpeed"`
    ProfileVisibility               string   `json:"profileVisibility"`
    ActivityStartVisibility         string   `json:"activityStartVisibility"`
    ActivityMapVisibility           string   `json:"activityMapVisibility"`
    CourseVisibility                string   `json:"courseVisibility"`
    ActivityHeartRateVisibility     string   `json:"activityHeartRateVisibility"`
    ActivityPowerVisibility         string   `json:"activityPowerVisibility"`
    BadgeVisibility                 string   `json:"badgeVisibility"`
    ShowAge                         bool     `json:"showAge"`
    ShowWeight                      bool     `json:"showWeight"`
    ShowHeight                      bool     `json:"showHeight"`
    ShowWeightClass                 bool     `json:"showWeightClass"`
    ShowAgeRange                    bool     `json:"showAgeRange"`
    ShowGender                      bool     `json:"showGender"`
    ShowActivityClass               bool     `json:"showActivityClass"`
    ShowVO2Max                      bool     `json:"showVo2Max"`
    ShowPersonalRecords             bool     `json:"showPersonalRecords"`
    ShowLast12Months                bool     `json:"showLast12Months"`
    ShowLifetimeTotals              bool     `json:"showLifetimeTotals"`
    ShowUpcomingEvents              bool     `json:"showUpcomingEvents"`
    ShowRecentFavorites             bool     `json:"showRecentFavorites"`
    ShowRecentDevice                bool     `json:"showRecentDevice"`
    ShowRecentGear                  bool     `json:"showRecentGear"`
    ShowBadges                      bool     `json:"showBadges"`
    OtherActivity                   *string  `json:"otherActivity"`
    OtherPrimaryActivity            *string  `json:"otherPrimaryActivity"`
    OtherMotivation                 *string  `json:"otherMotivation"`
    UserRoles                       []string `json:"userRoles"`
    NameApproved                    bool     `json:"nameApproved"`
    UserProfileFullName             string   `json:"userProfileFullName"`
    MakeGolfScorecardsPrivate       bool     `json:"makeGolfScorecardsPrivate"`
    AllowGolfLiveScoring            bool     `json:"allowGolfLiveScoring"`
    AllowGolfScoringByConnections   bool     `json:"allowGolfScoringByConnections"`
    UserLevel                       int      `json:"userLevel"`
    UserPoint                       int      `json:"userPoint"`
    LevelUpdateDate                 string   `json:"levelUpdateDate"`
    LevelIsViewed                   bool     `json:"levelIsViewed"`
    LevelPointThreshold             int      `json:"levelPointThreshold"`
    UserPointOffset                 int      `json:"userPointOffset"`
    UserPro                         bool     `json:"userPro"`
}
```

### User Settings
- **Endpoint**: `GET /userprofile-service/userprofile/user-settings`
- **Purpose**: Get user's account and device settings

```go
type PowerFormat struct {
    FormatID      int     `json:"formatId"`
    FormatKey     string  `json:"formatKey"`
    MinFraction   int     `json:"minFraction"`
    MaxFraction   int     `json:"maxFraction"`
    GroupingUsed  bool    `json:"groupingUsed"`
    DisplayFormat *string `json:"displayFormat"`
}

type FirstDayOfWeek struct {
    DayID               int  `json:"dayId"`
    DayName             string `json:"dayName"`
    SortOrder           int  `json:"sortOrder"`
    IsPossibleFirstDay  bool `json:"isPossibleFirstDay"`
}

type WeatherLocation struct {
    UseFixedLocation *bool   `json:"useFixedLocation"`
    Latitude         *float64 `json:"latitude"`
    Longitude        *float64 `json:"longitude"`
    LocationName     *string `json:"locationName"`
    ISOCountryCode   *string `json:"isoCountryCode"`
    PostalCode       *string `json:"postalCode"`
}

type HydrationContainer struct {
    Volume *float64 `json:"volume"`
    Name   *string  `json:"name"`
    Type   *string  `json:"type"`
}

type UserData struct {
    Gender                              string               `json:"gender"`
    Weight                              float64              `json:"weight"`
    Height                              float64              `json:"height"`
    TimeFormat                          string               `json:"timeFormat"`
    BirthDate                           string               `json:"birthDate"`
    MeasurementSystem                   string               `json:"measurementSystem"`
    ActivityLevel                       *string              `json:"activityLevel"`
    Handedness                          string               `json:"handedness"`
    PowerFormat                         PowerFormat          `json:"powerFormat"`
    HeartRateFormat                     PowerFormat          `json:"heartRateFormat"`
    FirstDayOfWeek                      FirstDayOfWeek       `json:"firstDayOfWeek"`
    VO2MaxRunning                       *float64             `json:"vo2MaxRunning"`
    VO2MaxCycling                       *float64             `json:"vo2MaxCycling"`
    LactateThresholdSpeed               *float64             `json:"lactateThresholdSpeed"`
    LactateThresholdHeartRate           *float64             `json:"lactateThresholdHeartRate"`
    DiveNumber                          *int                 `json:"diveNumber"`
    IntensityMinutesCalcMethod          string               `json:"intensityMinutesCalcMethod"`
    ModerateIntensityMinutesHRZone      int                  `json:"moderateIntensityMinutesHrZone"`
    VigorousIntensityMinutesHRZone      int                  `json:"vigorousIntensityMinutesHrZone"`
    HydrationMeasurementUnit            string               `json:"hydrationMeasurementUnit"`
    HydrationContainers                 []HydrationContainer `json:"hydrationContainers"`
    HydrationAutoGoalEnabled            bool                 `json:"hydrationAutoGoalEnabled"`
    FirstbeatMaxStressScore             *float64             `json:"firstbeatMaxStressScore"`
    FirstbeatCyclingLTTimestamp         *int                 `json:"firstbeatCyclingLtTimestamp"`
    FirstbeatRunningLTTimestamp         *int                 `json:"firstbeatRunningLtTimestamp"`
    ThresholdHeartRateAutoDetected      bool                 `json:"thresholdHeartRateAutoDetected"`
    FTPAutoDetected                     *bool                `json:"ftpAutoDetected"`
    TrainingStatusPausedDate            *string              `json:"trainingStatusPausedDate"`
    WeatherLocation                     *WeatherLocation     `json:"weatherLocation"`
    GolfDistanceUnit                    *string              `json:"golfDistanceUnit"`
    GolfElevationUnit                   *string              `json:"golfElevationUnit"`
    GolfSpeedUnit                       *string              `json:"golfSpeedUnit"`
    ExternalBottomTime                  *float64             `json:"externalBottomTime"`
}

type UserSleep struct {
    SleepTime        int  `json:"sleepTime"`
    DefaultSleepTime bool `json:"defaultSleepTime"`
    WakeTime         int  `json:"wakeTime"`
    DefaultWakeTime  bool `json:"defaultWakeTime"`
}

type UserSleepWindow struct {
    SleepWindowFrequency                   string `json:"sleepWindowFrequency"`
    StartSleepTimeSecondsFromMidnight      int    `json:"startSleepTimeSecondsFromMidnight"`
    EndSleepTimeSecondsFromMidnight        int    `json:"endSleepTimeSecondsFromMidnight"`
}

type UserSettings struct {
    ID                int                `json:"id"`
    UserData          UserData           `json:"userData"`
    UserSleep         UserSleep          `json:"userSleep"`
    ConnectDate       *string            `json:"connectDate"`
    SourceType        *string            `json:"sourceType"`
    UserSleepWindows  []UserSleepWindow  `json:"userSleepWindows,omitempty"`
}
```

## Wellness & Health Data Endpoints

### Daily Sleep Data
- **Endpoint**: `GET /wellness-service/wellness/dailySleepData/{username}`
- **Query Parameters**: `date`, `nonSleepBufferMinutes`
- **Purpose**: Get detailed sleep data for a specific date

```go
type Score struct {
    QualifierKey         string   `json:"qualifierKey"`
    OptimalStart         *float64 `json:"optimalStart"`
    OptimalEnd           *float64 `json:"optimalEnd"`
    Value                *int     `json:"value"`
    IdealStartInSeconds  *float64 `json:"idealStartInSeconds"`
    IdealEndInSeconds    *float64 `json:"idealEndInSeconds"`
}

type SleepScores struct {
    TotalDuration    Score `json:"totalDuration"`
    Stress           Score `json:"stress"`
    AwakeCount       Score `json:"awakeCount"`
    Overall          Score `json:"overall"`
    REMPercentage    Score `json:"remPercentage"`
    Restlessness     Score `json:"restlessness"`
    LightPercentage  Score `json:"lightPercentage"`
    DeepPercentage   Score `json:"deepPercentage"`
}

type DailySleepDTO struct {
    ID                               int          `json:"id"`
    UserProfilePK                    int          `json:"userProfilePk"`
    CalendarDate                     string       `json:"calendarDate"`
    SleepTimeSeconds                 int          `json:"sleepTimeSeconds"`
    NapTimeSeconds                   int          `json:"napTimeSeconds"`
    SleepWindowConfirmed             bool         `json:"sleepWindowConfirmed"`
    SleepWindowConfirmationType      string       `json:"sleepWindowConfirmationType"`
    SleepStartTimestampGMT           int64        `json:"sleepStartTimestampGmt"`
    SleepEndTimestampGMT             int64        `json:"sleepEndTimestampGmt"`
    SleepStartTimestampLocal         int64        `json:"sleepStartTimestampLocal"`
    SleepEndTimestampLocal           int64        `json:"sleepEndTimestampLocal"`
    DeviceREMCapable                 bool         `json:"deviceRemCapable"`
    Retro                            bool         `json:"retro"`
    UnmeasurableSleepSeconds         *int         `json:"unmeasurableSleepSeconds"`
    DeepSleepSeconds                 *int         `json:"deepSleepSeconds"`
    LightSleepSeconds                *int         `json:"lightSleepSeconds"`
    REMSleepSeconds                  *int         `json:"remSleepSeconds"`
    AwakeSleepSeconds                *int         `json:"awakeSleepSeconds"`
    SleepFromDevice                  *bool        `json:"sleepFromDevice"`
    SleepVersion                     *int         `json:"sleepVersion"`
    AwakeCount                       *int         `json:"awakeCount"`
    SleepScores                      *SleepScores `json:"sleepScores"`
    AutoSleepStartTimestampGMT       *int64       `json:"autoSleepStartTimestampGmt"`
    AutoSleepEndTimestampGMT         *int64       `json:"autoSleepEndTimestampGmt"`
    SleepQualityTypePK               *int         `json:"sleepQualityTypePk"`
    SleepResultTypePK                *int         `json:"sleepResultTypePk"`
    AverageSPO2Value                 *float64     `json:"averageSpO2Value"`
    LowestSPO2Value                  *int         `json:"lowestSpO2Value"`
    HighestSPO2Value                 *int         `json:"highestSpO2Value"`
    AverageSPO2HRSleep               *float64     `json:"averageSpO2HrSleep"`
    AverageRespirationValue          *float64     `json:"averageRespirationValue"`
    LowestRespirationValue           *float64     `json:"lowestRespirationValue"`
    HighestRespirationValue          *float64     `json:"highestRespirationValue"`
    AvgSleepStress                   *float64     `json:"avgSleepStress"`
    AgeGroup                         *string      `json:"ageGroup"`
    SleepScoreFeedback               *string      `json:"sleepScoreFeedback"`
    SleepScoreInsight                *string      `json:"sleepScoreInsight"`
}

type SleepMovement struct {
    StartGMT      string  `json:"startGmt"`
    EndGMT        string  `json:"endGmt"`
    ActivityLevel float64 `json:"activityLevel"`
}

type SleepData struct {
    DailySleepDTO *DailySleepDTO   `json:"dailySleepDto"`
    SleepMovement []SleepMovement  `json:"sleepMovement"`
    REMSleepData  interface{}      `json:"remSleepData"`
    SleepLevels   interface{}      `json:"sleepLevels"`
    SleepRestlessMoments interface{} `json:"sleepRestlessMoments"`
    RestlessMomentsCount interface{} `json:"restlessMomentsCount"`
    WellnessSpO2SleepSummaryDTO interface{} `json:"wellnessSpO2SleepSummaryDTO"`
    WellnessEpochSPO2DataDTOList interface{} `json:"wellnessEpochSPO2DataDTOList"`
    WellnessEpochRespirationDataDTOList interface{} `json:"wellnessEpochRespirationDataDTOList"`
    SleepStress interface{} `json:"sleepStress"`
}
```

### Daily Stress Data
- **Endpoint**: `GET /wellness-service/wellness/dailyStress/{date}`
- **Purpose**: Get Body Battery and stress data for a specific date

```go
type DailyBodyBatteryStress struct {
    UserProfilePK              int           `json:"userProfilePk"`
    CalendarDate               string        `json:"calendarDate"`
    StartTimestampGMT          string        `json:"startTimestampGmt"`
    EndTimestampGMT            string        `json:"endTimestampGmt"`
    StartTimestampLocal        string        `json:"startTimestampLocal"`
    EndTimestampLocal          string        `json:"endTimestampLocal"`
    MaxStressLevel             int           `json:"maxStressLevel"`
    AvgStressLevel             int           `json:"avgStressLevel"`
    StressChartValueOffset     int           `json:"stressChartValueOffset"`
    StressChartYAxisOrigin     int           `json:"stressChartYAxisOrigin"`
    StressValuesArray          [][]int       `json:"stressValuesArray"`
    BodyBatteryValuesArray     [][]interface{} `json:"bodyBatteryValuesArray"`
}
```

### Body Battery Events
- **Endpoint**: `GET /wellness-service/wellness/bodyBattery/events/{date}`
- **Purpose**: Get Body Battery events (sleep events) for a specific date

```go
type BodyBatteryEvent struct {
    EventType                 string `json:"eventType"`
    EventStartTimeGMT         string `json:"eventStartTimeGmt"`
    TimezoneOffset            int    `json:"timezoneOffset"`
    DurationInMilliseconds    int    `json:"durationInMilliseconds"`
    BodyBatteryImpact         int    `json:"bodyBatteryImpact"`
    FeedbackType              string `json:"feedbackType"`
    ShortFeedback             string `json:"shortFeedback"`
}

type BodyBatteryData struct {
    Event                    *BodyBatteryEvent `json:"event"`
    ActivityName             *string           `json:"activityName"`
    ActivityType             *string           `json:"activityType"`
    ActivityID               *string           `json:"activityId"`
    AverageStress            *float64          `json:"averageStress"`
    StressValuesArray        [][]int           `json:"stressValuesArray"`
    BodyBatteryValuesArray   [][]interface{}   `json:"bodyBatteryValuesArray"`
}
```

### HRV Data
- **Endpoint**: `GET /hrv-service/hrv/{date}`
- **Purpose**: Get detailed HRV data for a specific date

```go
type HRVBaseline struct {
    LowUpper     int     `json:"lowUpper"`
    BalancedLow  int     `json:"balancedLow"`
    BalancedUpper int    `json:"balancedUpper"`
    MarkerValue  float64 `json:"markerValue"`
}

type HRVSummary struct {
    CalendarDate        string       `json:"calendarDate"`
    WeeklyAvg          int          `json:"weeklyAvg"`
    LastNightAvg       *int         `json:"lastNightAvg"`
    LastNight5MinHigh  int          `json:"lastNight5MinHigh"`
    Baseline           HRVBaseline  `json:"baseline"`
    Status             string       `json:"status"`
    FeedbackPhrase     string       `json:"feedbackPhrase"`
    CreateTimeStamp    string       `json:"createTimeStamp"`
}

type HRVReading struct {
    HRVValue          int    `json:"hrvValue"`
    ReadingTimeGMT    string `json:"readingTimeGmt"`
    ReadingTimeLocal  string `json:"readingTimeLocal"`
}

type HRVData struct {
    UserProfilePK              int          `json:"userProfilePk"`
    HRVSummary                 HRVSummary   `json:"hrvSummary"`
    HRVReadings                []HRVReading `json:"hrvReadings"`
    StartTimestampGMT          string       `json:"startTimestampGmt"`
    EndTimestampGMT            string       `json:"endTimestampGmt"`
    StartTimestampLocal        string       `json:"startTimestampLocal"`
    EndTimestampLocal          string       `json:"endTimestampLocal"`
    SleepStartTimestampGMT     string       `json:"sleepStartTimestampGmt"`
    SleepEndTimestampGMT       string       `json:"sleepEndTimestampGmt"`
    SleepStartTimestampLocal   string       `json:"sleepStartTimestampLocal"`
    SleepEndTimestampLocal     string       `json:"sleepEndTimestampLocal"`
}
```

### Weight Data
- **Endpoint**: `GET /weight-service/weight/dayview/{date}` (single day)
- **Endpoint**: `GET /weight-service/weight/range/{start}/{end}?includeAll=true` (date range)
- **Purpose**: Get weight measurements and body composition data

```go
type WeightData struct {
    SamplePK         int64   `json:"samplePk"`
    CalendarDate     string  `json:"calendarDate"`
    Weight           int     `json:"weight"` // in grams
    SourceType       string  `json:"sourceType"`
    WeightDelta      float64 `json:"weightDelta"`
    TimestampGMT     int64   `json:"timestampGmt"`
    Date             int64   `json:"date"`
    BMI              *float64 `json:"bmi"`
    BodyFat          *float64 `json:"bodyFat"`
    BodyWater        *float64 `json:"bodyWater"`
    BoneMass         *int     `json:"boneMass"` // in grams
    MuscleMass       *int     `json:"muscleMass"` // in grams
    PhysiqueRating   *float64 `json:"physiqueRating"`
    VisceralFat      *float64 `json:"visceralFat"`
    MetabolicAge     *int     `json:"metabolicAge"`
}

type WeightResponse struct {
    DateWeightList []WeightData `json:"dateWeightList"`
}

type WeightSummary struct {
    AllWeightMetrics []WeightData `json:"allWeightMetrics"`
}

type WeightRangeResponse struct {
    DailyWeightSummaries []WeightSummary `json:"dailyWeightSummaries"`
}
```

## Stats Endpoints

### Daily Steps
- **Endpoint**: `GET /usersummary-service/stats/steps/daily/{start}/{end}`
- **Purpose**: Get daily step counts and distances

```go
type DailySteps struct {
    CalendarDate   string `json:"calendarDate"`
    TotalSteps     *int   `json:"totalSteps"`
    TotalDistance  *int   `json:"totalDistance"`
    StepGoal       int    `json:"stepGoal"`
}
```

### Weekly Steps
- **Endpoint**: `GET /usersummary-service/stats/steps/weekly/{end}/{period}`
- **Purpose**: Get weekly step summaries

```go
type WeeklySteps struct {
    CalendarDate            string  `json:"calendarDate"`
    TotalSteps              int     `json:"totalSteps"`
    AverageSteps            float64 `json:"averageSteps"`
    AverageDistance         float64 `json:"averageDistance"`
    TotalDistance           float64 `json:"totalDistance"`
    WellnessDataDaysCount   int     `json:"wellnessDataDaysCount"`
}
```

### Daily Stress
- **Endpoint**: `GET /usersummary-service/stats/stress/daily/{start}/{end}`
- **Purpose**: Get daily stress level summaries

```go
type DailyStress struct {
    CalendarDate          string `json:"calendarDate"`
    OverallStressLevel    int    `json:"overallStressLevel"`
    RestStressDuration    *int   `json:"restStressDuration"`
    LowStressDuration     *int   `json:"lowStressDuration"`
    MediumStressDuration  *int   `json:"mediumStressDuration"`
    HighStressDuration    *int   `json:"highStressDuration"`
}
```

### Weekly Stress
- **Endpoint**: `GET /usersummary-service/stats/stress/weekly/{end}/{period}`
- **Purpose**: Get weekly stress level summaries

```go
type WeeklyStress struct {
    CalendarDate string `json:"calendarDate"`
    Value        int    `json:"value"`
}
```

### Daily Intensity Minutes
- **Endpoint**: `GET /usersummary-service/stats/im/daily/{start}/{end}`
- **Purpose**: Get daily intensity minutes

```go
type DailyIntensityMinutes struct {
    CalendarDate   string `json:"calendarDate"`
    WeeklyGoal     int    `json:"weeklyGoal"`
    ModerateValue  *int   `json:"moderateValue"`
    VigorousValue  *int   `json:"vigorousValue"`
}
```

### Weekly Intensity Minutes
- **Endpoint**: `GET /usersummary-service/stats/im/weekly/{start}/{end}`
- **Purpose**: Get weekly intensity minutes

```go
type WeeklyIntensityMinutes struct {
    CalendarDate   string `json:"calendarDate"`
    WeeklyGoal     int    `json:"weeklyGoal"`
    ModerateValue  *int   `json:"moderateValue"`
    VigorousValue  *int   `json:"vigorousValue"`
}
```

### Daily Sleep Score
- **Endpoint**: `GET /wellness-service/stats/daily/sleep/score/{start}/{end}`
- **Purpose**: Get daily sleep quality scores

```go
type DailySleep struct {
    CalendarDate string `json:"calendarDate"`
    Value        *int   `json:"value"`
}
```

### Daily HRV
- **Endpoint**: `GET /hrv-service/hrv/daily/{start}/{end}`
- **Purpose**: Get daily HRV summaries

```go
type DailyHRV struct {
    CalendarDate        string       `json:"calendarDate"`
    WeeklyAvg          *int         `json:"weeklyAvg"`
    LastNightAvg       *int         `json:"lastNightAvg"`
    LastNight5MinHigh  *int         `json:"lastNight5MinHigh"`
    Baseline           *HRVBaseline `json:"baseline"`
    Status             string       `json:"status"`
    FeedbackPhrase     string       `json:"feedbackPhrase"`
    CreateTimeStamp    string       `json:"createTimeStamp"`
}
```

### Daily Hydration
- **Endpoint**: `GET /usersummary-service/stats/hydration/daily/{start}/{end}`
- **Purpose**: Get daily hydration data

```go
type DailyHydration struct {
    CalendarDate string  `json:"calendarDate"`
    ValueInML    float64 `json:"valueInMl"`
    GoalInML     float64 `json:"goalInMl"`
}
```

## File Upload/Download Endpoints

### Upload Activity
- **Endpoint**: `POST /upload-service/upload`
- **Content-Type**: `multipart/form-data`
- **Purpose**: Upload FIT files or other activity data

```go
type UploadResponse struct {
    DetailedImportResult struct {
        UploadID      int64  `json:"uploadId"`
        UploadUUID    struct {
            UUID string `json:"uuid"`
        } `json:"uploadUuid"`
        Owner           int    `json:"owner"`
        FileSize        int    `json:"fileSize"`
        ProcessingTime  int    `json:"processingTime"`
        CreationDate    string `json:"creationDate"`
        IPAddress       *string `json:"ipAddress"`
        FileName        string `json:"fileName"`
        Report          *string `json:"report"`
        Successes       []interface{} `json:"successes"`
        Failures        []interface{} `json:"failures"`
    } `json:"detailedImportResult"`
}
```

### Download Activity
- **Endpoint**: `GET /download-service/files/activity/{activityId}`
- **Purpose**: Download activity data in various formats
- **Returns**: Binary data (FIT, GPX, TCX, etc.)

## SSO Endpoints

### SSO Embed
- **Endpoint**: `GET /sso/embed`
- **Query Parameters**: Various SSO parameters
- **Purpose**: Initialize SSO session

### SSO Sign In
- **Endpoint**: `GET /sso/signin`
- **Endpoint**: `POST /sso/signin`
- **Purpose**: Authenticate user credentials

### MFA Verification
- **Endpoint**: `POST /sso/verifyMFA/loginEnterMfaCode`
- **Purpose**: Verify multi-factor authentication code

## Common Response Patterns

### Error Response
```go
type ErrorResponse struct {
    Message string `json:"message"`
    Code    string `json:"code,omitempty"`
}
```

### Paginated Response Pattern
Many endpoints support pagination with these common patterns:
- Date ranges: `{start}/{end}`
- Period-based: `{end}/{period}`
- Page size limits vary by endpoint (typically 28-52 items)

### Stats Response with Values
Some stats endpoints return data in this nested format:
```go
type StatsResponse struct {
    CalendarDate string                 `json:"calendarDate"`
    Values       map[string]interface{} `json:"values"`
}
```

## Authentication Headers
All API requests require:
- `Authorization: Bearer {oauth2_access_token}`
- `User-Agent: GCM-iOS-5.7.2.1` (or similar)

## Additional Endpoints and Data Types

### Activity Data Endpoints

Based on the codebase structure, there are likely additional activity-related endpoints that follow these patterns:

#### Activity List
- **Endpoint**: `GET /activitylist-service/activities/search/activities`
- **Purpose**: Search and list user activities

```go
type ActivitySummary struct {
    ActivityID          int64   `json:"activityId"`
    ActivityName        string  `json:"activityName"`
    Description         *string `json:"description"`
    StartTimeLocal      string  `json:"startTimeLocal"`
    StartTimeGMT        string  `json:"startTimeGMT"`
    ActivityType        struct {
        TypeID   int    `json:"typeId"`
        TypeKey  string `json:"typeKey"`
        ParentTypeID *int `json:"parentTypeId"`
    } `json:"activityType"`
    EventType           struct {
        TypeID  int    `json:"typeId"`
        TypeKey string `json:"typeKey"`
    } `json:"eventType"`
    Distance            *float64 `json:"distance"`
    Duration            *float64 `json:"duration"`
    ElapsedDuration     *float64 `json:"elapsedDuration"`
    MovingDuration      *float64 `json:"movingDuration"`
    ElevationGain       *float64 `json:"elevationGain"`
    ElevationLoss       *float64 `json:"elevationLoss"`
    AverageSpeed        *float64 `json:"averageSpeed"`
    MaxSpeed            *float64 `json:"maxSpeed"`
    StartLatitude       *float64 `json:"startLatitude"`
    StartLongitude      *float64 `json:"startLongitude"`
    HasPolyline         bool     `json:"hasPolyline"`
    OwnerID             int      `json:"ownerId"`
    Calories            *float64 `json:"calories"`
    BMRCalories         *float64 `json:"bmrCalories"`
    AverageHR           *int     `json:"averageHR"`
    MaxHR               *int     `json:"maxHR"`
    AverageRunCadence   *float64 `json:"averageRunCadence"`
    MaxRunCadence       *float64 `json:"maxRunCadence"`
}

type ActivitySearchResponse struct {
    Activities []ActivitySummary `json:"activities"`
}
```

#### Activity Details
- **Endpoint**: `GET /activity-service/activity/{activityId}`
- **Purpose**: Get detailed information about a specific activity

```go
type ActivityDetails struct {
    ActivityID                    int64    `json:"activityId"`
    ActivityName                  string   `json:"activityName"`
    Description                   *string  `json:"description"`
    StartTimeLocal                string   `json:"startTimeLocal"`
    StartTimeGMT                  string   `json:"startTimeGMT"`
    ActivityType                  struct {
        TypeID         int     `json:"typeId"`
        TypeKey        string  `json:"typeKey"`
        ParentTypeID   *int    `json:"parentTypeId"`
        IsHidden       bool    `json:"isHidden"`
        Restricted     bool    `json:"restricted"`
        TrailRun       bool    `json:"trailRun"`
    } `json:"activityType"`
    Distance                      *float64 `json:"distance"`
    Duration                      *float64 `json:"duration"`
    ElapsedDuration              *float64 `json:"elapsedDuration"`
    MovingDuration               *float64 `json:"movingDuration"`
    ElevationGain                *float64 `json:"elevationGain"`
    ElevationLoss                *float64 `json:"elevationLoss"`
    MinElevation                 *float64 `json:"minElevation"`
    MaxElevation                 *float64 `json:"maxElevation"`
    AverageSpeed                 *float64 `json:"averageSpeed"`
    MaxSpeed                     *float64 `json:"maxSpeed"`
    Calories                     *float64 `json:"calories"`
    BMRCalories                  *float64 `json:"bmrCalories"`
    AverageHR                    *int     `json:"averageHR"`
    MaxHR                        *int     `json:"maxHR"`
    AverageRunCadence            *float64 `json:"averageRunCadence"`
    MaxRunCadence                *float64 `json:"maxRunCadence"`
    AverageBikeCadence           *float64 `json:"averageBikeCadence"`
    MaxBikeCadence               *float64 `json:"maxBikeCadence"`
    AveragePower                 *float64 `json:"averagePower"`
    MaxPower                     *float64 `json:"maxPower"`
    NormalizedPower              *float64 `json:"normalizedPower"`
    TrainingStressScore          *float64 `json:"trainingStressScore"`
    IntensityFactor              *float64 `json:"intensityFactor"`
    LeftRightBalance             *struct {
        Left  float64 `json:"left"`
        Right float64 `json:"right"`
    } `json:"leftRightBalance"`
    AvgStrokes                   *float64 `json:"avgStrokes"`
    AvgStrokeDistance            *float64 `json:"avgStrokeDistance"`
    PoolLength                   *float64 `json:"poolLength"`
    StrokesLengthType            *string  `json:"strokesLengthType"`
    ActivityTrainingLoad         *float64 `json:"activityTrainingLoad"`
    Weather                      *struct {
        Temp            *float64 `json:"temp"`
        ApparentTemp    *float64 `json:"apparentTemp"`
        DewPoint        *float64 `json:"dewPoint"`
        RelativeHumidity *int    `json:"relativeHumidity"`
        WindDirection   *int     `json:"windDirection"`
        WindSpeed       *float64 `json:"windSpeed"`
        PressureAltimeter *float64 `json:"pressureAltimeter"`
        WeatherCondition *string  `json:"weatherCondition"`
    } `json:"weather"`
    SplitSummaries              []struct {
        SplitType        string   `json:"splitType"`
        SplitIndex       int      `json:"splitIndex"`
        StartTimeGMT     string   `json:"startTimeGMT"`
        Distance         float64  `json:"distance"`
        Duration         float64  `json:"duration"`
        MovingDuration   *float64 `json:"movingDuration"`
        ElevationChange  *float64 `json:"elevationChange"`
        AverageSpeed     *float64 `json:"averageSpeed"`
        MaxSpeed         *float64 `json:"maxSpeed"`
        AverageHR        *int     `json:"averageHR"`
        MaxHR            *int     `json:"maxHR"`
        AveragePower     *float64 `json:"averagePower"`
        MaxPower         *float64 `json:"maxPower"`
        Calories         *float64 `json:"calories"`
    } `json:"splitSummaries"`
}
```

### Device Management Endpoints

#### Device List
- **Endpoint**: `GET /device-service/deviceregistration/devices`
- **Purpose**: Get list of user's registered devices

```go
type Device struct {
    DeviceID                int64   `json:"deviceId"`
    DeviceTypePK            int     `json:"deviceTypePk"`
    DeviceTypeID            int     `json:"deviceTypeId"`
    DeviceVersionPK         int     `json:"deviceVersionPk"`
    ApplicationVersions     []struct {
        ApplicationTypePK int    `json:"applicationTypePk"`
        VersionString     string `json:"versionString"`
        ApplicationKey    string `json:"applicationKey"`
    } `json:"applicationVersions"`
    LastSyncTimeStamp       *string `json:"lastSyncTimeStamp"`
    ImageURL                string  `json:"imageUrl"`
    DeviceRegistrationDate  string  `json:"deviceRegistrationDate"`
    DeviceSettingsURL       *string `json:"deviceSettingsUrl"`
    DisplayName             string  `json:"displayName"`
    PartNumber              string  `json:"partNumber"`
    SoftwareVersionString   string  `json:"softwareVersionString"`
    UnitID                  string  `json:"unitId"`
    PrimaryDevice           bool    `json:"primaryDevice"`
}

type DeviceListResponse struct {
    Devices []Device `json:"devices"`
}
```

### Social/Community Endpoints

#### Social Profile
- **Endpoint**: `GET /userprofile-service/socialProfile/{profileId}`
- **Purpose**: Get public profile information for other users

#### Connections/Friends
- **Endpoint**: `GET /userprofile-service/connection-service/connections`
- **Purpose**: Get user's connections/friends list

```go
type Connection struct {
    ProfileID       int     `json:"profileId"`
    UserProfileID   int     `json:"userProfileId"`
    DisplayName     string  `json:"displayName"`
    FullName        *string `json:"fullName"`
    ProfileImageURL *string `json:"profileImageUrl"`
    Location        *string `json:"location"`
    ConnectionDate  string  `json:"connectionDate"`
    UserPro         bool    `json:"userPro"`
}

type ConnectionsResponse struct {
    Connections []Connection `json:"connections"`
}
```

### Nutrition/Hydration Endpoints

#### Log Hydration
- **Endpoint**: `POST /wellness-service/wellness/hydrationLog`
- **Purpose**: Log hydration intake

```go
type HydrationLogRequest struct {
    CalendarDate string  `json:"calendarDate"`
    ValueInML    float64 `json:"valueInMl"`
    TimestampGMT int64   `json:"timestampGmt"`
}

type HydrationLogResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message,omitempty"`
}
```

### Goals and Badges Endpoints

#### User Goals
- **Endpoint**: `GET /userprofile-service/userprofile/personal-information/goals`
- **Purpose**: Get user's fitness goals

```go
type Goals struct {
    WeeklyStepGoal          *int     `json:"weeklyStepGoal"`
    WeeklyIntensityMinutes  *int     `json:"weeklyIntensityMinutes"`
    WeeklyFloorsClimbedGoal *int     `json:"weeklyFloorsClimbedGoal"`
    WeeklyWorkoutGoal       *int     `json:"weeklyWorkoutGoal"`
    DailyHydrationGoal      *float64 `json:"dailyHydrationGoal"`
    WeightGoal              *struct {
        Weight     float64 `json:"weight"`
        TargetDate string  `json:"targetDate"`
        GoalType   string  `json:"goalType"`
    } `json:"weightGoal"`
}
```

#### Badges
- **Endpoint**: `GET /badge-service/badges/{profileId}`
- **Purpose**: Get user's earned badges

```go
type Badge struct {
    BadgeKey        string   `json:"badgeKey"`
    BadgeTypeID     int      `json:"badgeTypeId"`
    BadgeTypeName   string   `json:"badgeTypeName"`
    BadgePoints     int      `json:"badgePoints"`
    EarnedDate      string   `json:"earnedDate"`
    BadgeImageURL   string   `json:"badgeImageUrl"`
    ViewableBy      []string `json:"viewableBy"`
    BadgeCategory   string   `json:"badgeCategory"`
    AssociatedGoal  *string  `json:"associatedGoal"`
}

type BadgesResponse struct {
    Badges []Badge `json:"badges"`
}
```

## Wellness Insights and Trends

### Wellness Dashboard
- **Endpoint**: `GET /wellness-service/wellness/wellness-dashboard/{date}`
- **Purpose**: Get comprehensive wellness dashboard data

```go
type WellnessDashboard struct {
    CalendarDate        string `json:"calendarDate"`
    StepsData           *struct {
        TotalSteps    int `json:"totalSteps"`
        StepGoal      int `json:"stepGoal"`
        PercentGoal   int `json:"percentGoal"`
    } `json:"stepsData"`
    IntensityMinutes    *struct {
        WeeklyGoal      int `json:"weeklyGoal"`
        ModerateMinutes int `json:"moderateMinutes"`
        VigorousMinutes int `json:"vigorousMinutes"`
        TotalMinutes    int `json:"totalMinutes"`
        PercentGoal     int `json:"percentGoal"`
    } `json:"intensityMinutes"`
    FloorsClimbed       *struct {
        FloorsClimbed int `json:"floorsClimbed"`
        GoalFloors    int `json:"goalFloors"`
        PercentGoal   int `json:"percentGoal"`
    } `json:"floorsClimbed"`
    CaloriesData        *struct {
        TotalCalories  int `json:"totalCalories"`
        ActiveCalories int `json:"activeCalories"`
        BMRCalories    int `json:"bmrCalories"`
        CaloriesGoal   int `json:"caloriesGoal"`
    } `json:"caloriesData"`
}
```

## Training and Performance

### Training Status
- **Endpoint**: `GET /metrics-service/metrics/training-status/{profileId}`
- **Purpose**: Get training status and load information

```go
type TrainingStatus struct {
    TrainingStatusKey           string   `json:"trainingStatusKey"`
    LoadRatio                   *float64 `json:"loadRatio"`
    TrainingLoad                *float64 `json:"trainingLoad"`
    TrainingLoadFocus           *string  `json:"trainingLoadFocus"`
    TrainingEffectLabel         *string  `json:"trainingEffectLabel"`
    AnaerobicTrainingEffect     *float64 `json:"anaerobicTrainingEffect"`
    AerobicTrainingEffect       *float64 `json:"aerobicTrainingEffect"`
    TrainingEffectMessage       *string  `json:"trainingEffectMessage"`
    FitnessLevel               *string  `json:"fitnessLevel"`
    RecoveryTime               *int     `json:"recoveryTime"`
    RecoveryInfo               *string  `json:"recoveryInfo"`
}
```

### VO2 Max
- **Endpoint**: `GET /metrics-service/metrics/vo2max/{profileId}`
- **Purpose**: Get VO2 Max measurements and trends

```go
type VO2Max struct {
    ActivityType     string   `json:"activityType"`
    VO2MaxValue      *float64 `json:"vo2MaxValue"`
    FitnessAge       *int     `json:"fitnessAge"`
    FitnessLevel     string   `json:"fitnessLevel"`
    LastMeasurement  string   `json:"lastMeasurement"`
    Generic          *float64 `json:"generic"`
    Running          *float64 `json:"running"`
    Cycling          *float64 `json:"cycling"`
}
```

## Golf Endpoints

### Golf Scorecard
- **Endpoint**: `GET /golf-service/golf/scorecard/{scorecardId}`
- **Purpose**: Get golf scorecard details

```go
type GolfScorecard struct {
    ScorecardID      int64  `json:"scorecardId"`
    CourseID         int    `json:"courseId"`
    CourseName       string `json:"courseName"`
    PlayedDate       string `json:"playedDate"`
    TotalScore       int    `json:"totalScore"`
    TotalStrokes     int    `json:"totalStrokes"`
    CoursePar        int    `json:"coursePar"`
    CourseRating     float64 `json:"courseRating"`
    CourseSlope      int    `json:"courseSlope"`
    HandicapIndex    *float64 `json:"handicapIndex"`
    PlayingHandicap  *int    `json:"playingHandicap"`
    NetScore         *int    `json:"netScore"`
    Holes            []struct {
        HoleNumber   int     `json:"holeNumber"`
        Par          int     `json:"par"`
        Strokes      int     `json:"strokes"`
        HoleHandicap int     `json:"holeHandicap"`
        Distance     int     `json:"distance"`
        Score        int     `json:"score"`
        NetStrokes   *int    `json:"netStrokes"`
    } `json:"holes"`
}
```

## Pagination and Limits

### Common Pagination Parameters
- `limit`: Number of items per page (varies by endpoint)
- `start`: Start index or date
- `end`: End index or date

### Rate Limiting
The API implements rate limiting. Common limits observed:
- OAuth token requests: Limited per hour
- Data requests: Typically allow reasonable polling intervals
- File uploads: Size and frequency restrictions

## Error Codes and Handling

### Common HTTP Status Codes
- `200`: Success
- `204`: No Content (successful request with no data)
- `400`: Bad Request
- `401`: Unauthorized (token expired/invalid)
- `403`: Forbidden (insufficient permissions)
- `404`: Not Found
- `429`: Too Many Requests (rate limited)
- `500`: Internal Server Error

### Error Response Format
```go
type APIError struct {
    HTTPStatusCode int    `json:"httpStatusCode,omitempty"`
    HTTPStatus     string `json:"httpStatus,omitempty"`
    RequestURL     string `json:"requestUrl,omitempty"`
    ErrorMessage   string `json:"errorMessage"`
    ValidationErrors []struct {
        PropertyName string `json:"propertyName"`
        Message      string `json:"message"`
    } `json:"validationErrors,omitempty"`
}
```

## Data Synchronization

### Sync Status
- **Endpoint**: `GET /device-service/deviceservice/device-info/sync-status`
- **Purpose**: Check device synchronization status

```go
type SyncStatus struct {
    LastSyncTime     *string `json:"lastSyncTime"`
    SyncInProgress   bool    `json:"syncInProgress"`
    PendingDataTypes []struct {
        DataType     string `json:"dataType"`
        RecordCount  int    `json:"recordCount"`
        LastUpdate   string `json:"lastUpdate"`
    } `json:"pendingDataTypes"`
}
```

## Time Zones and Localization

### Supported Date Formats
- ISO 8601: `YYYY-MM-DD`
- ISO 8601 with time: `YYYY-MM-DDTHH:MM:SS.sssZ`
- Unix timestamp (milliseconds)

### Timezone Handling
All endpoints return both GMT and local timestamps where applicable:
- `timestampGmt`: UTC timestamp
- `timestampLocal`: Local timezone timestamp
- Timezone offset information included for conversion

This comprehensive documentation covers all the major endpoints and data structures available through the Garmin Connect API as implemented in the Garth library.