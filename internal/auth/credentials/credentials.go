package credentials

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadEnvCredentials loads credentials from .env file
func LoadEnvCredentials() (email, password, domain string, err error) {
	// Determine project root (assuming .env is in the project root)
	projectRoot := "/home/sstent/Projects/go-garth"
	envPath := filepath.Join(projectRoot, ".env")

	// Load .env file
	if err := godotenv.Load(envPath); err != nil {
		return "", "", "", fmt.Errorf("error loading .env file from %s: %w", envPath, err)
	}

	email = os.Getenv("GARMIN_EMAIL")
	password = os.Getenv("GARMIN_PASSWORD")
	domain = os.Getenv("GARMIN_DOMAIN")

	if email == "" {
		return "", "", "", fmt.Errorf("GARMIN_EMAIL not found in .env file")
	}
	if password == "" {
		return "", "", "", fmt.Errorf("GARMIN_PASSWORD not found in .env file")
	}
	if domain == "" {
		domain = "garmin.com" // default value
	}

	return email, password, domain, nil
}