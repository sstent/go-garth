## Brief overview
Project-specific rules for go-garth development. Provides implementation guidelines with Python reference and credential handling requirements.

## Project Context
- Reference Python implementation in 'ReferenceCode' directory when debugging issues:
  - `garth/` for core functionality patterns
  - `python-garminconnect/` for API integration examples

## Development Workflow
- Restrict all file modifications to `go-garth/` directory subtree

## Security Guidelines
- Always store credentials in root `.env` file
- Never hardcode credentials in source files
- Ensure `.gitignore` includes `.env` to prevent accidental commits

## Code Standards
- Follow Go standard naming conventions (camelCase for variables)
- Document public methods with GoDoc-style comments
- Write table-driven tests for critical functionality