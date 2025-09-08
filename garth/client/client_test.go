package client_test

import (
	"net/http"
	"testing"
	"time"

	"garmin-connect/garth/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"garmin-connect/garth/client"
)

func TestClient_GetUserProfile(t *testing.T) {
	// Create mock server returning user profile
	server := testutils.MockJSONResponse(http.StatusOK, `{
		"userName": "testuser",
		"displayName": "Test User",
		"fullName": "Test User",
		"location": "Test Location"
	}`)
	defer server.Close()

	// Create client with test configuration
	c := &client.Client{
		Domain:     server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
		AuthToken:  "Bearer testtoken",
	}

	// Get user profile
	profile, err := c.GetUserProfile()

	// Verify response
	require.NoError(t, err)
	assert.Equal(t, "testuser", profile.UserName)
	assert.Equal(t, "Test User", profile.DisplayName)
}
