package garth

import (
	"context"
	"net/http"
	"time"
)

// Authenticator defines the authentication interface
type Authenticator interface {
	// Login authenticates with Garmin services
	Login(ctx context.Context, username, password, mfaToken string) (*Token, error)

	// RefreshToken refreshes an expired access token
	RefreshToken(ctx context.Context, refreshToken string) (*Token, error)

	// GetClient returns an authenticated HTTP client
	GetClient() *http.Client
}

// ClientOptions configures the Authenticator
type ClientOptions struct {
	TokenURL string        // Token exchange endpoint
	Storage  TokenStorage  // Token storage implementation
	Timeout  time.Duration // HTTP client timeout
}
