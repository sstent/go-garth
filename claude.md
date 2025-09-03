// Fixed authentication flow based on Python garth implementation

// fetchOAuth1Token should be the first step and get the initial OAuth1 token
func (a *GarthAuthenticator) fetchOAuth1Token(ctx context.Context) (string, error) {
    // Step 1: Initial OAuth1 request - this should NOT have parameters initially
    oauth1URL := "https://connect.garmin.com/oauthConfirm"

    req, err := http.NewRequestWithContext(ctx, "GET", oauth1URL, nil)
    if err != nil {
        return "", fmt.Errorf("failed to create OAuth1 request: %w", err)
    }

    // Set proper headers
    req.Header.Set("User-Agent", a.userAgent)
    req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
    
    resp, err := a.client.Do(req)
    if err != nil {
        return "", fmt.Errorf("OAuth1 request failed: %w", err)
    }
    defer resp.Body.Close()

    // Handle redirect case - OAuth1 token often comes from redirect location
    if resp.StatusCode >= 300 && resp.StatusCode < 400 {
        location := resp.Header.Get("Location")
        if location != "" {
            if u, err := url.Parse(location); err == nil {
                if token := u.Query().Get("oauth_token"); token != "" {
                    return token, nil
                }
            }
        }
    }

    // If no redirect, parse response body
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("failed to read OAuth1 response: %w", err)
    }

    // Look for oauth_token in various formats
    patterns := []string{
        `oauth_token=([^&\s"']+)`,
        `"oauth_token":\s*"([^"]+)"`,
        `'oauth_token':\s*'([^']+)'`,
        `oauth_token["']?\s*[:=]\s*["']?([^"'\s&]+)`,
    }

    for _, pattern := range patterns {
        re := regexp.MustCompile(pattern)
        if matches := re.FindStringSubmatch(string(body)); len(matches) > 1 {
            return matches[1], nil
        }
    }

    // Debug: save response for analysis
    filename := fmt.Sprintf("oauth1_response_%d.html", time.Now().Unix())
    if writeErr := os.WriteFile(filename, body, 0644); writeErr == nil {
        return "", fmt.Errorf("OAuth1 token not found (response saved to %s)", filename)
    }

    return "", fmt.Errorf("OAuth1 token not found in response")
}

// Updated Login method with correct flow
func (a *GarthAuthenticator) Login(ctx context.Context, username, password, mfaToken string) (*Token, error) {
    // Step 1: Get OAuth1 token FIRST
    oauth1Token, err := a.fetchOAuth1Token(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get OAuth1 token: %w", err)
    }
    a.oauth1Token = oauth1Token

    // Step 2: Now get login parameters with OAuth1 context
    authToken, tokenType, err := a.fetchLoginParamsWithOAuth1(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get login params: %w", err)
    }

    a.csrfToken = authToken

    // Step 3: Authenticate with all tokens
    token, err := a.authenticate(ctx, username, password, mfaToken, authToken, tokenType)
    if err != nil {
        return nil, err
    }

    // Save token to storage
    if err := a.storage.SaveToken(token); err != nil {
        return nil, fmt.Errorf("failed to save token: %w", err)
    }

    return token, nil
}

