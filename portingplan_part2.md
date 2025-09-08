# Complete Garth Python to Go Port - Implementation Plan

## Current Status
The Go port has excellent architecture (85% complete) but needs implementation of core API methods and data models. All structure, error handling, and utilities are in place.

## Phase 1: Core API Implementation (Priority 1 - Week 1)

### Task 1.1: Implement Client.ConnectAPI Method
**File:** `garth/client/client.go`
**Reference:** `src/garth/http.py` lines 206-217

Add this method to the Client struct:

```go
func (c *Client) ConnectAPI(path, method string, data interface{}) (interface{}, error) {
    url := fmt.Sprintf("https://connectapi.%s%s", c.Domain, path)
    
    var body io.Reader
    if data != nil && (method == "POST" || method == "PUT") {
        jsonData, err := json.Marshal(data)
        if err != nil {
            return nil, &errors.APIError{GarthHTTPError: errors.GarthHTTPError{
                GarthError: errors.GarthError{Message: "Failed to marshal request data", Cause: err}}}
        }
        body = bytes.NewReader(jsonData)
    }
    
    req, err := http.NewRequest(method, url, body)
    if err != nil {
        return nil, &errors.APIError{GarthHTTPError: errors.GarthHTTPError{
            GarthError: errors.GarthError{Message: "Failed to create request", Cause: err}}}
    }
    
    req.Header.Set("Authorization", c.AuthToken)
    req.Header.Set("User-Agent", "GCM-iOS-5.7.2.1")
    if body != nil {
        req.Header.Set("Content-Type", "application/json")
    }
    
    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, &errors.APIError{GarthHTTPError: errors.GarthHTTPError{
            GarthError: errors.GarthError{Message: "API request failed", Cause: err}}}
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == 204 {
        return nil, nil
    }
    
    if resp.StatusCode >= 400 {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return nil, &errors.APIError{GarthHTTPError: errors.GarthHTTPError{
            StatusCode: resp.StatusCode,
            Response: string(bodyBytes),
            GarthError: errors.GarthError{Message: "API error"}}}
    }
    
    var result interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, &errors.IOError{GarthError: errors.GarthError{
            Message: "Failed to parse response", Cause: err}}
    }
    
    return result, nil
}
```

### Task 1.2: Add File Download/Upload Methods
**File:** `garth/client/client.go`
**Reference:** `src/garth/http.py` lines 219-230, 232-244

```go
func (c *Client) Download(path string) ([]byte, error) {
    resp, err := c.ConnectAPI(path, "GET", nil)
    if err != nil {
        return nil, err
    }
    
    url := fmt.Sprintf("https://connectapi.%s%s", c.Domain, path)
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", c.AuthToken)
    req.Header.Set("User-Agent", "GCM-iOS-5.7.2.1")
    
    httpResp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer httpResp.Body.Close()
    
    return io.ReadAll(httpResp.Body)
}

func (c *Client) Upload(filePath, uploadPath string) (map[string]interface{}, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, &errors.IOError{GarthError: errors.GarthError{
            Message: "Failed to open file", Cause: err}}
    }
    defer file.Close()
    
    var b bytes.Buffer
    writer := multipart.NewWriter(&b)
    part, err := writer.CreateFormFile("file", filepath.Base(filePath))
    if err != nil {
        return nil, err
    }
    
    _, err = io.Copy(part, file)
    if err != nil {
        return nil, err
    }
    writer.Close()
    
    url := fmt.Sprintf("https://connectapi.%s%s", c.Domain, uploadPath)
    req, err := http.NewRequest("POST", url, &b)
    if err != nil {
        return nil, err
    }
    
    req.Header.Set("Authorization", c.AuthToken)
    req.Header.Set("Content-Type", writer.FormDataContentType())
    
    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    return result, nil
}
```

## Phase 2: Data Model Implementation (Week 1-2)

### Task 2.1: Complete Body Battery Implementation
**File:** `garth/data/body_battery.go`
**Reference:** `src/garth/data/body_battery/daily_stress.py` lines 55-77

Replace the stub `Get()` method:

```go
func (d *DailyBodyBatteryStress) Get(day time.Time, client *client.Client) (interface{}, error) {
    dateStr := day.Format("2006-01-02")
    path := fmt.Sprintf("/wellness-service/wellness/dailyStress/%s", dateStr)
    
    response, err := client.ConnectAPI(path, "GET", nil)
    if err != nil {
        return nil, err
    }
    
    if response == nil {
        return nil, nil
    }
    
    responseMap, ok := response.(map[string]interface{})
    if !ok {
        return nil, &errors.IOError{GarthError: errors.GarthError{
            Message: "Invalid response format"}}
    }
    
    snakeResponse := utils.CamelToSnakeDict(responseMap)
    
    jsonBytes, err := json.Marshal(snakeResponse)
    if err != nil {
        return nil, err
    }
    
    var result DailyBodyBatteryStress
    if err := json.Unmarshal(jsonBytes, &result); err != nil {
        return nil, err
    }
    
    return &result, nil
}
```

