package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"garmin-connect/garth/types"
	"garmin-connect/garth/utils"
)

// GetOAuth1Token retrieves an OAuth1 token using the provided ticket
func GetOAuth1Token(domain, ticket string) (*types.OAuth1Token, error) {
	consumer, err := utils.LoadOAuthConsumer()
	if err != nil {
		return nil, fmt.Errorf("failed to load OAuth consumer: %w", err)
	}

	baseURL := fmt.Sprintf("https://connectapi.%s/oauth-service/oauth/", domain)
	loginURL := fmt.Sprintf("https://sso.%s/sso/embed", domain)
	tokenURL := fmt.Sprintf("%spreauthorized?ticket=%s&login-url=%s&accepts-mfa-tokens=true",
		baseURL, ticket, url.QueryEscape(loginURL))

	// Parse URL to extract query parameters for signing
	parsedURL, err := url.Parse(tokenURL)
	if err != nil {
		return nil, err
	}

	// Extract query parameters
	queryParams := make(map[string]string)
	for key, values := range parsedURL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}

	// Create OAuth1 signed request
	baseURLForSigning := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path
	authHeader := utils.CreateOAuth1AuthorizationHeader("GET", baseURLForSigning, queryParams,
		consumer.ConsumerKey, consumer.ConsumerSecret, "", "")

	req, err := http.NewRequest("GET", tokenURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("User-Agent", "com.garmin.android.apps.connectmobile")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	bodyStr := string(body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("OAuth1 request failed with status %d: %s", resp.StatusCode, bodyStr)
	}

	// Parse query string response - handle both & and ; separators
	bodyStr = strings.ReplaceAll(bodyStr, ";", "&")
	values, err := url.ParseQuery(bodyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OAuth1 response: %w", err)
	}

	oauthToken := values.Get("oauth_token")
	oauthTokenSecret := values.Get("oauth_token_secret")

	if oauthToken == "" || oauthTokenSecret == "" {
		return nil, fmt.Errorf("missing oauth_token or oauth_token_secret in response")
	}

	return &types.OAuth1Token{
		OAuthToken:       oauthToken,
		OAuthTokenSecret: oauthTokenSecret,
		MFAToken:         values.Get("mfa_token"),
		Domain:           domain,
	}, nil
}

// ExchangeToken exchanges an OAuth1 token for an OAuth2 token
func ExchangeToken(oauth1Token *types.OAuth1Token) (*types.OAuth2Token, error) {
	consumer, err := utils.LoadOAuthConsumer()
	if err != nil {
		return nil, fmt.Errorf("failed to load OAuth consumer: %w", err)
	}

	exchangeURL := fmt.Sprintf("https://connectapi.%s/oauth-service/oauth/exchange/user/2.0", oauth1Token.Domain)

	// Prepare form data
	formData := url.Values{}
	if oauth1Token.MFAToken != "" {
		formData.Set("mfa_token", oauth1Token.MFAToken)
	}

	// Convert form data to map for OAuth signing
	formParams := make(map[string]string)
	for key, values := range formData {
		if len(values) > 0 {
			formParams[key] = values[0]
		}
	}

	// Create OAuth1 signed request
	authHeader := utils.CreateOAuth1AuthorizationHeader("POST", exchangeURL, formParams,
		consumer.ConsumerKey, consumer.ConsumerSecret, oauth1Token.OAuthToken, oauth1Token.OAuthTokenSecret)

	req, err := http.NewRequest("POST", exchangeURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", authHeader)
	req.Header.Set("User-Agent", "com.garmin.android.apps.connectmobile")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("OAuth2 exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var oauth2Token types.OAuth2Token
	if err := json.Unmarshal(body, &oauth2Token); err != nil {
		return nil, fmt.Errorf("failed to decode OAuth2 token: %w", err)
	}

	// Set expiration time
	if oauth2Token.ExpiresIn > 0 {
		oauth2Token.ExpiresAt = time.Now().Add(time.Duration(oauth2Token.ExpiresIn) * time.Second)
	}

	return &oauth2Token, nil
}
