// Package sso implements the Garmin SSO login flow. It orchestrates CSRF,
// ticket exchange, MFA placeholders, and token retrieval, delegating OAuth
// details to internal/auth/oauth. The internal client consumes this package.
// Note: This is an internal package and not intended for direct external use.
package auth
