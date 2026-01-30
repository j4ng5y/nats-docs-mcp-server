# Implementation Plan: Multi-Transport Support

## Overview

This implementation plan adds multi-transport support to the NATS Documentation MCP Server, enabling it to run with STDIO, SSE, or StreamableHTTP transports. The implementation follows TDD principles and maintains backward compatibility with existing STDIO-only deployments.

## Tasks

- [ ] 1. Extend configuration structure with transport settings
  - [ ] 1.1 Add transport fields to Config struct
    - Add `TransportType string` field (default: "stdio")
    - Add `Host string` field (default: "localhost")
    - Add `Port int` field (default: 0)
    - _Requirements: 1.1, 5.1, 5.2_
  
  - [ ] 1.2 Write property test for configuration precedence
    - **Property 1: Configuration Precedence**
    - **Validates: Requirements 1.2, 1.3, 1.4, 5.4, 5.5, 5.6, 5.7, 5.8, 5.9**
  
  - [ ] 1.3 Add GetTransportAddress() method to Config
    - Return empty string for STDIO transport
    - Return "host:port" for network transports
    - _Requirements: 5.1, 5.2_
  
  - [ ] 1.4 Write unit tests for GetTransportAddress()
    - Test STDIO returns empty string
    - Test SSE returns "host:port"
    - Test StreamableHTTP returns "host:port"
    - _Requirements: 5.1, 5.2_

- [ ] 2. Implement transport validation
  - [ ] 2.1 Add ValidateTransport() method to Config
    - Validate transport type is one of: "stdio", "sse", "streamablehttp"
    - Validate port is between 1-65535 for network transports
    - Validate host is non-empty if explicitly set
    - Return descriptive error messages
    - _Requirements: 1.6, 1.7, 6.1, 6.2, 6.3, 6.4, 6.5_
  
  - [ ] 2.2 Write property test for transport type validation
    - **Property 2: Transport Type Validation**
    - **Validates: Requirements 1.6, 1.7**
  
  - [ ] 2.3 Write property test for network transport port requirement
    - **Property 3: Network Transport Port Requirement**
    - **Validates: Requirements 3.2, 4.2, 5.1, 6.1, 6.2, 6.3**
  
  - [ ] 2.4 Write property test for host validation
    - **Property 4: Host Validation**
    - **Validates: Requirements 6.4**
  
  - [ ] 2.5 Write property test for STDIO transport configuration independence
    - **Property 5: STDIO Transport Configuration Independence**
    - **Validates: Requirements 2.2, 5.10**
  
  - [ ] 2.6 Write property test for validation error descriptiveness
    - **Property 7: Validation Error Descriptiveness**
    - **Validates: Requirements 6.5**

- [ ] 3. Update configuration loading functions
  - [ ] 3.1 Update Load() to read transport environment variables
    - Read TRANSPORT_TYPE, HOST, PORT from environment
    - Apply to Config struct
    - _Requirements: 1.3, 5.5, 5.8_
  
  - [ ] 3.2 Update LoadFromFile() to read transport config fields
    - Read transport_type, host, port from YAML
    - Apply to Config struct
    - _Requirements: 1.4, 5.6, 5.9_
  
  - [ ] 3.3 Update LoadWithFlags() to accept transport flags
    - Accept transport, host, port from flags map
    - Apply with highest precedence
    - _Requirements: 1.2, 5.4, 5.7_
  
  - [ ] 3.4 Write unit tests for configuration loading
    - Test environment variable loading
    - Test config file loading
    - Test flag loading with precedence
    - Test default values
    - _Requirements: 1.2, 1.3, 1.4, 1.5, 5.3_

- [ ] 4. Checkpoint - Ensure configuration tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 5. Create transport interface and implementations
  - [ ] 5.1 Create internal/server/transport.go file
    - Define TransportStarter interface with Start(), Shutdown(), Type() methods
    - _Requirements: 7.1, 7.2, 8.1_
  
  - [ ] 5.2 Implement StdioTransport
    - Implement Start() to call server.ServeStdio()
    - Implement Shutdown() (no-op for STDIO)
    - Implement Type() to return "stdio"
    - _Requirements: 2.1, 2.3, 2.4, 2.5_
  
  - [ ] 5.3 Implement SSETransport
    - Store address and server instance
    - Implement Start() to create and start SSE server
    - Implement Shutdown() to stop SSE server
    - Implement Type() to return "sse"
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7_
  
  - [ ] 5.4 Implement StreamableHTTPTransport
    - Store address and server instance
    - Implement Start() to create and start StreamableHTTP server
    - Implement Shutdown() to stop HTTP server
    - Implement Type() to return "streamablehttp"
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7_
  
  - [ ] 5.5 Write unit tests for transport implementations
    - Test StdioTransport.Type() returns "stdio"
    - Test SSETransport.Type() returns "sse"
    - Test StreamableHTTPTransport.Type() returns "streamablehttp"
    - _Requirements: 2.1, 3.1, 4.1_

