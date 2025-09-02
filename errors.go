package garth

import (
	"net/http"
	"net/url"
)

// RequestError captures request context for enhanced error reporting
type RequestError struct {
	Method     string
	URL        string
	Query      url.Values
	StatusCode int
	Cause      error
}

// NewRequestError creates a new RequestError instance
func NewRequestError(method, url string, query url.Values, statusCode int, cause error) *RequestError {
	return &RequestError{
		Method:     method,
		URL:        url,
		Query:      query,
		StatusCode: statusCode,
		Cause:      cause,
	}
}

// Error implements the error interface for RequestError
func (e *RequestError) Error() string {
	return e.Cause.Error()
}

// GetStatusCode returns the HTTP status code
func (e *RequestError) GetStatusCode() int {
	return e.StatusCode
}

// GetType returns the error category
func (e *RequestError) GetType() string {
	return "request_error"
}

// GetCause returns the underlying error
func (e *RequestError) GetCause() error {
	return e.Cause
}

// Unwrap returns the underlying error
func (e *RequestError) Unwrap() error {
	return e.Cause
}

// RequestContext returns the request context details
func (e *RequestError) RequestContext() (method, url string, query url.Values) {
	return e.Method, e.URL, e.Query
}

// Helper functions for common error checks

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	if e, ok := err.(Error); ok {
		return e.GetStatusCode() == http.StatusNotFound
	}
	return false
}

// IsAuthenticationError checks if an error is an authentication error
func IsAuthenticationError(err error) bool {
	if e, ok := err.(Error); ok {
		return e.GetStatusCode() == http.StatusUnauthorized
	}
	return false
}

// IsRateLimitError checks if an error is a rate limit error
func IsRateLimitError(err error) bool {
	if e, ok := err.(Error); ok {
		return e.GetStatusCode() == http.StatusTooManyRequests
	}
	return false
}
