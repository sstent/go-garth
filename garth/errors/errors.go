package errors

import "fmt"

// AuthenticationError represents authentication failures
type AuthenticationError struct {
	Message string
	Cause   error
}

func (e *AuthenticationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("authentication error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("authentication error: %s", e.Message)
}

// OAuthError represents OAuth token-related errors
type OAuthError struct {
	Message string
	Cause   error
}

func (e *OAuthError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("OAuth error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("OAuth error: %s", e.Message)
}

// APIError represents errors from API calls
type APIError struct {
	StatusCode int
	Response   string
	Cause      error
}

func (e *APIError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("API error (status %d): %s: %v", e.StatusCode, e.Response, e.Cause)
	}
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Response)
}

// IOError represents file I/O errors
type IOError struct {
	Message string
	Cause   error
}

func (e *IOError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("I/O error: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("I/O error: %s", e.Message)
}

// ValidationError represents input validation failures
type ValidationError struct {
	Message string
	Field   string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error for %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}
