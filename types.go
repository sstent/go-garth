package garth

import (
	"errors"
	"fmt"
	"time"
)

// Unified Token Definitions

// OAuth1Token represents OAuth1 credentials
type OAuth1Token struct {
	Token  string
	Secret string
}

// OAuth2Token represents OAuth2 credentials
type OAuth2Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	ExpiresAt    int64  `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// IsExpired checks if the token has expired
func (t *OAuth2Token) IsExpired() bool {
	return time.Now().After(time.Unix(t.ExpiresAt, 0))
}

// Token represents unified authentication credentials
type Token struct {
	Domain      string       `json:"domain"`
	OAuth1Token *OAuth1Token `json:"oauth1_token"`
	OAuth2Token *OAuth2Token `json:"oauth2_token"`
	UserProfile *UserProfile `json:"user_profile"`
}

// IsExpired checks if the OAuth2 token has expired
func (t *Token) IsExpired() bool {
	if t.OAuth2Token == nil {
		return true
	}
	return t.OAuth2Token.IsExpired()
}

// NeedsRefresh checks if token needs refresh (within 5 min expiry window)
func (t *Token) NeedsRefresh() bool {
	if t.OAuth2Token == nil {
		return true
	}
	return time.Now().Add(5 * time.Minute).After(time.Unix(t.OAuth2Token.ExpiresAt, 0))
}

// UserProfile represents Garmin user profile information
type UserProfile struct {
	Username    string `json:"username"`
	ProfileID   string `json:"profile_id"`
	DisplayName string `json:"display_name"`
}

// ClientOptions contains configuration for the authenticator
type ClientOptions struct {
	SSOURL    string        // SSO endpoint
	TokenURL  string        // Token exchange endpoint
	Storage   TokenStorage  // Token storage implementation
	Timeout   time.Duration // HTTP client timeout
	Domain    string        // Garmin domain (default: garmin.com)
	UserAgent string        // User-Agent header (default: GCMv3)
}

// TokenStorage defines the interface for token storage
type TokenStorage interface {
	StoreToken(token *Token) error
	GetToken() (*Token, error)
	ClearToken() error
}

// Error interface defines common error behavior for Garth
type Error interface {
	error
	GetStatusCode() int
	GetType() string
	GetCause() error
	Unwrap() error
}

// ErrTokenNotFound is returned when a token is not available in storage
var ErrTokenNotFound = errors.New("token not found")

// AuthError represents Garmin authentication errors
type AuthError struct {
	StatusCode int    `json:"status_code"` // HTTP status code
	Message    string `json:"message"`     // Human-readable error message
	Type       string `json:"type"`        // Garmin error type identifier
	Cause      error  `json:"cause"`       // Underlying error
	CSRF       string `json:"csrf"`        // CSRF token for MFA flow
}

// GetStatusCode returns the HTTP status code
func (e *AuthError) GetStatusCode() int {
	return e.StatusCode
}

// GetType returns the error category
func (e *AuthError) GetType() string {
	return e.Type
}

// Error implements the error interface for AuthError
func (e *AuthError) Error() string {
	msg := fmt.Sprintf("garmin auth error %d: %s", e.StatusCode, e.Message)
	if e.Cause != nil {
		msg += " (" + e.Cause.Error() + ")"
	}
	return msg
}

// Unwrap returns the underlying error
func (e *AuthError) Unwrap() error {
	return e.Cause
}

// GetCause returns the underlying error (implements Error interface)
func (e *AuthError) GetCause() error {
	return e.Cause
}

// APIError represents errors from API operations
type APIError struct {
	StatusCode int    // HTTP status code
	Message    string // Error description
	Cause      error  // Underlying error
	ErrorType  string // Specific error category
}

// GetStatusCode returns the HTTP status code
func (e *APIError) GetStatusCode() int {
	return e.StatusCode
}

// GetType returns the error category
func (e *APIError) GetType() string {
	if e.ErrorType == "" {
		return "api_error"
	}
	return e.ErrorType
}

// Error implements the error interface for APIError
func (e *APIError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("API error (%d): %s: %v", e.StatusCode, e.Message, e.Cause)
	}
	return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Message)
}

// Unwrap returns the underlying error
func (e *APIError) Unwrap() error {
	return e.Cause
}

// GetCause returns the underlying error (implements Error interface)
func (e *APIError) GetCause() error {
	return e.Cause
}

// AuthenticatorSetter interface for storage that needs authenticator reference
type AuthenticatorSetter interface {
	SetAuthenticator(a Authenticator)
}