- [ ] 6. Implement transport factory
  - [ ] 6.1 Implement NewTransport() function
    - Accept Config and logger parameters
    - Switch on transport type
    - Create appropriate transport implementation
    - Validate network transport has port configured
    - Return error for invalid transport type
    - _Requirements: 7.1, 7.2, 7.3_
  
  - [ ] 6.2 Write property test for transport initialization correctness
    - **Property 6: Transport Initialization Correctness**
    - **Validates: Requirements 7.2**
  
  - [ ] 6.3 Write unit tests for transport factory
    - Test NewTransport() with "stdio" creates StdioTransport
    - Test NewTransport() with "sse" and valid config creates SSETransport
    - Test NewTransport() with "streamablehttp" and valid config creates StreamableHTTPTransport
    - Test NewTransport() with invalid type returns error
    - Test NewTransport() with network transport and no port returns error
    - _Requirements: 7.2, 7.3_

- [ ] 7. Checkpoint - Ensure transport tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 8. Update Server struct to use transport abstraction
  - [ ] 8.1 Add transport field to Server struct
    - Add `transport TransportStarter` field
    - _Requirements: 7.1, 7.2_
  
  - [ ] 8.2 Update NewServer() to create transport
    - Call NewTransport() with config and logger
    - Store transport in Server struct
    - Return error if transport creation fails
    - _Requirements: 7.1, 7.2, 7.3_
  
  - [ ] 8.3 Update Server.Start() to use transport
    - Remove direct call to server.ServeStdio()
    - Call transport.Start() with context and mcpServer
    - Log transport type and address
    - _Requirements: 7.4, 7.5, 7.6_
  
  - [ ] 8.4 Update Server.Shutdown() to use transport
    - Call transport.Shutdown() with context
    - Log shutdown completion with transport type
    - _Requirements: 8.1, 8.6_
  
  - [ ] 8.5 Write unit tests for server transport integration
    - Test NewServer() creates transport
    - Test Start() calls transport.Start()
    - Test Shutdown() calls transport.Shutdown()
    - _Requirements: 7.1, 7.2, 8.1_

- [ ] 9. Update main.go with transport CLI flags
  - [ ] 9.1 Add transport CLI flags
    - Add --transport/-t flag for transport type
    - Add --host flag for network host
    - Add --port/-p flag for network port
    - _Requirements: 1.2, 5.4, 5.7_
  
  - [ ] 9.2 Update configuration override logic
    - Override transport type from flag if provided
    - Override host from flag if provided
    - Override port from flag if provided
    - _Requirements: 1.2, 5.4, 5.7_
  
  - [ ] 9.3 Add transport validation call
    - Call cfg.ValidateTransport() after loading config
    - Return error if validation fails
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 7.1_
  
  - [ ] 9.4 Update startup logging
    - Log transport type and address at startup
    - _Requirements: 7.6_

- [ ] 10. Checkpoint - Ensure integration tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 11. Update documentation
  - [ ] 11.1 Update README.md with transport information
    - Add "Transport Types" section describing STDIO, SSE, StreamableHTTP
    - Add use cases for each transport type
    - Add configuration examples for each transport
    - Add command-line examples for each transport
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.7_
  
  - [ ] 11.2 Create config.example.yaml with transport examples
    - Add commented examples for each transport type
    - Include host and port configuration
    - _Requirements: 9.6_
  
  - [ ] 11.3 Update CONTRIBUTING.md with transport testing guidance
    - Add instructions for testing each transport type
    - Add examples of running server with different transports
    - _Requirements: 9.3, 9.4, 9.5_

- [ ] 12. Create integration tests
  - [ ] 12.1 Write integration test for STDIO transport
    - Test full server startup with STDIO transport
    - Test server responds to MCP requests
    - Test graceful shutdown
    - _Requirements: 2.1, 2.3, 2.4, 2.5, 8.2, 10.5_
  
  - [ ] 12.2 Write integration test for SSE transport
    - Test full server startup with SSE transport
    - Test server binds to configured port
    - Test SSE endpoint is accessible
    - Test graceful shutdown
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 8.3, 8.4_
  
  - [ ] 12.3 Write integration test for StreamableHTTP transport
    - Test full server startup with StreamableHTTP transport
    - Test server binds to configured port
    - Test HTTP endpoint is accessible
    - Test graceful shutdown
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 8.3, 8.4_
  
  - [ ] 12.4 Write integration test for backward compatibility
    - Test server starts with no transport configuration
    - Test defaults to STDIO transport
    - Test existing config files work without modification
    - _Requirements: 10.1, 10.2, 10.3, 10.4_

- [ ] 13. Final checkpoint - Ensure all tests pass
  - Run all unit tests: `go test -v ./...`
  - Run all property tests: `go test -v ./... -run Property`
  - Run all integration tests: `go test -v -tags=integration ./...`
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- All tasks are required for comprehensive implementation
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- Integration tests validate end-to-end functionality
- Follow TDD principles: write tests first, then implementation
- Use GoReleaser for building binaries after implementation is complete
