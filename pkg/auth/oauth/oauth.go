package oauth

import (
	"github.com/sstent/go-garth/pkg/garmin"
	garthoauth "github.com/sstent/go-garth/pkg/garth/auth/oauth"
)

// GetOAuth1Token retrieves an OAuth1 token using the provided ticket
func GetOAuth1Token(domain, ticket string) (*garmin.OAuth1Token, error) {
	token, err := garthoauth.GetOAuth1Token(domain, ticket)
	if err != nil {
		return nil, err
	}
	return (*garmin.OAuth1Token)(token), nil
}

// ExchangeToken exchanges an OAuth1 token for an OAuth2 token
func ExchangeToken(oauth1Token *garmin.OAuth1Token) (*garmin.OAuth2Token, error) {
	token, err := garthoauth.ExchangeToken(oauth1Token)
	if err != nil {
		return nil, err
	}
	return (*garmin.OAuth2Token)(token), nil
}
