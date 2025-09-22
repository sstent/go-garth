package data

import (
	"testing"
	"time"

	types "github.com/sstent/go-garth/internal/models/types"
	"github.com/sstent/go-garth/shared/interfaces"
	"github.com/sstent/go-garth/shared/models"

	"github.com/stretchr/testify/assert"
)

func TestVO2MaxData_Get(t *testing.T) {
	// Setup
	runningVO2 := 45.0
	cyclingVO2 := 50.0
	settings := &models.UserSettings{
		ID: 12345,
		UserData: models.UserData{
			VO2MaxRunning: &runningVO2,
			VO2MaxCycling: &cyclingVO2,
		},
	}

	vo2Data := NewVO2MaxData()

	// Mock the get function
	vo2Data.GetFunc = func(day time.Time, c interfaces.APIClient) (interface{}, error) {
		vo2Profile := &types.VO2MaxProfile{
			UserProfilePK: settings.ID,
			LastUpdated:   time.Now(),
		}

		if settings.UserData.VO2MaxRunning != nil && *settings.UserData.VO2MaxRunning > 0 {
			vo2Profile.Running = &types.VO2MaxEntry{
				Value:        *settings.UserData.VO2MaxRunning,
				ActivityType: "running",
				Date:         day,
				Source:       "user_settings",
			}
		}

		if settings.UserData.VO2MaxCycling != nil && *settings.UserData.VO2MaxCycling > 0 {
			vo2Profile.Cycling = &types.VO2MaxEntry{
				Value:        *settings.UserData.VO2MaxCycling,
				ActivityType: "cycling",
				Date:         day,
				Source:       "user_settings",
			}
		}
		return vo2Profile, nil
	}

	// Test
	result, err := vo2Data.Get(time.Now(), nil) // client is not used in this mocked get

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, result)

	profile, ok := result.(*types.VO2MaxProfile)
	assert.True(t, ok)
	assert.Equal(t, 12345, profile.UserProfilePK)
	assert.NotNil(t, profile.Running)
	assert.Equal(t, 45.0, profile.Running.Value)
	assert.Equal(t, "running", profile.Running.ActivityType)
	assert.NotNil(t, profile.Cycling)
	assert.Equal(t, 50.0, profile.Cycling.Value)
	assert.Equal(t, "cycling", profile.Cycling.ActivityType)
}
