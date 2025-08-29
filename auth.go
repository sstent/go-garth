package garth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// GarthAuthenticator implements the Authenticator interface
type GarthAuthenticator struct {
	client    *http.Client
	tokenURL  string
	storage   TokenStorage
	userAgent string
}

// NewAuthenticator creates a new Garth authentication client
func NewAuthenticator(opts ClientOptions) Authenticator {
	// Create HTTP client with reasonable defaults
	client := &http.Client{
		Timeout: opts.Timeout,
	}

	return &GarthAuthenticator{
		client:    client,
		tokenURL:  opts.TokenURL,
		storage:   opts.Storage,
		userAgent: "GarthAuthenticator/1.0",
	}
}

// Login authenticates with Garmin services
func (a *GarthAuthenticator) Login(ctx context.Context, username, password, mfaToken string) (*Token, error) {
	lt, execution, err := a.fetchLoginParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get login params: %w", err)
	}

	token, err := a.authenticate(ctx, username, password, mfaToken, lt, execution)
	if err != nil {
		return nil, err
	}

	// Save token to storage
	if err := a.storage.SaveToken(token); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	return token, nil
}

// RefreshToken refreshes an expired access token
func (a *GarthAuthenticator) RefreshToken(ctx context.Context, refreshToken string) (*Token, error) {
	if refreshToken == "" {
		return nil, &AuthError{
			StatusCode: http.StatusBadRequest,
			Message:    "Refresh token is required",
			Type:       "invalid_request",
		}
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, "POST", a.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, &AuthError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to create refresh request",
			Cause:      err,
		}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", a.userAgent)
	req.SetBasicAuth("garmin-connect", "garmin-connect-secret")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, &AuthError{
			StatusCode: http.StatusBadGateway,
			Message:    "Refresh request failed",
			Cause:      err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &AuthError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("Token refresh failed: %s", body),
			Type:       "token_refresh_failure",
		}
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, &AuthError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse token response",
			Cause:      err,
		}
	}

	token.Expiry = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	return &token, nil
}

// GetClient returns an authenticated HTTP client
func (a *GarthAuthenticator) GetClient() *http.Client {
	// This would be a client with middleware that automatically
	// adds authentication headers and handles token refresh
	return a.client
}

// fetchLoginParams retrieves required tokens from Garmin login page
func (a *GarthAuthenticator) fetchLoginParams(ctx context.Context) (lt, execution string, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://sso.garmin.com/sso/signin?service=https://connect.garmin.com", nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create login page request: %w", err)
	}

	req.Header = a.getBrowserHeaders()

	resp, err := a.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("login page request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read login page response: %w", err)
	}

	lt, err = extractParam(`name="lt"\s+value="([^"]+)"`, string(body))
	if err != nil {
		return "", "", fmt.Errorf("lt param not found: %w", err)
	}

	execution, err = extractParam(`name="execution"\s+value="([^"]+)"`, string(body))
	if err != nil {
		return "", "", fmt.Errorf("execution param not found: %w", err)
	}

	return lt, execution, nil
}

