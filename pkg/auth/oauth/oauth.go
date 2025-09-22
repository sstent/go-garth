package oauth

import (
	"github.com/sstent/go-garth/internal/auth/oauth"
	"github.com/sstent/go-garth/internal/models/types"
	"github.com/sstent/go-garth/pkg/garmin"
)

// GetOAuth1Token retrieves an OAuth1 token using the provided ticket
func GetOAuth1Token(domain, ticket string) (*garmin.OAuth1Token, error) {
	token, err := oauth.GetOAuth1Token(domain, ticket)
	if err != nil {
		return nil, err
	}
	return (*garmin.OAuth1Token)(token), nil
}

// ExchangeToken exchanges an OAuth1 token for an OAuth2 token
func ExchangeToken(oauth1Token *garmin.OAuth1Token) (*garmin.OAuth2Token, error) {
	token, err := oauth.ExchangeToken((*types.OAuth1Token)(oauth1Token))
	if err != nil {
		return nil, err
	}
	return (*garmin.OAuth2Token)(token), nil
}
