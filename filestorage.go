package garth

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
)

// fileStorage implements TokenStorage using a file
type fileStorage struct {
	path string
	auth Authenticator // Reference to authenticator for token refreshes
}

// NewFileStorage creates a new file-based token storage
func NewFileStorage(path string) TokenStorage {
	return &fileStorage{path: path}
}

// SetAuthenticator sets the authenticator for token refreshes
func (s *fileStorage) SetAuthenticator(a Authenticator) {
	s.auth = a
}

// GetToken retrieves token from file, refreshing if expired
func (s *fileStorage) GetToken() (*Token, error) {
	token, err := s.loadToken()
	if err != nil {
		return nil, err
	}

	// Refresh token if expired
	if token.IsExpired() {
		refreshed, err := s.auth.RefreshToken(context.Background(), token.RefreshToken)
		if err != nil {
			return nil, err
		}
		if err := s.SaveToken(refreshed); err != nil {
			return nil, err
		}
		return refreshed, nil
	}
	return token, nil
}

// loadToken loads token from file without refreshing
func (s *fileStorage) loadToken() (*Token, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}

	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	if token.AccessToken == "" || token.RefreshToken == "" {
		return nil, os.ErrNotExist
	}

	return &token, nil
}

// SaveToken saves token to file
func (s *fileStorage) SaveToken(token *Token) error {
	// Create directory if needed
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0600)
}
