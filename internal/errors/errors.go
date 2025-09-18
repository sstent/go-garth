package errors

import "fmt"

// GarthError represents the base error type for all custom errors in Garth
type GarthError struct {
	Message string
	Cause   error
}

func (e *GarthError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("garth error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("garth error: %s", e.Message)
}

// GarthHTTPError represents HTTP-related errors in API calls
type GarthHTTPError struct {
	GarthError
	StatusCode int
	Response   string
}

func (e *GarthHTTPError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("HTTP error (%d): %s: %v", e.StatusCode, e.Response, e.Cause)
	}
	return fmt.Sprintf("HTTP error (%d): %s", e.StatusCode, e.Response)
}

// AuthenticationError represents authentication failures
type AuthenticationError struct {
	GarthError
}

func (e *AuthenticationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("authentication error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("authentication error: %s", e.Message)
}

// OAuthError represents OAuth token-related errors
type OAuthError struct {
	GarthError
}

func (e *OAuthError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("OAuth error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("OAuth error: %s", e.Message)
}

// APIError represents errors from API calls
type APIError struct {
	GarthHTTPError
}

// IOError represents file I/O errors
type IOError struct {
	GarthError
}

func (e *IOError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("I/O error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("I/O error: %s", e.Message)
}

// ValidationError represents input validation failures
type ValidationError struct {
	GarthError
	Field string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error for %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}
