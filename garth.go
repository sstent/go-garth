package garth

import (
	"context"
	"net/http"
	"os"
	"strconv"
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

// NewClientOptionsFromEnv creates ClientOptions from environment variables
func NewClientOptionsFromEnv() ClientOptions {
	// Default configuration
	opts := ClientOptions{
		TokenURL: "https://connectapi.garmin.com/oauth-service/oauth/token",
		Timeout:  30 * time.Second,
	}

	// Override from environment variables
	if url := os.Getenv("GARTH_TOKEN_URL"); url != "" {
		opts.TokenURL = url
	}

	if timeoutStr := os.Getenv("GARTH_TIMEOUT"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			opts.Timeout = time.Duration(timeout) * time.Second
		}
	}

	// Default to memory storage
	opts.Storage = NewMemoryStorage()

	return opts
}
