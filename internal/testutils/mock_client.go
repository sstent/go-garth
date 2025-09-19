package testutils

import (
	"errors"
	"io"
	"net/url"

	"go-garth/internal/api/client"
)

// MockClient simulates API client for tests
type MockClient struct {
	RealClient *client.Client
	FailEvery  int
	counter    int
}

func (mc *MockClient) ConnectAPI(path string, method string, params url.Values, body io.Reader) ([]byte, error) {
	mc.counter++
	if mc.FailEvery != 0 && mc.counter%mc.FailEvery == 0 {
		return nil, errors.New("simulated error")
	}
	return mc.RealClient.ConnectAPI(path, method, params, body)
}
