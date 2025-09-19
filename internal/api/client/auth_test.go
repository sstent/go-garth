package client_test

import (
	"testing"

	"go-garth/internal/api/client"
	"go-garth/internal/auth/credentials"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Login_Functional(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping functional test in short mode")
	}

	// Load credentials from .env file
	email, password, domain, err := credentials.LoadEnvCredentials()
	require.NoError(t, err, "Failed to load credentials from .env file. Please ensure GARMIN_EMAIL, GARMIN_PASSWORD, and GARMIN_DOMAIN are set.")

	// Create client
	c, err := client.NewClient(domain)
	require.NoError(t, err, "Failed to create client")

	// Perform login
	err = c.Login(email, password)
	require.NoError(t, err, "Login failed")

	// Verify login
	assert.NotEmpty(t, c.AuthToken, "AuthToken should not be empty after login")
	assert.NotEmpty(t, c.Username, "Username should not be empty after login")

	// Logout for cleanup
	err = c.Logout()
	assert.NoError(t, err, "Logout failed")
}