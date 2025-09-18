package client

import (
	"io"
	"net/url"
)

// HTTPClient defines the interface for HTTP operations
type HTTPClient interface {
	ConnectAPI(path string, method string, params url.Values, body io.Reader) ([]byte, error)
}
