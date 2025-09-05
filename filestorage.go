package garth

import (
	"encoding/json"
	"os"
	"sync"
)

// FileStorage implements TokenStorage using a JSON file
type FileStorage struct {
	mu   sync.RWMutex
	path string
}

// NewFileStorage creates a new file-based token storage
func NewFileStorage(path string) *FileStorage {
	return &FileStorage{
		path: path,
	}
}

// GetToken retrieves token from file
func (s *FileStorage) GetToken() (*Token, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrTokenNotFound
		}
		return nil, err
	}

	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

// StoreToken saves token to file
func (s *FileStorage) StoreToken(token *Token) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0600)
}

// ClearToken removes the token file
func (s *FileStorage) ClearToken() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
