- when developing go code use the standard go tools
- always execute commands from the root of the project
- in ACT mode if you hit 5 errors in a row return to plan mode
- Both backend and frontend applications must read from the root `.env` file
- always prefer small targeted changds over big changes
- use the python implementation a core reference before making changes. That version works and will provide insights for solving issues.

**Testing Requirements**
- **MANDATORY**: Run all tests in containerized environments
- Follow testing framework conventions (pytest for Python, Jest for React)
- Include unit, integration, and end-to-end tests
- Test data should be minimal and focused
- Separate test types into different directories


**TESTING**
- when asked to resolve errors:
    - iterate until all builds complete with no errors
    - the commmand should be run from the root of the project:  go build ./...


## Go Formatting

When work is done the following shouldbe run

- go mod tidy
- go fmt ./...
- go vet ./...
- go test ./...



**Rule 13: Database Configuration**
- Use environment variables for database configuration
- Implement proper connection pooling
- Follow database naming conventions
- Mount database data as Docker volumes for persistence


**Project Structure**:
- **MANDATORY**: All new code must follow the standardized project structure
- **MANDATORY**: Core backend logic only in `/src/` directory
- **MANDATORY**: Frontend code only in `/frontend/` directory
- **MANDATORY**: All Docker files in `/docker/` directory

**Package Management**:
- **MANDATORY**: Use UV for Python package management
- **MANDATORY**: Use pnpm for React package management
- **MANDATORY**: Dependencies declared in appropriate configuration files


**Code Quality**:
- **MANDATORY**: Run linting before building
- **MANDATORY**: Resolve all errors before deployment
- **MANDATORY**: Follow language-specific coding standards

**Project Organization**:
- **FORBIDDEN**: Data files committed to git
- **FORBIDDEN**: Configuration secrets in code
- **FORBIDDEN**: Environment variables in subdirectories


**Pre-Deployment Checklist**:
1. All tests passing
2. Linting and type checking completed
3. Environment variables properly configured
4. Database migrations applied
5. Security scan completed


**Do at every iteration**
- check the rules
- when you finish a task or sub-task update the todo list




# GoLang Rules

## Interface and Mock Management Rules

- Always implement interfaces with concrete types - Never use anonymous structs with function fields to satisfy interfaces
- Create shared mock implementations - Don't duplicate mock types across test files; use a central test_helpers.go file
- One mock per interface - Each interface should have exactly one shared mock implementation for consistency
- Use factory functions for mocks - Provide NewMockXXX() functions to create mock instances with sensible defaults
- Make mocks configurable when needed - Use function fields or configuration structs for tests requiring custom behavior

## Test Organization Rules

- Centralize test utilities - Keep all shared test helpers, mocks, and utilities in test_helpers.go or similar
- Follow consistent patterns - Use the same approach for creating clients, sessions, and mocks across all test files
- Don't pass nil for required interfaces - Always provide a valid mock implementation unless specifically testing nil handling
- Use descriptive mock names - Name mocks clearly (e.g., MockAuthenticator, MockHTTPClient) to indicate their purpose

## Code Change Management Rules

- Update all dependent code when changing interfaces - If you modify an interface, check and update ALL implementations and tests
- Run full test suite after interface changes - Interface changes can break code in unexpected places
- Search codebase for interface usage - Use grep or IDE search to find all references to changed interfaces
- Update mocks when adding interface methods - New interface methods must be implemented in all mocks

## Testing Best Practices Rules

- Test with realistic mock data - Don't use empty strings or zero values unless specifically testing edge cases
- Set proper expiration times - Mock sessions should have future expiration times unless testing expiration
- Verify mock behavior is called - Add call tracking to mocks when behavior verification is important
- Use table-driven tests for multiple scenarios - Test different mock behaviors using test case structs
- Mock external dependencies only - Don't mock your own business logic; mock external services, APIs, and I/O

## Documentation and Review Rules

- Document shared test utilities - Add comments explaining when and how to use shared mocks
- Include interface changes in code reviews - Flag interface modifications for extra review attention
- Update README/docs when adding test patterns - Document new testing patterns for other developers
- Add examples for complex mock usage - Show how to use configurable mocks in comments or examples

## Error Prevention Rules

- Compile-test interface changes immediately - Run go build after any interface modification
- Write interface compliance tests - Add var _ InterfaceName = (*MockType)(nil) to verify mock implements interface
- Set up pre-commit hooks - Use tools like golangci-lint to catch interface issues before commit
- Review test failures carefully - Interface-related test failures often indicate broader architectural issues

## Refactoring Safety Rules

- Create shared mocks before removing duplicates - Implement the centralized solution before removing old code
- Update one test file at a time - Incremental changes make it easier to identify and fix issues
- Keep backup of working tests - Commit working tests before major refactoring
- Test each change independently - Verify each file works after updating it
- Use version control effectively - Make small, focused commits that are easy to review and revert

## Learning Points

- Interface Implementation: Go interfaces must be implemented by concrete types, not anonymous structs with function fields
- Test Organization: Shared test helpers reduce duplication and improve maintainability
- Dependency Injection: Using interfaces makes code more testable but requires proper mock implementation

