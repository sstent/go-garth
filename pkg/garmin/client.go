package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sstent/go-garth/garth/errors"
	"github.com/sstent/go-garth/garth/sso"
	"github.com/sstent/go-garth/garth/types"
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

	// Extract host without scheme if present
	if strings.Contains(domain, "://") {
		if u, err := url.Parse(domain); err == nil {
			domain = u.Host
		}
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
	// Extract host without scheme if present
	host := c.Domain
	if strings.Contains(host, "://") {
		if u, err := url.Parse(host); err == nil {
			host = u.Host
		}
	}

	ssoClient := sso.NewClient(host)
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
func (c *Client) GetUserProfile() (*types.UserProfile, error) {
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

	var profile types.UserProfile
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

// ConnectAPI makes a raw API request to the Garmin Connect API
func (c *Client) ConnectAPI(path string, method string, params url.Values, body io.Reader) ([]byte, error) {
	u := &url.URL{
		Scheme:   "https",
		Host:     fmt.Sprintf("connectapi.%s", c.Domain),
		Path:     path,
		RawQuery: params.Encode(),
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, &errors.APIError{
			GarthHTTPError: errors.GarthHTTPError{
				GarthError: errors.GarthError{
					Message: "Failed to create request",
					Cause:   err,
				},
			},
		}
	}

	req.Header.Set("Authorization", c.AuthToken)
	req.Header.Set("User-Agent", "garth-go-client/1.0")
	req.Header.Set("Accept", "application/json")

	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, &errors.APIError{
			GarthHTTPError: errors.GarthHTTPError{
				GarthError: errors.GarthError{
					Message: "Request failed",
					Cause:   err,
				},
			},
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &errors.APIError{
			GarthHTTPError: errors.GarthHTTPError{
				StatusCode: resp.StatusCode,
				Response:   string(bodyBytes),
				GarthError: errors.GarthError{
					Message: fmt.Sprintf("API request failed with status %d: %s",
						resp.StatusCode, tryReadErrorBody(bytes.NewReader(bodyBytes))),
				},
			},
		}
	}

	return io.ReadAll(resp.Body)
}

func tryReadErrorBody(r io.Reader) string {
	body, err := io.ReadAll(r)
	if err != nil {
		return "failed to read error response"
	}
	return string(body)
}

// Upload sends a file to Garmin Connect
func (c *Client) Upload(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return &errors.IOError{
			GarthError: errors.GarthError{
				Message: "Failed to open file",
				Cause:   err,
			},
		}
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return &errors.IOError{
			GarthError: errors.GarthError{
				Message: "Failed to create form file",
				Cause:   err,
			},
		}
	}

	if _, err := io.Copy(part, file); err != nil {
		return &errors.IOError{
			GarthError: errors.GarthError{
				Message: "Failed to copy file content",
				Cause:   err,
			},
		}
	}

	if err := writer.Close(); err != nil {
		return &errors.IOError{
			GarthError: errors.GarthError{
				Message: "Failed to close multipart writer",
				Cause:   err,
			},
		}
	}

	_, err = c.ConnectAPI("/upload-service/upload", "POST", nil, body)
	if err != nil {
		return &errors.APIError{
			GarthHTTPError: errors.GarthHTTPError{
				GarthError: errors.GarthError{
					Message: "File upload failed",
					Cause:   err,
				},
			},
		}
	}

	return nil
}

// Download retrieves a file from Garmin Connect
func (c *Client) Download(activityID string, filePath string) error {
	params := url.Values{}
	params.Add("activityId", activityID)

	resp, err := c.ConnectAPI("/download-service/export", "GET", params, nil)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, resp, 0644); err != nil {
		return &errors.IOError{
			GarthError: errors.GarthError{
				Message: "Failed to save file",
				Cause:   err,
			},
		}
	}

	return nil
}

// GetActivities retrieves activities filtered by date range
func (c *Client) GetActivities(start, end time.Time) ([]types.Activity, error) {
	activitiesURL := fmt.Sprintf("https://connectapi.%s/activitylist-service/activities/search/activities", c.Domain)

	params := url.Values{}
	if !start.IsZero() {
		params.Add("startDate", start.Format("2006-01-02"))
	}
	if !end.IsZero() {
		params.Add("endDate", end.Format("2006-01-02"))
	}
	activitiesURL += "?" + params.Encode()

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