// New method to fetch login params with OAuth1 context
func (a *GarthAuthenticator) fetchLoginParamsWithOAuth1(ctx context.Context) (token string, tokenType string, err error) {
    // Build login URL with OAuth1 context
    params := url.Values{}
    params.Set("id", "gauth-widget")
    params.Set("embedWidget", "true")
    params.Set("gauthHost", "https://sso.garmin.com/sso")
    params.Set("service", "https://connect.garmin.com/oauthConfirm")
    params.Set("source", "https://sso.garmin.com/sso")
    params.Set("redirectAfterAccountLoginUrl", "https://connect.garmin.com/oauthConfirm")
    params.Set("redirectAfterAccountCreationUrl", "https://connect.garmin.com/oauthConfirm")
    params.Set("consumeServiceTicket", "false")
    params.Set("generateExtraServiceTicket", "true")
    params.Set("clientId", "GarminConnect")
    params.Set("locale", "en_US")
    
    // Add OAuth1 token if we have it
    if a.oauth1Token != "" {
        params.Set("oauth_token", a.oauth1Token)
    }

    loginURL := "https://sso.garmin.com/sso/signin?" + params.Encode()

    req, err := http.NewRequestWithContext(ctx, "GET", loginURL, nil)
    if err != nil {
        return "", "", fmt.Errorf("failed to create login page request: %w", err)
    }

    // Set headers with proper referrer chain
    req.Header.Set("User-Agent", a.userAgent)
    req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
    req.Header.Set("Accept-Language", "en-US,en;q=0.9")
    req.Header.Set("Accept-Encoding", "gzip, deflate, br")
    req.Header.Set("Referer", "https://connect.garmin.com/oauthConfirm")
    req.Header.Set("Origin", "https://connect.garmin.com")

    resp, err := a.client.Do(req)
    if err != nil {
        return "", "", fmt.Errorf("login page request failed: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", "", fmt.Errorf("failed to read login page response: %w", err)
    }

    bodyStr := string(body)

    // Extract CSRF/lt token
    token, tokenType, err = getCSRFTokenWithType(bodyStr)
    if err != nil {
        // Save for debugging
        filename := fmt.Sprintf("login_page_oauth1_%d.html", time.Now().Unix())
        if writeErr := os.WriteFile(filename, body, 0644); writeErr == nil {
            return "", "", fmt.Errorf("authentication token not found with OAuth1 context: %w (HTML saved to %s)", err, filename)
        }
        return "", "", fmt.Errorf("authentication token not found with OAuth1 context: %w", err)
    }

    return token, tokenType, nil
}

// Enhanced authentication method
func (a *GarthAuthenticator) authenticate(ctx context.Context, username, password, mfaToken, authToken, tokenType string) (*Token, error) {
    data := url.Values{}
    data.Set("username", username)
    data.Set("password", password)
    data.Set("embed", "true")
    data.Set("rememberme", "on")
    
    // Set the correct token field based on type
    if tokenType == "lt" {
        data.Set("lt", authToken)
    } else {
        data.Set("_csrf", authToken)
    }
    
    data.Set("_eventId", "submit")
    data.Set("geolocation", "")
    data.Set("clientId", "GarminConnect")
    data.Set("service", "https://connect.garmin.com/oauthConfirm") // This should match OAuth1 context
    data.Set("webhost", "https://connect.garmin.com")
    data.Set("fromPage", "oauth")
    data.Set("locale", "en_US")
    data.Set("id", "gauth-widget")
    data.Set("redirectAfterAccountLoginUrl", "https://connect.garmin.com/oauthConfirm")
    data.Set("redirectAfterAccountCreationUrl", "https://connect.garmin.com/oauthConfirm")

    loginURL := "https://sso.garmin.com/sso/signin"
    req, err := http.NewRequestWithContext(ctx, "POST", loginURL, strings.NewReader(data.Encode()))
    if err != nil {
        return nil, fmt.Errorf("failed to create SSO request: %w", err)
    }

    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("Accept", "application/json")
    req.Header.Set("User-Agent", a.userAgent)
    req.Header.Set("Referer", "https://sso.garmin.com/sso/signin?id=gauth-widget&embedWidget=true&gauthHost=https://sso.garmin.com/sso")

    resp, err := a.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("SSO request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusPreconditionFailed {
        body, _ := io.ReadAll(resp.Body)
        return a.handleMFA(ctx, username, password, mfaToken, string(body))
    }

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("authentication failed with status: %d, response: %s", resp.StatusCode, body)
    }

    var authResponse struct {
        Ticket string `json:"ticket"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
        return nil, fmt.Errorf("failed to parse SSO response: %w", err)
    }

    if authResponse.Ticket == "" {
        return nil, errors.New("empty ticket in SSO response")
    }

    return a.exchangeTicketForToken(ctx, authResponse.Ticket)
}