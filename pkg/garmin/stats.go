package garmin

import (
	"time"

	"go-garth/internal/stats"
)

// Stats is an interface for stats data types.
type Stats = stats.Stats

// NewDailySteps creates a new DailySteps stats type.
func NewDailySteps() Stats {
	return stats.NewDailySteps()
}

// NewDailyStress creates a new DailyStress stats type.
func NewDailyStress() Stats {
	return stats.NewDailyStress()
}

// NewDailyHydration creates a new DailyHydration stats type.
func NewDailyHydration() Stats {
	return stats.NewDailyHydration()
}

// NewDailyIntensityMinutes creates a new DailyIntensityMinutes stats type.
func NewDailyIntensityMinutes() Stats {
	return stats.NewDailyIntensityMinutes()
}

// NewDailySleep creates a new DailySleep stats type.
func NewDailySleep() Stats {
	return stats.NewDailySleep()
}

// NewDailyHRV creates a new DailyHRV stats type.
func NewDailyHRV() Stats {
	return stats.NewDailyHRV()
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