### Task 2.2: Complete Sleep Data Implementation
**File:** `garth/data/sleep.go`
**Reference:** `src/garth/data/sleep.py` lines 91-107

```go
func (d *DailySleepDTO) Get(day time.Time, client *client.Client) (interface{}, error) {
    dateStr := day.Format("2006-01-02")
    path := fmt.Sprintf("/wellness-service/wellness/dailySleepData/%s?nonSleepBufferMinutes=60&date=%s", 
        client.Username, dateStr)
    
    response, err := client.ConnectAPI(path, "GET", nil)
    if err != nil {
        return nil, err
    }
    
    if response == nil {
        return nil, nil
    }
    
    responseMap := response.(map[string]interface{})
    snakeResponse := utils.CamelToSnakeDict(responseMap)
    
    dailySleepDto, exists := snakeResponse["daily_sleep_dto"].(map[string]interface{})
    if !exists || dailySleepDto["id"] == nil {
        return nil, nil // No sleep data
    }
    
    jsonBytes, err := json.Marshal(snakeResponse)
    if err != nil {
        return nil, err
    }
    
    var result struct {
        DailySleepDTO *DailySleepDTO   `json:"daily_sleep_dto"`
        SleepMovement []SleepMovement  `json:"sleep_movement"`
    }
    
    if err := json.Unmarshal(jsonBytes, &result); err != nil {
        return nil, err
    }
    
    return result, nil
}
```

### Task 2.3: Complete HRV Implementation
**File:** `garth/data/hrv.go`
**Reference:** `src/garth/data/hrv.py` lines 68-78

```go
func (h *HRVData) Get(day time.Time, client *client.Client) (interface{}, error) {
    dateStr := day.Format("2006-01-02")
    path := fmt.Sprintf("/hrv-service/hrv/%s", dateStr)
    
    response, err := client.ConnectAPI(path, "GET", nil)
    if err != nil {
        return nil, err
    }
    
    if response == nil {
        return nil, nil
    }
    
    responseMap := response.(map[string]interface{})
    snakeResponse := utils.CamelToSnakeDict(responseMap)
    
    jsonBytes, err := json.Marshal(snakeResponse)
    if err != nil {
        return nil, err
    }
    
    var result HRVData
    if err := json.Unmarshal(jsonBytes, &result); err != nil {
        return nil, err
    }
    
    return &result, nil
}
```

### Task 2.4: Complete Weight Implementation
**File:** `garth/data/weight.go`
**Reference:** `src/garth/data/weight.py` lines 39-52 and 54-74

```go
func (w *WeightData) Get(day time.Time, client *client.Client) (interface{}, error) {
    dateStr := day.Format("2006-01-02")
    path := fmt.Sprintf("/weight-service/weight/dayview/%s", dateStr)
    
    response, err := client.ConnectAPI(path, "GET", nil)
    if err != nil {
        return nil, err
    }
    
    if response == nil {
        return nil, nil
    }
    
    responseMap := response.(map[string]interface{})
    dayWeightList, exists := responseMap["dateWeightList"].([]interface{})
    if !exists || len(dayWeightList) == 0 {
        return nil, nil
    }
    
    // Get first weight entry
    firstEntry := dayWeightList[0].(map[string]interface{})
    snakeResponse := utils.CamelToSnakeDict(firstEntry)
    
    jsonBytes, err := json.Marshal(snakeResponse)
    if err != nil {
        return nil, err
    }
    
    var result WeightData
    if err := json.Unmarshal(jsonBytes, &result); err != nil {
        return nil, err
    }
    
    return &result, nil
}
```

## Phase 3: Stats Module Implementation (Week 2)

### Task 3.1: Create Stats Base
**File:** `garth/stats/base.go` (new file)
**Reference:** `src/garth/stats/_base.py`

