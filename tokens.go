package garth

// TokenStorage defines the interface for token persistence
type TokenStorage interface {
	// GetToken retrieves the stored token
	GetToken() (*Token, error)

	// SaveToken stores a new token
	SaveToken(token *Token) error
}
