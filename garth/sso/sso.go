package sso

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"garmin-connect/garth/oauth"
	"garmin-connect/garth/types"
)

var (
	csrfRegex   = regexp.MustCompile(`name="_csrf"\s+value="(.+?)"`)
	titleRegex  = regexp.MustCompile(`<title>(.+?)</title>`)
	ticketRegex = regexp.MustCompile(`embed\?ticket=([^"]+)"`)
)

// MFAContext preserves state for resuming MFA login
type MFAContext struct {
	SigninURL string
	CSRFToken string
	Ticket    string
}

// Client represents an SSO client
type Client struct {
	Domain     string
	HTTPClient *http.Client
}

// NewClient creates a new SSO client
func NewClient(domain string) *Client {
	return &Client{
		Domain:     domain,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Login performs the SSO authentication flow
func (c *Client) Login(email, password string) (*types.OAuth2Token, *MFAContext, error) {
	fmt.Printf("Logging in to Garmin Connect (%s) using SSO flow...\n", c.Domain)

	// Step 1: Set up SSO parameters
	ssoURL := fmt.Sprintf("https://sso.%s/sso", c.Domain)
	ssoEmbedURL := fmt.Sprintf("%s/embed", ssoURL)

	ssoEmbedParams := url.Values{
		"id":          {"gauth-widget"},
		"embedWidget": {"true"},
		"gauthHost":   {ssoURL},
	}

	signinParams := url.Values{
		"id":                              {"gauth-widget"},
		"embedWidget":                     {"true"},
		"gauthHost":                       {ssoEmbedURL},
		"service":                         {ssoEmbedURL},
		"source":                          {ssoEmbedURL},
		"redirectAfterAccountLoginUrl":    {ssoEmbedURL},
		"redirectAfterAccountCreationUrl": {ssoEmbedURL},
	}

	// Step 2: Initialize SSO session
	fmt.Println("Initializing SSO session...")
	embedURL := fmt.Sprintf("https://sso.%s/sso/embed?%s", c.Domain, ssoEmbedParams.Encode())
	req, err := http.NewRequest("GET", embedURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create embed request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize SSO: %w", err)
	}
	resp.Body.Close()

	// Step 3: Get signin page and CSRF token
	fmt.Println("Getting signin page...")
	signinURL := fmt.Sprintf("https://sso.%s/sso/signin?%s", c.Domain, signinParams.Encode())
	req, err = http.NewRequest("GET", signinURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create signin request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Referer", embedURL)

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get signin page: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read signin response: %w", err)
	}

	// Extract CSRF token
	csrfToken := extractCSRFToken(string(body))
	if csrfToken == "" {
		return nil, nil, fmt.Errorf("failed to find CSRF token")
	}
	fmt.Printf("Found CSRF token: %s\n", csrfToken[:10]+"...")

	// Step 4: Submit login form
	fmt.Println("Submitting login credentials...")
	formData := url.Values{
		"username": {email},
		"password": {password},
		"embed":    {"true"},
		"_csrf":    {csrfToken},
	}

	req, err = http.NewRequest("POST", signinURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Referer", signinURL)

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to submit login: %w", err)
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read login response: %w", err)
	}

	// Check login result
	title := extractTitle(string(body))
	fmt.Printf("Login response title: %s\n", title)

	// Handle MFA requirement
	if strings.Contains(title, "MFA") {
		fmt.Println("MFA required - returning context for ResumeLogin")
		ticket := extractTicket(string(body))
		return nil, &MFAContext{
			SigninURL: signinURL,
			CSRFToken: csrfToken,
			Ticket:    ticket,
		}, nil
	}

	if title != "Success" {
		return nil, nil, fmt.Errorf("login failed, unexpected title: %s", title)
	}

	// Step 5: Extract ticket for OAuth flow
	fmt.Println("Extracting OAuth ticket...")
	ticket := extractTicket(string(body))
	if ticket == "" {
		return nil, nil, fmt.Errorf("failed to find OAuth ticket")
	}
	fmt.Printf("Found ticket: %s\n", ticket[:10]+"...")

	// Step 6: Get OAuth1 token
	oauth1Token, err := oauth.GetOAuth1Token(c.Domain, ticket)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get OAuth1 token: %w", err)
	}
	fmt.Println("Got OAuth1 token")

	// Step 7: Exchange for OAuth2 token
	oauth2Token, err := oauth.ExchangeToken(oauth1Token)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to exchange for OAuth2 token: %w", err)
	}
	fmt.Printf("Got OAuth2 token: %s\n", oauth2Token.TokenType)

	return oauth2Token, nil, nil
}

// ResumeLogin completes authentication after MFA challenge
func (c *Client) ResumeLogin(mfaCode string, ctx *MFAContext) (*types.OAuth2Token, error) {
	fmt.Println("Resuming login with MFA code...")

	// Submit MFA form
	formData := url.Values{
		"mfa-code": {mfaCode},
		"embed":    {"true"},
		"_csrf":    {ctx.CSRFToken},
	}

	req, err := http.NewRequest("POST", ctx.SigninURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create MFA request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	req.Header.Set("Referer", ctx.SigninURL)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to submit MFA: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read MFA response: %w", err)
	}

	// Verify MFA success
	title := extractTitle(string(body))
	if title != "Success" {
		return nil, fmt.Errorf("MFA failed, unexpected title: %s", title)
	}

	// Continue with ticket flow
	fmt.Println("Extracting OAuth ticket after MFA...")
	ticket := extractTicket(string(body))
	if ticket == "" {
		return nil, fmt.Errorf("failed to find OAuth ticket after MFA")
	}

	// Get OAuth1 token
	oauth1Token, err := oauth.GetOAuth1Token(c.Domain, ticket)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth1 token: %w", err)
	}

	// Exchange for OAuth2 token
	return oauth.ExchangeToken(oauth1Token)
}

// extractCSRFToken extracts CSRF token from HTML
func extractCSRFToken(html string) string {
	matches := csrfRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractTitle extracts page title from HTML
func extractTitle(html string) string {
	matches := titleRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// extractTicket extracts OAuth ticket from HTML
func extractTicket(html string) string {
	matches := ticketRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
