package data

import (
	"fmt"
	"time"

	shared "go-garth/shared/interfaces"
	types "go-garth/internal/models/types"
)

// VO2MaxData implements the Data interface for VO2 max retrieval
type VO2MaxData struct {
	shared.BaseData
}

// NewVO2MaxData creates a new VO2MaxData instance
func NewVO2MaxData() *VO2MaxData {
	vo2 := &VO2MaxData{}
	vo2.GetFunc = vo2.get
	return vo2
}

// get implements the specific VO2 max data retrieval logic
func (v *VO2MaxData) get(day time.Time, c shared.APIClient) (interface{}, error) {
	// Primary approach: Get from user settings (most reliable)
	settings, err := c.GetUserSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	// Extract VO2 max data from user settings
	vo2Profile := &types.VO2MaxProfile{
		UserProfilePK: settings.ID,
		LastUpdated:   time.Now(),
	}

	// Add running VO2 max if available
	if settings.UserData.VO2MaxRunning != nil && *settings.UserData.VO2MaxRunning > 0 {
		vo2Profile.Running = &types.VO2MaxEntry{
			Value:        *settings.UserData.VO2MaxRunning,
			ActivityType: "running",
			Date:         day,
			Source:       "user_settings",
		}
	}

	// Add cycling VO2 max if available
	if settings.UserData.VO2MaxCycling != nil && *settings.UserData.VO2MaxCycling > 0 {
		vo2Profile.Cycling = &types.VO2MaxEntry{
			Value:        *settings.UserData.VO2MaxCycling,
			ActivityType: "cycling",
			Date:         day,
			Source:       "user_settings",
		}
	}

	// If no VO2 max data found, still return valid empty profile
	return vo2Profile, nil
}

// List implements concurrent fetching for multiple days
// Note: VO2 max typically doesn't change daily, so this returns the same values
func (v *VO2MaxData) List(end time.Time, days int, c shared.APIClient, maxWorkers int) ([]interface{}, []error) {
	// For VO2 max, we want current values from user settings
	vo2Data, err := v.get(end, c)
	if err != nil {
		return nil, []error{err}
	}

	// Return the same VO2 max data for all requested days
	results := make([]interface{}, days)
	for i := 0; i < days; i++ {
		results[i] = vo2Data
	}

	return results, nil
}

// GetCurrentVO2Max is a convenience method to get current VO2 max values
func GetCurrentVO2Max(c shared.APIClient) (*types.VO2MaxProfile, error) {
	vo2Data := NewVO2MaxData()
	result, err := vo2Data.get(time.Now(), c)
	if err != nil {
		return nil, err
	}

	vo2Profile, ok := result.(*types.VO2MaxProfile)
	if !ok {
		return nil, fmt.Errorf("unexpected result type")
	}

	return vo2Profile, nil
}
