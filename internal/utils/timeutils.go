package utils

import (
	"time"
)

// SetDefaultLocation sets the default time location for conversions
func SetDefaultLocation(loc *time.Location) {
	// defaultLocation = loc
}

// ToLocalTime converts UTC time to local time using default location
func ToLocalTime(utcTime time.Time) time.Time {
	// return utcTime.In(defaultLocation)
	return utcTime // TODO: Implement proper time zone conversion
}

// ToUTCTime converts local time to UTC
func ToUTCTime(localTime time.Time) time.Time {
	return localTime.UTC()
}