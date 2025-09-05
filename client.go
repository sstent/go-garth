package garth

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// AuthTransport implements http.RoundTripper to inject authentication headers
type AuthTransport struct {
	base      http.RoundTripper
	auth      *GarthAuthenticator
	storage   TokenStorage
	userAgent string
	mutex     sync.Mutex // Protects refreshing token
}

// NewAuthTransport creates a new authenticated transport with specified storage
func NewAuthTransport(auth *GarthAuthenticator, storage TokenStorage, base http.RoundTripper) *AuthTransport {
	if base == nil {
		base = http.DefaultTransport
	}

	if storage == nil {
		storage = NewFileStorage("garmin_session.json")
	}

	return &AuthTransport{
		base:      base,
		auth:      auth,
		storage:   storage,
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	}
}

// RoundTrip executes a single HTTP transaction with authentication
func (t *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone request to avoid modifying the original
	req = cloneRequest(req)

	// Get current token
	token, err := t.storage.GetToken()
	if err != nil {
		return nil, &AuthError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Token not available",
			Cause:      err,
		}
	}

	// Refresh token if expired
	if token.IsExpired() {
		newToken, err := t.refreshToken(req.Context(), token)
		if err != nil {
			return nil, err
		}
		token = newToken
	}

	// Add Authorization header
	req.Header.Set("Authorization", "Bearer "+token.OAuth2Token.AccessToken)
	req.Header.Set("User-Agent", t.userAgent)
	req.Header.Set("Referer", "https://sso.garmin.com/sso/signin")

	// Execute request with retry logic
	var resp *http.Response
	maxRetries := 3
	backoff := 200 * time.Millisecond // Initial backoff duration

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err = t.base.RoundTrip(req)
		if err != nil {
			// Network error, retry with backoff
			time.Sleep(backoff)
			backoff *= 2 // Exponential backoff
			continue
		}

		// Handle token expiration during request (e.g. token revoked)
		if resp.StatusCode == http.StatusUnauthorized {
			resp.Body.Close()
			// Refresh token and update request
			token, err = t.refreshToken(req.Context(), token)
			if err != nil {
				return nil, err
			}
			req.Header.Set("Authorization", "Bearer "+token.OAuth2Token.AccessToken)
			continue
		}

		// Retry server errors (5xx) and rate limits (429)
		if resp.StatusCode >= 500 && resp.StatusCode < 600 || resp.StatusCode == http.StatusTooManyRequests {
			resp.Body.Close()
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		// Successful response
		return resp, nil
	}

	// Return last error or response if max retries exceeded
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// refreshToken handles token refresh with mutex protection
func (t *AuthTransport) refreshToken(ctx context.Context, token *Token) (*Token, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Check again in case another goroutine refreshed while waiting
	currentToken, err := t.storage.GetToken()
	if err != nil {
		return nil, err
	}
	if !currentToken.IsExpired() {
		return currentToken, nil
	}

	// Perform refresh
	newToken, err := t.auth.RefreshToken(ctx, token.OAuth2Token.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Save new token
	if err := t.storage.StoreToken(newToken); err != nil {
		return nil, err
	}

	return newToken, nil
}

// NewDefaultAuthTransport creates a transport with persistent storage
func NewDefaultAuthTransport(auth *GarthAuthenticator) *AuthTransport {
	return NewAuthTransport(auth, NewFileStorage("garmin_session.json"), nil)
}

// NewMemoryAuthTransport creates a transport with in-memory storage (for testing)
func NewMemoryAuthTransport(auth *GarthAuthenticator) *AuthTransport {
	return NewAuthTransport(auth, NewMemoryStorage(), nil)
}

// cloneRequest returns a clone of the provided HTTP request
func cloneRequest(r *http.Request) *http.Request {
	// Shallow copy of the struct
	clone := *r
	// Deep copy of the headers
	clone.Header = make(http.Header, len(r.Header))
	for k, v := range r.Header {
		clone.Header[k] = v
	}
	return &clone
}
