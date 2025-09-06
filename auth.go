package garth

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"net/http/cookiejar"
)

type GarthAuthenticator struct {
	client      *http.Client
	tokenURL    string
	storage     TokenStorage
	userAgent   string
	csrfToken   string
	oauth1Token string
	domain      string
}

func NewAuthenticator(opts ClientOptions) Authenticator {
	// Set default domain if not provided
	if opts.Domain == "" {
		opts.Domain = "garmin.com"
	}

	// Enhanced transport with better TLS settings and compression
	baseTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:               tls.VersionTLS12,
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		},
		Proxy:               http.ProxyFromEnvironment,
		DisableCompression:  false, // Enable compression
		MaxIdleConns:        10,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		jar = nil
	}

	client := &http.Client{
		Timeout:   opts.Timeout,
		Transport: baseTransport,
		Jar:       jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			// Preserve headers during redirects
			if len(via) > 0 {
				for key, values := range via[0].Header {
					if key == "Authorization" || key == "Cookie" {
						continue // Let the jar handle cookies
					}
					req.Header[key] = values
				}
			}
			if jar != nil {
				for _, v := range via {
					if v.Response != nil {
						if cookies := v.Response.Cookies(); len(cookies) > 0 {
							jar.SetCookies(req.URL, cookies)
						}
					}
				}
			}
			return nil
		},
	}

	auth := &GarthAuthenticator{
		client:    client,
		tokenURL:  opts.TokenURL,
		storage:   opts.Storage,
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36", // More realistic user agent
		domain:    opts.Domain,
	}

	if setter, ok := opts.Storage.(interface{ SetAuthenticator(a Authenticator) }); ok {
		setter.SetAuthenticator(auth)
	}

	return auth
}

// Enhanced browser headers to bypass Cloudflare
func (a *GarthAuthenticator) getRealisticBrowserHeaders(referer string) http.Header {
	headers := http.Header{
		"User-Agent":                {a.userAgent},
		"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"Accept-Language":           {"en-US,en;q=0.9"},
		"Accept-Encoding":           {"gzip, deflate, br"},
		"Cache-Control":             {"max-age=0"},
		"Sec-Ch-Ua":                 {`"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`},
		"Sec-Ch-Ua-Mobile":          {"?0"},
		"Sec-Ch-Ua-Platform":        {`"Windows"`},
		"Sec-Fetch-Dest":            {"document"},
		"Sec-Fetch-Mode":            {"navigate"},
		"Sec-Fetch-Site":            {"none"},
		"Sec-Fetch-User":            {"?1"},
		"Upgrade-Insecure-Requests": {"1"},
		"DNT":                       {"1"},
	}

	if referer != "" {
		headers.Set("Referer", referer)
		headers.Set("Sec-Fetch-Site", "same-origin")
	}

	return headers
}

func (a *GarthAuthenticator) Login(ctx context.Context, username, password, mfaToken string) (*Token, error) {
	// Add delay to simulate human behavior
	time.Sleep(time.Duration(500+rand.Intn(1000)) * time.Millisecond)

	// Step 1: Get login ticket (lt) from SSO signin page
	authToken, tokenType, err := a.getLoginTicket(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get login ticket: %w", err)
	}

	// Step 2: Authenticate with credentials
	serviceTicket, err := a.authenticate(ctx, username, password, "", authToken, tokenType)
	if err != nil {
		// Check if MFA is required
		if authErr, ok := err.(*AuthError); ok && authErr.Type == "mfa_required" {
			if mfaToken == "" {
				return nil, errors.New("MFA required but no token provided")
			}
			log.Printf("MFA required, handling with token: %s", mfaToken)
			// Handle MFA authentication
			serviceTicket, err = a.handleMFA(ctx, username, password, mfaToken, authErr.CSRF)
			if err != nil {
				return nil, fmt.Errorf("MFA authentication failed: %w", err)
			}
			log.Printf("MFA authentication successful, service ticket obtained")
		} else {
			return nil, err
		}
	}

	// Step 3: Exchange service ticket for access token
	token, err := a.exchangeServiceTicketForToken(ctx, serviceTicket)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange service ticket for token: %w", err)
	}

	if err := a.storage.StoreToken(token); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	return token, nil
}

