package garmin

import types "go-garth/internal/models/types"

// GarminTime represents Garmin's timestamp format with custom JSON parsing
type GarminTime = types.GarminTime

// SessionData represents saved session information
type SessionData = types.SessionData

// ActivityType represents the type of activity
type ActivityType = types.ActivityType

// EventType represents the event type of an activity
type EventType = types.EventType

// Activity represents a Garmin Connect activity
type Activity = types.Activity

// UserProfile represents a Garmin user profile
type UserProfile = types.UserProfile

// OAuth1Token represents OAuth1 token response
type OAuth1Token = types.OAuth1Token

// OAuth2Token represents OAuth2 token response
type OAuth2Token = types.OAuth2Token

// DetailedSleepData represents comprehensive sleep data
type DetailedSleepData = types.DetailedSleepData

// SleepLevel represents different sleep stages
type SleepLevel = types.SleepLevel

// SleepMovement represents movement during sleep
type SleepMovement = types.SleepMovement

// SleepScore represents detailed sleep scoring
type SleepScore = types.SleepScore

// SleepScoreBreakdown represents breakdown of sleep score
type SleepScoreBreakdown = types.SleepScoreBreakdown

// HRVBaseline represents HRV baseline data
type HRVBaseline = types.HRVBaseline

// DailyHRVData represents comprehensive daily HRV data
type DailyHRVData = types.DailyHRVData

// BodyBatteryEvent represents events that impact Body Battery
type BodyBatteryEvent = types.BodyBatteryEvent

// DetailedBodyBatteryData represents comprehensive Body Battery data
type DetailedBodyBatteryData = types.DetailedBodyBatteryData

// TrainingStatus represents current training status
type TrainingStatus = types.TrainingStatus

// TrainingLoad represents training load data
type TrainingLoad = types.TrainingLoad

// FitnessAge represents fitness age calculation
type FitnessAge = types.FitnessAge

// VO2MaxData represents VO2 max data
type VO2MaxData = types.VO2MaxData

// VO2MaxEntry represents a single VO2 max entry
type VO2MaxEntry = types.VO2MaxEntry

// HeartRateZones represents heart rate zone data
type HeartRateZones = types.HeartRateZones

// HRZone represents a single heart rate zone
type HRZone = types.HRZone

// WellnessData represents additional wellness metrics
type WellnessData = types.WellnessData

// SleepData represents sleep summary data
type SleepData = types.SleepData

// HrvData represents Heart Rate Variability data
type HrvData = types.HrvData

// StressData represents stress level data
type StressData = types.StressData

// BodyBatteryData represents Body Battery data
type BodyBatteryData = types.BodyBatteryData
