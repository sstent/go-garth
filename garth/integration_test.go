package garth_test

import (
	"testing"
	"time"

	"garmin-connect/garth/client"
	"garmin-connect/garth/data"
)

func TestBodyBatteryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	c, err := client.NewClient("garmin.com")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Load test session
	err = c.LoadSession("test_session.json")
	if err != nil {
		t.Skip("No test session available")
	}

	bb := &data.DailyBodyBatteryStress{}
	result, err := bb.Get(time.Now().AddDate(0, 0, -1), c)

	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if result != nil {
		bbData := result.(*data.DailyBodyBatteryStress)
		if bbData.UserProfilePK == 0 {
			t.Error("UserProfilePK is zero")
		}
	}
}