// exchangeServiceTicketForToken exchanges service ticket for access token
func (a *GarthAuthenticator) exchangeServiceTicketForToken(ctx context.Context, ticket string) (*Token, error) {
	callbackURL := fmt.Sprintf("https://connect.%s/oauthConfirm?ticket=%s", a.domain, ticket)
	req, err := http.NewRequestWithContext(ctx, "GET", callbackURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create callback request: %w", err)
	}

	// Use realistic browser headers
	req.Header = a.getRealisticBrowserHeaders("https://sso.garmin.com")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("callback request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("callback failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read callback response: %w", err)
	}

	// Extract tokens from embedded JavaScript
	accessToken, err := extractParam(`"accessToken":"([^"]+)"`, string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to extract access token: %w", err)
	}

	refreshToken, err := extractParam(`"refreshToken":"([^"]+)"`, string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to extract refresh token: %w", err)
	}

	expiresAt, err := extractParam(`"expiresAt":(\d+)`, string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to extract expiresAt: %w", err)
	}

	expiresAtInt, err := strconv.ParseInt(expiresAt, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expiresAt: %w", err)
	}

	return &Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAtInt,
		Domain:       a.domain,
	}, nil
}

func (a *GarthAuthenticator) getLoginTicket(ctx context.Context) (string, string, error) {
	params := url.Values{}
	params.Set("id", "gauth-widget")
	params.Set("embedWidget", "true")
	params.Set("gauthHost", fmt.Sprintf("https://sso.%s/sso", a.domain))
	params.Set("service", fmt.Sprintf("https://connect.%s/oauthConfirm", a.domain))
	params.Set("source", fmt.Sprintf("https://sso.%s/sso", a.domain))
	params.Set("redirectAfterAccountLoginUrl", fmt.Sprintf("https://connect.%s/oauthConfirm", a.domain))
	params.Set("redirectAfterAccountCreationUrl", fmt.Sprintf("https://connect.%s/oauthConfirm", a.domain))
	params.Set("consumeServiceTicket", "false")
	params.Set("generateExtraServiceTicket", "true")
	params.Set("clientId", "GarminConnect")
	params.Set("locale", "en_US")

	if a.oauth1Token != "" {
		params.Set("oauth_token", a.oauth1Token)
	}

	loginURL := "https://sso.garmin.com/sso/signin?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", loginURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create login page request: %w", err)
	}

	// Use realistic browser headers with proper referer chain
	for key, values := range a.getRealisticBrowserHeaders("") {
		req.Header[key] = values
	}

	// Add some randomness to the request timing
	time.Sleep(time.Duration(100+rand.Intn(200)) * time.Millisecond)

	resp, err := a.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("login page request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 {
		// Cloudflare blocked us, try with different headers
		return "", "", fmt.Errorf("blocked by Cloudflare - try using different IP or wait before retrying")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read login page response: %w", err)
	}

	return getCSRFToken(string(body))
}

