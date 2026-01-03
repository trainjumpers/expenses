# Server Development Guide

## Build, Lint, and Test Commands

### Build & Development
- `just install` - Install required Go tools (air, goose, ginkgo)
- `just server` - Start server with hot-reload using air

### Formatting & Linting
- `just format` - Format all Go files (`go fmt ./...` and `go mod tidy`)
- `go fmt ./...` - Format code in current directory
- `go mod tidy` - Clean up dependencies

### Testing
- `just test` - Run all e2e tests (runs migrations, seeds, then ginkgo tests)
- `just test "focus string"` - Run specific test by focus string
- `ginkgo -r -race ./...` - Run tests without migrations

For running a single test:
- Use `just test "YourTestName"` where `YourTestName` matches the `Describe` or `It` block name
- Focus strings support regex patterns, e.g., `just test "AccountService.*"` to run all account service tests

### Database Migrations
- `just db-create <name>` - Create new migration file
- `just db-upgrade` - Apply pending migrations
- `just db-downgrade` - Revert last migration
- `just db-seed` - Apply seed data for testing

## Architecture & Code Organization

### Directory Structure
```
internal/
├── api/controller/     # HTTP handlers, request/response, error handling
├── config/              # Configuration management
├── errors/              # Custom error types with HTTP status codes
├── service/             # Business logic with debug logging
├── models/              # Request/response models
├── repository/          # Database access layer
├── validator/           # Input validation
└── database/
    ├── migrations/      # Goose migration files
    └── seed/            # Seed data (test/prod)
```

### Layered Architecture
- **Controller**: Entry point, handles HTTP requests/responses, logs start/end
- **Service**: Business logic, no framework dependencies, extensive debug logging
- **Repository**: Database operations, use `database/sql` with `pgx` driver
- **Validator**: Input validation using `go-playground/validator`

## Code Style Guidelines

### Naming Conventions
- **Files**: `snake_case.go` (e.g., `user_service.go`)
- **Tests**: `<name>_test.go` (e.g., `user_service_test.go`)
- **Packages**: lowercase, single word, no underscores
- **Exports**: PascalCase (e.g., `UserService`, `CreateUser`)
- **Privates**: camelCase (e.g., `dbConn`, `userRepo`)

### Error Handling
- Never ignore errors - always handle explicitly
- Wrap errors with context: `fmt.Errorf("failed to read file: %w", err)`
- Use custom error types from `internal/errors/` for HTTP status mapping
- Controller layer logs errors, returns appropriate status codes

### Imports
- Group imports: standard library, third-party, local
- Keep imports sorted alphabetically within groups
- Use absolute imports for local packages (e.g., `expenses/internal/service`)

### Logging
- Use `go.uber.org/zap` for structured logging
- Service layer: extensive debug logging for business logic
- Controller layer: log request start/end for monitoring
- Never log sensitive data (passwords, tokens)

### Testing (Ginkgo Framework)
- Use `Describe` for test suites, `It` for test cases
- **Unit tests**: Focus on service layer with mocked repository
- **Integration tests**: Full stack with real database via goose migrations
- Mock repositories using `go.uber.org/mock` for service tests
- Test data: use seed files in `internal/database/seed/test/`
- Test files must end in `_test.go`

### Dependency Injection
- Use `google/wire` for compile-time dependency injection
- Define interfaces in `internal/repository/` and `internal/service/`
- Wire generation: `wire gen ./internal/wire`

### Common Patterns
- **Functional Options**: For structs with many optional parameters
- **Context**: Always pass `context.Context` as first argument to I/O functions
- **Middleware**: Use for auth, logging, CORS via gin middleware chain
- **Defers**: Always cleanup resources (connections, statements) with defer

### Anti-patterns to Avoid
- Never panic in production code
- Don't use global variables
- Avoid shadowing variables in inner scopes
- Don't ignore return values, especially errors
- Avoid deep nesting - use early returns
- Never mock external services in integration tests

### Security
- Always use prepared statements to prevent SQL injection
- Validate all input at the controller/validator layer
- Hash passwords using bcrypt (check `golang.org/x/crypto/bcrypt`)
- Use JWT for authentication (check `golang.org/x/crypto/jwt/v5`)
- Sanitize user input before processing

## Development Workflow

1. Install tools: `just install`
2. Create migration: `just db-create <name>`
3. Write models in `internal/models/`
4. Write repository in `internal/repository/`
5. Write service in `internal/service/` with business logic
6. Write controller in `internal/api/controller/`
7. Write tests using ginkgo with focus strings
8. Run specific test: `just test "TestName"`
9. Format code: `just format`

## Notes
- Use `wire` for DI - regenerate wire files after adding dependencies
- Database is PostgreSQL with `pgx/v5` driver
- Use `goose` for migrations and seeding
- Tests use real database (not mocked) via e2e_test.sh
- Hot-reload enabled with `air` during development
- Use `context` for cancellation and timeouts in long-running operations
