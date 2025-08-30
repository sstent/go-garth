## Step-by-Step Fix Guide for Junior Developer

### Step 1: Add OAuth1 Token Fetching

```go
// Add to auth.go
func (a *GarthAuthenticator) fetchOAuth1Token(ctx context.Context) (string, error) {
    oauth1URL := "https://connect.garmin.com/oauthConfirm"
    
    req, err := http.NewRequestWithContext(ctx, "GET", oauth1URL, nil)
    if err != nil {
        return "", fmt.Errorf("failed to create OAuth1 request: %w", err)
    }
    
    // Set proper headers
    req.Header.Set("User-Agent", a.userAgent)
    
    resp, err := a.client.Do(req)
    if err != nil {
        return "", fmt.Errorf("OAuth1 request failed: %w", err)
    }
    defer resp.Body.Close()
    
    // Extract oauth_token from response
    // Implementation needed based on Garmin's response format
    
    return "", nil // Replace with actual token extraction
}
```

### Step 2: Fix Login URL Construction

```go
// Replace the simple loginURL with complete parameters
func (a *GarthAuthenticator) buildLoginURL() string {
    params := url.Values{}
    params.Set("service", "https://connect.garmin.com/oauthConfirm")
    params.Set("webhost", "https://connect.garmin.com")
    params.Set("source", "https://connect.garmin.com/signin")
    params.Set("redirectAfterAccountLoginUrl", "https://connect.garmin.com/oauthConfirm")
    params.Set("redirectAfterAccountCreationUrl", "https://connect.garmin.com/oauthConfirm")
    params.Set("gauthHost", "https://sso.garmin.com/sso")
    params.Set("locale", "en_US")
    params.Set("id", "gauth-widget")
    params.Set("clientId", "GarminConnect")
    params.Set("consumeServiceTicket", "false")
    params.Set("generateExtraServiceTicket", "true")
    
    return "https://sso.garmin.com/sso/signin?" + params.Encode()
}
```

### Step 3: Add CSRF Token Extraction

```go
// Add CSRF token extraction to fetchLoginParams
func (a *GarthAuthenticator) fetchLoginParams(ctx context.Context) (lt, execution, csrf string, err error) {
    loginURL := a.buildLoginURL()
    req, err := http.NewRequestWithContext(ctx, "GET", loginURL, nil)
    if err != nil {
        return "", "", "", fmt.Errorf("failed to create login page request: %w", err)
    }

    req.Header = a.getEnhancedBrowserHeaders(loginURL)

    resp, err := a.client.Do(req)
    if err != nil {
        return "", "", "", fmt.Errorf("login page request failed: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", "", "", fmt.Errorf("failed to read login page response: %w", err)
    }

    bodyStr := string(body)
    
    lt, err = extractParam(`name="lt"\s+value="([^"]+)"`, bodyStr)
    if err != nil {
        return "", "", "", fmt.Errorf("lt param not found: %w", err)
    }

    execution, err = extractParam(`name="execution"\s+value="([^"]+)"`, bodyStr)
    if err != nil {
        return "", "", "", fmt.Errorf("execution param not found: %w", err)
    }

    // Add CSRF token extraction
    csrf, err = extractParam(`name="_csrf"\s+value="([^"]+)"`, bodyStr)
    if err != nil {
        return "", "", "", fmt.Errorf("csrf param not found: %w", err)
    }

    return lt, execution, csrf, nil
}
```

### Step 4: Update Authentication Method Signature

```go
// Update method signatures to handle the complete flow
func (a *GarthAuthenticator) Login(ctx context.Context, username, password, mfaToken string) (*Token, error) {
    // Step 1: Fetch OAuth1 token (if not already done)
    oauth1Token, err := a.fetchOAuth1Token(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get OAuth1 token: %w", err)
    }
    
    // Step 2: Get login parameters including CSRF
    lt, execution, csrf, err := a.fetchLoginParams(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get login params: %w", err)
    }

    // Step 3: Perform SSO login
    token, err := a.authenticate(ctx, username, password, mfaToken, lt, execution, csrf)
    if err != nil {
        return nil, err
    }

    // Step 4: Save token to storage
    if err := a.storage.SaveToken(token); err != nil {
        return nil, fmt.Errorf("failed to save token: %w", err)
    }

    return token, nil
}
```

### Step 5: Complete MFA Implementation

```go
// Complete the MFA handling
func (a *GarthAuthenticator) handleMFA(ctx context.Context, username, password, mfaToken, csrf string) (*Token, error) {
    if mfaToken == "" {
        return nil, &AuthError{
            StatusCode: http.StatusPreconditionFailed,
            Message:    "MFA token required but not provided",
            Type:       "mfa_required",
        }
    }

    // Prepare MFA verification data
    data := url.Values{}
    data.Set("username", username)
    data.Set("password", password)
    data.Set("mfaToken", mfaToken)
    data.Set("_csrf", csrf)
    data.Set("embed", "true")
    data.Set("rememberme", "on")
    data.Set("_eventId", "submit")
    // ... add other required fields based on Python version

    mfaURL := "https://sso.garmin.com/sso/verifyMFA" // Adjust URL as needed
    req, err := http.NewRequestWithContext(ctx, "POST", mfaURL, strings.NewReader(data.Encode()))
    if err != nil {
        return nil, &AuthError{
            StatusCode: http.StatusInternalServerError,
            Message:    "Failed to create MFA request",
            Cause:      err,
        }
    }

    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    req.Header.Set("Accept", "application/json")
    req.Header.Set("User-Agent", a.userAgent)

    resp, err := a.client.Do(req)
    if err != nil {
        return nil, &AuthError{
            StatusCode: http.StatusBadGateway,
            Message:    "MFA request failed",
            Cause:      err,
        }
    }
    defer resp.Body.Close()

    // Handle MFA response and extract service ticket
    // Implementation needed based on actual Garmin MFA response format
    
    return nil, nil // Replace with actual token extraction
}
```

## Critical Missing Components

1. **Session State Management**: The Python version maintains session state across requests, while the Go version doesn't properly handle cookies and session continuity.

2. **Error Response Parsing**: The Go version needs to parse Garmin-specific error responses to handle different failure scenarios.

3. **Complete OAuth Flow Integration**: The token exchange process needs to be properly integrated with the SSO login flow.

## Testing Strategy

1. **Start with Manual Testing**: Use tools like Postman or curl to trace the actual Garmin authentication flow
2. **Add Comprehensive Logging**: Log all HTTP requests/responses during development
3. **Create Mock Server Tests**: Build tests that simulate Garmin's responses
4. **Integration Testing**: Test with real credentials in a controlled environment

## Priority Order for Fixes

1. **High Priority**: Complete URL construction and CSRF handling
2. **High Priority**: OAuth1 token fetching integration
3. **Medium Priority**: Complete MFA implementation
4. **Medium Priority**: Proper error handling and response parsing
5. **Low Priority**: Session state management optimization

The key insight is that Garmin's authentication is a multi-step process that requires careful state management and parameter handling. The Python version succeeds because it follows this complete flow, while the Go version shortcuts several critical steps.