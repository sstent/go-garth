# System Architecture

## Overview
The Garth client library follows a layered architecture with clear separation between:
- API Client Layer
- Data Models Layer
- Stats Endpoints Layer
- Utilities Layer

## Key Components
- **Client**: Handles authentication, session management, and HTTP requests
- **Data Models**: Represent Garmin API responses (e.g., DailyBodyBatteryStress)
- **Stats**: Implement endpoints for different metrics (steps, stress, etc.)
- **Test Utilities**: Mock implementations for testing

## Data Flow
1. User authenticates via `client.Login()`
2. Session tokens are stored in `Client` struct
3. Stats requests are made through `stats.Stats.List()` method
4. Responses are parsed into data models
5. Sessions can be saved/loaded using `SaveSession()`/`LoadSession()`

## Critical Paths
- Authentication: `garth/sso/sso.go`
- API Requests: `garth/client/client.go`
- Stats Implementation: `garth/stats/base.go`
- Data Models: `garth/types/types.go`