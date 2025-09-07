package client

import (
	"time"
)

// OAuth1Token represents OAuth 1.0a credentials
type OAuth1Token struct {
	Token       string
	TokenSecret string
	CreatedAt   time.Time
}

// Expired checks if token is expired (OAuth1 tokens typically don't expire but we'll implement for consistency)
func (t *OAuth1Token) Expired() bool {
	return false // OAuth1 tokens don't typically expire
}

// OAuth2Token represents OAuth 2.0 credentials
type OAuth2Token struct {
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresIn    int
	ExpiresAt    time.Time
}

// Expired checks if token is expired
func (t *OAuth2Token) Expired() bool {
	return time.Now().After(t.ExpiresAt)
}

// RefreshIfNeeded refreshes token if expired (implementation pending)
func (t *OAuth2Token) RefreshIfNeeded(client *Client) error {
	// Placeholder for token refresh logic
	return nil
}
