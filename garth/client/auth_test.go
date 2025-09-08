package client_test

import (
	"net/http"
	"testing"

	"garmin-connect/garth/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"garmin-connect/garth/client"
	"garmin-connect/garth/errors"
)

func TestClient_Login_Success(t *testing.T) {
	// Create mock SSO server
	ssoServer := testutils.MockJSONResponse(http.StatusOK, `{
		"access_token": "test_token",
		"token_type": "Bearer",
		"expires_in": 3600
	}`)
	defer ssoServer.Close()

	// Create client with test configuration
	c, err := client.NewClient("example.com")
	require.NoError(t, err)
	c.Domain = ssoServer.URL

	// Perform login
	err = c.Login("test@example.com", "password")

	// Verify login
	require.NoError(t, err)
	assert.Equal(t, "Bearer test_token", c.AuthToken)
}

func TestClient_Login_Failure(t *testing.T) {
	// Create mock SSO server returning error
	ssoServer := testutils.MockJSONResponse(http.StatusUnauthorized, `{
		"error": "invalid_credentials"
	}`)
	defer ssoServer.Close()

	// Create client with test configuration
	c, err := client.NewClient("example.com")
	require.NoError(t, err)
	c.Domain = ssoServer.URL

	// Perform login
	err = c.Login("test@example.com", "wrongpassword")

	// Verify error
	require.Error(t, err)
	assert.IsType(t, &errors.AuthenticationError{}, err)
	assert.Contains(t, err.Error(), "SSO login failed")
}
