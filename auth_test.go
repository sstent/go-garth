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
	if username == "" || password == "" {
		t.Fatal("GARMIN_USERNAME or GARMIN_PASSWORD not set in .env")
	}

	// Add timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create token storage (using memory storage for this test)
	storage := NewMemoryStorage()

	// Create authenticator
	auth := NewAuthenticator(ClientOptions{
		Storage:  storage,
		TokenURL: "https://connectapi.garmin.com/oauth-service/oauth/token",
		Timeout:  30 * time.Second,
	})

	// Perform authentication with timeout context
	token, err := auth.Login(ctx, username, password, "")
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}

	log.Printf("Authentication successful! Token details:")
	log.Printf("Access Token: %s", token.AccessToken)
	log.Printf("Expires: %s", token.Expiry.Format(time.RFC3339))
	log.Printf("Refresh Token: %s", token.RefreshToken)

	// Verify token storage
	storedToken, err := storage.GetToken()
	if err != nil {
		t.Fatalf("Token storage verification failed: %v", err)
	}
	if storedToken.AccessToken != token.AccessToken {
		t.Fatal("Stored token doesn't match authenticated token")
	}

	log.Println("Token storage verification successful")
}
