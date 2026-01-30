# Implementation Plan: NATS Documentation MCP Server

## Overview

This implementation plan breaks down the NATS Documentation MCP Server into discrete, incremental coding tasks. Each task builds on previous work, starting with core infrastructure and progressing through documentation handling, MCP tool implementation, and finally build/release automation.

## Tasks

- [x] 1. Set up project structure and dependencies
  - Create Go module with `go mod init github.com/j4ng5y/nats-docs-mcp-server`
  - Add dependencies: mark3labs/mcp-go, HTML parser (golang.org/x/net/html), HTTP client
  - Create directory structure: cmd/server, internal/index, internal/fetcher, internal/parser, internal/config
  - Set up basic logging with structured logger (e.g., slog)
  - NOTE: Use GoReleaser exclusively for all builds (never use `go build` directly)
  - _Requirements: 5.1, 8.1_

- [x] 2. Implement configuration management
  - [x] 2.1 Create configuration struct and defaults
    - Define Config struct with all settings (log level, docs URL, timeouts, etc.)
    - Implement default values for all optional parameters
    - _Requirements: 8.5_
  
  - [x] 2.2 Implement configuration loading from multiple sources
    - Load from environment variables
    - Load from config file (YAML format)
    - Load from command-line flags
    - Implement precedence: flags > config file > env vars > defaults
    - _Requirements: 8.1, 8.2, 8.3, 5.5_
  
  - [x] 2.3 Implement configuration validation
    - Validate all configuration values on startup
    - Check for invalid log levels, negative timeouts, invalid URLs
    - Return descriptive errors for invalid configuration
    - _Requirements: 8.4_
  
  - [x] 2.4 Write property test for configuration loading
    - **Property 9: Configuration Loading**
    - **Validates: Requirements 8.1, 8.2, 8.3, 5.5**
  
  - [x] 2.5 Write property test for configuration validation
    - **Property 10: Configuration Validation**
    - **Validates: Requirements 8.4**
  
  - [x] 2.6 Write property test for configuration defaults
    - **Property 11: Configuration Defaults**
    - **Validates: Requirements 8.5**

- [x] 3. Implement documentation fetcher
  - [x] 3.1 Create HTTP client with timeout and retry logic
    - Configure HTTP client with timeouts
    - Implement exponential backoff retry mechanism
    - Add rate limiting for concurrent requests
    - _Requirements: 4.5_
  
  - [x] 3.2 Implement documentation page fetching
    - Fetch single page by URL
    - Handle HTTP errors and timeouts
    - Log fetch attempts and results
    - _Requirements: 4.1_
  
  - [x] 3.3 Implement documentation discovery
    - Fetch main docs page and parse navigation
    - Extract all documentation page URLs
    - Support concurrent fetching with rate limiting
    - _Requirements: 4.1_
  
  - [x] 3.4 Write property test for retry backoff behavior
    - **Property 8: Retry Backoff Behavior**
    - **Validates: Requirements 4.5**
  
  - [x] 3.5 Write unit tests for fetcher error handling
    - Test HTTP errors, timeouts, network failures
    - Test rate limiting behavior
    - _Requirements: 4.5_

- [x] 4. Implement HTML parser
  - [x] 4.1 Create HTML parser for documentation pages
    - Parse HTML using golang.org/x/net/html
    - Extract title, headings (h1-h6), and content
    - Preserve code blocks and lists
    - Handle special characters and encoding
    - _Requirements: 4.3, 3.4_
  
  - [x] 4.2 Implement content extraction and cleaning
    - Remove navigation, footer, and non-content elements
    - Extract main content area
    - Clean up whitespace
    - _Requirements: 4.3_
  
  - [x] 4.3 Write property test for document parsing completeness
    - **Property 7: Document Parsing Completeness**
    - **Validates: Requirements 4.3**
  
  - [x] 4.4 Write unit tests for HTML parser edge cases
    - Test malformed HTML handling
    - Test empty pages
    - Test pages with special characters
    - _Requirements: 4.3_

- [x] 5. Implement documentation index
  - [x] 5.1 Create document storage data structures
    - Implement Document struct with ID, title, URL, content, sections
    - Create DocumentStore with map-based storage
    - Implement thread-safe access with sync.RWMutex
    - _Requirements: 4.2_
  
  - [x] 5.2 Implement search index with TF-IDF
    - Build inverted index: term → document IDs
    - Calculate term frequencies per document
    - Calculate inverse document frequencies
    - Implement TF-IDF scoring for relevance ranking
    - _Requirements: 2.3, 4.2_
  
  - [x] 5.3 Implement indexing operations
    - Add document to index
    - Update existing document
    - Tokenize and normalize text (lowercase, split on whitespace)
    - _Requirements: 4.2, 4.3_
  
  - [x] 5.4 Write property test for indexed documents are searchable
    - **Property 6: Indexed Documents Are Searchable**
    - **Validates: Requirements 4.2**
  
  - [x] 5.5 Write unit tests for indexing operations
    - Test adding single document
    - Test updating document
    - Test tokenization and normalization
    - _Requirements: 4.2_

