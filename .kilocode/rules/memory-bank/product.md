# Product Description

## Why this project exists
Garth provides a Go client library for interacting with Garmin Connect API, enabling developers to:
- Access personal fitness data programmatically
- Build custom health tracking applications
- Integrate Garmin data with other systems

## Problems it solves
- Simplifies authentication with Garmin's complex SSO system
- Provides structured access to health metrics (steps, sleep, stress, etc.)
- Handles session persistence and token management
- Abstracts API complexities behind a clean Go interface

## How it should work
1. Users authenticate with their Garmin credentials
2. The client manages session tokens automatically
3. Developers can request daily/weekly metrics through simple method calls
4. Data is returned as structured Go types
5. Sessions can be saved/loaded for persistent access

## User Experience Goals
- Minimal setup required
- Intuitive method names matching Garmin's metrics
- Clear error handling and documentation
- Support for both individual data points and bulk retrieval
- Mock client for testing without live API calls