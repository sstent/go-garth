// Package garth provides a Go client for the Garmin Connect API.
//
// This client supports authentication, user profile management, activity tracking,
// workout management, and other Garmin Connect services.
//
// Features:
// - OAuth 2.0 authentication with MFA support
// - User profile operations (retrieve, update, delete)
// - Activity management (create, read, update, delete)
// - Workout management (CRUD operations, scheduling, templates)
// - Comprehensive error handling
// - Automatic token refresh
//
// Usage:
//  1. Create an Authenticator instance with your credentials
//  2. Obtain an access token
//  3. Create an APIClient using the authenticator
//  4. Use service methods to interact with Garmin Connect API
//
// Example:
//
//	opts := garth.NewClientOptionsFromEnv()
//	auth := garth.NewBasicAuthenticator(opts)
//	token, err := auth.Login(ctx, "username", "password", "")
//	client := garth.NewAPIClient(auth)
//
//	// Get user profile
//	profile, err := client.Profile().Get(ctx)
//
// For more details, see the documentation for each service.
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