```go
package stats

import (
    "fmt"
    "time"
    "garmin-connect/garth/client"
    "garmin-connect/garth/utils"
)

type Stats interface {
    List(end time.Time, period int, client *client.Client) ([]interface{}, error)
}

type BaseStats struct {
    Path     string
    PageSize int
}

func (b *BaseStats) List(end time.Time, period int, client *client.Client) ([]interface{}, error) {
    endDate := utils.FormatEndDate(end)
    
    if period > b.PageSize {
        // Handle pagination - get first page
        page, err := b.fetchPage(endDate, b.PageSize, client)
        if err != nil || len(page) == 0 {
            return page, err
        }
        
        // Get remaining pages recursively
        remainingStart := endDate.AddDate(0, 0, -b.PageSize)
        remainingPeriod := period - b.PageSize
        remainingData, err := b.List(remainingStart, remainingPeriod, client)
        if err != nil {
            return page, err
        }
        
        return append(remainingData, page...), nil
    }
    
    return b.fetchPage(endDate, period, client)
}

func (b *BaseStats) fetchPage(end time.Time, period int, client *client.Client) ([]interface{}, error) {
    var start time.Time
    var path string
    
    if strings.Contains(b.Path, "daily") {
        start = end.AddDate(0, 0, -(period - 1))
        path = strings.Replace(b.Path, "{start}", start.Format("2006-01-02"), 1)
        path = strings.Replace(path, "{end}", end.Format("2006-01-02"), 1)
    } else {
        path = strings.Replace(b.Path, "{end}", end.Format("2006-01-02"), 1)
        path = strings.Replace(path, "{period}", fmt.Sprintf("%d", period), 1)
    }
    
    response, err := client.ConnectAPI(path, "GET", nil)
    if err != nil {
        return nil, err
    }
    
    if response == nil {
        return []interface{}{}, nil
    }
    
    responseSlice, ok := response.([]interface{})
    if !ok || len(responseSlice) == 0 {
        return []interface{}{}, nil
    }
    
    var results []interface{}
    for _, item := range responseSlice {
        itemMap := item.(map[string]interface{})
        
        // Handle nested "values" structure
        if values, exists := itemMap["values"]; exists {
            valuesMap := values.(map[string]interface{})
            for k, v := range valuesMap {
                itemMap[k] = v
            }
            delete(itemMap, "values")
        }
        
        snakeItem := utils.CamelToSnakeDict(itemMap)
        results = append(results, snakeItem)
    }
    
    return results, nil
}
```

### Task 3.2: Create Individual Stats Types
**Files:** Create these files in `garth/stats/`
**Reference:** All files in `src/garth/stats/`

**`steps.go`** (Reference: `src/garth/stats/steps.py`):
```go
package stats

import "time"

const BASE_STEPS_PATH = "/usersummary-service/stats/steps"

type DailySteps struct {
    CalendarDate  time.Time `json:"calendar_date"`
    TotalSteps    *int      `json:"total_steps"`
    TotalDistance *int      `json:"total_distance"`
    StepGoal      int       `json:"step_goal"`
    BaseStats
}

func NewDailySteps() *DailySteps {
    return &DailySteps{
        BaseStats: BaseStats{
            Path:     BASE_STEPS_PATH + "/daily/{start}/{end}",
            PageSize: 28,
        },
    }
}

type WeeklySteps struct {
    CalendarDate            time.Time `json:"calendar_date"`
    TotalSteps              int       `json:"total_steps"`
    AverageSteps            float64   `json:"average_steps"`
    AverageDistance         float64   `json:"average_distance"`
    TotalDistance           float64   `json:"total_distance"`
    WellnessDataDaysCount   int       `json:"wellness_data_days_count"`
    BaseStats
}

func NewWeeklySteps() *WeeklySteps {
    return &WeeklySteps{
        BaseStats: BaseStats{
            Path:     BASE_STEPS_PATH + "/weekly/{end}/{period}",
            PageSize: 52,
        },
    }
}
```

**`stress.go`** (Reference: `src/garth/stats/stress.py`):
```go
package stats

import "time"

const BASE_STRESS_PATH = "/usersummary-service/stats/stress"

type DailyStress struct {
    CalendarDate         time.Time `json:"calendar_date"`
    OverallStressLevel   int       `json:"overall_stress_level"`
    RestStressDuration   *int      `json:"rest_stress_duration"`
    LowStressDuration    *int      `json:"low_stress_duration"`
    MediumStressDuration *int      `json:"medium_stress_duration"`
    HighStressDuration   *int      `json:"high_stress_duration"`
    BaseStats
}

func NewDailyStress() *DailyStress {
    return &DailyStress{
        BaseStats: BaseStats{
            Path:     BASE_STRESS_PATH + "/daily/{start}/{end}",
            PageSize: 28,
        },
    }
}
```

