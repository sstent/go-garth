package garth_test

import (
	"testing"
	"time"

	"garmin-connect/garth/client"
	"garmin-connect/garth/data"
	"garmin-connect/garth/stats"
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

func TestStatsEndpoints(t *testing.T) {
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

	tests := []struct {
		name string
		stat stats.Stats
	}{
		{"Steps", stats.NewDailySteps()},
		{"Stress", stats.NewDailyStress()},
		{"Hydration", stats.NewDailyHydration()},
		{"IntensityMinutes", stats.NewDailyIntensityMinutes()},
		{"Sleep", stats.NewDailySleep()},
		{"HRV", stats.NewDailyHRV()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			end := time.Now().AddDate(0, 0, -1)
			results, err := tt.stat.List(end, 1, c)
			if err != nil {
				t.Errorf("List failed: %v", err)
			}
			if len(results) == 0 {
				t.Logf("No data returned for %s", tt.name)
				return
			}

			// Basic validation that we got some data
			resultMap, ok := results[0].(map[string]interface{})
			if !ok {
				t.Errorf("Expected map for %s result, got %T", tt.name, results[0])
				return
			}

			if len(resultMap) == 0 {
				t.Errorf("Empty result map for %s", tt.name)
			}
		})
	}
}
