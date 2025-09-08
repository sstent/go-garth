package garth

import (
	"garmin-connect/garth/client"
	"garmin-connect/garth/data"
	"garmin-connect/garth/errors"
	"garmin-connect/garth/stats"
	"garmin-connect/garth/types"
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

// Activity type
type Activity = types.Activity

// Error types
type APIError = errors.APIError
type IOError = errors.IOError
type AuthError = errors.AuthenticationError
type OAuthError = errors.OAuthError
type ValidationError = errors.ValidationError

// Main functions
var (
	NewClient = client.NewClient
	Login     = client.Login
)
