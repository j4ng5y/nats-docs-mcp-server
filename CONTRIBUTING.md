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
- Tag format: `Feature: {feature-name}, Property {N}: {property_text}`
- Test both success and failure cases
- Use smart generators that constrain to valid input space
- Document which requirements each property validates

**Example property test:**

```go
// Feature: syncp-documentation-support, Property 1: Index Independence
// VALIDATES: Requirements 2.1
func TestProperty_IndexIndependence(t *testing.T) {
    properties := gopter.NewProperties(nil)

    properties.Property("NATS and Syncp indices are independent",
        prop.ForAll(
            func() bool {
                // Setup dual indices
                manager := index.NewManager()

                // Index NATS docs
                natsDoc := &index.Document{ID: "nats-1", Title: "Test", Content: "jetstream"}
                manager.IndexNATS([]*index.Document{natsDoc})

                // Index Syncp docs
                syncpDoc := &index.Document{ID: "syncp-1", Title: "Test", Content: "control-plane"}
                manager.IndexSyncp([]*index.Document{syncpDoc})

                // Verify independence - results from each index should only contain respective docs
                natsResults, _ := manager.GetNATSIndex().Search("jetstream", 10)
                syncpResults, _ := manager.GetSyncpIndex().Search("control-plane", 10)

                return len(natsResults) > 0 && len(syncpResults) > 0
            },
        ),
    )

    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

**Dual-Source Testing Patterns:**

When testing dual-source features (Syncp support), follow these patterns:

1. **Independence Testing**: Verify that NATS and Syncp components operate independently
   ```go
   // Each source's results should not include the other source's documents
   // Test by checking result source metadata
   ```

2. **Classification Testing**: Verify query classification routes to correct sources
   ```go
   // NATS keywords → NATS source
   // Syncp keywords → Syncp source
   // Both/ambiguous → Both sources
   ```

3. **Merge Testing**: Verify results from multiple sources merge correctly
   ```go
   // Results should be sorted by relevance score across both sources
   // Source metadata should be preserved
   // No duplicate results should appear
   ```

4. **Graceful Degradation Testing**: Verify system continues if one source fails
   ```go
   // Syncp fetch failure → Continue with NATS-only
   // NATS fetch failure → Fail (primary source)
   ```

5. **Backward Compatibility Testing**: Verify NATS-only mode still works
   ```go
   // With syncp.enabled = false, all features work as before
   // No breaking changes to tool interfaces
   // Same results as original implementation
   ```

### Testing Transport Implementations

The server supports multiple transport types. When adding features or making changes, test all transport types:

**Testing STDIO Transport (Default):**

```bash
# Run the server with stdio (default)
./nats-docs-mcp-server --config config.yaml

# The server reads from stdin and writes to stdout using the MCP protocol
# Test by sending MCP requests via stdin
```

**Testing SSE Transport:**

```bash
# Start the server with SSE transport
./nats-docs-mcp-server --transport sse --host localhost --port 8080

# In another terminal, test the SSE endpoint
curl http://localhost:8080/sse

# The server will send MCP responses as Server-Sent Events
```

**Testing StreamableHTTP Transport:**

```bash
# Start the server with StreamableHTTP transport
./nats-docs-mcp-server --transport streamablehttp --host localhost --port 8080

# In another terminal, test the HTTP endpoint
curl -X POST http://localhost:8080/messages \
  -H "Content-Type: application/json" \
  -d '{...mcp request...}'

# The server will respond with MCP responses via HTTP and SSE
```

**Integration Testing All Transports:**

```bash
# Run integration tests
go test -v -tags=integration ./...

# This tests all transport implementations end-to-end
```

**Manual Validation Checklist:**

- [ ] STDIO transport starts without flags (defaults to STDIO)
- [ ] SSE transport starts with `--transport sse --port 8080`
- [ ] StreamableHTTP transport starts with `--transport streamablehttp --port 8080`
- [ ] Transport type is logged at startup
- [ ] Transport address is logged for network transports
- [ ] Server responds to MCP requests via each transport
- [ ] Graceful shutdown works for all transports
- [ ] Port conflicts are handled correctly
- [ ] Invalid transport type returns error

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
- Property tests in `_test.go` files alongside unit tests
- Each package may have separate test files for unit and property tests

**Dual-Source Package Organization:**
- `internal/classifier/` - Query classification logic and tests
- `internal/search/` - Multi-source search orchestration and tests
- `internal/index/` - Dual index management and tests (Manager struct)
- `internal/fetcher/` - MultiSourceFetcher for dual sources and tests
- Each package includes comprehensive unit tests and property tests

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
│   ├── classifier/      # Query classification (NATS/Syncp routing)
│   ├── config/          # Configuration management
│   ├── fetcher/         # Documentation fetching (dual-source support)
│   ├── parser/          # HTML parsing
│   ├── index/           # Search indexing and management (TF-IDF)
│   ├── search/          # Multi-source search orchestration
│   ├── logger/          # Structured logging
│   └── server/          # MCP server core
├── .github/
│   └── workflows/       # CI/CD workflows
├── .kiro/
│   ├── specs/           # Feature specifications
│   │   └── syncp-documentation-support/  # Syncp feature specs
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