// authenticate performs the authentication flow
func (a *GarthAuthenticator) authenticate(ctx context.Context, username, password, mfaToken, lt, execution string) (*Token, error) {
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("embed", "true")
	data.Set("rememberme", "on")
	data.Set("lt", lt)
	data.Set("execution", execution)
	data.Set("_eventId", "submit")
	data.Set("geolocation", "")
	data.Set("clientId", "GarminConnect")
	data.Set("service", "https://connect.garmin.com")
	data.Set("webhost", "https://connect.garmin.com")
	data.Set("fromPage", "oauth")
	data.Set("locale", "en_US")
	data.Set("id", "gauth-widget")
	data.Set("redirectAfterAccountLoginUrl", "https://connect.garmin.com/oauthConfirm")
	data.Set("redirectAfterAccountCreationUrl", "https://connect.garmin.com/oauthConfirm")

	loginURL := "https://sso.garmin.com/sso/signin"
	req, err := http.NewRequestWithContext(ctx, "POST", loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create SSO request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", a.userAgent)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SSO request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusPreconditionFailed {
		return a.handleMFA(ctx, username, password, mfaToken, "")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentication failed with status: %d", resp.StatusCode)
	}

	var authResponse struct {
		Ticket string `json:"ticket"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return nil, fmt.Errorf("failed to parse SSO response: %w", err)
	}

	if authResponse.Ticket == "" {
		return nil, errors.New("empty ticket in SSO response")
	}

	return a.exchangeTicketForToken(ctx, authResponse.Ticket)
}

// exchangeTicketForToken exchanges an SSO ticket for an access token
func (a *GarthAuthenticator) exchangeTicketForToken(ctx context.Context, ticket string) (*Token, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", ticket)
	data.Set("redirect_uri", "https://connect.garmin.com")

	req, err := http.NewRequestWithContext(ctx, "POST", a.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", a.userAgent)
	req.SetBasicAuth("garmin-connect", "garmin-connect-secret")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %d %s", resp.StatusCode, body)
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	token.Expiry = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	return &token, nil
}

// handleMFA processes multi-factor authentication
func (a *GarthAuthenticator) handleMFA(ctx context.Context, username, password, mfaToken, responseBody string) (*Token, error) {
	// Extract CSRF token from response body
	csrfToken, err := extractParam(`name="_csrf"\s+value="([^"]+)"`, responseBody)
	if err != nil {
		return nil, &AuthError{
			StatusCode: http.StatusPreconditionFailed,
			Message:    "MFA CSRF token not found",
			Cause:      err,
		}
	}

	// Prepare MFA request
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("mfaToken", mfaToken)
	data.Set("embed", "true")
	data.Set("rememberme", "on")
	data.Set("_csrf", csrfToken)
	data.Set("_eventId", "submit")
	data.Set("geolocation", "")
	data.Set("clientId", "GarminConnect")
	data.Set("service", "https://connect.garmin.com")
	data.Set("webhost", "https://connect.garmin.com")
	data.Set("fromPage", "oauth")
	data.Set("locale", "en_US")
	data.Set("id", "gauth-widget")
	data.Set("redirectAfterAccountLoginUrl", "https://connect.garmin.com/oauthConfirm")
	data.Set("redirectAfterAccountCreationUrl", "https://connect.garmin.com/oauthConfirm")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://sso.garmin.com/sso/signin", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, &AuthError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to create MFA request",
			Cause:      err,
		}
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", a.userAgent)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, &AuthError{
			StatusCode: http.StatusBadGateway,
			Message:    "MFA request failed",
			Cause:      err,
		}
	}
	defer resp.Body.Close()

	// Handle MFA response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &AuthError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("MFA failed: %s", body),
			Type:       "mfa_failure",
		}
	}

	// Parse MFA response
	var mfaResponse struct {
		Ticket string `json:"ticket"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&mfaResponse); err != nil {
		return nil, &AuthError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse MFA response",
			Cause:      err,
		}
	}

	if mfaResponse.Ticket == "" {
		return nil, &AuthError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Invalid MFA response - ticket missing",
			Type:       "invalid_mfa_response",
		}
	}

	return a.exchangeTicketForToken(ctx, mfaResponse.Ticket)
}

// extractParam helper to extract regex pattern
func extractParam(pattern, body string) (string, error) {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(body)
	if len(matches) < 2 {
		return "", fmt.Errorf("pattern not found: %s", pattern)
	}
	return matches[1], nil
}

// getBrowserHeaders returns browser-like headers for requests
func (a *GarthAuthenticator) getBrowserHeaders() http.Header {
	return http.Header{
		"User-Agent":                {a.userAgent},
		"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8"},
		"Accept-Language":           {"en-US,en;q=0.9"},
		"Accept-Encoding":           {"gzip, deflate, br"},
		"Connection":                {"keep-alive"},
		"Cache-Control":             {"max-age=0"},
		"Sec-Fetch-Site":            {"none"},
		"Sec-Fetch-Mode":            {"navigate"},
		"Sec-Fetch-User":            {"?1"},
		"Sec-Fetch-Dest":            {"document"},
		"DNT":                       {"1"},
		"Upgrade-Insecure-Requests": {"1"},
	}
}
