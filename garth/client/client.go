package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"time"

	"garmin-connect/garth/errors"
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
		return nil, &errors.IOError{
			GarthError: errors.GarthError{
				Message: "Failed to create cookie jar",
				Cause:   err,
			},
		}
	}

	return &Client{
		Domain: domain,
		HTTPClient: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return &errors.APIError{
						GarthHTTPError: errors.GarthHTTPError{
							GarthError: errors.GarthError{
								Message: "Too many redirects",
							},
						},
					}
				}
				return nil
			},
		},
	}, nil
}

// Login authenticates to Garmin Connect using SSO
func (c *Client) Login(email, password string) error {
	ssoClient := sso.NewClient(c.Domain)
	oauth2Token, mfaContext, err := ssoClient.Login(email, password)
	if err != nil {
		return &errors.AuthenticationError{
			GarthError: errors.GarthError{
				Message: "SSO login failed",
				Cause:   err,
			},
		}
	}

	// Handle MFA required
	if mfaContext != nil {
		return &errors.AuthenticationError{
			GarthError: errors.GarthError{
				Message: "MFA required - not implemented yet",
			},
		}
	}

	c.OAuth2Token = oauth2Token
	c.AuthToken = fmt.Sprintf("%s %s", oauth2Token.TokenType, oauth2Token.AccessToken)

	// Get user profile to set username
	profile, err := c.GetUserProfile()
	if err != nil {
		return &errors.AuthenticationError{
			GarthError: errors.GarthError{
				Message: "Failed to get user profile after login",
				Cause:   err,
			},
		}
	}
	c.Username = profile.UserName

	return nil
}

// GetUserProfile retrieves the current user's full profile
func (c *Client) GetUserProfile() (*UserProfile, error) {
	profileURL := fmt.Sprintf("https://connectapi.%s/userprofile-service/socialProfile", c.Domain)

	req, err := http.NewRequest("GET", profileURL, nil)
	if err != nil {
		return nil, &errors.APIError{
			GarthHTTPError: errors.GarthHTTPError{
				GarthError: errors.GarthError{
					Message: "Failed to create profile request",
					Cause:   err,
				},
			},
		}
	}

	req.Header.Set("Authorization", c.AuthToken)
	req.Header.Set("User-Agent", "com.garmin.android.apps.connectmobile")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, &errors.APIError{
			GarthHTTPError: errors.GarthHTTPError{
				GarthError: errors.GarthError{
					Message: "Failed to get user profile",
					Cause:   err,
				},
			},
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, &errors.APIError{
			GarthHTTPError: errors.GarthHTTPError{
				StatusCode: resp.StatusCode,
				Response:   string(body),
				GarthError: errors.GarthError{
					Message: "Profile request failed",
				},
			},
		}
	}

	var profile UserProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, &errors.IOError{
			GarthError: errors.GarthError{
				Message: "Failed to parse profile",
				Cause:   err,
			},
		}
	}

	return &profile, nil
}

// GetActivities retrieves recent activities
func (c *Client) GetActivities(limit int) ([]types.Activity, error) {
	if limit <= 0 {
		limit = 10
	}

	activitiesURL := fmt.Sprintf("https://connectapi.%s/activitylist-service/activities/search/activities?limit=%d&start=0", c.Domain, limit)

	req, err := http.NewRequest("GET", activitiesURL, nil)
	if err != nil {
		return nil, &errors.APIError{
			GarthHTTPError: errors.GarthHTTPError{
				GarthError: errors.GarthError{
					Message: "Failed to create activities request",
					Cause:   err,
				},
			},
		}
	}

	req.Header.Set("Authorization", c.AuthToken)
	req.Header.Set("User-Agent", "com.garmin.android.apps.connectmobile")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, &errors.APIError{
			GarthHTTPError: errors.GarthHTTPError{
				GarthError: errors.GarthError{
					Message: "Failed to get activities",
					Cause:   err,
				},
			},
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, &errors.APIError{
			GarthHTTPError: errors.GarthHTTPError{
				StatusCode: resp.StatusCode,
				Response:   string(body),
				GarthError: errors.GarthError{
					Message: "Activities request failed",
				},
			},
		}
	}

	var activities []types.Activity
	if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
		return nil, &errors.IOError{
			GarthError: errors.GarthError{
				Message: "Failed to parse activities",
				Cause:   err,
			},
		}
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
		return &errors.IOError{
			GarthError: errors.GarthError{
				Message: "Failed to marshal session",
				Cause:   err,
			},
		}
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return &errors.IOError{
			GarthError: errors.GarthError{
				Message: "Failed to write session file",
				Cause:   err,
			},
		}
	}

	return nil
}

// LoadSession loads a session from a file
func (c *Client) LoadSession(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return &errors.IOError{
			GarthError: errors.GarthError{
				Message: "Failed to read session file",
				Cause:   err,
			},
		}
	}

	var session types.SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return &errors.IOError{
			GarthError: errors.GarthError{
				Message: "Failed to unmarshal session",
				Cause:   err,
			},
		}
	}

	c.Domain = session.Domain
	c.Username = session.Username
	c.AuthToken = session.AuthToken

	return nil
}
