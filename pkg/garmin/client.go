package garmin

import (
	"fmt"
	"time"

	internalClient "garmin-connect/internal/api/client"
	"garmin-connect/internal/types"
	"garmin-connect/pkg/garmin/activities"
	"garmin-connect/pkg/garmin/health"
	"garmin-connect/pkg/garmin/stats"
)

// Client is the main Garmin Connect client type
type Client struct {
	Client *internalClient.Client
}

// NewClient creates a new Garmin Connect client
func NewClient(domain string) (*Client, error) {
	c, err := internalClient.NewClient(domain)
	if err != nil {
		return nil, err
	}
	return &Client{Client: c},
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

// GetActivities retrieves recent activities
func (c *Client) GetActivities(opts activities.ActivityOptions) ([]Activity, error) {
	// TODO: Map ActivityOptions to internalClient.Client.GetActivities parameters
	// For now, just call the internal client's GetActivities with a dummy limit
	internalActivities, err := c.Client.GetActivities(opts.Limit)
	if err != nil {
		return nil, err
	}

	var garminActivities []Activity
	for _, act := range internalActivities {
		garminActivities = append(garminActivities, Activity{
			ActivityID:   act.ActivityID,
			ActivityName: act.ActivityName,
			ActivityType: act.ActivityType,
			Starttime:    act.Starttime,
			Distance:     act.Distance,
			Duration:     act.Duration,
		})
	}
	return garminActivities, nil
}

// GetActivity retrieves details for a specific activity ID
func (c *Client) GetActivity(activityID int) (*activities.ActivityDetail, error) {
	// TODO: Implement internalClient.Client.GetActivity
	return nil, fmt.Errorf("not implemented")
}

// DownloadActivity downloads activity data
func (c *Client) DownloadActivity(activityID int, opts activities.DownloadOptions) error {
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
func (c *Client) GetSleepData(startDate, endDate time.Time) ([]health.SleepData, error) {
	// TODO: Implement internalClient.Client.GetSleepData
	return nil, fmt.Errorf("not implemented")
}

// GetHrvData retrieves HRV data for a specified number of days
func (c *Client) GetHrvData(days int) ([]health.HrvData, error) {
	// TODO: Implement internalClient.Client.GetHrvData
	return nil, fmt.Errorf("not implemented")
}

// GetStressData retrieves stress data
func (c *Client) GetStressData(startDate, endDate time.Time) ([]health.StressData, error) {
	// TODO: Implement internalClient.Client.GetStressData
	return nil, fmt.Errorf("not implemented")
}

// GetBodyBatteryData retrieves Body Battery data
func (c *Client) GetBodyBatteryData(startDate, endDate time.Time) ([]health.BodyBatteryData, error) {
	// TODO: Implement internalClient.Client.GetBodyBatteryData
	return nil, fmt.Errorf("not implemented")
}

// GetStepsData retrieves steps data for a specified date range
func (c *Client) GetStepsData(startDate, endDate time.Time) ([]stats.StepsData, error) {
	// TODO: Implement internalClient.Client.GetStepsData
	return nil, fmt.Errorf("not implemented")
}

// GetDistanceData retrieves distance data for a specified date range
func (c *Client) GetDistanceData(startDate, endDate time.Time) ([]stats.DistanceData, error) {
	// TODO: Implement internalClient.Client.GetDistanceData
	return nil, fmt.Errorf("not implemented")
}

// GetCaloriesData retrieves calories data for a specified date range
func (c *Client) GetCaloriesData(startDate, endDate time.Time) ([]stats.CaloriesData, error) {
	// TODO: Implement internalClient.Client.GetCaloriesData
	return nil, fmt.Errorf("not implemented")
}

// OAuth1Token returns the OAuth1 token
func (c *Client) OAuth1Token() *types.OAuth1Token {
	return c.Client.OAuth1Token
}

// OAuth2Token returns the OAuth2 token
func (c *Client) OAuth2Token() *types.OAuth2Token {
	return c.Client.OAuth2Token
}
