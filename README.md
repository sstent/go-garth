# Garmin Connect Go Client

[![Go Reference](https://pkg.go.dev/badge/github.com/sstent/go-garth/pkg/garmin.svg)](https://pkg.go.dev/github.com/sstent/go-garth/pkg/garmin)

Go port of the Garth Python library for accessing Garmin Connect data. Provides full API coverage with improved performance and type safety.

## Installation
```bash
go get github.com/sstent/go-garth/pkg/garmin
```

## Basic Usage
```go
package main

import (
	"fmt"
	"time"
	"github.com/sstent/go-garth/pkg/garmin"
)

func main() {
	// Create client and authenticate
	client, err := garmin.NewClient("garmin.com")
	if err != nil {
		panic(err)
	}

	err = client.Login("your@email.com", "password")
	if err != nil {
		panic(err)
	}

	// List recent activities with filtering
	opts := garmin.ActivityOptions{
		Limit:       10,
		Offset:      0,
		ActivityType: "running", // optional filter
		DateFrom:    time.Now().AddDate(0, 0, -30), // last 30 days
		DateTo:      time.Now(),
	}
	activities, err := client.ListActivities(opts)
	if err != nil {
		panic(err)
	}

	for _, activity := range activities {
		fmt.Printf("%s: %s (%.2f km)\n",
			activity.StartTimeLocal.Format("2006-01-02"),
			activity.ActivityName,
			activity.Distance/1000)
	}

	// Get detailed activity information
	if len(activities) > 0 {
		activityDetail, err := client.GetActivity(activities[0].ActivityID)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Activity details: %+v\n", activityDetail)
	}

	// Search for activities
	searchResults, err := client.SearchActivities("morning run")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Found %d activities matching search\n", len(searchResults))

	// Get fitness age
	fitnessAge, err := client.GetFitnessAge()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Fitness Age: %d (Chronological: %d)\n",
		fitnessAge.FitnessAge, fitnessAge.ChronologicalAge)

	// Get health data ranges
	start := time.Now().AddDate(0, 0, -7)
	end := time.Now()

	sleepData, err := client.GetSleepData(start, end)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Sleep records: %d\n", len(sleepData))

	hrvData, err := client.GetHrvData(start, end)
	if err != nil {
		panic(err)
	}
	fmt.Printf("HRV records: %d\n", len(hrvData))

	stressData, err := client.GetStressData(start, end)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Stress records: %d\n", len(stressData))

	bodyBatteryData, err := client.GetBodyBatteryData(start, end)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Body Battery records: %d\n", len(bodyBatteryData))

	stepsData, err := client.GetStepsData(start, end)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Steps records: %d\n", len(stepsData))

	distanceData, err := client.GetDistanceData(start, end)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Distance records: %d\n", len(distanceData))

	caloriesData, err := client.GetCaloriesData(start, end)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Calories records: %d\n", len(caloriesData))
}
```

## Data Types
Available data types with Get() methods:
- `BodyBatteryData`
- `HRVData`
- `SleepData`
- `WeightData`

## Stats Types
Available stats with List() methods:

### Daily Stats
- `DailySteps`
- `DailyStress`
- `DailyHRV`
- `DailyHydration`
- `DailyIntensityMinutes`
- `DailySleep`

### Weekly Stats
- `WeeklySteps`
- `WeeklyStress`
- `WeeklyHRV`

## Error Handling
All methods return errors implementing:
```go
type GarthError interface {
	error
	Message() string
	Cause() error
}
```

Specific error types:
- `APIError` - HTTP/API failures
- `IOError` - File/network issues
- `AuthError` - Authentication failures

## Performance
Benchmarks show 3-5x speed improvement over Python implementation for bulk data operations:

```
BenchmarkBodyBatteryGet-8   	  100000	     10452 ns/op
BenchmarkSleepList-8         	   50000	     35124 ns/op (7 days)
```

## Documentation
Full API docs: [https://pkg.go.dev/github.com/sstent/go-garth/pkg/garmin](https://pkg.go.dev/github.com/sstent/go-garth/pkg/garmin)

## CLI Tool
Includes `cmd/garth` CLI for data export. Supports both daily and weekly stats:

```bash
# Daily steps
go run cmd/garth/main.go --data steps --period daily --start 2023-01-01 --end 2023-01-07

# Weekly stress
go run cmd/garth/main.go --data stress --period weekly --start 2023-01-01 --end 2023-01-28
```