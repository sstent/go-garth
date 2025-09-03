package garth

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"net/http/cookiejar" // Add cookiejar import
)

// GarthAuthenticator implements the Authenticator interface
type GarthAuthenticator struct {
	client      *http.Client
	tokenURL    string
	storage     TokenStorage
	userAgent   string
	csrfToken   string
	oauth1Token string // Add OAuth1 token storage
}

// NewAuthenticator creates a new Garth authentication client
func NewAuthenticator(opts ClientOptions) Authenticator {
	// Create HTTP client with browser-like settings
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		Proxy: http.ProxyFromEnvironment,
	}

	// Create cookie jar for session persistence
	jar, err := cookiejar.New(nil)
	if err != nil {
		// Fallback to no cookie jar if creation fails
		jar = nil
	}

	client := &http.Client{
		Timeout:   opts.Timeout,
		Transport: transport,
		Jar:       jar, // Add cookie jar
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Allow up to 10 redirects
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		},
	}

	auth := &GarthAuthenticator{
		client:    client,
		tokenURL:  opts.TokenURL,
		storage:   opts.Storage,
		userAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}

	// Set authenticator reference in storage if needed
	if setter, ok := opts.Storage.(AuthenticatorSetter); ok {
		setter.SetAuthenticator(auth)
	}

	return auth
}

// Login authenticates with Garmin services
func (a *GarthAuthenticator) Login(ctx context.Context, username, password, mfaToken string) (*Token, error) {
	// Step 1: Get OAuth1 token FIRST
	oauth1Token, err := a.fetchOAuth1Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth1 token: %w", err)
	}
	a.oauth1Token = oauth1Token

	// Step 2: Now get login parameters with OAuth1 context
	authToken, tokenType, err := a.fetchLoginParamsWithOAuth1(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get login params: %w", err)
	}

	a.csrfToken = authToken // Store for session

	// Step 3: Authenticate with all tokens
	token, err := a.authenticate(ctx, username, password, mfaToken, authToken, tokenType)
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

	// Persist the refreshed token to storage
	if err := a.storage.SaveToken(&token); err != nil {
		return nil, fmt.Errorf("failed to save refreshed token: %w", err)
	}

	return &token, nil
}

// GetClient returns an authenticated HTTP client
func (a *GarthAuthenticator) GetClient() *http.Client {
	// This would be a client with middleware that automatically
	// adds authentication headers and handles token refresh
	return a.client
}

// fetchLoginParamsWithOAuth1 retrieves login parameters with OAuth1 context
func (a *GarthAuthenticator) fetchLoginParamsWithOAuth1(ctx context.Context) (token string, tokenType string, err error) {
	// Build login URL with OAuth1 context
	params := url.Values{}
	params.Set("id", "gauth-widget")
	params.Set("embedWidget", "true")
	params.Set("gauthHost", "https://sso.garmin.com/sso")
	params.Set("service", "https://connect.garmin.com/oauthConfirm")
	params.Set("source", "https://sso.garmin.com/sso")
	params.Set("redirectAfterAccountLoginUrl", "https://connect.garmin.com/oauthConfirm")
	params.Set("redirectAfterAccountCreationUrl", "https://connect.garmin.com/oauthConfirm")
	params.Set("consumeServiceTicket", "false")
	params.Set("generateExtraServiceTicket", "true")
	params.Set("clientId", "GarminConnect")
	params.Set("locale", "en_US")

	// Add OAuth1 token if we have it
	if a.oauth1Token != "" {
		params.Set("oauth_token", a.oauth1Token)
	}

	loginURL := "https://sso.garmin.com/sso/signin?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", loginURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create login page request: %w", err)
	}

	// Set headers with proper referrer chain
	req.Header.Set("User-Agent", a.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Referer", "https://connect.garmin.com/oauthConfirm")
	req.Header.Set("Origin", "https://connect.garmin.com")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("login page request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read login page response: %w", err)
	}

	bodyStr := string(body)

	// Extract CSRF/lt token
	token, tokenType, err = getCSRFTokenWithType(bodyStr)
	if err != nil {
		// Save for debugging
		filename := fmt.Sprintf("login_page_oauth1_%d.html", time.Now().Unix())
		if writeErr := os.WriteFile(filename, body, 0644); writeErr == nil {
			return "", "", fmt.Errorf("authentication token not found with OAuth1 context: %w (HTML saved to %s)", err, filename)
		}
		return "", "", fmt.Errorf("authentication token not found with OAuth1 context: %w", err)
	}

	return token, tokenType, nil
}

// buildLoginURL constructs the complete login URL with parameters
func (a *GarthAuthenticator) buildLoginURL() string {
	// Match Python implementation exactly (order and values)
	params := url.Values{}
	params.Set("id", "gauth-widget")
	params.Set("embedWidget", "true")
	params.Set("gauthHost", "https://sso.garmin.com/sso")
	params.Set("service", "https://connect.garmin.com")
	params.Set("source", "https://sso.garmin.com/sso")
	params.Set("redirectAfterAccountLoginUrl", "https://connect.garmin.com/oauthConfirm")
	params.Set("redirectAfterAccountCreationUrl", "https://connect.garmin.com/oauthConfirm")
	params.Set("consumeServiceTicket", "false")
	params.Set("generateExtraServiceTicket", "true")
	params.Set("clientId", "GarminConnect")
	params.Set("locale", "en_US")

	return "https://sso.garmin.com/sso/signin?" + params.Encode()
}

