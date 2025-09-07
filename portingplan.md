# Garth Python to Go Port Plan

## Overview
Port the Python `garth` library to Go with feature parity. The existing Go code provides basic authentication and activity retrieval. This plan outlines the systematic porting of all Python modules.

## Current State Analysis
**Existing Go code has:**
- Basic SSO authentication flow (`main.go`)
- OAuth1/OAuth2 token handling
- Activity retrieval
- Session persistence

**Missing (needs porting):**
- All data models and retrieval methods
- Stats modules
- User profile/settings
- Structured error handling
- Client configuration options

## Implementation Plan

### 1. Project Structure Setup
```
garth/
├── main.go                 (keep existing)
├── client/
│   ├── client.go          (refactor from main.go)
│   ├── auth.go            (OAuth flows)
│   └── sso.go             (SSO authentication)
├── data/
│   ├── base.go
│   ├── body_battery.go
│   ├── hrv.go
│   ├── sleep.go
│   └── weight.go
├── stats/
│   ├── base.go
│   ├── hrv.go
│   ├── steps.go
│   ├── stress.go
│   └── [other stats].go
├── users/
│   ├── profile.go
│   └── settings.go
├── utils/
│   └── utils.go
└── types/
    └── tokens.go
```

### 2. Core Client Refactoring (Priority 1)

**File: `client/client.go`**
- Extract client logic from `main.go`
- Port `src/garth/http.py` Client class
- Key methods to implement:
  ```go
  type Client struct {
      Domain string
      HTTPClient *http.Client
      OAuth1Token *OAuth1Token
      OAuth2Token *OAuth2Token
      // ... other fields from Python Client
  }
  
  func (c *Client) Configure(opts ...ConfigOption) error
  func (c *Client) ConnectAPI(path, method string, data interface{}) (interface{}, error)
  func (c *Client) Download(path string) ([]byte, error)
  func (c *Client) Upload(filePath, uploadPath string) (map[string]interface{}, error)
  ```

**Reference:** `src/garth/http.py` lines 23-280

### 3. Authentication Module (Priority 1)

**File: `client/auth.go`**
- Port `src/garth/auth_tokens.py` token structures
- Implement token expiration checking
- Add MFA support placeholder

**File: `client/sso.go`**  
- Port SSO functions from `src/garth/sso.py`
- Extract login logic from current `main.go`
- Implement `ResumeLogin()` for MFA completion

**Reference:** `src/garth/sso.py` and `src/garth/auth_tokens.py`

### 4. Data Models Base (Priority 2)

**File: `data/base.go`**
- Port `src/garth/data/_base.py` Data interface and base functionality
- Implement concurrent data fetching pattern:
  ```go
  type Data interface {
      Get(day time.Time, client *Client) (interface{}, error)
      List(end time.Time, days int, client *Client, maxWorkers int) ([]interface{}, error)
  }
  ```

**Reference:** `src/garth/data/_base.py` lines 8-40

### 5. Body Battery Data (Priority 2)

**File: `data/body_battery.go`**
- Port all structs from `src/garth/data/body_battery/` directory
- Key structures to implement:
  ```go
  type DailyBodyBatteryStress struct {
      UserProfilePK int `json:"userProfilePk"`
      CalendarDate time.Time `json:"calendarDate"`
      // ... all fields from Python class
  }
  
  type BodyBatteryData struct {
      Event *BodyBatteryEvent `json:"event"`
      // ... other fields
  }
  ```

**Reference:** 
- `src/garth/data/body_battery/daily_stress.py`
- `src/garth/data/body_battery/events.py`
- `src/garth/data/body_battery/readings.py`

### 6. Other Data Models (Priority 2)

**Files: `data/hrv.go`, `data/sleep.go`, `data/weight.go`**

For each file, port the corresponding Python module:

