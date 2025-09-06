package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/sstent/go-garth"
)

func main() {
	// Load environment variables from project root
	projectRoot := "../.."
	if err := godotenv.Load(projectRoot + "/.env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get credentials
	username := os.Getenv("GARMIN_USERNAME")
	password := os.Getenv("GARMIN_PASSWORD")
	if username == "" || password == "" {
		log.Fatal("GARMIN_USERNAME or GARMIN_PASSWORD not set in .env")
	}

	// Create storage and authenticator
	storage := garth.NewMemoryStorage()
	auth := garth.NewAuthenticator(garth.ClientOptions{
		Storage:  storage,
		TokenURL: "https://connectapi.garmin.com/oauth-service/oauth/token",
		Timeout:  120,
	})

	// Perform authentication
	fmt.Println("Starting authentication...")
	token, err := auth.Login(context.Background(), username, password, "")
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}

	fmt.Println("\nAuthentication successful! Token details:")
	fmt.Printf("Access Token: %s\n", token.AccessToken)
	fmt.Printf("Expires At: %d\n", token.ExpiresAt)
	fmt.Printf("Refresh Token: %s\n", token.RefreshToken)

	// Verify token storage
	storedToken, err := storage.GetToken()
	if err != nil {
		log.Fatalf("Token storage verification failed: %v", err)
	}
	if storedToken.AccessToken != token.AccessToken {
		log.Fatal("Stored token doesn't match authenticated token")
	}

	fmt.Println("Token storage verification successful")
}
