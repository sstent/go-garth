package client_test

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"testing"
	"time"

	"go-garth/internal/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-garth/internal/api/client"
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
	u, _ := url.Parse(server.URL)
	c, err := client.NewClient(u.Host)
	require.NoError(t, err)
	c.Domain = u.Host
	require.NoError(t, err)
	c.HTTPClient = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	c.AuthToken = "Bearer testtoken"

	// Get user profile
	profile, err := c.GetUserProfile()

	// Verify response
	require.NoError(t, err)
	assert.Equal(t, "testuser", profile.UserName)
	assert.Equal(t, "Test User", profile.DisplayName)
}