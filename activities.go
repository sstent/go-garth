package garth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Activity represents a summary of a Garmin activity
type Activity struct {
	ActivityID int64     `json:"activityId"`
	Name       string    `json:"activityName"`
	Type       string    `json:"activityType"`
	StartTime  time.Time `json:"startTime"`
	Distance   float64   `json:"distance"`
	Duration   float64   `json:"duration"`
	Calories   int       `json:"calories"`
}

// ActivityDetails contains detailed information about an activity
type ActivityDetails struct {
	ActivityID      int64           `json:"activityId"`
	Name            string          `json:"activityName"`
	Description     string          `json:"description"`
	Type            string          `json:"activityType"`
	StartTime       time.Time       `json:"startTime"`
	Distance        float64         `json:"distance"`
	Duration        float64         `json:"duration"`
	Calories        int             `json:"calories"`
	ElevationGain   float64         `json:"elevationGain"`
	ElevationLoss   float64         `json:"elevationLoss"`
	MaxHeartRate    int             `json:"maxHeartRate"`
	AvgHeartRate    int             `json:"avgHeartRate"`
	MaxSpeed        float64         `json:"maxSpeed"`
	AvgSpeed        float64         `json:"avgSpeed"`
	Steps           int             `json:"steps"`
	Stress          int             `json:"stress"`
	TotalSteps      int             `json:"totalSteps"`
	Device          json.RawMessage `json:"device"`
	Location        json.RawMessage `json:"location"`
	Weather         json.RawMessage `json:"weather"`
	HeartRateZones  json.RawMessage `json:"heartRateZones"`
	TrainingEffect  json.RawMessage `json:"trainingEffect"`
	ActivityMetrics json.RawMessage `json:"activityMetrics"`
}

// ActivityListOptions provides filtering options for listing activities
type ActivityListOptions struct {
	Limit        int
	StartDate    time.Time
	EndDate      time.Time
	ActivityType string
	NameContains string
}

// ActivityUpdate represents fields that can be updated on an activity
type ActivityUpdate struct {
	Name        string    `json:"activityName,omitempty"`
	Description string    `json:"description,omitempty"`
	Type        string    `json:"activityType,omitempty"`
	StartTime   time.Time `json:"startTime,omitempty"`
	Distance    float64   `json:"distance,omitempty"`
	Duration    float64   `json:"duration,omitempty"`
}

// ActivityService provides access to activity operations
type ActivityService struct {
	client *APIClient
}

// NewActivityService creates a new ActivityService instance
func NewActivityService(client *APIClient) *ActivityService {
	return &ActivityService{client: client}
}

// List retrieves a list of activities for the current user with optional filters
func (s *ActivityService) List(ctx context.Context, opts ActivityListOptions) ([]Activity, error) {
	params := url.Values{}
	if opts.Limit > 0 {
		params.Set("limit", strconv.Itoa(opts.Limit))
	}
	if !opts.StartDate.IsZero() {
		params.Set("startDate", opts.StartDate.Format(time.RFC3339))
	}
	if !opts.EndDate.IsZero() {
		params.Set("endDate", opts.EndDate.Format(time.RFC3339))
	}
	if opts.ActivityType != "" {
		params.Set("activityType", opts.ActivityType)
	}
	if opts.NameContains != "" {
		params.Set("nameContains", opts.NameContains)
	}

	path := "/activitylist-service/activities/search/activities"
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to get activities list",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to read activities response",
			Cause:      err,
		}
	}

	var activities []Activity
	if err := json.Unmarshal(body, &activities); err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse activities data",
			Cause:      err,
		}
	}

	return activities, nil
}

// Create creates a new activity
func (s *ActivityService) Create(ctx context.Context, activity Activity) (*Activity, error) {
	jsonBody, err := json.Marshal(activity)
	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to marshal activity",
			Cause:      err,
		}
	}

	resp, err := s.client.Post(ctx, "/activity-service/activity", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to create activity",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to read activity response",
			Cause:      err,
		}
	}

	var createdActivity Activity
	if err := json.Unmarshal(body, &createdActivity); err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse activity data",
			Cause:      err,
		}
	}

	return &createdActivity, nil
}

// Update updates an existing activity
func (s *ActivityService) Update(ctx context.Context, activityID int64, update ActivityUpdate) (*Activity, error) {
	jsonBody, err := json.Marshal(update)
	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to marshal activity update",
			Cause:      err,
		}
	}

	path := "/activity-service/activity/" + strconv.FormatInt(activityID, 10)
	resp, err := s.client.Put(ctx, path, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to update activity",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to read activity response",
			Cause:      err,
		}
	}

	var updatedActivity Activity
	if err := json.Unmarshal(body, &updatedActivity); err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse activity data",
			Cause:      err,
		}
	}

	return &updatedActivity, nil
}

// Delete deletes an existing activity
func (s *ActivityService) Delete(ctx context.Context, activityID int64) error {
	path := "/activity-service/activity/" + strconv.FormatInt(activityID, 10)
	resp, err := s.client.Delete(ctx, path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to delete activity",
		}
	}

	return nil
}

// Get retrieves detailed information about a specific activity
func (s *ActivityService) Get(ctx context.Context, activityID int64) (*ActivityDetails, error) {
	path := "/activity-service/activity/" + strconv.FormatInt(activityID, 10)

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to get activity details",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to read activity response",
			Cause:      err,
		}
	}

	var details ActivityDetails
	if err := json.Unmarshal(body, &details); err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse activity data",
			Cause:      err,
		}
	}

	return &details, nil
}

// Export exports an activity in the specified format (gpx, tcx, original)
func (s *ActivityService) Export(ctx context.Context, activityID int64, format string) (io.ReadCloser, error) {
	path := "/download-service/export/" + format + "/activity/" + strconv.FormatInt(activityID, 10)

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to export activity",
		}
	}

	return resp.Body, nil
}