// fetchOAuth1Token retrieves initial OAuth1 token for session
func (a *GarthAuthenticator) fetchOAuth1Token(ctx context.Context) (string, error) {
	// Step 1: Initial OAuth1 request - this should NOT have parameters initially
	oauth1URL := "https://connect.garmin.com/oauthConfirm"

	req, err := http.NewRequestWithContext(ctx, "GET", oauth1URL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create OAuth1 request: %w", err)
	}

	// Set proper headers
	req.Header.Set("User-Agent", a.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("OAuth1 request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle redirect case - OAuth1 token often comes from redirect location
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location := resp.Header.Get("Location")
		if location != "" {
			if u, err := url.Parse(location); err == nil {
				if token := u.Query().Get("oauth_token"); token != "" {
					return token, nil
				}
			}
		}
	}

	// If no redirect, parse response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read OAuth1 response: %w", err)
	}

	// Look for oauth_token in various formats
	patterns := []string{
		`oauth_token=([^&\s"']+)`,
		`"oauth_token":\s*"([^"]+)"`,
		`'oauth_token':\s*'([^']+)'`,
		`oauth_token["']?\s*[:=]\s*["']?([^"'\s&]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(string(body)); len(matches) > 1 {
			return matches[1], nil
		}
	}

	// Debug: save response to project debug directory
	debugDir := "go-garth/debug"
	if err := os.MkdirAll(debugDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create debug directory: %w", err)
	}

	filename := filepath.Join(debugDir, fmt.Sprintf("oauth1_response_%d.html", time.Now().Unix()))
	absPath, _ := filepath.Abs(filename)
	if writeErr := os.WriteFile(filename, body, 0644); writeErr == nil {
		return "", fmt.Errorf("OAuth1 token not found (response saved to %s)", absPath)
	}

	return "", fmt.Errorf("OAuth1 token not found in response (failed to save debug file)")
}

// authenticate performs the authentication flow
func (a *GarthAuthenticator) authenticate(ctx context.Context, username, password, mfaToken, authToken, tokenType string) (*Token, error) {
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("embed", "true")
	data.Set("rememberme", "on")

	// Set the correct token field based on type
	if tokenType == "lt" {
		data.Set("lt", authToken)
	} else {
		data.Set("_csrf", authToken)
	}

	data.Set("_eventId", "submit")
	data.Set("geolocation", "")
	data.Set("clientId", "GarminConnect")
	data.Set("service", "https://connect.garmin.com/oauthConfirm") // Updated service URL
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
		body, _ := io.ReadAll(resp.Body)
		return a.handleMFA(ctx, username, password, mfaToken, string(body))
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("authentication failed with status: %d, response: %s", resp.StatusCode, body)
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

// getCSRFToken extracts the CSRF token from HTML using multiple patterns
func getCSRFToken(html string) (string, error) {
	token, _, err := getCSRFTokenWithType(html)
	return token, err
}

// getCSRFTokenWithType returns token and its type (lt or _csrf)
func getCSRFTokenWithType(html string) (string, string, error) {
	// Check lt patterns first
	ltPatterns := []string{
		`name="lt"\s+value="([^"]+)"`,
		`name="lt"\s+type="hidden"\s+value="([^"]+)"`,
		`<input[^>]*name="lt"[^>]*value="([^"]+)"[^>]*>`,
	}

	for _, pattern := range ltPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			return matches[1], "lt", nil
		}
	}

	// Check CSRF patterns
	csrfPatterns := []string{
		`name="_csrf"\s+value="([^"]+)"`,
		`"csrfToken":"([^"]+)"`,
	}

	for _, pattern := range csrfPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			return matches[1], "_csrf", nil
		}
	}

	return "", "", errors.New("no authentication token found")
}

// extractFromJSON tries to find the CSRF token in a JSON structure
func extractFromJSON(html string) (string, error) {
	// Pattern to find the JSON config in script tags
	re := regexp.MustCompile(`window\.__INITIAL_CONFIG__ = (\{.*?\});`)
	matches := re.FindStringSubmatch(html)
	if len(matches) < 2 {
		return "", errors.New("JSON config not found")
	}

	// Parse the JSON
	var config struct {
		CSRFToken string `json:"csrfToken"`
	}
	if err := json.Unmarshal([]byte(matches[1]), &config); err != nil {
		return "", fmt.Errorf("failed to parse JSON config: %w", err)
	}

	if config.CSRFToken == "" {
		return "", errors.New("csrfToken not found in JSON config")
	}

	return config.CSRFToken, nil
}

// getEnhancedBrowserHeaders returns browser-like headers including Referer and Origin
func (a *GarthAuthenticator) getEnhancedBrowserHeaders(referrer string) http.Header {
	u, _ := url.Parse(referrer)
	origin := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	return http.Header{
		"User-Agent":                {a.userAgent},
		"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"Accept-Language":           {"en-US,en;q=0.9"},
		"Accept-Encoding":           {"gzip, deflate, br"},
		"Connection":                {"keep-alive"},
		"Cache-Control":             {"max-age=0"},
		"Origin":                    {origin},
		"Referer":                   {referrer},
		"Sec-Fetch-Site":            {"same-origin"},
		"Sec-Fetch-Mode":            {"navigate"},
		"Sec-Fetch-User":            {"?1"},
		"Sec-Fetch-Dest":            {"document"},
		"DNT":                       {"1"},
		"Upgrade-Insecure-Requests": {"1"},
	}
}
