# Technology Stack

## Language and Runtime

- **Language:** Go 1.22+
- **Build:** Static binaries with CGO_ENABLED=0 for portability

## Core Dependencies

- **MCP SDK:** [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - Model Context Protocol implementation
- **HTML Parser:** golang.org/x/net/html - Documentation parsing
- **HTTP Client:** Standard library net/http with custom retry logic
- **Logging:** [github.com/rs/zerolog](https://github.com/rs/zerolog) - Fast, structured logging
- **CLI Framework:** [github.com/spf13/cobra](https://github.com/spf13/cobra) - Command-line interface
- **Configuration:** [github.com/spf13/viper](https://github.com/spf13/viper) - Configuration management

## Testing Framework

- **Unit Tests:** Standard Go testing package
- **Property-Based Tests:** [leanovate/gopter](https://github.com/leanovate/gopter)
  - Minimum 100 iterations per property test
  - Tag format: `Feature: nats-docs-mcp-server, Property {N}: {property_text}`

## Test-Driven Development (TDD)

**Follow TDD principles strictly for all implementation work:**

1. **Write a failing test first** - Before implementing any function or feature, write a test that fails
2. **Implement the minimum code** - Write just enough code to make the test pass
3. **Verify the test passes** - Run the test to confirm it now passes
4. **Refactor if needed** - Clean up code while keeping tests green

**TDD Workflow:**
```bash
# 1. Write failing test
# Edit: internal/config/config_test.go

# 2. Run test (should fail)
go test -v ./internal/config

# 3. Implement function
# Edit: internal/config/config.go

# 4. Run test again (should pass)
go test -v ./internal/config

# 5. Refactor and retest
go test -v ./internal/config
```

**Benefits:**
- Ensures 100% test coverage from the start
- Validates requirements before implementation
- Catches regressions immediately
- Documents expected behavior through tests

## Integration Testing

**Run integration tests between major features to validate component interactions.**

Integration tests verify that multiple components work together correctly. Run these at key milestones:

**When to Run Integration Tests:**
- After completing a major feature (e.g., documentation indexing, MCP server core)
- Before moving to the next major component
- After significant refactoring
- Before creating a release

**Integration Test Approach:**
```bash
# Run all tests including integration tests
go test -v ./...

# Run integration tests with specific tag
go test -v -tags=integration ./...

# Run integration tests for specific component
go test -v ./internal/server/integration_test.go
```

**Integration Test Scenarios:**
- Fetcher → Parser → Index: Verify full documentation ingestion pipeline
- Index → Search: Verify search returns correctly indexed documents
- MCP Server → Tool Handlers → Index: Verify end-to-end tool invocations
- Configuration → All Components: Verify config propagates correctly

**Checkpoints in Implementation Plan:**
- Task 6: After documentation indexing is complete
- Task 13: After all core functionality is implemented
- Task 18: Final validation before release

Integration tests ensure components work together as designed and catch interface mismatches early.

## Build and Release

- **Build Tool:** GoReleaser for cross-platform binary creation
- **CI/CD:** GitHub Actions for automated testing and releases
- **Platforms:** Linux, macOS, Windows (AMD64 and ARM64)

## Common Commands

```bash
# Initialize module
go mod init

# Install dependencies
go mod download

# Run tests
go test -v ./...

# Run specific test
go test -v ./internal/index

# Run linter (requires golangci-lint)
golangci-lint run

# Build binaries (use GoReleaser exclusively)
goreleaser build --snapshot --clean

# Test full release process locally
goreleaser release --snapshot --clean

# Run server (after building with GoReleaser)
./dist/nats-docs-mcp-server_linux_amd64_v1/nats-docs-mcp-server --config config.yaml
```

## Build Process

**Use GoReleaser exclusively for all builds.** Do not use `go build` directly for production binaries.

- **Development builds:** `goreleaser build --snapshot --clean`
- **Local testing:** `goreleaser release --snapshot --clean`
- **Production releases:** Automated via GitHub Actions on version tags

## Configuration

- **Config Library:** Use github.com/spf13/viper for all configuration management
- **Config File:** YAML format
- **Environment Variables:** Supported for all settings
- **CLI Flags:** Override config file and env vars
- **Precedence:** CLI flags > config file > env vars > defaults

## Best Practices

**Follow 12-Factor App Principles:**
1. **Codebase:** One codebase tracked in version control
2. **Dependencies:** Explicitly declare and isolate dependencies (go.mod)
3. **Config:** Store config in environment variables (use Viper)
4. **Backing services:** Treat backing services as attached resources
5. **Build, release, run:** Strictly separate build and run stages (GoReleaser)
6. **Processes:** Execute the app as stateless processes
7. **Port binding:** Export services via port binding (stdio for MCP)
8. **Concurrency:** Scale out via the process model
9. **Disposability:** Maximize robustness with fast startup and graceful shutdown
10. **Dev/prod parity:** Keep development and production as similar as possible
11. **Logs:** Treat logs as event streams (use zerolog to stdout/stderr)
12. **Admin processes:** Run admin/management tasks as one-off processes


## Caching Strategy

**Session-Based In-Memory Caching:**
- All documentation is fetched once at server startup and cached in memory
- Cache persists for the entire server session (no expiration during runtime)
- Cache is NOT persisted to disk between sessions
- Each server restart fetches fresh documentation from docs.nats.io
- Zero network requests after initial startup fetch
- Fast response times for all search and retrieval operations

**Benefits:**
- Always fresh documentation (fetched at startup)
- No stale data concerns (cache cleared on restart)
- Simple implementation (no cache invalidation logic needed)
- No disk I/O after startup
- Predictable memory usage (~15-75 MB)

**Trade-offs:**
- Startup time includes documentation fetch (5-30 seconds)
- Memory footprint includes full documentation cache
- No offline operation (requires network at startup)
