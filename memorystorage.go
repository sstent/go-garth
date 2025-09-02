package garth

import (
	"sync"
)

// MemoryStorage implements TokenStorage using an in-memory cache
type MemoryStorage struct {
	mu    sync.RWMutex
	token *Token
}

// NewMemoryStorage creates a new in-memory token storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{}
}

// GetToken retrieves token from memory
func (s *MemoryStorage) GetToken() (*Token, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.token == nil {
		return nil, ErrTokenNotFound
	}
	return s.token, nil
}

// SaveToken saves token to memory
func (s *MemoryStorage) SaveToken(token *Token) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.token = token
	return nil
}
