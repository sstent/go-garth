package interfaces

import (
	"io"
	"net/url"
	"time"

	types "go-garth/internal/models/types"
	"go-garth/shared/models"
)

// APIClient defines the interface for making API calls that data packages need.
type APIClient interface {
	ConnectAPI(path string, method string, params url.Values, body io.Reader) ([]byte, error)
	GetUsername() string
	GetUserSettings() (*models.UserSettings, error)
	GetUserProfile() (*types.UserProfile, error)
	GetWellnessData(startDate, endDate time.Time) ([]types.WellnessData, error)
}
