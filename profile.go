package garth

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// Profile represents a user's Garmin profile
type Profile struct {
	UserID       string `json:"userId"`
	Username     string `json:"username"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	EmailAddress string `json:"emailAddress"`
	Country      string `json:"country"`
	City         string `json:"city"`
	State        string `json:"state"`
	ProfileImage string `json:"profileImage"`
}

// ProfileService provides access to user profile operations
type ProfileService struct {
	client *APIClient
}

// NewProfileService creates a new ProfileService instance
func NewProfileService(client *APIClient) *ProfileService {
	return &ProfileService{client: client}
}

// Get fetches the current user's profile
func (s *ProfileService) Get(ctx context.Context) (*Profile, error) {
	resp, err := s.client.Get(ctx, "/userprofile-service/userprofile")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to get user profile",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to read profile response",
			Cause:      err,
		}
	}

	var profile Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse profile data",
			Cause:      err,
		}
	}

	return &profile, nil
}

// UpdateSettings updates the user's profile settings
func (s *ProfileService) UpdateSettings(ctx context.Context, settings map[string]interface{}) error {
	// Serialize settings to JSON
	jsonData, err := json.Marshal(settings)
	if err != nil {
		return &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to serialize settings",
			Cause:      err,
		}
	}

	// Convert JSON data to a Reader
	reader := bytes.NewReader(jsonData)

	resp, err := s.client.Post(ctx, "/userprofile-service/userprofile/settings", reader)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to update settings",
		}
	}

	return nil
}

// Delete deletes the user's profile
func (s *ProfileService) Delete(ctx context.Context) error {
	resp, err := s.client.Delete(ctx, "/userprofile-service/userprofile", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to delete profile",
		}
	}

	return nil
}

// GetPublic retrieves public profile information for a user
func (s *ProfileService) GetPublic(ctx context.Context, userID string) (*Profile, error) {
	resp, err := s.client.Get(ctx, "/userprofile-service/userprofile/public/"+userID)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "Failed to get public profile",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to read profile response",
			Cause:      err,
		}
	}

	var profile Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		return nil, &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to parse profile data",
			Cause:      err,
		}
	}

	return &profile, nil
}
