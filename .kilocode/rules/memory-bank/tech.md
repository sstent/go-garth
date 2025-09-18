# Technologies and Development Setup

## Core Technologies
- **Language**: Go (Golang)
- **Package Manager**: Go Modules (go.mod)
- **HTTP Client**: net/http with custom enhancements
- **Testing**: Standard testing package

## Development Setup
1. Install Go (version 1.20+)
2. Clone the repository
3. Run `go mod tidy` to install dependencies
4. Set up Garmin credentials in environment variables or session file

## Technical Constraints
- Requires Garmin Connect account for integration tests
- Limited to Garmin's public API endpoints
- No official SDK support from Garmin

## Dependencies
- Standard library packages only
- No external dependencies beyond Go standard library

## Tool Usage Patterns
- Use `go test` for running unit tests
- Use `go test -short` to skip integration tests
- Use `go build` to build the package