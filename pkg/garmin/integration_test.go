package garmin_test

import (
	"testing"
	"time"

	"go-garth/internal/api/client"
	"go-garth/internal/data"
	"go-garth/internal/stats"
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
		{"DailySteps", stats.NewDailySteps()},
		{"DailyStress", stats.NewDailyStress()},
		{"DailyHydration", stats.NewDailyHydration()},
		{"DailyIntensityMinutes", stats.NewDailyIntensityMinutes()},
		{"DailySleep", stats.NewDailySleep()},
		{"DailyHRV", stats.NewDailyHRV()},
		{"WeeklySteps", stats.NewWeeklySteps()},
		{"WeeklyStress", stats.NewWeeklyStress()},
		{"WeeklyHRV", stats.NewWeeklyHRV()},
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

func TestPagination(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	c, err := client.NewClient("garmin.com")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = c.LoadSession("test_session.json")
	if err != nil {
		t.Skip("No test session available")
	}

	tests := []struct {
		name   string
		stat   stats.Stats
		period int
	}{
		{"DailySteps_30", stats.NewDailySteps(), 30},
		{"WeeklySteps_60", stats.NewWeeklySteps(), 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			end := time.Now().AddDate(0, 0, -1)
			results, err := tt.stat.List(end, tt.period, c)
			if err != nil {
				t.Errorf("List failed: %v", err)
			}
			if len(results) != tt.period {
				t.Errorf("Expected %d results, got %d", tt.period, len(results))
			}
		})
	}
}
