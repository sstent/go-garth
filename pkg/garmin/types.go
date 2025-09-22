package garmin

import garth "github.com/sstent/go-garth/pkg/garth/types"

// GarminTime represents Garmin's timestamp format with custom JSON parsing
type GarminTime = garth.GarminTime

// SessionData represents saved session information
type SessionData = garth.SessionData

// ActivityType represents the type of activity
type ActivityType = garth.ActivityType

// EventType represents the event type of an activity
type EventType = garth.EventType

// Activity represents a Garmin Connect activity
type Activity = garth.Activity

// UserProfile represents a Garmin user profile
type UserProfile = garth.UserProfile

// OAuth1Token represents OAuth1 token response
type OAuth1Token = garth.OAuth1Token

// OAuth2Token represents OAuth2 token response
type OAuth2Token = garth.OAuth2Token

// DetailedSleepData represents comprehensive sleep data
type DetailedSleepData = garth.DetailedSleepData

// SleepLevel represents different sleep stages
type SleepLevel = garth.SleepLevel

// SleepMovement represents movement during sleep
type SleepMovement = garth.SleepMovement

// SleepScore represents detailed sleep scoring
type SleepScore = garth.SleepScore

// SleepScoreBreakdown represents breakdown of sleep score
type SleepScoreBreakdown = garth.SleepScoreBreakdown

// HRVBaseline represents HRV baseline data
type HRVBaseline = garth.HRVBaseline

// DailyHRVData represents comprehensive daily HRV data
type DailyHRVData = garth.DailyHRVData

// BodyBatteryEvent represents events that impact Body Battery
type BodyBatteryEvent = garth.BodyBatteryEvent

// DetailedBodyBatteryData represents comprehensive Body Battery data
type DetailedBodyBatteryData = garth.DetailedBodyBatteryData

// TrainingStatus represents current training status
type TrainingStatus = garth.TrainingStatus

// TrainingLoad represents training load data
type TrainingLoad = garth.TrainingLoad

// FitnessAge represents fitness age calculation
type FitnessAge = garth.FitnessAge

// VO2MaxData represents VO2 max data
type VO2MaxData = garth.VO2MaxData

// VO2MaxEntry represents a single VO2 max entry
type VO2MaxEntry = garth.VO2MaxEntry

// HeartRateZones represents heart rate zone data
type HeartRateZones = garth.HeartRateZones

// HRZone represents a single heart rate zone
type HRZone = garth.HRZone

// WellnessData represents additional wellness metrics
type WellnessData = garth.WellnessData

// SleepData represents sleep summary data
type SleepData = garth.SleepData

// HrvData represents Heart Rate Variability data
type HrvData = garth.HrvData

// StressData represents stress level data
type StressData = garth.StressData

// BodyBatteryData represents Body Battery data
type BodyBatteryData = garth.BodyBatteryData

// StepsData represents steps statistics
type StepsData = garth.StepsData

// DistanceData represents distance statistics
type DistanceData = garth.DistanceData

// CaloriesData represents calories statistics
type CaloriesData = garth.CaloriesData
