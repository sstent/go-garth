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
type DailyHydration = stats.DailyHydration
type DailyIntensityMinutes = stats.DailyIntensityMinutes
type DailySleep = stats.DailySleep

// Main functions
var (
	NewClient = client.NewClient
	Login     = client.Login
)