Create similar files for:
- `hydration.go` → Reference `src/garth/stats/hydration.py`
- `intensity_minutes.go` → Reference `src/garth/stats/intensity_minutes.py`
- `sleep.go` → Reference `src/garth/stats/sleep.py`
- `hrv.go` → Reference `src/garth/stats/hrv.py`

## Phase 4: Complete Data Interface Implementation (Week 2)

### Task 4.1: Fix BaseData List Implementation
**File:** `garth/data/base.go`

Update the List method to properly use the BaseData pattern:

```go
func (b *BaseData) List(end time.Time, days int, c *client.Client, maxWorkers int) ([]interface{}, []error) {
    if maxWorkers < 1 {
        maxWorkers = 10 // Match Python's MAX_WORKERS
    }

    dates := utils.DateRange(end, days)

    var wg sync.WaitGroup
    workCh := make(chan time.Time, days)
    resultsCh := make(chan result, days)

    type result struct {
        data interface{}
        err  error
    }

    // Worker function
    worker := func() {
        defer wg.Done()
        for date := range workCh {
            data, err := b.Get(date, c)
            resultsCh <- result{data: data, err: err}
        }
    }

    // Start workers
    wg.Add(maxWorkers)
    for i := 0; i < maxWorkers; i++ {
        go worker()
    }

    // Send work
    go func() {
        for _, date := range dates {
            workCh <- date
        }
        close(workCh)
    }()

    // Close results channel when workers are done
    go func() {
        wg.Wait()
        close(resultsCh)
    }()

    var results []interface{}
    var errs []error

    for r := range resultsCh {
        if r.err != nil {
            errs = append(errs, r.err)
        } else if r.data != nil {
            results = append(results, r.data)
        }
    }

    return results, errs
}
```

## Phase 5: Testing and Documentation (Week 3)

### Task 5.1: Create Integration Tests
**File:** `garth/integration_test.go` (new file)

```go
package garth_test

import (
    "testing"
    "time"
    "garmin-connect/garth/client"
    "garmin-connect/garth/data"
)

func TestBodyBatteryIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    c, err := client.NewClient("garmin.com")
    require.NoError(t, err)
    
    // Load test session
    err = c.LoadSession("test_session.json")
    if err != nil {
        t.Skip("No test session available")
    }
    
    bb := &data.DailyBodyBatteryStress{}
    result, err := bb.Get(time.Now().AddDate(0, 0, -1), c)
    
    assert.NoError(t, err)
    if result != nil {
        bbData := result.(*data.DailyBodyBatteryStress)
        assert.NotZero(t, bbData.UserProfilePK)
    }
}
```

### Task 5.2: Update Package Exports
**File:** `garth/__init__.go` (new file)

Create a package-level API that matches Python's `__init__.py`:

```go
package garth

import (
    "garmin-connect/garth/client"
    "garmin-connect/garth/data"
    "garmin-connect/garth/stats"
)

// Re-export main types for convenience
type Client = client.Client

// Data types
type BodyBatteryData = data.DailyBodyBatteryStress
type HRVData = data.HRVData
type SleepData = data.DailySleepDTO
type WeightData = data.WeightData

// Stats types
type DailySteps = stats.DailySteps
type DailyStress = stats.DailyStress
type DailyHRV = stats.DailyHRV

// Main functions
var (
    NewClient = client.NewClient
    Login     = client.Login
)
```

## Implementation Checklist

### Week 1 (Core Implementation):
- [ ] Client.ConnectAPI method
- [ ] Download/Upload methods  
- [ ] Body Battery Get() implementation
- [ ] Sleep Data Get() implementation
- [ ] End-to-end test with real API

### Week 2 (Complete Feature Set):
- [ ] HRV and Weight Get() implementations
- [ ] Complete stats module (all 7 types)
- [ ] BaseData List() method fix
- [ ] Integration tests

### Week 3 (Polish and Documentation):
- [ ] Package-level exports
- [ ] README with examples
- [ ] Performance testing vs Python
- [ ] CLI tool verification

## Key Implementation Notes

1. **Error Handling**: Use the existing comprehensive error types
2. **Date Formats**: Always use `time.Time` and convert to "2006-01-02" for API calls
3. **Response Parsing**: Always use `utils.CamelToSnakeDict` before unmarshaling
4. **Concurrency**: The existing BaseData.List() handles worker pools correctly
5. **Testing**: Use `testutils.MockJSONResponse` for unit tests

## Success Criteria

Port is complete when:
- All Python data models have working Get() methods
- All Python stats types are implemented  
- CLI tool outputs same format as Python
- Integration tests pass against real API
- Performance is equal or better than Python

**Estimated Effort:** 2-3 weeks for junior developer with this detailed plan.