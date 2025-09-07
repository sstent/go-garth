package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Client represents the Garmin Connect client
type Client struct {
	domain     string
	httpClient *http.Client
	username   string
	authToken  string
}

// SessionData represents saved session information
type SessionData struct {
	Domain    string `json:"domain"`
	Username  string `json:"username"`
	AuthToken string `json:"auth_token"`
}

// Activity represents a Garmin Connect activity
type Activity struct {
	ActivityID      int64   `json:"activityId"`
	ActivityName    string  `json:"activityName"`
	Description     string  `json:"description"`
	StartTimeLocal  string  `json:"startTimeLocal"`
	StartTimeGMT    string  `json:"startTimeGMT"`
	ActivityType    string  `json:"activityType"`
	EventType       string  `json:"eventType"`
	Distance        float64 `json:"distance"`
	Duration        float64 `json:"duration"`
	ElapsedDuration float64 `json:"elapsedDuration"`
	MovingDuration  float64 `json:"movingDuration"`
	ElevationGain   float64 `json:"elevationGain"`
	ElevationLoss   float64 `json:"elevationLoss"`
	AverageSpeed    float64 `json:"averageSpeed"`
	MaxSpeed        float64 `json:"maxSpeed"`
	Calories        float64 `json:"calories"`
	AverageHR       int     `json:"averageHR"`
	MaxHR           int     `json:"maxHR"`
}

// OAuth1Token represents OAuth1 token response
type OAuth1Token struct {
	OAuthToken       string `json:"oauth_token"`
	OAuthTokenSecret string `json:"oauth_token_secret"`
	MFAToken         string `json:"mfa_token,omitempty"`
	Domain           string `json:"domain"`
}

// OAuth2Token represents OAuth2 token response
type OAuth2Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// OAuth consumer credentials
type OAuthConsumer struct {
	ConsumerKey    string `json:"consumer_key"`
	ConsumerSecret string `json:"consumer_secret"`
}

var (
	csrfRegex     = regexp.MustCompile(`name="_csrf"\s+value="(.+?)"`)
	titleRegex    = regexp.MustCompile(`<title>(.+?)</title>`)
	ticketRegex   = regexp.MustCompile(`embed\?ticket=([^"]+)"`)
	oauthConsumer *OAuthConsumer
)

// loadOAuthConsumer loads OAuth consumer credentials
func loadOAuthConsumer() (*OAuthConsumer, error) {
	if oauthConsumer != nil {
		return oauthConsumer, nil
	}

	// First try to get from S3 (like the Python library)
	resp, err := http.Get("https://thegarth.s3.amazonaws.com/oauth_consumer.json")
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			var consumer OAuthConsumer
			if err := json.NewDecoder(resp.Body).Decode(&consumer); err == nil {
				oauthConsumer = &consumer
				return oauthConsumer, nil
			}
		}
	}

	// Fallback to hardcoded values (these are the same ones used by garth)
	// These are not secret - they're used by the Garmin Connect mobile app
	oauthConsumer = &OAuthConsumer{
		ConsumerKey:    "fc320c35-fbdc-4308-b5c6-8e41a8b2e0c8",
		ConsumerSecret: "8b344b8c-5bd5-4b7b-9c98-ad76a6bbf0e7",
	}
	return oauthConsumer, nil
}

// OAuth1 signing functions
func generateNonce() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func generateTimestamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func percentEncode(s string) string {
	return url.QueryEscape(s)
}

func createSignatureBaseString(method, baseURL string, params map[string]string) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var paramStrs []string
	for _, key := range keys {
		paramStrs = append(paramStrs, percentEncode(key)+"="+percentEncode(params[key]))
	}
	paramString := strings.Join(paramStrs, "&")

	return method + "&" + percentEncode(baseURL) + "&" + percentEncode(paramString)
}

func createSigningKey(consumerSecret, tokenSecret string) string {
	return percentEncode(consumerSecret) + "&" + percentEncode(tokenSecret)
}

