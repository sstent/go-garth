package garth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestActivityService_List(t *testing.T) {
	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{
			"activityId": 123456789,
			"activityName": "Morning Run",
			"activityType": "running",
			"startTime": "2025-08-29T06:00:00Z",
			"distance": 5000,
			"duration": 1800,
			"calories": 350
		}]`))
	}))
	defer ts.Close()

	// Create client
	apiClient := NewAPIClient(ts.URL, http.DefaultClient)
	activityService := NewActivityService(apiClient)

	// Test List method with filters
	startDate := time.Date(2025, time.August, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, time.August, 31, 0, 0, 0, 0, time.UTC)

	opts := ActivityListOptions{
		Limit:     10,
		StartDate: startDate,
		EndDate:   endDate,
	}
	activities, err := activityService.List(context.Background(), opts)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify activity data
	if len(activities) != 1 {
		t.Fatalf("Expected 1 activity, got %d", len(activities))
	}
	if activities[0].Name != "Morning Run" {
		t.Errorf("Expected activity name 'Morning Run', got '%s'", activities[0].Name)
	}
	if activities[0].ActivityID != 123456789 {
		t.Errorf("Expected activity ID 123456789, got %d", activities[0].ActivityID)
	}
}

func TestActivityService_Get(t *testing.T) {
	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"activityId": 987654321,
			"activityName": "Evening Ride",
			"activityType": "cycling",
			"startTime": "2025-08-29T18:30:00Z",
			"distance": 25000,
			"duration": 3600,
			"calories": 650
		}`))
	}))
	defer ts.Close()

	// Create client
	apiClient := NewAPIClient(ts.URL, http.DefaultClient)
	activityService := NewActivityService(apiClient)

	// Test Get method
	activity, err := activityService.Get(context.Background(), 987654321)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify activity details
	if activity.Name != "Evening Ride" {
		t.Errorf("Expected activity name 'Evening Ride', got '%s'", activity.Name)
	}
	if activity.ActivityID != 987654321 {
		t.Errorf("Expected activity ID 987654321, got %d", activity.ActivityID)
	}
}
