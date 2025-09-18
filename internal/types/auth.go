package types

import "time"

// OAuthConsumer represents OAuth consumer credentials
type OAuthConsumer struct {
	ConsumerKey    string `json:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret"`
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