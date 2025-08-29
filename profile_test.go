package garth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProfileService_Get(t *testing.T) {
	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"userId": "12345",
			"username": "testuser",
			"firstName": "Test",
			"lastName": "User",
			"emailAddress": "test@example.com",
			"country": "US",
			"city": "Seattle",
			"state": "WA",
			"profileImage": "https://example.com/avatar.jpg"
		}`))
	}))
	defer ts.Close()

	// Create client
	apiClient := NewAPIClient(ts.URL, http.DefaultClient)
	profileService := NewProfileService(apiClient)

	// Test Get method
	profile, err := profileService.Get(context.Background())
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify profile data
	if profile.UserID != "12345" {
		t.Errorf("Expected UserID '12345', got '%s'", profile.UserID)
	}
	if profile.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got '%s'", profile.Username)
	}
	if profile.EmailAddress != "test@example.com" {
		t.Errorf("Expected Email 'test@example.com', got '%s'", profile.EmailAddress)
	}
}

func TestProfileService_UpdateSettings(t *testing.T) {
	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	// Create client
	apiClient := NewAPIClient(ts.URL, http.DefaultClient)
	profileService := NewProfileService(apiClient)

	// Test UpdateSettings method
	settings := map[string]interface{}{
		"preferences": map[string]string{
			"units": "metric",
			"theme": "dark",
		},
	}
	err := profileService.UpdateSettings(context.Background(), settings)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
