// Package oauth provides public wrappers around the internal OAuth helpers.
// It exposes functions to obtain OAuth1 tokens and exchange them for OAuth2
// tokens compatible with the public garmin package types. External consumers
// should use this package when they need token bootstrapping independent of
// a fully initialized client.
package oauth
