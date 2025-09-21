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

	"go-garth/internal/auth/sso"
	"go-garth/internal/errors"
	types "go-garth/internal/models/types"
	shared "go-garth/shared/interfaces"
	models "go-garth/shared/models"
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

// Verify that Client implements shared.APIClient
var _ shared.APIClient = (*Client)(nil)

// GetUsername returns the authenticated username
func (c *Client) GetUsername() string {
	return c.Username
}

// GetUserSettings retrieves the current user's settings
func (c *Client) GetUserSettings() (*models.UserSettings, error) {
	scheme := "https"
	if strings.HasPrefix(c.Domain, "127.0.0.1") {
		scheme = "http"
	}
	host := c.Domain
	if !strings.HasPrefix(c.Domain, "127.0.0.1") {
		host = "connectapi." + c.Domain
	}
	settingsURL := fmt.Sprintf("%s://%s/userprofile-service/userprofile/user-settings", scheme, host)

	req, err := http.NewRequest("GET", settingsURL, nil)
	if err != nil {
		return nil, &errors.APIError{
			GarthHTTPError: errors.GarthHTTPError{
				GarthError: errors.GarthError{
					Message: "Failed to create user settings request",
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
					Message: "Failed to get user settings",
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
					Message: "User settings request failed",
				},
			},
		}
	}

	var settings models.UserSettings
	if err := json.NewDecoder(resp.Body).Decode(&settings); err != nil {
		return nil, &errors.IOError{
			GarthError: errors.GarthError{
				Message: "Failed to parse user settings",
				Cause:   err,
			},
		}
	}

	return &settings, nil
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

// Logout clears the current session and tokens.
func (c *Client) Logout() error {
	c.AuthToken = ""
	c.Username = ""
	c.OAuth1Token = nil
	c.OAuth2Token = nil

	// Clear cookies
	if c.HTTPClient != nil && c.HTTPClient.Jar != nil {
		// Create a dummy URL for the domain to clear all cookies associated with it
		dummyURL, err := url.Parse(fmt.Sprintf("https://%s", c.Domain))
		if err == nil {
			c.HTTPClient.Jar.SetCookies(dummyURL, []*http.Cookie{})
		}
	}
	return nil
}

// GetUserProfile retrieves the current user's full profile
func (c *Client) GetUserProfile() (*types.UserProfile, error) {
	scheme := "https"
	if strings.HasPrefix(c.Domain, "127.0.0.1") {
		scheme = "http"
	}
	host := c.Domain
	if !strings.HasPrefix(c.Domain, "127.0.0.1") {
		host = "connectapi." + c.Domain
	}
	profileURL := fmt.Sprintf("%s://%s/userprofile-service/socialProfile", scheme, host)

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
	scheme := "https"
	if strings.HasPrefix(c.Domain, "127.0.0.1") {
		scheme = "http"
	}
	u := &url.URL{
		Scheme:   scheme,
		Host:     c.Domain,
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
func (c *Client) Download(activityID string, format string, filePath string) error {
	params := url.Values{}
	params.Add("activityId", activityID)
	// Add format parameter if provided and not empty
	if format != "" {
		params.Add("format", format)
	}

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

// GetActivities retrieves recent activities
func (c *Client) GetActivities(limit int) ([]types.Activity, error) {
	if limit <= 0 {
		limit = 10
	}

	scheme := "https"
	if strings.HasPrefix(c.Domain, "127.0.0.1") {
		scheme = "http"
	}

	activitiesURL := fmt.Sprintf("%s://connectapi.%s/activitylist-service/activities/search/activities?limit=%d&start=0", scheme, c.Domain, limit)

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

func (c *Client) GetSleepData(startDate, endDate time.Time) ([]types.SleepData, error) {
	// TODO: Implement GetSleepData
	return nil, fmt.Errorf("GetSleepData not implemented")
}

// GetHrvData retrieves HRV data for a specified number of days
func (c *Client) GetHrvData(days int) ([]types.HrvData, error) {
	// TODO: Implement GetHrvData
	return nil, fmt.Errorf("GetHrvData not implemented")
}

// GetStressData retrieves stress data
func (c *Client) GetStressData(startDate, endDate time.Time) ([]types.StressData, error) {
	// TODO: Implement GetStressData
	return nil, fmt.Errorf("GetStressData not implemented")
}

// GetBodyBatteryData retrieves Body Battery data
func (c *Client) GetBodyBatteryData(startDate, endDate time.Time) ([]types.BodyBatteryData, error) {
	// TODO: Implement GetBodyBatteryData
	return nil, fmt.Errorf("GetBodyBatteryData not implemented")
}

// GetStepsData retrieves steps data for a specified date range
func (c *Client) GetStepsData(startDate, endDate time.Time) ([]types.StepsData, error) {
	// TODO: Implement GetStepsData
	return nil, fmt.Errorf("GetStepsData not implemented")
}

// GetDistanceData retrieves distance data for a specified date range
func (c *Client) GetDistanceData(startDate, endDate time.Time) ([]types.DistanceData, error) {
	// TODO: Implement GetDistanceData
	return nil, fmt.Errorf("GetDistanceData not implemented")
}

// GetCaloriesData retrieves calories data for a specified date range
func (c *Client) GetCaloriesData(startDate, endDate time.Time) ([]types.CaloriesData, error) {
	// TODO: Implement GetCaloriesData
	return nil, fmt.Errorf("GetCaloriesData not implemented")
}

// GetVO2MaxData retrieves VO2 max data using the modern approach via user settings
func (c *Client) GetVO2MaxData(startDate, endDate time.Time) ([]types.VO2MaxData, error) {
	// Get user settings which contains current VO2 max values
	settings, err := c.GetUserSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	// Create VO2MaxData for the date range
	var results []types.VO2MaxData
	current := startDate
	for !current.After(endDate) {
		vo2Data := types.VO2MaxData{
			Date:          current,
			UserProfilePK: settings.ID,
		}

		// Set VO2 max values if available
		if settings.UserData.VO2MaxRunning != nil {
			vo2Data.VO2MaxRunning = settings.UserData.VO2MaxRunning
		}
		if settings.UserData.VO2MaxCycling != nil {
			vo2Data.VO2MaxCycling = settings.UserData.VO2MaxCycling
		}

		results = append(results, vo2Data)
		current = current.AddDate(0, 0, 1)
	}

	return results, nil
}

// GetCurrentVO2Max retrieves the current VO2 max values from user profile
func (c *Client) GetCurrentVO2Max() (*types.VO2MaxProfile, error) {
	settings, err := c.GetUserSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	profile := &types.VO2MaxProfile{
		UserProfilePK: settings.ID,
		LastUpdated:   time.Now(),
	}

	// Add running VO2 max if available
	if settings.UserData.VO2MaxRunning != nil && *settings.UserData.VO2MaxRunning > 0 {
		profile.Running = &types.VO2MaxEntry{
			Value:        *settings.UserData.VO2MaxRunning,
			ActivityType: "running",
			Date:         time.Now(),
			Source:       "user_settings",
		}
	}

	// Add cycling VO2 max if available
	if settings.UserData.VO2MaxCycling != nil && *settings.UserData.VO2MaxCycling > 0 {
		profile.Cycling = &types.VO2MaxEntry{
			Value:        *settings.UserData.VO2MaxCycling,
			ActivityType: "cycling",
			Date:         time.Now(),
			Source:       "user_settings",
		}
	}

	return profile, nil
}

// GetHeartRateZones retrieves heart rate zone data
func (c *Client) GetHeartRateZones() (*types.HeartRateZones, error) {
	scheme := "https"
	if strings.HasPrefix(c.Domain, "127.0.0.1") {
		scheme = "http"
	}

	hrzURL := fmt.Sprintf("%s://connectapi.%s/userprofile-service/userprofile/heartRateZones", scheme, c.Domain)

	req, err := http.NewRequest("GET", hrzURL, nil)
	if err != nil {
		return nil, &errors.APIError{
			GarthHTTPError: errors.GarthHTTPError{
				GarthError: errors.GarthError{
					Message: "Failed to create HR zones request",
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
					Message: "Failed to get HR zones data",
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
					Message: "HR zones request failed",
				},
			},
		}
	}

	var hrZones types.HeartRateZones
	if err := json.NewDecoder(resp.Body).Decode(&hrZones); err != nil {
		return nil, &errors.IOError{
			GarthError: errors.GarthError{
				Message: "Failed to parse HR zones data",
				Cause:   err,
			},
		}
	}

	return &hrZones, nil
}

// GetWellnessData retrieves comprehensive wellness data for a specified date range
func (c *Client) GetWellnessData(startDate, endDate time.Time) ([]types.WellnessData, error) {
	scheme := "https"
	if strings.HasPrefix(c.Domain, "127.0.0.1") {
		scheme = "http"
	}

	params := url.Values{}
	params.Add("startDate", startDate.Format("2006-01-02"))
	params.Add("endDate", endDate.Format("2006-01-02"))

	wellnessURL := fmt.Sprintf("%s://connectapi.%s/wellness-service/wellness/daily/wellness?%s", scheme, c.Domain, params.Encode())

	req, err := http.NewRequest("GET", wellnessURL, nil)
	if err != nil {
		return nil, &errors.APIError{
			GarthHTTPError: errors.GarthHTTPError{
				GarthError: errors.GarthError{
					Message: "Failed to create wellness data request",
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
					Message: "Failed to get wellness data",
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
					Message: "Wellness data request failed",
				},
			},
		}
	}

	var wellnessData []types.WellnessData
	if err := json.NewDecoder(resp.Body).Decode(&wellnessData); err != nil {
		return nil, &errors.IOError{
			GarthError: errors.GarthError{
				Message: "Failed to parse wellness data",
				Cause:   err,
			},
		}
	}

	return wellnessData, nil
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

// GetDetailedSleepData retrieves comprehensive sleep data for a date
func (c *Client) GetDetailedSleepData(date time.Time) (*types.DetailedSleepData, error) {
	dateStr := date.Format("2006-01-02")
	path := fmt.Sprintf("/wellness-service/wellness/dailySleepData/%s?date=%s&nonSleepBufferMinutes=60",
		c.Username, dateStr)

	data, err := c.ConnectAPI(path, "GET", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get detailed sleep data: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var response struct {
		DailySleepDTO                       *types.DetailedSleepData `json:"dailySleepDTO"`
		SleepMovement                       []types.SleepMovement    `json:"sleepMovement"`
		RemSleepData                        bool                     `json:"remSleepData"`
		SleepLevels                         []types.SleepLevel       `json:"sleepLevels"`
		SleepRestlessMoments                []interface{}            `json:"sleepRestlessMoments"`
		RestlessMomentsCount                int                      `json:"restlessMomentsCount"`
		WellnessSpO2SleepSummaryDTO         interface{}              `json:"wellnessSpO2SleepSummaryDTO"`
		WellnessEpochSPO2DataDTOList        []interface{}            `json:"wellnessEpochSPO2DataDTOList"`
		WellnessEpochRespirationDataDTOList []interface{}            `json:"wellnessEpochRespirationDataDTOList"`
		SleepStress                         interface{}              `json:"sleepStress"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse detailed sleep response: %w", err)
	}

	if response.DailySleepDTO == nil {
		return nil, nil
	}

	// Populate additional data
	response.DailySleepDTO.SleepMovement = response.SleepMovement
	response.DailySleepDTO.SleepLevels = response.SleepLevels

	return response.DailySleepDTO, nil
}

// GetDailyHRVData retrieves comprehensive daily HRV data for a date
func (c *Client) GetDailyHRVData(date time.Time) (*types.DailyHRVData, error) {
	dateStr := date.Format("2006-01-02")
	path := fmt.Sprintf("/wellness-service/wellness/dailyHrvData/%s?date=%s",
		c.Username, dateStr)

	data, err := c.ConnectAPI(path, "GET", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get HRV data: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var response struct {
		HRVSummary  types.DailyHRVData `json:"hrvSummary"`
		HRVReadings []types.HRVReading `json:"hrvReadings"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse HRV response: %w", err)
	}

	// Combine summary and readings
	response.HRVSummary.HRVReadings = response.HRVReadings
	return &response.HRVSummary, nil
}

// GetDetailedBodyBatteryData retrieves comprehensive Body Battery data for a date
func (c *Client) GetDetailedBodyBatteryData(date time.Time) (*types.DetailedBodyBatteryData, error) {
	dateStr := date.Format("2006-01-02")

	// Get main Body Battery data
	path1 := fmt.Sprintf("/wellness-service/wellness/dailyStress/%s", dateStr)
	data1, err := c.ConnectAPI(path1, "GET", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get Body Battery stress data: %w", err)
	}

	// Get Body Battery events
	path2 := fmt.Sprintf("/wellness-service/wellness/bodyBattery/%s", dateStr)
	data2, err := c.ConnectAPI(path2, "GET", nil, nil)
	if err != nil {
		// Events might not be available, continue without them
		data2 = []byte("[]")
	}

	var result types.DetailedBodyBatteryData
	if len(data1) > 0 {
		if err := json.Unmarshal(data1, &result); err != nil {
			return nil, fmt.Errorf("failed to parse Body Battery data: %w", err)
		}
	}

	var events []types.BodyBatteryEvent
	if len(data2) > 0 {
		if err := json.Unmarshal(data2, &events); err == nil {
			result.Events = events
		}
	}

	return &result, nil
}

// GetTrainingStatus retrieves current training status
func (c *Client) GetTrainingStatus(date time.Time) (*types.TrainingStatus, error) {
	dateStr := date.Format("2006-01-02")
	path := fmt.Sprintf("/metrics-service/metrics/trainingStatus/%s", dateStr)

	data, err := c.ConnectAPI(path, "GET", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get training status: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var result types.TrainingStatus
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse training status: %w", err)
	}

	return &result, nil
}

// GetTrainingLoad retrieves training load data
func (c *Client) GetTrainingLoad(date time.Time) (*types.TrainingLoad, error) {
	dateStr := date.Format("2006-01-02")
	endDate := date.AddDate(0, 0, 6).Format("2006-01-02") // Get week of data
	path := fmt.Sprintf("/metrics-service/metrics/trainingLoad/%s/%s", dateStr, endDate)

	data, err := c.ConnectAPI(path, "GET", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get training load: %w", err)
	}

	if len(data) == 0 {
		return nil, nil
	}

	var results []types.TrainingLoad
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, fmt.Errorf("failed to parse training load: %w", err)
	}

	if len(results) == 0 {
		return nil, nil
	}

	return &results[0], nil
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

// RefreshSession refreshes the authentication tokens
func (c *Client) RefreshSession() error {
	// TODO: Implement token refresh logic
	return fmt.Errorf("RefreshSession not implemented")
}
