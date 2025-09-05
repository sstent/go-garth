package garth

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func TestRealAuthentication(t *testing.T) {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		t.Fatalf("Error loading .env file: %v", err)
	}

	// Get credentials from environment
	username := os.Getenv("GARMIN_USERNAME")
	password := os.Getenv("GARMIN_PASSWORD")
	mfaToken := os.Getenv("GARMIN_MFA_TOKEN") // Optional MFA token
	if username == "" || password == "" {
		t.Fatal("GARMIN_USERNAME or GARMIN_PASSWORD not set in .env")
	}

	// Add timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create token storage (using memory storage for this test)
	storage := NewMemoryStorage()

	// Create authenticator
	auth := NewAuthenticator(ClientOptions{
		Storage:  storage,
		TokenURL: "https://connectapi.garmin.com/oauth-service/oauth/token",
		Timeout:  30 * time.Second,
	})

	// Test authentication with and without MFA
	testCases := []struct {
		name     string
		mfaToken string
	}{
		{"Without MFA", ""},
		{"With MFA", mfaToken},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Perform authentication
			token, err := auth.Login(ctx, username, password, tc.mfaToken)
			if err != nil {
				if tc.mfaToken != "" && err.Error() == "MFA required but no token provided" {
					t.Skip("Skipping MFA test since no token provided")
				}
				t.Fatalf("Authentication failed: %v", err)
			}

			log.Printf("Authentication successful! Token details:")
			log.Printf("Access Token: %s", token.OAuth2Token.AccessToken)
			log.Printf("Expires At: %d", token.OAuth2Token.ExpiresAt)
			log.Printf("Refresh Token: %s", token.OAuth2Token.RefreshToken)

			// Verify token storage
			storedToken, err := storage.GetToken()
			if err != nil {
				t.Fatalf("Token storage verification failed: %v", err)
			}
			if storedToken.OAuth2Token.AccessToken != token.OAuth2Token.AccessToken {
				t.Fatal("Stored token doesn't match authenticated token")
			}

			log.Println("Token storage verification successful")

			// Test token refresh
			newToken, err := auth.RefreshToken(ctx, token.OAuth2Token.RefreshToken)
			if err != nil {
				t.Fatalf("Token refresh failed: %v", err)
			}
			if newToken.OAuth2Token.AccessToken == token.OAuth2Token.AccessToken {
				t.Fatal("Refreshed token should be different from original")
			}
			log.Println("Token refresh successful")
		})
	}
}
