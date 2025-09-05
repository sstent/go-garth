package garth

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

// CloudflareBypass provides methods to handle Cloudflare protection
type CloudflareBypass struct {
	client     *http.Client
	userAgent  string
	maxRetries int
}

// NewCloudflareBypass creates a new CloudflareBypass instance
func NewCloudflareBypass(client *http.Client) *CloudflareBypass {
	return &CloudflareBypass{
		client:     client,
		userAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		maxRetries: 3,
	}
}

// MakeRequest performs a request with Cloudflare bypass techniques
func (cf *CloudflareBypass) MakeRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt < cf.maxRetries; attempt++ {
		// Clone the request to avoid modifying the original
		clonedReq := cf.cloneRequest(req)

		// Apply bypass headers
		cf.applyBypassHeaders(clonedReq)

		// Add random delay to simulate human behavior
		if attempt > 0 {
			delay := time.Duration(1000+rand.Intn(2000)) * time.Millisecond
			time.Sleep(delay)
		}

		resp, err = cf.client.Do(clonedReq)
		if err != nil {
			continue
		}

		// Check if we got blocked by Cloudflare
		if cf.isCloudflareBlocked(resp) {
			resp.Body.Close()
			if attempt < cf.maxRetries-1 {
				// Try different user agent on retry
				cf.rotateUserAgent()
				continue
			}
			return nil, fmt.Errorf("blocked by Cloudflare after %d attempts", cf.maxRetries)
		}

		return resp, nil
	}

	return nil, err
}

// applyBypassHeaders adds headers to bypass Cloudflare
func (cf *CloudflareBypass) applyBypassHeaders(req *http.Request) {
	req.Header.Set("User-Agent", cf.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Sec-Ch-Ua", `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("DNT", "1")

	// Add some randomized headers
	if rand.Float32() < 0.7 {
		req.Header.Set("Pragma", "no-cache")
	}
	if rand.Float32() < 0.5 {
		req.Header.Set("Connection", "keep-alive")
	}
}

// isCloudflareBlocked checks if the response indicates Cloudflare blocking
func (cf *CloudflareBypass) isCloudflareBlocked(resp *http.Response) bool {
	if resp.StatusCode == 403 {
		// Check for Cloudflare-specific headers or content
		if resp.Header.Get("Server") == "cloudflare" {
			return true
		}
		if resp.Header.Get("CF-Ray") != "" {
			return true
		}

		// Check response body for Cloudflare indicators
		if resp.ContentLength > 0 && resp.ContentLength < 50000 {
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				bodyStr := string(body)
				if contains(bodyStr, "cloudflare") || contains(bodyStr, "Attention Required") {
					return true
				}
			}
		}
	}
	return false
}

// rotateUserAgent changes the user agent for retry attempts
func (cf *CloudflareBypass) rotateUserAgent() {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
	}
	cf.userAgent = userAgents[rand.Intn(len(userAgents))]
}

// cloneRequest creates a copy of the HTTP request
func (cf *CloudflareBypass) cloneRequest(req *http.Request) *http.Request {
	cloned := req.Clone(req.Context())
	if cloned.Header == nil {
		cloned.Header = make(http.Header)
	}
	return cloned
}

// contains is a case-insensitive string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsAt(s, substr)))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
