package garth

import (
	"errors"
	"fmt"
	"time"
)

// TokenStorage defines the interface for token persistence
type TokenStorage interface {
	// GetToken retrieves the stored token
	GetToken() (*Token, error)

	// SaveToken stores a new token
	SaveToken(token *Token) error
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

// Token represents OAuth 2.0 tokens
type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type,omitempty"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in,omitempty"` // Duration in seconds
	Expiry       time.Time `json:"expiry"`               // Absolute time of expiration
}

// IsExpired checks if the token has expired
func (t *Token) IsExpired() bool {
	return time.Now().After(t.Expiry)
}

// AuthError represents Garmin authentication errors
type AuthError struct {
	StatusCode int    `json:"status_code"` // HTTP status code
	Message    string `json:"message"`     // Human-readable error message
	Type       string `json:"type"`        // Garmin error type identifier
	Cause      error  `json:"cause"`       // Underlying error
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
