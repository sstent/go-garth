package utils

import (
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
