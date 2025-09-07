# Implementation Plan for Steps 1 & 2: Project Structure and Client Refactoring

## Overview
This document provides a detailed implementation plan for refactoring the existing Go code from `main.go` into a proper modular structure as outlined in the porting plan.

## Current State Analysis

### Existing Code in main.go (Lines 1-761)
The current `main.go` contains:
- **Client struct** (lines 24-30) with domain, httpClient, username, authToken
- **Data models**: SessionData, ActivityType, EventType, Activity, OAuth1Token, OAuth2Token, OAuthConsumer
- **OAuth functions**: loadOAuthConsumer, generateNonce, generateTimestamp, percentEncode, createSignatureBaseString, createSigningKey, signRequest, createOAuth1AuthorizationHeader
- **SSO functions**: getCSRFToken, extractTicket, exchangeOAuth1ForOAuth2, Login, loadEnvCredentials
- **Client methods**: NewClient, getUserProfile, GetActivities, SaveSession, LoadSession
- **Main function** with authentication flow and activity retrieval

## Step 1: Project Structure Setup

### Directory Structure to Create
```
garmin-connect/
├── client/
│   ├── client.go      # Core client logic
│   ├── auth.go        # Authentication handling
│   └── sso.go         # SSO authentication
├── data/
│   └── base.go        # Base data models and interfaces
├── types/
│   └── tokens.go      # Token structures
├── utils/
│   └── utils.go       # Utility functions
├── errors/
│   └── errors.go      # Custom error types
├── cmd/
│   └── garth/
│       └── main.go    # CLI tool (refactored from current main.go)
└── main.go            # Keep original temporarily for testing
```

## Step 2: Core Client Refactoring - Detailed Implementation

### 2.1 Create `types/tokens.go`
**Purpose**: Centralize all token-related structures

```go
package types

import "time"

// OAuth1Token represents OAuth1 token response
type OAuth1Token struct {
    OAuthToken       string `json:"oauth_token"`
    OAuthTokenSecret string `json:"oauth_token_secret"`
    MFAToken         string `json:"mfa_token,omitempty"`
    Domain           string `json:"domain"`
}

// OAuth2Token represents OAuth2 token response
type OAuth2Token struct {
    AccessToken  string    `json:"access_token"`
    TokenType    string    `json:"token_type"`
    ExpiresIn    int       `json:"expires_in"`
    RefreshToken string    `json:"refresh_token"`
    Scope        string    `json:"scope"`
    CreatedAt    time.Time // Added for expiration tracking
}

// OAuthConsumer represents OAuth consumer credentials
type OAuthConsumer struct {
    ConsumerKey    string `json:"consumer_key"`
    ConsumerSecret string `json:"consumer_secret"`
}

// SessionData represents saved session information
type SessionData struct {
    Domain    string `json:"domain"`
    Username  string `json:"username"`
    AuthToken string `json:"auth_token"`
}
```

### 2.2 Create `client/client.go`
**Purpose**: Core client functionality and HTTP operations

```go
package client

import (
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha1"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/http/cookiejar"
    "net/url"
    "os"
    "regexp"
    "sort"
    "strconv"
    "strings"
    "time"
    
    "github.com/joho/godotenv"
    "garmin-connect/types"
)

// Client represents the Garmin Connect client
type Client struct {
    domain      string
    httpClient  *http.Client
    username    string
    authToken   string
    oauth1Token *types.OAuth1Token
    oauth2Token *types.OAuth2Token
}

// ConfigOption represents a client configuration option
type ConfigOption func(*Client)

// NewClient creates a new Garmin Connect client
func NewClient(domain string) (*Client, error) {
    jar, err := cookiejar.New(nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create cookie jar: %w", err)
    }
    
    return &Client{
        domain: domain,
        httpClient: &http.Client{
            Jar:     jar,
            Timeout: 30 * time.Second,
        },
    }, nil
}

// Configure applies configuration options to the client
func (c *Client) Configure(opts ...ConfigOption) error {
    for _, opt := range opts {
        opt(c)
    }
    return nil
}

// ConnectAPI makes authenticated API calls to Garmin Connect
func (c *Client) ConnectAPI(path, method string, data interface{}) (interface{}, error) {
    // Implementation based on Python http.py Client.connectapi()
    // Should handle authentication, retries, and error responses
}

// Download downloads data from Garmin Connect
func (c *Client) Download(path string) ([]byte, error) {
    // Implementation for downloading files/data
}

// Upload uploads data to Garmin Connect
func (c *Client) Upload(filePath, uploadPath string) (map[string]interface{}, error) {
    // Implementation for uploading files/data
}

// GetUserProfile retrieves the current user's profile
func (c *Client) GetUserProfile() error {
    // Extracted from main.go getUserProfile method
}

// GetActivities retrieves recent activities
func (c *Client) GetActivities(limit int) ([]Activity, error) {
    // Extracted from main.go GetActivities method
}

// SaveSession saves the current session to a file
func (c *Client) SaveSession(filename string) error {
    // Extracted from main.go SaveSession method
}

// LoadSession loads a session from a file
func (c *Client) LoadSession(filename string) error {
    // Extracted from main.go LoadSession method
}
```

