package garth

import (
	"github.com/sstent/go-garth/garth/client"
	"github.com/sstent/go-garth/garth/data"
	"github.com/sstent/go-garth/garth/errors"
	"github.com/sstent/go-garth/garth/stats"
	"github.com/sstent/go-garth/garth/types"
)

// Client is the main Garmin Connect client type
type Client = client.Client

// OAuth1Token represents OAuth 1.0 token
type OAuth1Token = types.OAuth1Token

// OAuth2Token represents OAuth 2.0 token
type OAuth2Token = types.OAuth2Token

// Data types
type (
	BodyBatteryData = data.DailyBodyBatteryStress
	HRVData         = data.HRVData
	SleepData       = data.DailySleepDTO
	WeightData      = data.WeightData
)

// Stats types
type (
	Stats                 = stats.Stats
	DailySteps            = stats.DailySteps
	DailyStress           = stats.DailyStress
	DailyHRV              = stats.DailyHRV
	DailyHydration        = stats.DailyHydration
	DailyIntensityMinutes = stats.DailyIntensityMinutes
	DailySleep            = stats.DailySleep
)

// Activity represents a Garmin activity
type Activity = types.Activity

// Error types
type (
	APIError        = errors.APIError
	IOError         = errors.IOError
	AuthError       = errors.AuthenticationError
	OAuthError      = errors.OAuthError
	ValidationError = errors.ValidationError
)

// Main functions
var (
	NewClient = client.NewClient
)

// Stats constructor functions
var (
	NewDailySteps            = stats.NewDailySteps
	NewDailyStress           = stats.NewDailyStress
	NewDailyHydration        = stats.NewDailyHydration
	NewDailyIntensityMinutes = stats.NewDailyIntensityMinutes
	NewDailySleep            = stats.NewDailySleep
	NewDailyHRV              = stats.NewDailyHRV
)
