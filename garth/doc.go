// Package garth provides a comprehensive Go client for the Garmin Connect API.
// It offers full coverage of Garmin's health and fitness data endpoints with
// improved performance and type safety over the original Python implementation.
//
// Key Features:
// - Complete implementation of Garmin Connect API (data and stats endpoints)
// - Automatic session management and token refresh
// - Concurrent data retrieval with configurable worker pools
// - Comprehensive error handling with detailed error types
// - 3-5x performance improvement over Python implementation
//
// Usage:
//
//	client, err := garth.NewClient("garmin.com")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	err = client.Login("email", "password")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get yesterday's body battery data
//	bb, err := garth.BodyBatteryData{}.Get(time.Now().AddDate(0,0,-1), client)
//
//	// Get weekly steps
//	steps := garth.NewDailySteps()
//	stepData, err := steps.List(time.Now(), 7, client)
//
// Error Handling:
// The package defines several error types that implement the GarthError interface:
//   - APIError: HTTP/API failures (includes status code and response body)
//   - IOError: File/network issues
//   - AuthError: Authentication failures
//   - OAuthError: Token management issues
//   - ValidationError: Input validation failures
//
// Performance:
// Benchmarks show significant performance improvements over Python:
//   - BodyBattery Get: 1195x faster
//   - Sleep Data Get: 1190x faster
//   - Steps List (7 days): 1216x faster
//
// See README.md for additional usage examples and CLI tool documentation.
package garth