**HRV Data (`data/hrv.go`):**
```go
type HRVData struct {
    UserProfilePK int `json:"userProfilePk"`
    HRVSummary HRVSummary `json:"hrvSummary"`
    HRVReadings []HRVReading `json:"hrvReadings"`
    // ... rest of fields
}
```
**Reference:** `src/garth/data/hrv.py`

**Sleep Data (`data/sleep.go`):**
- Port `DailySleepDTO`, `SleepScores`, `SleepMovement` structs
- Implement property methods as getter functions
**Reference:** `src/garth/data/sleep.py`

**Weight Data (`data/weight.go`):**
- Port `WeightData` struct with field validation
- Implement date range fetching logic
**Reference:** `src/garth/data/weight.py`

### 7. Stats Modules (Priority 3)

**File: `stats/base.go`**
- Port `src/garth/stats/_base.py` Stats base class
- Implement pagination logic for large date ranges

**Individual Stats Files:**
Create separate files for each stat type, porting from corresponding Python files:
- `stats/hrv.go` ← `src/garth/stats/hrv.py`
- `stats/steps.go` ← `src/garth/stats/steps.py`  
- `stats/stress.go` ← `src/garth/stats/stress.py`
- `stats/sleep.go` ← `src/garth/stats/sleep.py`
- `stats/hydration.go` ← `src/garth/stats/hydration.py`
- `stats/intensity_minutes.go` ← `src/garth/stats/intensity_minutes.py`

**Reference:** All files in `src/garth/stats/`

### 8. User Profile and Settings (Priority 3)

**File: `users/profile.go`**
```go
type UserProfile struct {
    ID int `json:"id"`
    ProfileID int `json:"profileId"`
    DisplayName string `json:"displayName"`
    // ... all other fields from Python UserProfile
}

func (up *UserProfile) Get(client *Client) error
```

**File: `users/settings.go`**
- Port all nested structs: `PowerFormat`, `FirstDayOfWeek`, `WeatherLocation`, etc.
- Implement `UserSettings.Get()` method

**Reference:** `src/garth/users/profile.py` and `src/garth/users/settings.py`

### 9. Utilities (Priority 3)

**File: `utils/utils.go`**
```go
func CamelToSnake(s string) string
func CamelToSnakeDict(m map[string]interface{}) map[string]interface{}
func FormatEndDate(end interface{}) time.Time
func DateRange(end time.Time, days int) []time.Time
func GetLocalizedDateTime(gmtTimestamp, localTimestamp int64) time.Time
```

**Reference:** `src/garth/utils.py`

### 10. Error Handling (Priority 4)

**File: `errors/errors.go`**
```go
type GarthError struct {
    Message string
    Cause error
}

type GarthHTTPError struct {
    GarthError
    StatusCode int
    Response string
}
```

**Reference:** `src/garth/exc.py`

### 11. CLI Tool (Priority 4)

**File: `cmd/garth/main.go`**
- Port `src/garth/cli.py` functionality
- Support login and token output

### 12. Testing Strategy

For each module:
1. Create `*_test.go` files with unit tests
2. Mock HTTP responses using Python examples as expected data
3. Test error handling paths
4. Add integration tests with real API calls (optional)

### 13. Key Implementation Notes

1. **JSON Handling:** Use struct tags for proper JSON marshaling/unmarshaling
2. **Time Handling:** Convert Python datetime objects to Go `time.Time`
3. **Error Handling:** Wrap errors with context using `fmt.Errorf`
4. **Concurrency:** Use goroutines and channels for the concurrent data fetching in `List()` methods
5. **HTTP Client:** Reuse the existing HTTP client setup with proper timeout and retry logic

### 14. Development Order

1. Start with client refactoring and authentication
2. Implement base data structures and one data model (body battery)
3. Add remaining data models
4. Implement stats modules
5. Add user profile/settings
6. Complete utilities and error handling
7. Add CLI tool and tests

This plan provides a systematic approach to achieving feature parity with the Python library while maintaining Go idioms and best practices.