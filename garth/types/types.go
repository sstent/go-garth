package types

import (
	"net/http"
	"time"
)

// Client represents the Garmin Connect client
type Client struct {
	Domain      string
	HTTPClient  *http.Client
	Username    string
	AuthToken   string
	OAuth1Token *OAuth1Token
	OAuth2Token *OAuth2Token
}

// SessionData represents saved session information
type SessionData struct {
	Domain    string `json:"domain"`
	Username  string `json:"username"`
	AuthToken string `json:"auth_token"`
}

// ActivityType represents the type of activity
type ActivityType struct {
	TypeID       int    `json:"typeId"`
	TypeKey      string `json:"typeKey"`
	ParentTypeID *int   `json:"parentTypeId,omitempty"`
}

// EventType represents the event type of an activity
type EventType struct {
	TypeID  int    `json:"typeId"`
	TypeKey string `json:"typeKey"`
}

// Activity represents a Garmin Connect activity
type Activity struct {
	ActivityID      int64        `json:"activityId"`
	ActivityName    string       `json:"activityName"`
	Description     string       `json:"description"`
	StartTimeLocal  string       `json:"startTimeLocal"`
	StartTimeGMT    string       `json:"startTimeGMT"`
	ActivityType    ActivityType `json:"activityType"`
	EventType       EventType    `json:"eventType"`
	Distance        float64      `json:"distance"`
	Duration        float64      `json:"duration"`
	ElapsedDuration float64      `json:"elapsedDuration"`
	MovingDuration  float64      `json:"movingDuration"`
	ElevationGain   float64      `json:"elevationGain"`
	ElevationLoss   float64      `json:"elevationLoss"`
	AverageSpeed    float64      `json:"averageSpeed"`
	MaxSpeed        float64      `json:"maxSpeed"`
	Calories        float64      `json:"calories"`
	AverageHR       float64      `json:"averageHR"`
	MaxHR           float64      `json:"maxHR"`
}

// OAuth1Token represents OAuth1 token response
type OAuth1Token struct {
	OAuthToken       string `json:"oauth_token"`
	OAuthTokenSecret string `json:"oauth_token_secret"`
	MFAToken         string `json:"mfa_token,omitempty"`
	Domain           string `json:"domain"`
}

// OAuth2Token represents OAuth2 token response
type OAuth2Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	RefreshToken string    `json:"refresh_token"`
	Scope        string    `json:"scope"`
	CreatedAt    time.Time // Used for expiration tracking
	ExpiresAt    time.Time // Computed expiration time
}

// OAuthConsumer represents OAuth consumer credentials
type OAuthConsumer struct {
	ConsumerKey    string `json:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret"`
}