- [x] 6. Checkpoint - Ensure documentation indexing works
  - Ensure all tests pass, ask the user if questions arise.

- [x] 7. Implement search functionality
  - [x] 7.1 Implement search query processing
    - Tokenize and normalize query
    - Look up terms in inverted index
    - Calculate relevance scores using TF-IDF
    - Rank results by score (descending)
    - _Requirements: 2.1, 2.2, 2.3_
  
  - [x] 7.2 Implement search result formatting
    - Create SearchResult struct with title, URL, summary, relevance
    - Generate summaries (first 200 chars or sentence with query term)
    - Limit results to configured maximum
    - _Requirements: 2.2_
  
  - [x] 7.3 Implement case-insensitive search
    - Normalize query and index terms to lowercase
    - Ensure case variations return same documents
    - _Requirements: 2.5_
  
  - [x] 7.4 Write property test for search result completeness
    - **Property 2: Search Result Completeness**
    - **Validates: Requirements 2.2, 2.3**
  
  - [x] 7.5 Write property test for case-insensitive search
    - **Property 3: Case-Insensitive Search**
    - **Validates: Requirements 2.5**
  
  - [x] 7.6 Write unit tests for search edge cases
    - Test empty query handling
    - Test no results scenario
    - Test single term vs multi-term queries
    - _Requirements: 2.4_

- [x] 8. Implement document retrieval
  - [x] 8.1 Implement document retrieval by ID
    - Look up document in index by ID or URL path
    - Return complete document with all metadata
    - Handle missing documents with error
    - _Requirements: 3.1, 3.2, 3.3_
  
  - [x] 8.2 Implement retrieval response formatting
    - Include title, URL, content, sections
    - Preserve formatting (code blocks, lists)
    - _Requirements: 3.4, 3.5_
  
  - [x] 8.3 Write property test for valid document retrieval
    - **Property 4: Valid Document Retrieval**
    - **Validates: Requirements 3.2, 3.4, 3.5**
  
  - [x] 8.4 Write property test for invalid document error handling
    - **Property 5: Invalid Document Error Handling**
    - **Validates: Requirements 3.3**
  
  - [x] 8.5 Write unit tests for retrieval operations
    - Test successful retrieval
    - Test missing document error
    - Test formatting preservation
    - _Requirements: 3.2, 3.3, 3.4_

- [x] 9. Implement MCP server core
  - [x] 9.1 Create MCP server initialization
    - Initialize mcp-go server with stdio transport
    - Set server name and version
    - Configure logging
    - _Requirements: 1.1, 1.3_
  
  - [x] 9.2 Implement server startup sequence
    - Load configuration
    - Initialize documentation fetcher
    - Fetch and index documentation
    - Start MCP server
    - Log startup progress
    - _Requirements: 1.3, 4.1_
  
  - [x] 9.3 Implement graceful shutdown
    - Handle interrupt signals (SIGINT, SIGTERM)
    - Clean up resources
    - Close connections
    - Log shutdown
    - _Requirements: 1.5_
  
  - [x] 9.4 Write unit tests for server lifecycle
    - Test initialization
    - Test startup sequence
    - Test shutdown
    - _Requirements: 1.3, 1.5_

- [x] 10. Implement MCP tools
  - [x] 10.1 Implement search_nats_docs tool
    - Define tool schema with query and limit parameters
    - Implement tool handler that calls search functionality
    - Format results as MCP tool response
    - Handle errors and return structured error responses
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_
  
  - [x] 10.2 Implement retrieve_nats_doc tool
    - Define tool schema with doc_id parameter
    - Implement tool handler that calls retrieval functionality
    - Format document as MCP tool response
    - Handle errors and return structured error responses
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_
  
  - [x] 10.3 Register tools with MCP server
    - Register search_nats_docs tool
    - Register retrieve_nats_doc tool
    - Ensure tools are listed in tools/list response
    - _Requirements: 1.2_
  
  - [x] 10.4 Write property test for tool error response structure
    - **Property 13: Tool Error Response Structure**
    - **Validates: Requirements 9.3**
  
  - [x] 10.5 Write unit tests for tool handlers
    - Test successful tool invocations
    - Test invalid parameter handling
    - Test error response formatting
    - _Requirements: 2.1, 2.2, 3.1, 3.2, 9.3_

