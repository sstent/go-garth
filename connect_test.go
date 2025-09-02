package garth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAPIClient_Get(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer ts.Close()

	// Create client
	client := NewAPIClient(ts.URL, http.DefaultClient)

	// Test successful request
	resp, err := client.Get(context.Background(), "/test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %d", resp.StatusCode)
	}
}

func TestAPIClient_Retry(t *testing.T) {
	retryCount := 0
	// Create a test server that fails first two requests
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		retryCount++
		if retryCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Create client with faster backoff for testing
	client := NewAPIClient(ts.URL, http.DefaultClient)
	client.SetRateLimit(10 * time.Millisecond)

	// Test retry logic
	resp, err := client.Get(context.Background(), "/retry-test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK after retries, got %d", resp.StatusCode)
	}
	if retryCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", retryCount)
	}
}

func TestAPIClient_ErrorHandling(t *testing.T) {
	// Create a test server that returns 404
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	// Create client
	client := NewAPIClient(ts.URL, http.DefaultClient)

	// Test error handling
	_, err := client.Get(context.Background(), "/not-found")
	if err == nil {
		t.Fatal("Expected error but got none")
	}

	// Check for RequestError wrapper
	reqErr, ok := err.(*RequestError)
	if !ok {
		t.Fatalf("Expected RequestError, got %T", err)
	}

	// Check the wrapped APIError
	apiErr, ok := reqErr.GetCause().(*APIError)
	if !ok {
		t.Fatalf("Expected APIError inside RequestError, got %T", reqErr.GetCause())
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 status, got %d", apiErr.StatusCode)
	}
}
