package garmin

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"time"

	internalClient "go-garth/internal/api/client"
	"go-garth/internal/errors"
	types "go-garth/internal/models/types"
	shared "go-garth/shared/interfaces"
	models "go-garth/shared/models"
)

// Client is the main Garmin Connect client type
type Client struct {
	Client *internalClient.Client
}

var _ shared.APIClient = (*Client)(nil)

// NewClient creates a new Garmin Connect client
func NewClient(domain string) (*Client, error) {
	c, err := internalClient.NewClient(domain)
	if err != nil {
		return nil, err
	}
	return &Client{Client: c}, nil
}

func (c *Client) InternalClient() *internalClient.Client {
	return c.Client
}

// ConnectAPI implements the APIClient interface
func (c *Client) ConnectAPI(path string, method string, params url.Values, body io.Reader) ([]byte, error) {
	return c.Client.ConnectAPI(path, method, params, body)
}

// GetUsername implements the APIClient interface
func (c *Client) GetUsername() string {
	return c.Client.GetUsername()
}

// GetUserSettings implements the APIClient interface
func (c *Client) GetUserSettings() (*models.UserSettings, error) {
	return c.Client.GetUserSettings()
}

// GetUserProfile implements the APIClient interface
func (c *Client) GetUserProfile() (*types.UserProfile, error) {
	return c.Client.GetUserProfile()
}

// GetWellnessData implements the APIClient interface
func (c *Client) GetWellnessData(startDate, endDate time.Time) ([]types.WellnessData, error) {
	return c.Client.GetWellnessData(startDate, endDate)
}

// Login authenticates to Garmin Connect
func (c *Client) Login(email, password string) error {
	return c.Client.Login(email, password)
}

// LoadSession loads a session from a file
func (c *Client) LoadSession(filename string) error {
	return c.Client.LoadSession(filename)
}

// SaveSession saves the current session to a file
func (c *Client) SaveSession(filename string) error {
	return c.Client.SaveSession(filename)
}

// RefreshSession refreshes the authentication tokens
func (c *Client) RefreshSession() error {
	return c.Client.RefreshSession()
}

// ListActivities retrieves recent activities
func (c *Client) ListActivities(opts ActivityOptions) ([]Activity, error) {
	// TODO: Map ActivityOptions to internalClient.Client.GetActivities parameters
	// For now, just call the internal client's GetActivities with a dummy limit
	internalActivities, err := c.Client.GetActivities(opts.Limit)
	if err != nil {
		return nil, err
	}

	var garminActivities []Activity
	for _, act := range internalActivities {
		garminActivities = append(garminActivities, Activity{
			ActivityID:     act.ActivityID,
			ActivityName:   act.ActivityName,
			ActivityType:   act.ActivityType,
			StartTimeLocal: act.StartTimeLocal,
			Distance:       act.Distance,
			Duration:       act.Duration,
		})
	}
	return garminActivities, nil
}

// GetActivity retrieves details for a specific activity ID
func (c *Client) GetActivity(activityID int) (*ActivityDetail, error) {
	// TODO: Implement internalClient.Client.GetActivity
	return nil, fmt.Errorf("not implemented")
}

// DownloadActivity downloads activity data
func (c *Client) DownloadActivity(activityID int, opts DownloadOptions) error {
	// TODO: Determine file extension based on format
	fileExtension := opts.Format
	if fileExtension == "csv" {
		fileExtension = "csv"
	} else if fileExtension == "gpx" {
		fileExtension = "gpx"
	} else if fileExtension == "tcx" {
		fileExtension = "tcx"
	} else {
		return fmt.Errorf("unsupported download format: %s", opts.Format)
	}

	// Construct filename
	filename := fmt.Sprintf("%d.%s", activityID, fileExtension)
	if opts.Filename != "" {
		filename = opts.Filename
	}

	// Construct output path
	outputPath := filename
	if opts.OutputDir != "" {
		outputPath = filepath.Join(opts.OutputDir, filename)
	}

	err := c.Client.Download(fmt.Sprintf("%d", activityID), opts.Format, outputPath)
	if err != nil {
		return err
	}

	// Basic validation: check if file is empty
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return &errors.IOError{
			GarthError: errors.GarthError{
				Message: "Failed to get file info after download",
				Cause:   err,
			},
		}
	}
	if fileInfo.Size() == 0 {
		return &errors.IOError{
			GarthError: errors.GarthError{
				Message: "Downloaded file is empty",
			},
		}
	}

	return nil
}

// SearchActivities searches for activities by a query string
func (c *Client) SearchActivities(query string) ([]Activity, error) {
	// TODO: Implement internalClient.Client.SearchActivities
	return nil, fmt.Errorf("not implemented")
}

// GetSleepData retrieves sleep data for a specified date range
func (c *Client) GetSleepData(date time.Time) (*types.DetailedSleepData, error) {
	return c.Client.GetDetailedSleepData(date)
}

// GetHrvData retrieves HRV data for a specified number of days
func (c *Client) GetHrvData(date time.Time) (*types.DailyHRVData, error) {
	return c.Client.GetDailyHRVData(date)
}

// GetStressData retrieves stress data
func (c *Client) GetStressData(startDate, endDate time.Time) ([]types.StressData, error) {
	return c.Client.GetStressData(startDate, endDate)
}

// GetBodyBatteryData retrieves Body Battery data
func (c *Client) GetBodyBatteryData(date time.Time) (*types.DetailedBodyBatteryData, error) {
	return c.Client.GetDetailedBodyBatteryData(date)
}

// GetStepsData retrieves steps data for a specified date range
func (c *Client) GetStepsData(startDate, endDate time.Time) ([]types.StepsData, error) {
	return c.Client.GetStepsData(startDate, endDate)
}

// GetDistanceData retrieves distance data for a specified date range
func (c *Client) GetDistanceData(startDate, endDate time.Time) ([]types.DistanceData, error) {
	return c.Client.GetDistanceData(startDate, endDate)
}

// GetCaloriesData retrieves calories data for a specified date range
func (c *Client) GetCaloriesData(startDate, endDate time.Time) ([]types.CaloriesData, error) {
	return c.Client.GetCaloriesData(startDate, endDate)
}

// GetVO2MaxData retrieves VO2 max data for a specified date range
func (c *Client) GetVO2MaxData(startDate, endDate time.Time) ([]types.VO2MaxData, error) {
	return c.Client.GetVO2MaxData(startDate, endDate)
}

// GetHeartRateZones retrieves heart rate zone data
func (c *Client) GetHeartRateZones() (*types.HeartRateZones, error) {
	return c.Client.GetHeartRateZones()
}

// GetTrainingStatus retrieves current training status
func (c *Client) GetTrainingStatus(date time.Time) (*types.TrainingStatus, error) {
	return c.Client.GetTrainingStatus(date)
}

// GetTrainingLoad retrieves training load data
func (c *Client) GetTrainingLoad(date time.Time) (*types.TrainingLoad, error) {
	return c.Client.GetTrainingLoad(date)
}

// GetFitnessAge retrieves fitness age calculation
func (c *Client) GetFitnessAge() (*types.FitnessAge, error) {
	// TODO: Implement GetFitnessAge in internalClient.Client
	return nil, fmt.Errorf("GetFitnessAge not implemented in internalClient.Client")
}

// OAuth1Token returns the OAuth1 token
func (c *Client) OAuth1Token() *types.OAuth1Token {
	return c.Client.OAuth1Token
}

// OAuth2Token returns the OAuth2 token
func (c *Client) OAuth2Token() *types.OAuth2Token {
	return c.Client.OAuth2Token
}
