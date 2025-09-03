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
	"regexp"
	"strings"
	"time"

	"net/http/cookiejar" // Add cookiejar import
)

// GarthAuthenticator implements the Authenticator interface
type GarthAuthenticator struct {
	client    *http.Client
	tokenURL  string
	storage   TokenStorage
	userAgent string
	// Add CSRF token for session management
	csrfToken string
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
	// Fetch OAuth1 token to initialize session
	if _, err := a.fetchOAuth1Token(ctx); err != nil {
		return nil, fmt.Errorf("failed to get OAuth1 token: %w", err)
	}

	// Get login parameters including CSRF token
	authToken, tokenType, err := a.fetchLoginParams(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get login params: %w", err)
	}

	a.csrfToken = authToken // Store for session

	// Call authenticate with the extracted token and its type
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

// fetchLoginParams retrieves required tokens from Garmin login page and returns token + type
func (a *GarthAuthenticator) fetchLoginParams(ctx context.Context) (token string, tokenType string, err error) {
	// Step 1: Set cookies by accessing the embed endpoint
	embedURL := "https://sso.garmin.com/sso/embed?" + url.Values{
		"id":          []string{"gauth-widget"},
		"embedWidget": []string{"true"},
		"gauthHost":   []string{"https://sso.garmin.com/sso"},
	}.Encode()

	embedReq, err := http.NewRequestWithContext(ctx, "GET", embedURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create embed request: %w", err)
	}
	embedReq.Header = a.getEnhancedBrowserHeaders(embedURL)

	_, err = a.client.Do(embedReq)
	if err != nil {
		return "", "", fmt.Errorf("embed request failed: %w", err)
	}

	// Step 2: Get login parameters including CSRF token
	loginURL := a.buildLoginURL()
	req, err := http.NewRequestWithContext(ctx, "GET", loginURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create login page request: %w", err)
	}

	req.Header = a.getEnhancedBrowserHeaders(loginURL)

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

	// Use our robust CSRF token extractor with multiple patterns
	token, err = getCSRFToken(bodyStr)
	if err != nil {
		filename := fmt.Sprintf("login_page_%d.html", time.Now().Unix())
		if writeErr := os.WriteFile(filename, body, 0644); writeErr == nil {
			return "", "", fmt.Errorf("authentication token not found: %w (HTML saved to %s)", err, filename)
		}
		return "", "", fmt.Errorf("authentication token not found: %w (failed to save HTML for debugging)", err)
	}

	// Determine token type (lt or _csrf)
	if strings.Contains(token, "LT-") {
		return token, "lt", nil
	}
	return token, "_csrf", nil
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
	oauth1URL := "https://connect.garmin.com/oauthConfirm"

	req, err := http.NewRequestWithContext(ctx, "GET", oauth1URL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create OAuth1 request: %w", err)
	}

	req.Header.Set("User-Agent", a.userAgent)

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("OAuth1 request failed: %w", err)
	}
	defer resp.Body.Close()

	// Extract oauth_token from Location header or response body
	if location := resp.Header.Get("Location"); location != "" {
		if u, err := url.Parse(location); err == nil {
			if token := u.Query().Get("oauth_token"); token != "" {
				return token, nil
			}
		}
	}

	// Or extract from HTML response
	body, _ := io.ReadAll(resp.Body)
	tokenPattern := regexp.MustCompile(`oauth_token=([^&\s"]+)`)
	matches := tokenPattern.FindStringSubmatch(string(body))
	if len(matches) > 1 {
		return matches[1], nil
	}

	return "", fmt.Errorf("OAuth1 token not found")
}

// authenticate performs the authentication flow
func (a *GarthAuthenticator) authenticate(ctx context.Context, username, password, mfaToken, authToken, tokenType string) (*Token, error) {
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("embed", "true")
	data.Set("rememberme", "on")
	// Use correct token field based on token type
	if tokenType == "lt" {
		data.Set("lt", authToken)
	} else {
		data.Set("_csrf", authToken)
	}
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
	// Priority order based on what Garmin actually uses
	patterns := []string{
		// 1. Login Ticket (lt) - PRIMARY pattern used by Garmin
		`name="lt"\s+value="([^"]+)"`,
		`name="lt"\s+type="hidden"\s+value="([^"]+)"`,
		`<input[^>]*name="lt"[^>]*value="([^"]+)"[^>]*>`,

		// 2. CSRF tokens (backup patterns)
		`name="_csrf"\s+value="([^"]+)"`,
		`"csrfToken":"([^"]+)"`,

		// 3. Other possible patterns
		`name=["']_csrf["']\s+value=["']([^"']+)["']`,
		`value=["']([^"']+)["']\s+name=["']_csrf["']`,
		`id="__csrf"\s+value="([^"]+)"`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", errors.New("no authentication token found")
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
