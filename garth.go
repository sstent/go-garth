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
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Authenticator defines the authentication interface
type Authenticator interface {
	// Login authenticates with Garmin services using OAuth1/OAuth2 hybrid flow
	Login(ctx context.Context, username, password, mfaToken string) (*Token, error)

	// RefreshToken refreshes an expired OAuth2 access token
	RefreshToken(ctx context.Context, refreshToken string) (*Token, error)

	// ExchangeToken exchanges OAuth1 token for OAuth2 token
	ExchangeToken(ctx context.Context, oauth1Token *OAuth1Token) (*Token, error)

	// GetClient returns an authenticated HTTP client
	GetClient() *http.Client
}

// ClientOptions configures the Authenticator
type ClientOptions struct {
	SSOURL    string        // SSO endpoint
	TokenURL  string        // Token exchange endpoint
	Storage   TokenStorage  // Token storage implementation
	Timeout   time.Duration // HTTP client timeout
	Domain    string        // Garmin domain (default: garmin.com)
	UserAgent string        // User-Agent header (default: GCMv3)
}

// NewClientOptionsFromEnv creates ClientOptions from environment variables
func NewClientOptionsFromEnv() ClientOptions {
	// Default configuration
	opts := ClientOptions{
		SSOURL:    "https://sso.garmin.com/sso",
		TokenURL:  "https://connectapi.garmin.com/oauth-service",
		Domain:    "garmin.com",
		UserAgent: "GCMv3",
		Timeout:   30 * time.Second,
	}

	// Override from environment variables
	if url := os.Getenv("GARTH_SSO_URL"); url != "" {
		opts.SSOURL = url
	}
	if url := os.Getenv("GARTH_TOKEN_URL"); url != "" {
		opts.TokenURL = url
	}
	if domain := os.Getenv("GARTH_DOMAIN"); domain != "" {
		opts.Domain = domain
	}
	if ua := os.Getenv("GARTH_USER_AGENT"); ua != "" {
		opts.UserAgent = ua
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

// getRequestToken retrieves OAuth1 request token
// getRequestToken retrieves OAuth1 request token
func getRequestToken(a Authenticator, ctx context.Context) (*OAuth1Token, error) {
	// Implementation will be added in next step
	return nil, errors.New("not implemented")
}

// authorizeRequestToken authorizes OAuth1 token through SSO
func authorizeRequestToken(a Authenticator, ctx context.Context, token *OAuth1Token) error {
	// Implementation will be added in next step
	return errors.New("not implemented")
}