func (a *GarthAuthenticator) authenticate(ctx context.Context, username, password, mfaToken, authToken, tokenType string) (string, error) {
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)
	data.Set("embed", "true")
	data.Set("rememberme", "on")

	if tokenType == "lt" {
		data.Set("lt", authToken)
	} else {
		data.Set("_csrf", authToken)
	}

	data.Set("_eventId", "submit")
	data.Set("geolocation", "")
	data.Set("clientId", "GarminConnect")
	data.Set("service", fmt.Sprintf("https://connect.%s/oauthConfirm", a.domain))
	data.Set("webhost", fmt.Sprintf("https://connect.%s", a.domain))
	data.Set("fromPage", "oauth")
	data.Set("locale", "en_US")
	data.Set("id", "gauth-widget")
	data.Set("redirectAfterAccountLoginUrl", fmt.Sprintf("https://connect.%s/oauthConfirm", a.domain))
	data.Set("redirectAfterAccountCreationUrl", fmt.Sprintf("https://connect.%s/oauthConfirm", a.domain))

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://sso.%s/sso/signin", a.domain), strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create SSO request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", a.userAgent)
	req.Header.Set("Referer", fmt.Sprintf("https://sso.%s/sso/signin", a.domain))
	req.Header.Set("Origin", fmt.Sprintf("https://sso.%s", a.domain))

	// Add some delay to simulate typing
	time.Sleep(time.Duration(800+rand.Intn(400)) * time.Millisecond)

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("SSO request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusPreconditionFailed {
		body, _ := io.ReadAll(resp.Body)
		csrfToken, err := extractParam(`name="_csrf"\s+value="([^"]+)"`, string(body))
		if err != nil {
			return "", &AuthError{
				StatusCode: http.StatusPreconditionFailed,
				Message:    "MFA CSRF token not found",
				Cause:      err,
				Type:       "mfa_required",
				CSRF:       "", // Will be set below
			}
		}
		return "", &AuthError{
			StatusCode: http.StatusPreconditionFailed,
			Message:    "MFA required",
			Type:       "mfa_required",
			CSRF:       csrfToken,
		}
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("authentication failed with status: %d, response: %s", resp.StatusCode, body)
	}

	var authResponse struct {
		Ticket string `json:"ticket"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return "", fmt.Errorf("failed to parse SSO response: %w", err)
	}

	if authResponse.Ticket == "" {
		return "", errors.New("empty ticket in SSO response")
	}

	return authResponse.Ticket, nil
}

func (a *GarthAuthenticator) getEnhancedBrowserHeaders(referrer string) http.Header {
	u, _ := url.Parse(referrer)
	origin := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	return http.Header{
		"User-Agent":                {a.userAgent},
		"Accept":                    {"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		"Accept-Language":           {"en-US,en;q=0.9"},
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

func getCSRFToken(html string) (string, string, error) {
	// More robust regex patterns to handle variations in HTML structure
	patterns := []struct {
		regex     string
		tokenType string
	}{
		// Pattern for login ticket (lt)
		{`<input[^>]*name="lt"[^>]*value="([^"]+)"`, "lt"},
		// Pattern for CSRF token in hidden input
		{`<input[^>]*name="_csrf"[^>]*value="([^"]+)"`, "_csrf"},
		// Pattern for CSRF token in meta tag
		{`<meta[^>]*name="_csrf"[^>]*content="([^"]+)"`, "_csrf"},
		// Pattern for CSRF token in JSON payload
		{`"csrfToken"\s*:\s*"([^"]+)"`, "_csrf"},
	}

	for _, p := range patterns {
		re := regexp.MustCompile(p.regex)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			return matches[1], p.tokenType, nil
		}
	}

	// If we get here, we didn't find a token
	// Log and save the response for debugging
	log.Printf("Failed to find authentication token in HTML response")
	debugFilename := fmt.Sprintf("debug/oauth1_response_%d.html", time.Now().UnixNano())
	if err := os.WriteFile(debugFilename, []byte(html), 0644); err != nil {
		log.Printf("Failed to write debug file: %v", err)
	}
	return "", "", fmt.Errorf("no authentication token found in HTML response; response written to %s", debugFilename)
}

// RefreshToken implements token refresh functionality
func (a *GarthAuthenticator) RefreshToken(ctx context.Context, refreshToken string) (*Token, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, "POST", a.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", a.userAgent)
	req.SetBasicAuth("garmin-connect", "garmin-connect-secret")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("refresh failed: %d %s", resp.StatusCode, body)
	}

	var response struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).Unix()
	return &Token{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		ExpiresAt:    expiresAt,
		Domain:       a.domain,
	}, nil
}

// GetClient returns the HTTP client used for authentication
func (a *GarthAuthenticator) GetClient() *http.Client {
	return a.client
}

// handleMFA processes multi-factor authentication
func (a *GarthAuthenticator) handleMFA(ctx context.Context, username, password, mfaToken, csrfToken string) (string, error) {
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
	data.Set("service", fmt.Sprintf("https://connect.%s", a.domain))
	data.Set("webhost", fmt.Sprintf("https://connect.%s", a.domain))
	data.Set("fromPage", "oauth")
	data.Set("locale", "en_US")
	data.Set("id", "gauth-widget")
	data.Set("redirectAfterAccountLoginUrl", fmt.Sprintf("https://connect.%s/oauthConfirm", a.domain))
	data.Set("redirectAfterAccountCreationUrl", fmt.Sprintf("https://connect.%s/oauthConfirm", a.domain))

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://sso.%s/sso/signin", a.domain), strings.NewReader(data.Encode()))
	if err != nil {
		return "", &AuthError{
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
		return "", &AuthError{
			StatusCode: http.StatusBadGateway,
			Message:    "MFA request failed",
			Cause:      err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", &AuthError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("MFA failed: %s", body),
			Type:       "mfa_failure",
		}
	}

	var mfaResponse struct {
		Ticket string `json:"ticket"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&mfaResponse); err != nil {
		return "", &AuthError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse MFA response",
			Cause:      err,
		}
	}

	if mfaResponse.Ticket == "" {
		return "", &AuthError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Invalid MFA response - ticket missing",
			Type:       "invalid_mfa_response",
		}
	}

	return mfaResponse.Ticket, nil
}

// Configure updates authenticator settings
func (a *GarthAuthenticator) Configure(domain string) {
	a.domain = domain
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