- [x] 11. Implement error handling and logging
  - [x] 11.1 Implement structured error types
    - Define error codes (DOCUMENT_NOT_FOUND, INVALID_QUERY, etc.)
    - Create ToolError struct with code, message, details
    - Implement error formatting functions
    - _Requirements: 9.3_
  
  - [x] 11.2 Implement logging throughout application
    - Add structured logging to all major operations
    - Log tool invocations with parameters
    - Log errors with context
    - Support configurable log levels
    - _Requirements: 9.1, 9.2, 9.4_
  
  - [x] 11.3 Write property test for log level configuration
    - **Property 12: Log Level Configuration**
    - **Validates: Requirements 9.2**
  
  - [x] 11.4 Write unit tests for error handling
    - Test error code assignment
    - Test error message formatting
    - Test error response structure
    - _Requirements: 9.3_

- [x] 12. Implement concurrent client handling
  - [x] 12.1 Ensure thread-safe index access
    - Use sync.RWMutex for index operations
    - Allow concurrent reads, exclusive writes
    - _Requirements: 1.4_
  
  - [x] 12.2 Implement connection management
    - Track active connections
    - Clean up resources on disconnect
    - Handle concurrent tool invocations
    - _Requirements: 1.4, 1.5_
  
  - [x] 12.3 Write property test for concurrent client handling
    - **Property 1: Concurrent Client Handling**
    - **Validates: Requirements 1.4, 1.5**
  
  - [x] 12.4 Write integration tests for concurrency
    - Test multiple concurrent clients
    - Test concurrent tool invocations
    - Test resource cleanup on disconnect
    - _Requirements: 1.4, 1.5_

- [x] 13. Checkpoint - Ensure all core functionality works
  - Ensure all tests pass, ask the user if questions arise.

- [x] 14. Create main entry point
  - [x] 14.1 Implement main.go with CLI
    - Parse command-line flags (--config, --log-level, --help, --version)
    - Load configuration
    - Initialize and start server
    - Handle errors and exit codes
    - _Requirements: 5.4, 8.1_
  
  - [x] 14.2 Implement version information
    - Embed version, commit, and build date via ldflags
    - Display version with --version flag
    - _Requirements: 5.4_
  
  - [x] 14.3 Write integration tests for main entry point
    - Test server starts successfully
    - Test configuration loading
    - Test error handling and exit codes
    - _Requirements: 5.4_

- [x] 15. Set up GoReleaser configuration
  - [x] 15.1 Create .goreleaser.yaml
    - Configure builds for Linux, macOS, Windows
    - Configure architectures: amd64, arm64
    - Set up ldflags for version information
    - Configure CGO_ENABLED=0 for static linking
    - _Requirements: 6.1, 6.2, 5.1_
  
  - [x] 15.2 Configure archives and checksums
    - Set up tar.gz for Unix, zip for Windows
    - Include README, LICENSE, example config
    - Generate checksums file
    - _Requirements: 6.3, 6.4_
  
  - [x] 15.3 Configure changelog generation
    - Set up changelog from git commits
    - Filter out non-user-facing commits
    - _Requirements: 6.5_

- [x] 16. Set up GitHub Actions workflow
  - [x] 16.1 Create .github/workflows/release.yaml
    - Trigger on version tags (v*)
    - Set up Go environment
    - Run tests before release
    - Execute GoReleaser
    - _Requirements: 7.1, 7.2_
  
  - [x] 16.2 Configure release creation
    - Create GitHub release with artifacts
    - Attach checksums and changelog
    - Handle build failures
    - _Requirements: 7.3, 7.4, 7.5_
  
  - [x] 16.3 Add CI workflow for pull requests
    - Run tests on every PR
    - Run linting (golangci-lint)
    - Check build succeeds

- [x] 17. Create documentation
  - [x] 17.1 Write README.md
    - Add project overview and features
    - Document installation instructions
    - Document configuration options
    - Provide usage examples
    - Include troubleshooting section
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5_
  
  - [x] 17.2 Create config.example.yaml
    - Include all configuration options with comments
    - Provide sensible example values
    - _Requirements: 10.4_
  
  - [x] 17.3 Write CONTRIBUTING.md
    - Document development setup
    - Explain how to run tests
    - Describe release process

- [x] 18. Final checkpoint - Complete testing and validation
  - Run all tests (unit, property, integration)
  - Test manual installation and usage
  - Verify GoReleaser configuration with --snapshot
  - Ensure all documentation is complete
  - Ask the user if questions arise.

## Notes

- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation at key milestones
- Property tests validate universal correctness properties with minimum 100 iterations
- Unit tests validate specific examples and edge cases
- The implementation uses mark3labs/mcp-go SDK for MCP protocol handling
- All property tests should be tagged with: `Feature: nats-docs-mcp-server, Property {N}: {property_text}`
