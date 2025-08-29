package garth

import (
	"context"
	"io"
	"net/http"
	"time"
)

// APIClient manages API requests to Garmin Connect
type APIClient struct {
	baseURL    string
	httpClient *http.Client
	rateLimit  time.Duration
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

// Get executes a GET request
func (c *APIClient) Get(ctx context.Context, path string) (*http.Response, error) {
	return c.request(ctx, http.MethodGet, path, nil)
}

// Post executes a POST request
func (c *APIClient) Post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.request(ctx, http.MethodPost, path, body)
}

func (c *APIClient) request(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	// Rate limiting
	time.Sleep(c.rateLimit)

	var resp *http.Response
	var err error
	maxRetries := 3
	backoff := 500 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		req, createErr := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
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

	if err != nil {
		return nil, &APIError{
			StatusCode: http.StatusBadGateway,
			Message:    "Request failed after retries",
			Cause:      err,
		}
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    "API request failed",
		}
	}

	return resp, nil
}
