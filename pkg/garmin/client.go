package garmin

import (
	internalClient "garmin-connect/internal/api/client"
	"garmin-connect/internal/types"
)

// Client is the main Garmin Connect client type
type Client struct {
	Client *internalClient.Client
}

// NewClient creates a new Garmin Connect client
func NewClient(domain string) (*Client, error) {
	c, err := internalClient.NewClient(domain)
	if err != nil {
		return nil, err
	}
	return &Client{Client: c}, nil
}

// Login authenticates to Garmin Connect
func (c *Client) Login(email, password string) error {
	return c.Client.Login(email, password)
}

// LoadSession loads a session from a file
func (c *Client) LoadSession(filename string) error {
	return c.Client.LoadSession(filename)
}

// SaveSession saves the current session to a file
func (c *Client) SaveSession(filename string) error {
	return c.Client.SaveSession(filename)
}

// GetActivities retrieves recent activities
func (c *Client) GetActivities(limit int) ([]Activity, error) {
	return c.Client.GetActivities(limit)
}

// OAuth1Token returns the OAuth1 token
func (c *Client) OAuth1Token() *types.OAuth1Token {
	return c.Client.OAuth1Token
}

// OAuth2Token returns the OAuth2 token
func (c *Client) OAuth2Token() *types.OAuth2Token {
	return c.Client.OAuth2Token
}
