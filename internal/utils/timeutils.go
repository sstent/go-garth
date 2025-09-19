package utils

import (
	"time"
	"strconv"
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

// SetDefaultLocation sets the default time location for conversions
func SetDefaultLocation(loc *time.Location) {
	defaultLocation = loc
}

// ParseTimestamp converts a millisecond timestamp to time.Time in default location
func ParseTimestamp(ts int) time.Time {
	return time.Unix(0, int64(ts)*int64(time.Millisecond)).In(defaultLocation)
}

// ToLocalTime converts UTC time to local time using default location
func ToLocalTime(utcTime time.Time) time.Time {
	return utcTime.In(defaultLocation)
}

// ToUTCTime converts local time to UTC
func ToUTCTime(localTime time.Time) time.Time {
	return localTime.UTC()
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