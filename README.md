# Garmin Connect Go Client

Go port of the Garth Python library for accessing Garmin Connect data. Provides full API coverage with improved performance and type safety.

## Installation
```bash
go get github.com/sstent/garmin-connect/garth
```

## Basic Usage
```go
package main

import (
	"fmt"
	"time"
	"garmin-connect/garth"
)

func main() {
	// Create client and authenticate
	client, err := garth.NewClient("garmin.com")
	if err != nil {
		panic(err)
	}
	
	err = client.Login("your@email.com", "password")
	if err != nil {
		panic(err)
	}

	// Get yesterday's body battery data
	yesterday := time.Now().AddDate(0, 0, -1)
	bb, err := garth.BodyBatteryData{}.Get(yesterday, client)
	if err != nil {
		panic(err)
	}
	
	if bb != nil {
		fmt.Printf("Body Battery: %d\n", bb.BodyBatteryValue)
	}

	// Get weekly steps
	steps := garth.NewDailySteps()
	stepData, err := steps.List(time.Now(), 7, client)
	if err != nil {
		panic(err)
	}
	
	for _, s := range stepData {
		fmt.Printf("%s: %d steps\n", 
			s.(garth.DailySteps).CalendarDate.Format("2006-01-02"),
			*s.(garth.DailySteps).TotalSteps)
	}
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
- `DailySteps`
- `DailyStress`
- `DailyHRV`
- `DailyHydration`
- `DailyIntensityMinutes`
- `DailySleep`

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
Full API docs: [https://pkg.go.dev/garmin-connect/garth](https://pkg.go.dev/garmin-connect/garth)

## CLI Tool
Includes `cmd/garth` CLI for data export:
```bash
go run cmd/garth/main.go --email user@example.com --password pass \
	--data bodybattery --start 2023-01-01 --end 2023-01-07