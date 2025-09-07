package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"

	"garmin-connect/garth/sso"
	"garmin-connect/garth/types"
)

// Client represents the Garmin Connect API client
type Client struct {
	Domain      string
	HTTPClient  *http.Client
	Username    string
	AuthToken   string
	OAuth1Token *types.OAuth1Token
	OAuth2Token *types.OAuth2Token
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
		Domain: domain,
		HTTPClient: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
	}, nil
}

// Login authenticates to Garmin Connect using SSO
func (c *Client) Login(email, password string) error {
	ssoClient := sso.NewClient(c.Domain)
	oauth2Token, err := ssoClient.Login(email, password)
	if err != nil {
		return fmt.Errorf("SSO login failed: %w", err)
	}

	c.OAuth2Token = oauth2Token
	c.AuthToken = fmt.Sprintf("%s %s", oauth2Token.TokenType, oauth2Token.AccessToken)

	// Get user profile to set username
	if err := c.GetUserProfile(); err != nil {
		return fmt.Errorf("failed to get user profile after login: %w", err)
	}

	return nil
}

// GetUserProfile retrieves the current user's profile
func (c *Client) GetUserProfile() error {
	profileURL := fmt.Sprintf("https://connectapi.%s/userprofile-service/socialProfile", c.Domain)

	req, err := http.NewRequest("GET", profileURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create profile request: %w", err)
	}

	req.Header.Set("Authorization", c.AuthToken)
	req.Header.Set("User-Agent", "com.garmin.android.apps.connectmobile")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get user profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("profile request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var profile map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return fmt.Errorf("failed to parse profile: %w", err)
	}

	if username, ok := profile["userName"].(string); ok {
		c.Username = username
		return nil
	}

	return fmt.Errorf("username not found in profile response")
}

// GetActivities retrieves recent activities
func (c *Client) GetActivities(limit int) ([]types.Activity, error) {
	if limit <= 0 {
		limit = 10
	}

	activitiesURL := fmt.Sprintf("https://connectapi.%s/activitylist-service/activities/search/activities?limit=%d&start=0", c.Domain, limit)

	req, err := http.NewRequest("GET", activitiesURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create activities request: %w", err)
	}

	req.Header.Set("Authorization", c.AuthToken)
	req.Header.Set("User-Agent", "com.garmin.android.apps.connectmobile")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("activities request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var activities []types.Activity
	if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
		return nil, fmt.Errorf("failed to parse activities: %w", err)
	}

	return activities, nil
}

// SaveSession saves the current session to a file
func (c *Client) SaveSession(filename string) error {
	session := types.SessionData{
		Domain:    c.Domain,
		Username:  c.Username,
		AuthToken: c.AuthToken,
	}

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// LoadSession loads a session from a file
func (c *Client) LoadSession(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read session file: %w", err)
	}

	var session types.SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("failed to unmarshal session: %w", err)
	}

	c.Domain = session.Domain
	c.Username = session.Username
	c.AuthToken = session.AuthToken

	return nil
}
