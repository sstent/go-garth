package garth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

// APIClient manages API requests to Garmin Connect
type APIClient struct {
	baseURL    string
	httpClient *http.Client
	rateLimit  time.Duration
	logger     ErrorLogger // Optional error logger
}

// NewAPIClient creates a new API client instance
func NewAPIClient(baseURL string, httpClient *http.Client) *APIClient {
	return &APIClient{
		baseURL:    baseURL,
		httpClient: httpClient,
		rateLimit:  500 * time.Millisecond, // Default rate limit
	}
}

// SetRateLimit configures request rate limiting
func (c *APIClient) SetRateLimit(limit time.Duration) {
	c.rateLimit = limit
}

// SetRequestsPerSecond configures the maximum number of requests per second
func (c *APIClient) SetRequestsPerSecond(rate float64) {
	interval := time.Duration(float64(time.Second) / rate)
	c.SetRateLimit(interval)
}

// Get executes a GET request
func (c *APIClient) Get(ctx context.Context, path string) (*http.Response, error) {
	return c.request(ctx, http.MethodGet, path, nil)
}

// Post executes a POST request
func (c *APIClient) Post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.request(ctx, http.MethodPost, path, body)
}

// Put executes a PUT request
func (c *APIClient) Put(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.request(ctx, http.MethodPut, path, body)
}

// Delete executes a DELETE request
func (c *APIClient) Delete(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.request(ctx, http.MethodDelete, path, body)
}

// handleResponse handles API response and error decoding
func handleResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(body),
		}
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return &APIError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to parse response",
				Cause:      err,
			}
		}
	}
	return nil
}

// GetJSON executes a GET request and decodes the JSON response
func (c *APIClient) GetJSON(ctx context.Context, path string, result interface{}) error {
	resp, err := c.Get(ctx, path)
	if err != nil {
		return err
	}
	return handleResponse(resp, result)
}

// PostJSON executes a POST request with JSON body and decodes the JSON response
func (c *APIClient) PostJSON(ctx context.Context, path string, body interface{}, result interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to marshal request body",
			Cause:      err,
		}
	}

	resp, err := c.Post(ctx, path, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	return handleResponse(resp, result)
}

// PutJSON executes a PUT request with JSON body and decodes the JSON response
func (c *APIClient) PutJSON(ctx context.Context, path string, body interface{}, result interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return &APIError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to marshal request body",
			Cause:      err,
		}
	}

	resp, err := c.Put(ctx, path, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}
	return handleResponse(resp, result)
}

// DeleteJSON executes a DELETE request and decodes the JSON response
func (c *APIClient) DeleteJSON(ctx context.Context, path string, result interface{}) error {
	resp, err := c.Delete(ctx, path, nil)
	if err != nil {
		return err
	}
	return handleResponse(resp, result)
}

// ErrorLogger defines an interface for logging errors
type ErrorLogger interface {
	Errorf(format string, args ...interface{})
}

// SetLogger sets the error logger for the API client
func (c *APIClient) SetLogger(logger ErrorLogger) {
	c.logger = logger
}

func (c *APIClient) request(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	// Rate limiting using token bucket algorithm
	if c.rateLimit > 0 {
		time.Sleep(c.rateLimit)
	}

	var resp *http.Response
	var err error
	var req *http.Request
	maxRetries := 3
	backoff := 500 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		var createErr error
		req, createErr = http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
		if createErr != nil {
			return nil, &APIError{
				StatusCode: http.StatusInternalServerError,
				Message:    "Failed to create request",
				Cause:      createErr,
			}
		}

		// Set common headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err = c.httpClient.Do(req)

		// Retry only on network errors or server-side issues
		if err != nil || (resp != nil && resp.StatusCode >= 500) {
			if i < maxRetries-1 {
				// Exponential backoff with jitter
				time.Sleep(backoff)
				backoff = time.Duration(float64(backoff) * 2.5)
				continue
			}
		}
		break
	}

	// Extract query parameters for error context
	var queryValues url.Values
	if req != nil {
		queryValues = req.URL.Query()
	}

	if err != nil {
		apiErr := &APIError{
			StatusCode: http.StatusBadGateway,
			Message:    "Request failed after retries",
			Cause:      err,
		}
		reqErr := NewRequestError(method, req.URL.String(), queryValues, http.StatusBadGateway, apiErr)

		// Log error if logger is configured
		if c.logger != nil {
			c.logger.Errorf("API request failed: %v, Method: %s, URL: %s", reqErr, method, req.URL.String())
		}

		return nil, reqErr
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{
			StatusCode: resp.StatusCode,
			Message:    "API request failed",
		}
		reqErr := NewRequestError(method, req.URL.String(), queryValues, resp.StatusCode, apiErr)

		// Log error if logger is configured
		if c.logger != nil {
			c.logger.Errorf("API request failed with status %d: %s, Method: %s, URL: %s",
				resp.StatusCode, apiErr.Message, method, req.URL.String())
		}

		return nil, reqErr
	}

	return resp, nil
}