func signRequest(consumerSecret, tokenSecret, baseString string) string {
	signingKey := createSigningKey(consumerSecret, tokenSecret)
	mac := hmac.New(sha1.New, []byte(signingKey))
	mac.Write([]byte(baseString))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func createOAuth1AuthorizationHeader(method, requestURL string, params map[string]string, consumerKey, consumerSecret, token, tokenSecret string) string {
	oauthParams := map[string]string{
		"oauth_consumer_key":     consumerKey,
		"oauth_nonce":            generateNonce(),
		"oauth_signature_method": "HMAC-SHA1",
		"oauth_timestamp":        generateTimestamp(),
		"oauth_version":          "1.0",
	}

	if token != "" {
		oauthParams["oauth_token"] = token
	}

	// Combine OAuth params with request params for signature
	allParams := make(map[string]string)
	for k, v := range oauthParams {
		allParams[k] = v
	}
	for k, v := range params {
		allParams[k] = v
	}

	// Parse URL to get base URL without query params
	parsedURL, _ := url.Parse(requestURL)
	baseURL := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path

	// Create signature base string
	baseString := createSignatureBaseString(method, baseURL, allParams)

	// Sign the request
	signature := signRequest(consumerSecret, tokenSecret, baseString)
	oauthParams["oauth_signature"] = signature

	// Build authorization header
	var headerParts []string
	for key, value := range oauthParams {
		headerParts = append(headerParts, percentEncode(key)+"=\""+percentEncode(value)+"\"")
	}
	sort.Strings(headerParts)

	return "OAuth " + strings.Join(headerParts, ", ")
}

// loadEnvCredentials loads credentials from .env file
func loadEnvCredentials() (email, password, domain string, err error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		return "", "", "", fmt.Errorf("error loading .env file: %w", err)
	}

	email = os.Getenv("GARMIN_EMAIL")
	password = os.Getenv("GARMIN_PASSWORD")
	domain = os.Getenv("GARMIN_DOMAIN")

	if email == "" {
		return "", "", "", fmt.Errorf("GARMIN_EMAIL not found in .env file")
	}
	if password == "" {
		return "", "", "", fmt.Errorf("GARMIN_PASSWORD not found in .env file")
	}
	if domain == "" {
		domain = "garmin.com" // default value
	}

	return email, password, domain, nil
}