### 2.3 Create `client/auth.go`
**Purpose**: Authentication and token management

```go
package client

import (
    "crypto/hmac"
    "crypto/rand"
    "crypto/sha1"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "sort"
    "strconv"
    "strings"
    "time"
    
    "garmin-connect/types"
)

var oauthConsumer *types.OAuthConsumer

// loadOAuthConsumer loads OAuth consumer credentials
func loadOAuthConsumer() (*types.OAuthConsumer, error) {
    // Extracted from main.go loadOAuthConsumer function
}

// OAuth1 signing functions (extract from main.go)
func generateNonce() string
func generateTimestamp() string
func percentEncode(s string) string
func createSignatureBaseString(method, baseURL string, params map[string]string) string
func createSigningKey(consumerSecret, tokenSecret string) string
func signRequest(consumerSecret, tokenSecret, baseString string) string
func createOAuth1AuthorizationHeader(method, requestURL string, params map[string]string, consumerKey, consumerSecret, token, tokenSecret string) string

// Token expiration checking
func (t *types.OAuth2Token) IsExpired() bool {
    return time.Since(t.CreatedAt) > time.Duration(t.ExpiresIn)*time.Second
}

// MFA support placeholder
func (c *Client) HandleMFA(mfaToken string) error {
    // Placeholder for MFA handling
    return fmt.Errorf("MFA not yet implemented")
}
```

### 2.4 Create `client/sso.go`
**Purpose**: SSO authentication flow

