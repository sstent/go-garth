package garth

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"sort"
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
	baseTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		Proxy: http.ProxyFromEnvironment,
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
		userAgent: "GCMv3",
		domain:    opts.Domain,
	}

	if setter, ok := opts.Storage.(interface{ SetAuthenticator(a Authenticator) }); ok {
		setter.SetAuthenticator(auth)
	}

	return auth
}

func (a *GarthAuthenticator) Login(ctx context.Context, username, password, mfaToken string) (*Token, error) {
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

	// Step 3: Get OAuth1 request token
	oauth1RequestToken, err := a.fetchOAuth1RequestToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth1 request token: %w", err)
	}

	// Step 4: Authorize OAuth1 request token (using the session from authentication)
	err = a.authorizeOAuth1Token(ctx, oauth1RequestToken)
	if err != nil {
		return nil, fmt.Errorf("failed to authorize OAuth1 token: %w", err)
	}

	// Step 5: Exchange service ticket for OAuth1 access token
	oauth1AccessToken, err := a.exchangeTicketForOAuth1Token(ctx, serviceTicket)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange ticket for OAuth1 access token: %w", err)
	}

	// Step 6: Exchange OAuth1 access token for OAuth2 token
	token, err := a.exchangeOAuth1ForOAuth2Token(ctx, oauth1AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange OAuth1 for OAuth2 token: %w", err)
	}

	if err := a.storage.StoreToken(token); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	return token, nil
}

func (a *GarthAuthenticator) fetchOAuth1RequestToken(ctx context.Context) (*OAuth1Token, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://connectapi.%s/oauth-service/oauth/request_token", a.domain), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Sign request with OAuth1 consumer credentials
	req.Header.Set("Authorization", a.buildOAuth1Header(req, nil, ""))

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &OAuth1Token{
		Token:  values.Get("oauth_token"),
		Secret: values.Get("oauth_token_secret"),
	}, nil
}

func (a *GarthAuthenticator) authorizeOAuth1Token(ctx context.Context, token *OAuth1Token) error {
	params := url.Values{}
	params.Set("oauth_token", token.Token)
	authURL := fmt.Sprintf("https://connect.%s/oauthConfirm?%s", a.domain, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", authURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create authorization request: %w", err)
	}

	req.Header = a.getEnhancedBrowserHeaders(authURL)

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("authorization request failed: %w", err)
	}
	defer resp.Body.Close()

	// We don't need the CSRF token anymore, so just check for success
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("authorization failed with status: %d, response: %s", resp.StatusCode, body)
	}

	return nil
}

func (a *GarthAuthenticator) exchangeTicketForOAuth1Token(ctx context.Context, ticket string) (*OAuth1Token, error) {
	data := url.Values{}
	data.Set("oauth_verifier", ticket)

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://connectapi.%s/oauth-service/oauth/access_token", a.domain), strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create access token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", a.buildOAuth1Header(req, nil, ""))

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("access token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("access token request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read access token response: %w", err)
	}

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, fmt.Errorf("failed to parse access token response: %w", err)
	}

	return &OAuth1Token{
		Token:  values.Get("oauth_token"),
		Secret: values.Get("oauth_token_secret"),
	}, nil
}

func (a *GarthAuthenticator) exchangeOAuth1ForOAuth2Token(ctx context.Context, oauth1Token *OAuth1Token) (*Token, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("https://connectapi.%s/oauth-service/oauth/exchange_token", a.domain), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create token exchange request: %w", err)
	}

	req.Header.Set("Authorization", a.buildOAuth1Header(req, oauth1Token, ""))

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token exchange request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status: %d", resp.StatusCode)
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if token.OAuth2Token != nil {
		token.OAuth2Token.ExpiresAt = time.Now().Add(time.Duration(token.OAuth2Token.ExpiresIn) * time.Second).Unix()
	}
	token.OAuth1Token = oauth1Token
	return &token, nil
}

func (a *GarthAuthenticator) buildOAuth1Header(req *http.Request, token *OAuth1Token, callback string) string {
	oauthParams := url.Values{}
	oauthParams.Set("oauth_consumer_key", "fc020df2-e33d-4ec5-987a-7fb6de2e3850")
	oauthParams.Set("oauth_signature_method", "HMAC-SHA1")
	oauthParams.Set("oauth_timestamp", fmt.Sprintf("%d", time.Now().Unix()))
	oauthParams.Set("oauth_nonce", fmt.Sprintf("%d", rand.Int63()))
	oauthParams.Set("oauth_version", "1.0")

	if token != nil {
		oauthParams.Set("oauth_token", token.Token)
	}
	if callback != "" {
		oauthParams.Set("oauth_callback", callback)
	}

	// Generate signature
	baseString := a.buildSignatureBaseString(req, oauthParams)
	signingKey := url.QueryEscape("secret_key_from_mobile_app") + "&"
	if token != nil {
		signingKey += url.QueryEscape(token.Secret)
	}

	mac := hmac.New(sha1.New, []byte(signingKey))
	mac.Write([]byte(baseString))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	oauthParams.Set("oauth_signature", signature)

	// Build header
	params := make([]string, 0, len(oauthParams))
	for k, v := range oauthParams {
		params = append(params, fmt.Sprintf(`%s="%s"`, k, url.QueryEscape(v[0])))
	}
	sort.Strings(params)

	return "OAuth " + strings.Join(params, ", ")
}

func (a *GarthAuthenticator) buildSignatureBaseString(req *http.Request, oauthParams url.Values) string {
	method := strings.ToUpper(req.Method)
	baseURL := req.URL.Scheme + "://" + req.URL.Host + req.URL.Path

	// Collect all parameters
	params := url.Values{}
	for k, v := range req.URL.Query() {
		params[k] = v
	}
	for k, v := range oauthParams {
		params[k] = v
	}

	// Sort parameters
	paramKeys := make([]string, 0, len(params))
	for k := range params {
		paramKeys = append(paramKeys, k)
	}
	sort.Strings(paramKeys)

	paramPairs := make([]string, 0, len(paramKeys))
	for _, k := range paramKeys {
		paramPairs = append(paramPairs, fmt.Sprintf("%s=%s", k, url.QueryEscape(params[k][0])))
	}

	queryString := strings.Join(paramPairs, "&")
	return fmt.Sprintf("%s&%s&%s", method, url.QueryEscape(baseURL), url.QueryEscape(queryString))
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

	req.Header.Set("User-Agent", a.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Referer", fmt.Sprintf("https://sso.%s/sso/signin?id=gauth-widget&embedWidget=true&gauthHost=https://sso.%s/sso", a.domain, a.domain))
	req.Header.Set("Origin", fmt.Sprintf("https://sso.%s", a.domain))

	resp, err := a.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("login page request failed: %w", err)
	}
	defer resp.Body.Close()

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

func (a *GarthAuthenticator) ExchangeToken(ctx context.Context, token *OAuth1Token) (*Token, error) {
	return a.exchangeOAuth1ForOAuth2Token(ctx, token)
}

// Removed exchangeTicketForToken method - no longer needed

func (a *GarthAuthenticator) getEnhancedBrowserHeaders(referrer string) http.Header {
	u, _ := url.Parse(referrer)
	origin := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	return http.Header{
		"User-Agent":                {"GCMv3"},
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

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %w", err)
	}

	if token.OAuth2Token != nil {
		token.OAuth2Token.ExpiresAt = time.Now().Add(time.Duration(token.OAuth2Token.ExpiresIn) * time.Second).Unix()
	}
	return &token, nil
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
	req.Header.Set("User-Agent", "GCMv3")

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
