package garth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Session contains authentication state
type Session struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
	// Add other session fields as needed
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.Expiry)
}

// DefaultSessionPath returns the default session storage path
func DefaultSessionPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".garminconnect", "session.json")
}

// Save saves session to file
func (s *Session) Save(path string) error {
	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// LoadSession loads session from file
func LoadSession(path string) (*Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var session Session
	err = json.Unmarshal(data, &session)
	if err != nil {
		return nil, err
	}

	// Validate session
	if session.AccessToken == "" || session.RefreshToken == "" {
		return nil, os.ErrNotExist
	}

	return &session, nil
}
