package garmin

import "go-garth/internal/types"

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