// NewClient creates a new Garmin Connect client
func NewClient(domain string) (*Client, error) {
	if domain == "" {
		domain = "garmin.com"
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	return &Client{
		domain: domain,
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				// Allow up to 10 redirects
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
	}, nil
}

// Login authenticates using Garmin's SSO flow (matching Python garth implementation)
func (c *Client) Login(email, password string) error {
	fmt.Printf("Logging in to Garmin Connect (%s) using SSO flow...\n", c.domain)

	// Step 1: Set up SSO parameters
	ssoURL := fmt.Sprintf("https://sso.%s/sso", c.domain)
	ssoEmbedURL := fmt.Sprintf("%s/embed", ssoURL)

	ssoEmbedParams := url.Values{
		"id":          {"gauth-widget"},
		"embedWidget": {"true"},
		"gauthHost":   {ssoURL},
	}

	signinParams := url.Values{
		"id":                              {"gauth-widget"},
		"embedWidget":                     {"true"},
		"gauthHost":                       {ssoEmbedURL},
		"service":                         {ssoEmbedURL},
		"source":                          {ssoEmbedURL},
		"redirectAfterAccountLoginUrl":    {ssoEmbedURL},
		"redirectAfterAccountCreationUrl": {ssoEmbedURL},
	}

	// Step 2: Initialize SSO session
	fmt.Println("Initializing SSO session...")
	embedURL := fmt.Sprintf("https://sso.%s/sso/embed?%s", c.domain, ssoEmbedParams.Encode())
	req, err := http.NewRequest("GET", embedURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create embed request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to initialize SSO: %w", err)
	}
	resp.Body.Close()

	// Step 3: Get signin page and CSRF token
	fmt.Println("Getting signin page...")
	signinURL := fmt.Sprintf("https://sso.%s/sso/signin?%s", c.domain, signinParams.Encode())
	req, err = http.NewRequest("GET", signinURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create signin request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Referer", embedURL)

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get signin page: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read signin response: %w", err)
	}

	// Extract CSRF token
	csrfToken := c.extractCSRFToken(string(body))
	if csrfToken == "" {
		return fmt.Errorf("failed to find CSRF token")
	}
	fmt.Printf("Found CSRF token: %s\n", csrfToken[:10]+"...")

	// Step 4: Submit login form
	fmt.Println("Submitting login credentials...")
	formData := url.Values{
		"username": {email},
		"password": {password},
		"embed":    {"true"},
		"_csrf":    {csrfToken},
	}

	req, err = http.NewRequest("POST", signinURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Referer", signinURL)

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to submit login: %w", err)
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read login response: %w", err)
	}

	// Check login result
	title := c.extractTitle(string(body))
	fmt.Printf("Login response title: %s\n", title)

	if strings.Contains(title, "MFA") {
		return fmt.Errorf("MFA required - not implemented yet")
	}

	if title != "Success" {
		return fmt.Errorf("login failed, unexpected title: %s", title)
	}

	// Step 5: Extract ticket for OAuth flow
	fmt.Println("Extracting OAuth ticket...")
	ticket := c.extractTicket(string(body))
	if ticket == "" {
		return fmt.Errorf("failed to find OAuth ticket")
	}
	fmt.Printf("Found ticket: %s\n", ticket[:10]+"...")

	// Step 6: Get OAuth1 token
	oauth1Token, err := c.getOAuth1Token(ticket)
	if err != nil {
		return fmt.Errorf("failed to get OAuth1 token: %w", err)
	}
	fmt.Println("Got OAuth1 token")

	// Step 7: Exchange for OAuth2 token
	oauth2Token, err := c.exchangeToken(oauth1Token)
	if err != nil {
		return fmt.Errorf("failed to exchange for OAuth2 token: %w", err)
	}
	fmt.Printf("Got OAuth2 token: %s\n", oauth2Token.TokenType)

	// Step 8: Set auth token and get user profile
	c.authToken = fmt.Sprintf("%s %s", oauth2Token.TokenType, oauth2Token.AccessToken)

	if err := c.getUserProfile(); err != nil {
		return fmt.Errorf("failed to get user profile: %w", err)
	}

	fmt.Println("SSO authentication successful!")
	return nil
}

func (c *Client) extractCSRFToken(html string) string {
	matches := csrfRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (c *Client) extractTitle(html string) string {
	matches := titleRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (c *Client) extractTicket(html string) string {
	matches := ticketRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (c *Client) getOAuth1Token(ticket string) (*OAuth1Token, error) {
	consumer, err := loadOAuthConsumer()
	if err != nil {
		return nil, fmt.Errorf("failed to load OAuth consumer: %w", err)
	}

	baseURL := fmt.Sprintf("https://connectapi.%s/oauth-service/oauth/", c.domain)
	loginURL := fmt.Sprintf("https://sso.%s/sso/embed", c.domain)
	tokenURL := fmt.Sprintf("%spreauthorized?ticket=%s&login-url=%s&accepts-mfa-tokens=true",
		baseURL, ticket, url.QueryEscape(loginURL))

	// Parse URL to extract query parameters for signing
	parsedURL, err := url.Parse(tokenURL)
	if err != nil {
		return nil, err
	}

	// Extract query parameters for OAuth signing
	queryParams := make(map[string]string)
	for key, values := range parsedURL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	// Create OAuth1 signed request
	baseURLForSigning := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path
	authHeader := createOAuth1AuthorizationHeader("GET", baseURLForSigning, queryParams,
		consumer.ConsumerKey, consumer.ConsumerSecret, "", "")

	req, err := http.NewRequest("GET", tokenURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("User-Agent", "com.garmin.android.apps.connectmobile")

	fmt.Printf("OAuth1 request URL: %s\n", tokenURL)
	fmt.Printf("OAuth1 authorization header: %s\n", authHeader[:50]+"...")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	bodyStr := string(body)
	fmt.Printf("OAuth1 response status: %d\n", resp.StatusCode)
	fmt.Printf("OAuth1 response: %s\n", bodyStr[:min(200, len(bodyStr))])

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("OAuth1 request failed with status %d: %s", resp.StatusCode, bodyStr)
	}

	// Parse query string response - handle both & and ; separators
	bodyStr = strings.ReplaceAll(bodyStr, ";", "&")
	values, err := url.ParseQuery(bodyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OAuth1 response: %w", err)
	}

	oauthToken := values.Get("oauth_token")
	oauthTokenSecret := values.Get("oauth_token_secret")

	if oauthToken == "" || oauthTokenSecret == "" {
		return nil, fmt.Errorf("missing oauth_token or oauth_token_secret in response")
	}

	return &OAuth1Token{
		OAuthToken:       oauthToken,
		OAuthTokenSecret: oauthTokenSecret,
		MFAToken:         values.Get("mfa_token"),
		Domain:           c.domain,
	}, nil
}

func (c *Client) exchangeToken(oauth1Token *OAuth1Token) (*OAuth2Token, error) {
	consumer, err := loadOAuthConsumer()
	if err != nil {
		return nil, fmt.Errorf("failed to load OAuth consumer: %w", err)
	}

	exchangeURL := fmt.Sprintf("https://connectapi.%s/oauth-service/oauth/exchange/user/2.0", c.domain)

	// Prepare form data
	formData := url.Values{}
	if oauth1Token.MFAToken != "" {
		formData.Set("mfa_token", oauth1Token.MFAToken)
	}

	// Convert form data to map for OAuth signing
	formParams := make(map[string]string)
	for key, values := range formData {
		if len(values) > 0 {
			formParams[key] = values[0]
		}
	}

	// Create OAuth1 signed request
	authHeader := createOAuth1AuthorizationHeader("POST", exchangeURL, formParams,
		consumer.ConsumerKey, consumer.ConsumerSecret, oauth1Token.OAuthToken, oauth1Token.OAuthTokenSecret)

	req, err := http.NewRequest("POST", exchangeURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", authHeader)
	req.Header.Set("User-Agent", "com.garmin.android.apps.connectmobile")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	fmt.Println("Attempting OAuth2 token exchange with signed request...")
	fmt.Printf("OAuth2 authorization header: %s\n", authHeader[:50]+"...")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Printf("OAuth2 exchange response status: %d\n", resp.StatusCode)
	fmt.Printf("OAuth2 exchange response: %s\n", string(body)[:min(500, len(string(body)))])

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("OAuth2 exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var oauth2Token OAuth2Token
	if err := json.Unmarshal(body, &oauth2Token); err != nil {
		return nil, fmt.Errorf("failed to decode OAuth2 token: %w", err)
	}

	return &oauth2Token, nil
}

// getUserProfile gets user profile information
func (c *Client) getUserProfile() error {
	fmt.Println("Getting user profile...")

	profileURL := fmt.Sprintf("https://connectapi.%s/userprofile-service/socialProfile", c.domain)

	req, err := http.NewRequest("GET", profileURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create profile request: %w", err)
	}

	req.Header.Set("Authorization", c.authToken)
	req.Header.Set("User-Agent", "com.garmin.android.apps.connectmobile")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get user profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Profile request failed. Status: %d, Response: %s\n", resp.StatusCode, string(body))
		return fmt.Errorf("profile request failed with status: %d", resp.StatusCode)
	}

	var profile map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return fmt.Errorf("failed to parse profile: %w", err)
	}

	if username, ok := profile["userName"].(string); ok {
		c.username = username
		fmt.Printf("Username: %s\n", c.username)
	} else {
		return fmt.Errorf("failed to extract username from profile")
	}

	return nil
}

// GetActivities retrieves recent activities
func (c *Client) GetActivities(limit int) ([]Activity, error) {
	if limit <= 0 {
		limit = 10
	}

	fmt.Printf("Getting last %d activities...\n", limit)

	activitiesURL := fmt.Sprintf("https://connectapi.%s/activitylist-service/activities/search/activities?limit=%d&start=0", c.domain, limit)

	req, err := http.NewRequest("GET", activitiesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create activities request: %w", err)
	}

	req.Header.Set("Authorization", c.authToken)
	req.Header.Set("User-Agent", "com.garmin.android.apps.connectmobile")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Activities request failed. Status: %d, Response: %s\n", resp.StatusCode, string(body))
		return nil, fmt.Errorf("activities request failed with status: %d", resp.StatusCode)
	}

	var activities []Activity
	if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
		return nil, fmt.Errorf("failed to parse activities: %w", err)
	}

	fmt.Printf("Retrieved %d activities\n", len(activities))
	return activities, nil
}

// SaveSession saves the current session to a file
func (c *Client) SaveSession(filename string) error {
	session := SessionData{
		Domain:    c.domain,
		Username:  c.username,
		AuthToken: c.authToken,
	}

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	fmt.Printf("Session saved to %s\n", filename)
	return nil
}

// LoadSession loads a session from a file
func (c *Client) LoadSession(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read session file: %w", err)
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("failed to unmarshal session: %w", err)
	}

	c.domain = session.Domain
	c.username = session.Username
	c.authToken = session.AuthToken

	fmt.Printf("Session loaded from %s\n", filename)
	return nil
}

// Helper function for min (Go 1.21+ has this built-in)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	// Load credentials from .env file
	email, password, domain, err := loadEnvCredentials()
	if err != nil {
		fmt.Printf("Failed to load credentials: %v\n", err)
		fmt.Println("Please create a .env file with GARMIN_EMAIL and GARMIN_PASSWORD")
		return
	}

	client, err := NewClient(domain)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}

	// Try to load existing session first
	sessionFile := "garmin_session.json"
	if err := client.LoadSession(sessionFile); err != nil {
		fmt.Println("No existing session found, logging in with credentials from .env...")

		if err := client.Login(email, password); err != nil {
			fmt.Printf("Login failed: %v\n", err)
			return
		}

		// Save session for future use
		if err := client.SaveSession(sessionFile); err != nil {
			fmt.Printf("Failed to save session: %v\n", err)
		}
	} else {
		fmt.Println("Loaded existing session")
	}

	// Test getting activities
	activities, err := client.GetActivities(5)
	if err != nil {
		fmt.Printf("Failed to get activities: %v\n", err)
		return
	}

	// Display activities
	fmt.Printf("\n=== Recent Activities ===\n")
	for i, activity := range activities {
		fmt.Printf("%d. %s\n", i+1, activity.ActivityName)
		fmt.Printf("   Type: %s\n", activity.ActivityType)
		fmt.Printf("   Date: %s\n", activity.StartTimeLocal)
		if activity.Distance > 0 {
			fmt.Printf("   Distance: %.2f km\n", activity.Distance/1000)
		}
		if activity.Duration > 0 {
			duration := time.Duration(activity.Duration) * time.Second
			fmt.Printf("   Duration: %v\n", duration.Round(time.Second))
		}
		fmt.Println()
	}
}
