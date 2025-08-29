package garth

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// fileStorage implements TokenStorage using a file
type fileStorage struct {
	path string
}

// GetToken retrieves token from file
func (s *fileStorage) GetToken() (*Token, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}

	var token Token
	err = json.Unmarshal(data, &token)
	if err != nil {
		return nil, err
	}

	// Check if token is expired
	if token.AccessToken == "" || token.RefreshToken == "" {
		return nil, os.ErrNotExist
	}

	return &token, nil
}

// SaveToken saves token to file
func (s *fileStorage) SaveToken(token *Token) error {
	// Create directory if it doesn't exist
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
