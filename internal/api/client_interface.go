package client

import (
	"io"
	"net/url"
	"time"

	types "go-garth/internal/models/types"
	"go-garth/internal/users"
)

// APIClient defines the interface for making API calls
type APIClient interface {
	ConnectAPI(path string, method string, params url.Values, body io.Reader) ([]byte, error)
	GetUserSettings() (*users.UserSettings, error)
	GetUserProfile() (*types.UserProfile, error)
	GetWellnessData(startDate, endDate time.Time) ([]types.WellnessData, error)
}