```go
package client

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"
    "os"
    "regexp"
    "strings"
    
    "github.com/joho/godotenv"
    "garmin-connect/types"
)

var (
    csrfRegex   = regexp.MustCompile(`name="_csrf"\s+value="(.+?)"`)
    titleRegex  = regexp.MustCompile(`<title>(.+?)</title>`)
    ticketRegex = regexp.MustCompile(`embed\?ticket=([^"]+)"`)
)

// Login performs SSO login with email and password
func (c *Client) Login(email, password string) error {
    // Extracted from main.go Login method
}

// ResumeLogin resumes login after MFA
func (c *Client) ResumeLogin(mfaToken string) error {
    // New method for MFA completion
}

// SSO helper functions (extract from main.go)
func getCSRFToken(respBody string) string
func extractTicket(respBody string) string
func exchangeOAuth1ForOAuth2(oauth1Token *types.OAuth1Token, domain string) (*types.OAuth2Token, error)
func loadEnvCredentials() (email, password, domain string, err error)
```

### 2.5 Create `data/base.go`
**Purpose**: Base data models and interfaces

```go
package data

import (
    "time"
    "garmin-connect/client"
)

// ActivityType represents the type of activity
type ActivityType struct {
    TypeID       int    `json:"typeId"`
    TypeKey      string `json:"typeKey"`
    ParentTypeID *int   `json:"parentTypeId,omitempty"`
}

// EventType represents the event type of an activity
type EventType struct {
    TypeID  int    `json:"typeId"`
    TypeKey string `json:"typeKey"`
}

// Activity represents a Garmin Connect activity
type Activity struct {
    ActivityID      int64        `json:"activityId"`
    ActivityName    string       `json:"activityName"`
    Description     string       `json:"description"`
    StartTimeLocal  string       `json:"startTimeLocal"`
    StartTimeGMT    string       `json:"startTimeGMT"`
    ActivityType    ActivityType `json:"activityType"`
    EventType       EventType    `json:"eventType"`
    Distance        float64      `json:"distance"`
    Duration        float64      `json:"duration"`
    ElapsedDuration float64      `json:"elapsedDuration"`
    MovingDuration  float64      `json:"movingDuration"`
    ElevationGain   float64      `json:"elevationGain"`
    ElevationLoss   float64      `json:"elevationLoss"`
    AverageSpeed    float64      `json:"averageSpeed"`
    MaxSpeed        float64      `json:"maxSpeed"`
    Calories        float64      `json:"calories"`
    AverageHR       float64      `json:"averageHR"`
    MaxHR           float64      `json:"maxHR"`
}

// Data interface for all data models
type Data interface {
    Get(day time.Time, client *client.Client) (interface{}, error)
    List(end time.Time, days int, client *client.Client, maxWorkers int) ([]interface{}, error)
}
```

### 2.6 Create `errors/errors.go`
**Purpose**: Custom error types for better error handling

```go
package errors

import "fmt"

// GarthError represents a general Garth error
type GarthError struct {
    Message string
    Cause   error
}

func (e *GarthError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}

// GarthHTTPError represents an HTTP-related error
type GarthHTTPError struct {
    GarthError
    StatusCode int
    Response   string
}

func (e *GarthHTTPError) Error() string {
    return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.GarthError.Error())
}
```

### 2.7 Create `utils/utils.go`
**Purpose**: Utility functions

```go
package utils

import (
    "strings"
    "time"
    "unicode"
)

// CamelToSnake converts CamelCase to snake_case
func CamelToSnake(s string) string {
    var result []rune
    for i, r := range s {
        if unicode.IsUpper(r) && i > 0 {
            result = append(result, '_')
        }
        result = append(result, unicode.ToLower(r))
    }
    return string(result)
}

// CamelToSnakeDict converts map keys from camelCase to snake_case
func CamelToSnakeDict(m map[string]interface{}) map[string]interface{} {
    result := make(map[string]interface{})
    for k, v := range m {
        result[CamelToSnake(k)] = v
    }
    return result
}

// FormatEndDate formats an end date interface to time.Time
func FormatEndDate(end interface{}) time.Time {
    switch v := end.(type) {
    case time.Time:
        return v
    case string:
        if t, err := time.Parse("2006-01-02", v); err == nil {
            return t
        }
    }
    return time.Now()
}

// DateRange generates a range of dates
func DateRange(end time.Time, days int) []time.Time {
    var dates []time.Time
    for i := 0; i < days; i++ {
        dates = append(dates, end.AddDate(0, 0, -i))
    }
    return dates
}

// GetLocalizedDateTime converts timestamps to localized time
func GetLocalizedDateTime(gmtTimestamp, localTimestamp int64) time.Time {
    // Implementation based on timezone offset
    return time.Unix(localTimestamp, 0)
}
```

### 2.8 Refactor `main.go`
**Purpose**: Simplified main function using the new client package

```go
package main

import (
    "fmt"
    "log"
    "os"
    
    "garmin-connect/client"
    "garmin-connect/data"
)

func main() {
    // Load credentials from .env file
    email, password, domain, err := loadEnvCredentials()
    if err != nil {
        log.Fatalf("Failed to load credentials: %v", err)
    }

    // Create client
    garminClient, err := client.NewClient(domain)
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    // Try to load existing session first
    sessionFile := "garmin_session.json"
    if err := garminClient.LoadSession(sessionFile); err != nil {
        fmt.Println("No existing session found, logging in with credentials from .env...")
        
        if err := garminClient.Login(email, password); err != nil {
            log.Fatalf("Login failed: %v", err)
        }
        
        // Save session for future use
        if err := garminClient.SaveSession(sessionFile); err != nil {
            fmt.Printf("Failed to save session: %v\n", err)
        }
    } else {
        fmt.Println("Loaded existing session")
    }

    // Test getting activities
    activities, err := garminClient.GetActivities(5)
    if err != nil {
        log.Fatalf("Failed to get activities: %v", err)
    }

    // Display activities
    displayActivities(activities)
}

func displayActivities(activities []data.Activity) {
    fmt.Printf("\n=== Recent Activities ===\n")
    for i, activity := range activities {
        fmt.Printf("%d. %s\n", i+1, activity.ActivityName)
        fmt.Printf("   Type: %s\n", activity.ActivityType.TypeKey)
        fmt.Printf("   Date: %s\n", activity.StartTimeLocal)
        if activity.Distance > 0 {
            fmt.Printf("   Distance: %.2f km\n", activity.Distance/1000)
        }
        if activity.Duration > 0 {
            duration := time.Duration(activity.Duration) * time.Second
            fmt.Printf("   Duration: %v\n", duration.Round(time.Second))
        }
        fmt.Println()
    }
}

func loadEnvCredentials() (email, password, domain string, err error) {
    // This function should be moved to client package eventually
    // For now, keep it here to maintain functionality
    if err := godotenv.Load(); err != nil {
        return "", "", "", fmt.Errorf("failed to load .env file: %w", err)
    }
    
    email = os.Getenv("GARMIN_EMAIL")
    password = os.Getenv("GARMIN_PASSWORD")
    domain = os.Getenv("GARMIN_DOMAIN")
    
    if domain == "" {
        domain = "garmin.com"
    }
    
    if email == "" || password == "" {
        return "", "", "", fmt.Errorf("GARMIN_EMAIL and GARMIN_PASSWORD must be set in .env file")
    }
    
    return email, password, domain, nil
}
```

## Implementation Order

1. **Create directory structure** first
2. **Create types/tokens.go** - Move all token structures
3. **Create errors/errors.go** - Define custom error types
4. **Create utils/utils.go** - Add utility functions
5. **Create client/auth.go** - Extract authentication logic
6. **Create client/sso.go** - Extract SSO logic  
7. **Create data/base.go** - Extract data models
8. **Create client/client.go** - Extract client logic
9. **Refactor main.go** - Update to use new packages
10. **Test the refactored code** - Ensure functionality is preserved

## Testing Strategy

After each major step:
1. Run `go build` to check for compilation errors
2. Test authentication flow if SSO logic was modified
3. Test activity retrieval if client methods were changed
4. Verify session save/load functionality

## Key Considerations

1. **Maintain backward compatibility** - Ensure existing functionality works
2. **Error handling** - Use new custom error types appropriately  
3. **Package imports** - Update import paths correctly
4. **Visibility** - Export only necessary functions/types (capitalize appropriately)
5. **Documentation** - Add package and function documentation

This plan provides a systematic approach to refactoring the existing code while maintaining functionality and preparing for the addition of new features from the Python library.