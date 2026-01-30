# Contributing to NATS Documentation MCP Server

Thank you for your interest in contributing! This document provides guidelines for developing and contributing to the project.

## Development Setup

### Prerequisites

- Go 1.22 or later
- GoReleaser (for building binaries)
- golangci-lint (for linting)
- Git

### Installation

```bash
# Install Go
# See https://golang.org/doc/install

# Install GoReleaser
# See https://goreleaser.com/install/

# Install golangci-lint
# See https://golangci-lint.run/usage/install/

# Clone the repository
git clone https://github.com/j4ng5y/nats-docs-mcp-server.git
cd nats-docs-mcp-server

# Download dependencies
go mod download
```

## Development Workflow

### Test-Driven Development (TDD)

This project follows TDD principles strictly:

1. **Write a failing test first** - Before implementing any function, write a test that fails
2. **Implement the minimum code** - Write just enough code to make the test pass
3. **Verify the test passes** - Run the test to confirm it now passes
4. **Refactor if needed** - Clean up code while keeping tests green

**Example TDD workflow:**

```bash
# 1. Write failing test
# Edit: internal/index/index_test.go

# 2. Run test (should fail)
go test -v ./internal/index

# 3. Implement function
# Edit: internal/index/index.go

# 4. Run test again (should pass)
go test -v ./internal/index

# 5. Refactor and retest
go test -v ./internal/index
```

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run tests with race detection
go test -v -race ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test -v ./internal/index

# Run specific test
go test -v -run TestSearchDocuments ./internal/index

# Run property-based tests only
go test -v -run Property ./...
```

### Property-Based Testing

This project uses property-based testing with [gopter](https://github.com/leanovate/gopter) to validate universal correctness properties.

**Property test requirements:**
- Minimum 100 iterations per property test
- Tag format: `Feature: nats-docs-mcp-server, Property {N}: {property_text}`
- Test both success and failure cases
- Use smart generators that constrain to valid input space

**Example property test:**

```go
func TestPropertySearchResultCompleteness(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    properties.Property("Property 2: Search Result Completeness", 
        prop.ForAll(
            func(query string) bool {
                // Test implementation
                return true
            },
            gen.AnyString(),
        ),
    )
    
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

### Building

**Always use GoReleaser for building - never use `go build` directly.**

```bash
# Development build (creates binaries in dist/)
KO_DOCKER_REPO=ghcr.io/j4ng5y goreleaser build --snapshot --clean

# Test full release process locally
KO_DOCKER_REPO=ghcr.io/j4ng5y goreleaser release --snapshot --clean

# Run the built binary
./dist/nats-docs-mcp-server_linux_amd64_v1/nats-docs-mcp-server --version
```

### Linting

```bash
# Run linter
golangci-lint run

# Run linter with auto-fix
golangci-lint run --fix
```

## Code Style

### General Guidelines

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting (automatically applied by most editors)
- Write clear, descriptive variable and function names
- Add comments for exported functions and types
- Keep functions small and focused on a single responsibility

### Package Organization

- `cmd/server/` - Main entry point only, minimal logic
- `internal/` - All implementation code
- Tests co-located with source files using `_test.go` suffix
- Property tests in separate `_property_test.go` files

### Error Handling

- Always check and handle errors
- Use structured error types with context
- Log errors with appropriate severity
- Return errors up the call stack, don't panic

### Logging

- Use zerolog for all logging
- Include context in log messages
- Use appropriate log levels:
  - `debug` - Detailed diagnostic information
  - `info` - General informational messages
  - `warn` - Warning messages for recoverable issues
  - `error` - Error messages for failures

## Project Structure

```
.
├── cmd/
│   └── server/          # Main entry point (main.go)
├── internal/
│   ├── config/          # Configuration management
│   ├── fetcher/         # Documentation fetching
│   ├── parser/          # HTML parsing
│   ├── index/           # Search indexing (TF-IDF)
│   ├── logger/          # Structured logging
│   └── server/          # MCP server core
├── .github/
│   └── workflows/       # CI/CD workflows
├── .kiro/
│   ├── specs/           # Feature specifications
│   └── steering/        # Project guidance
├── .goreleaser.yaml     # Build configuration
├── config.example.yaml  # Example configuration
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
├── README.md            # Project documentation
└── CONTRIBUTING.md      # This file
```

## Making Changes

### Branching Strategy

- `main` - Stable branch, always deployable
- Feature branches - `feature/description`
- Bug fix branches - `fix/description`

### Commit Messages

Follow conventional commits format:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation changes
- `test` - Test additions or changes
- `refactor` - Code refactoring
- `perf` - Performance improvements
- `chore` - Maintenance tasks

**Examples:**
```
feat(index): add TF-IDF relevance scoring
fix(fetcher): handle timeout errors correctly
docs(readme): update installation instructions
test(parser): add property test for HTML parsing
```

### Pull Request Process

1. Create a feature branch from `main`
2. Make your changes following TDD principles
3. Ensure all tests pass: `go test -v ./...`
4. Run linter: `golangci-lint run`
5. Test build: `KO_DOCKER_REPO=ghcr.io/j4ng5y goreleaser build --snapshot --clean`
6. Commit with conventional commit messages
7. Push to your fork
8. Open a pull request against `main`

**PR Requirements:**
- All tests must pass
- Code coverage should not decrease
- Linter must pass with no errors
- Build must succeed
- Include description of changes
- Reference any related issues

## Release Process

Releases are automated via GitHub Actions:

1. Update version in code if needed
2. Create and push a version tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
3. GitHub Actions will:
   - Run all tests
   - Build binaries for all platforms
   - Create GitHub release
   - Attach binaries and checksums

## Getting Help

- Open an issue for bugs or feature requests
- Check existing issues before creating new ones
- Provide detailed information and reproduction steps
- Be respectful and constructive

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on constructive feedback
- Assume good intentions

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.
