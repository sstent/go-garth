package interfaces

import (
	"io"
	"net/url"
	"time"

	garth "github.com/sstent/go-garth/pkg/garth/types"
	"github.com/sstent/go-garth/shared/models"
)

// APIClient defines the interface for making API calls that data packages need.
type APIClient interface {
	ConnectAPI(path string, method string, params url.Values, body io.Reader) ([]byte, error)
	GetUsername() string
	GetUserSettings() (*models.UserSettings, error)
	GetUserProfile() (*garth.UserProfile, error)
	GetWellnessData(startDate, endDate time.Time) ([]garth.WellnessData, error)
}
