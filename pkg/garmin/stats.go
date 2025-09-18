package garmin

import (
	"garmin-connect/internal/stats"
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
