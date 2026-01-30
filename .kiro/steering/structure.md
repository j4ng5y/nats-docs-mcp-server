# Project Structure

## Directory Organization

```
.
├── cmd/
│   └── server/          # Main entry point (main.go)
├── internal/
│   ├── config/          # Configuration management
│   ├── fetcher/         # Documentation fetching from docs.nats.io
│   ├── parser/          # HTML parsing and content extraction
│   ├── index/           # Documentation indexing and search (TF-IDF)
│   └── server/          # MCP server core and tool handlers
├── .github/
│   └── workflows/       # GitHub Actions CI/CD workflows
├── .kiro/
│   ├── specs/           # Feature specifications
│   └── steering/        # Project guidance documents
├── .goreleaser.yaml     # GoReleaser configuration
├── config.example.yaml  # Example configuration file
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
├── README.md            # Project documentation
├── CONTRIBUTING.md      # Development guidelines
└── LICENSE              # License file
```

## Code Organization Principles

- **cmd/**: Application entry points only, minimal logic
- **internal/**: All implementation code, not importable by external projects
- **Package naming:** Lowercase, single word when possible
- **File naming:** Lowercase with underscores for multi-word names

## Key Components

- **MCP Server Core:** Handles protocol communication via stdio transport
- **Documentation Fetcher:** HTTP client with retry logic and rate limiting
- **HTML Parser:** Extracts structured content from documentation pages
- **Documentation Index:** In-memory storage with TF-IDF search
- **Tool Handlers:** Implement search_nats_docs and retrieve_nats_doc tools
- **Configuration Manager:** Multi-source config loading with validation

## Testing Structure

- Tests co-located with source files using `_test.go` suffix
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- Integration tests in separate test files or subdirectories

## Spec-Driven Development

This project follows spec-driven development methodology:
1. **Requirements:** User stories and acceptance criteria
2. **Design:** Architecture, interfaces, and correctness properties
3. **Tasks:** Incremental implementation plan with property-based testing
4. All specs located in `.kiro/specs/nats-docs-mcp-server/`
