// Package client implements the low-level Garmin Connect HTTP client.
// It is responsible for authentication (via SSO helpers), request construction,
// header and cookie handling, error mapping, and JSON decoding. Higher-level
// public APIs in pkg/garmin delegate to this package for actual network I/O.
// Note: This is an internal package and not intended for direct external use.
package client
