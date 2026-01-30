# Design Document: Multi-Transport Support

## Overview

This design extends the NATS Documentation MCP Server to support multiple transport types for MCP protocol communication. Currently, the server only supports STDIO transport (standard input/output for local process communication). This enhancement adds support for SSE (Server-Sent Events) and StreamableHTTP transports, enabling the server to operate as a network service accessible to web clients, browser extensions, and remote MCP clients.

The design maintains backward compatibility with existing STDIO-only deployments while providing a flexible, configuration-driven approach to transport selection. The implementation leverages the mcp-go SDK's built-in transport abstractions to minimize code complexity and ensure protocol compliance.

**Key Design Principles:**
- Transport-agnostic server core (same MCP server instance works with any transport)
- Configuration-driven transport selection (CLI flags, env vars, config file)
- Graceful degradation (defaults to STDIO for backward compatibility)
- Clean separation between transport initialization and business logic
- Comprehensive validation before transport initialization

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Configuration Layer                      │
│  (CLI Flags → Config File → Env Vars → Defaults)           │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                  Transport Validator                         │
│  - Validates transport type                                  │
│  - Validates transport-specific config                       │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│                  Transport Factory                           │
│  - Creates appropriate transport based on config             │
│  - Initializes transport-specific settings                   │
└─────────────────────┬───────────────────────────────────────┘
                      │
          ┌───────────┴───────────┬───────────────────┐
          ▼                       ▼                   ▼
    ┌──────────┐          ┌──────────┐        ┌──────────────┐
    │  STDIO   │          │   SSE    │        │ StreamableHTTP│
    │Transport │          │Transport │        │   Transport   │
    └────┬─────┘          └────┬─────┘        └───────┬───────┘
         │                     │                      │
         └─────────────────────┴──────────────────────┘
                               │
                               ▼
                    ┌──────────────────────┐
                    │   MCP Server Core    │
                    │  (Transport-Agnostic)│
                    └──────────────────────┘
```

### Component Interaction Flow

**Startup Sequence:**
1. Load configuration (transport type, port, host, etc.)
2. Validate configuration (transport-specific validation)
3. Initialize MCP server core (transport-agnostic)
4. Fetch and index documentation (transport-agnostic)
5. Register MCP tools (transport-agnostic)
6. Create and start transport (transport-specific)

**Request Handling (All Transports):**
1. Transport receives client message
2. Transport deserializes JSON-RPC message
3. MCP server core routes to appropriate tool handler
4. Tool handler executes business logic
5. Tool handler returns result
6. Transport serializes and sends response

## Components and Interfaces

### 1. Configuration Component

**File:** `internal/config/config.go`

**Extended Config Structure:**
```go
type Config struct {
    // Existing fields
    LogLevel         string
    DocsBaseURL      string
    FetchTimeout     int
    MaxConcurrent    int
    CacheDir         string
    MaxSearchResults int
    
    // New transport fields
    TransportType    string // "stdio", "sse", or "streamablehttp"
    Host             string // Host to bind for network transports
    Port             int    // Port to bind for network transports
}
```

**New Functions:**
```go
// ValidateTransport validates transport-specific configuration
func (c *Config) ValidateTransport() error

// GetTransportAddress returns the full address for network transports
func (c *Config) GetTransportAddress() string
```

**Configuration Precedence:**
- CLI flags (highest priority)
- Config file
- Environment variables
- Defaults (lowest priority)

**Default Values:**
- `TransportType`: "stdio" (backward compatibility)
- `Host`: "localhost" (security by default)
- `Port`: 0 (must be explicitly configured for network transports)

### 2. Transport Factory Component

**New File:** `internal/server/transport.go`

**Purpose:** Encapsulates transport creation logic and provides a clean interface for initializing different transport types.

**Interface:**
```go
// TransportStarter defines the interface for starting a transport
type TransportStarter interface {
    Start(ctx context.Context, mcpServer *server.MCPServer) error
    Shutdown(ctx context.Context) error
}

// StdioTransport implements STDIO transport
type StdioTransport struct{}

// SSETransport implements SSE transport
type SSETransport struct {
    address string
    server  *server.SSEServer
}

// StreamableHTTPTransport implements StreamableHTTP transport
type StreamableHTTPTransport struct {
    address string
    server  *server.StreamableHTTPServer
}

// NewTransport creates the appropriate transport based on configuration
func NewTransport(cfg *config.Config, logger *slog.Logger) (TransportStarter, error)
```

**Transport Creation Logic:**
```go
func NewTransport(cfg *config.Config, logger *slog.Logger) (TransportStarter, error) {
    switch cfg.TransportType {
    case "stdio":
        return &StdioTransport{}, nil
    case "sse":
        if cfg.Port == 0 {
            return nil, fmt.Errorf("port must be configured for SSE transport")
        }
        return &SSETransport{
            address: cfg.GetTransportAddress(),
        }, nil
    case "streamablehttp":
        if cfg.Port == 0 {
            return nil, fmt.Errorf("port must be configured for StreamableHTTP transport")
        }
        return &StreamableHTTPTransport{
            address: cfg.GetTransportAddress(),
        }, nil
    default:
        return nil, fmt.Errorf("unsupported transport type: %s", cfg.TransportType)
    }
}
```

### 3. Server Component Updates

**File:** `internal/server/server.go`

**Modified Server Structure:**
```go
type Server struct {
    config      *config.Config
    index       *index.DocumentationIndex
    logger      *slog.Logger
    mcpServer   *server.MCPServer
    fetcher     *fetcher.DocumentationFetcher
    transport   TransportStarter  // New field
    initialized bool
}
```

**Modified Start Method:**
```go
func (s *Server) Start(ctx context.Context) error {
    if !s.initialized {
        return fmt.Errorf("server not initialized, call Initialize() first")
    }
    
    s.logger.Info("Starting MCP server", 
        "transport", s.config.TransportType,
        "address", s.config.GetTransportAddress())
    
    // Start the configured transport
    if err := s.transport.Start(ctx, s.mcpServer); err != nil {
        s.logger.Error("Transport start error", "error", err)
        return fmt.Errorf("transport start error: %w", err)
    }
    
    return nil
}
```

**Modified Shutdown Method:**
```go
func (s *Server) Shutdown(ctx context.Context) error {
    s.logger.Info("Shutting down server", "transport", s.config.TransportType)
    
    // Shutdown the transport
    if s.transport != nil {
        if err := s.transport.Shutdown(ctx); err != nil {
            s.logger.Error("Transport shutdown error", "error", err)
            return fmt.Errorf("transport shutdown error: %w", err)
        }
    }
    
    s.logger.Info("Server shutdown complete")
    return nil
}
```

### 4. Main Entry Point Updates

**File:** `cmd/server/main.go`

**New CLI Flags:**
```go
var (
    configFile    string
    logLevel      string
    showVersion   bool
    transportType string  // New flag
    host          string  // New flag
    port          int     // New flag
)

rootCmd.Flags().StringVarP(&transportType, "transport", "t", "", 
    "Transport type (stdio, sse, streamablehttp)")
rootCmd.Flags().StringVar(&host, "host", "", 
    "Host to bind for network transports (default: localhost)")
rootCmd.Flags().IntVarP(&port, "port", "p", 0, 
    "Port to bind for network transports")
```

**Configuration Override Logic:**
```go
// Override transport settings from command line flags if provided
if transportType != "" {
    cfg.TransportType = transportType
}
if host != "" {
    cfg.Host = host
}
if port != 0 {
    cfg.Port = port
}

// Validate transport configuration
if err := cfg.ValidateTransport(); err != nil {
    return fmt.Errorf("transport configuration validation failed: %w", err)
}
```

## Data Models

### Transport Configuration Model

```go
// TransportConfig holds transport-specific configuration
type TransportConfig struct {
    Type    string // "stdio", "sse", "streamablehttp"
    Host    string // For network transports
    Port    int    // For network transports
}

// Validate validates the transport configuration
func (tc *TransportConfig) Validate() error {
    // Validate transport type
    validTypes := map[string]bool{
        "stdio":          true,
        "sse":            true,
        "streamablehttp": true,
    }
    if !validTypes[tc.Type] {
        return fmt.Errorf("invalid transport type: %s (must be stdio, sse, or streamablehttp)", tc.Type)
    }
    
    // Validate network transport requirements
    if tc.Type == "sse" || tc.Type == "streamablehttp" {
        if tc.Port <= 0 || tc.Port > 65535 {
            return fmt.Errorf("port must be between 1 and 65535 for network transports, got: %d", tc.Port)
        }
        if tc.Host == "" {
            return fmt.Errorf("host cannot be empty for network transports")
        }
    }
    
    return nil
}

// Address returns the network address for network transports
func (tc *TransportConfig) Address() string {
    if tc.Type == "stdio" {
        return ""
    }
    return fmt.Sprintf("%s:%d", tc.Host, tc.Port)
}
```

### Transport Interface Model

```go
// TransportStarter defines the interface for all transport implementations
type TransportStarter interface {
    // Start initializes and starts the transport
    Start(ctx context.Context, mcpServer *server.MCPServer) error
    
    // Shutdown gracefully shuts down the transport
    Shutdown(ctx context.Context) error
    
    // Type returns the transport type name
    Type() string
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property Reflection

After analyzing all acceptance criteria, I've identified the following testable properties and examples. Here's the reflection to eliminate redundancy:

**Configuration Loading Properties:**
- Properties 1.2, 1.3, 1.4 (CLI flag, env var, config file) can be combined into a single comprehensive property about configuration source precedence
- Properties 5.4-5.9 (port and host configuration from different sources) are redundant with the general configuration loading property
- Keep: One property for configuration precedence across all sources

**Validation Properties:**
- Properties 1.7, 6.1, 6.2, 6.3, 6.4 all test validation error handling
- Property 6.5 tests that errors are descriptive (applies to all validation errors)
- These can be combined into comprehensive validation properties
- Keep: Separate properties for transport type validation and port/host validation (different concerns)

**Transport Type Properties:**
- Properties 1.6 (valid transport types) and 1.7 (invalid transport types) test the same validation logic
- Combine into one property: valid types accepted, invalid types rejected

**Network Transport Requirements:**
- Properties 3.2, 4.2, 5.1, 6.1 all test that network transports require port configuration
- These are redundant - combine into one property
- Keep: One property for network transport port requirement

**Default Behavior:**
- Properties 1.5, 5.3, 10.1, 10.2, 10.3 all test default values
- Combine into examples rather than properties (specific cases, not universal rules)

**STDIO Behavior:**
- Property 2.2 and 5.10 both test that STDIO doesn't require network configuration
- Combine into one property

**Redundant with SDK Behavior:**
- Properties 2.3, 2.4, 3.3-3.7, 4.3-4.7, 8.2-8.5 test mcp-go SDK behavior, not our code
- Remove from our property list (trust the SDK)

**Final Property Set:**
1. Configuration precedence property (combines 1.2-1.4, 5.4-5.9)
2. Transport type validation property (combines 1.6, 1.7)
3. Network transport port validation property (combines 3.2, 4.2, 5.1, 6.1-6.3)
4. Host validation property (6.4)
5. STDIO transport configuration property (combines 2.2, 5.10)
6. Transport initialization property (7.2)
7. Validation error descriptiveness property (6.5)

### Correctness Properties

Property 1: Configuration Precedence
*For any* configuration setting (transport type, port, host), when the same setting is provided through multiple sources (CLI flag, config file, environment variable), the value from the highest precedence source (CLI > config file > env var > default) should be used in the final configuration.
**Validates: Requirements 1.2, 1.3, 1.4, 5.4, 5.5, 5.6, 5.7, 5.8, 5.9**

Property 2: Transport Type Validation
*For any* transport type value, the configuration validation should accept exactly the three valid values ("stdio", "sse", "streamablehttp") and reject all other values with a descriptive error message listing the valid options.
**Validates: Requirements 1.6, 1.7**

Property 3: Network Transport Port Requirement
*For any* configuration where transport type is "sse" or "streamablehttp", validation should fail if port is not configured, or if port is less than 1, or if port is greater than 65535.
**Validates: Requirements 3.2, 4.2, 5.1, 6.1, 6.2, 6.3**

Property 4: Host Validation
*For any* configuration where host is explicitly set, validation should fail if the host value is an empty string.
**Validates: Requirements 6.4**

Property 5: STDIO Transport Configuration Independence
*For any* configuration where transport type is "stdio", validation should succeed regardless of whether port or host are configured (these settings should be ignored for STDIO transport).
**Validates: Requirements 2.2, 5.10**

Property 6: Transport Initialization Correctness
*For any* valid transport type ("stdio", "sse", "streamablehttp"), the transport factory should create an instance of the corresponding transport implementation type.
**Validates: Requirements 7.2**

Property 7: Validation Error Descriptiveness
*For any* configuration that fails validation, the returned error message should contain specific information about which validation rule failed and what the valid values or constraints are.
**Validates: Requirements 6.5**

## Error Handling

### Configuration Errors

**Invalid Transport Type:**
- Error: "invalid transport type: {value} (must be stdio, sse, or streamablehttp)"
- Action: Server exits with error code 1
- Logged at: ERROR level

**Missing Port for Network Transport:**
- Error: "port must be configured for {transport_type} transport"
- Action: Server exits with error code 1
- Logged at: ERROR level

**Invalid Port Value:**
- Error: "port must be between 1 and 65535 for network transports, got: {value}"
- Action: Server exits with error code 1
- Logged at: ERROR level

**Invalid Host Value:**
- Error: "host cannot be empty for network transports"
- Action: Server exits with error code 1
- Logged at: ERROR level

### Transport Initialization Errors

**Transport Creation Failure:**
- Error: "failed to create transport: {underlying_error}"
- Action: Server exits with error code 1
- Logged at: ERROR level

**Transport Start Failure:**
- Error: "transport start error: {underlying_error}"
- Action: Server exits with error code 1
- Logged at: ERROR level
- Context: Includes transport type and configuration

### Runtime Errors

**Network Binding Failure (SSE/StreamableHTTP):**
- Error: "failed to bind to {address}: {underlying_error}"
- Action: Server exits with error code 1
- Logged at: ERROR level
- Common causes: Port already in use, insufficient permissions

**Shutdown Timeout:**
- Error: "transport shutdown timeout exceeded"
- Action: Force shutdown after timeout
- Logged at: WARN level
- Timeout: 30 seconds (configurable via context)

### Error Recovery

**No Automatic Recovery:**
- Configuration errors require user intervention
- Transport initialization errors require user intervention
- Network binding errors require user intervention

**Graceful Degradation:**
- None applicable (server cannot function without valid transport)

**Error Reporting:**
- All errors logged with structured context (transport type, configuration values)
- Error messages include actionable guidance when possible
- Exit codes distinguish between configuration errors (1) and runtime errors (1)

## Testing Strategy

### Dual Testing Approach

This feature requires both unit tests and property-based tests to ensure comprehensive coverage:

**Unit Tests** focus on:
- Specific configuration examples (default values, specific transport types)
- Transport factory creation for each transport type
- Error message formatting and content
- Integration between components (config → validation → transport creation)

**Property-Based Tests** focus on:
- Configuration precedence across all possible values
- Validation rules across all possible inputs
- Transport type validation across all possible strings
- Port validation across all possible integers
- Host validation across all possible strings

### Property-Based Testing Configuration

**Testing Library:** [leanovate/gopter](https://github.com/leanovate/gopter)

**Test Configuration:**
- Minimum 100 iterations per property test
- Each property test tagged with: `Feature: multi-transport-support, Property {N}: {property_text}`

**Property Test Implementation:**

Each correctness property must be implemented as a single property-based test:

1. **Property 1: Configuration Precedence**
   - Generate random configuration values
   - Set same value in multiple sources with different values
   - Verify highest precedence source wins
   - Tag: `Feature: multi-transport-support, Property 1: Configuration precedence`

2. **Property 2: Transport Type Validation**
   - Generate random strings (valid and invalid transport types)
   - Verify valid types accepted, invalid types rejected with proper error
   - Tag: `Feature: multi-transport-support, Property 2: Transport type validation`

3. **Property 3: Network Transport Port Requirement**
   - Generate random transport types and port values
   - Verify network transports require valid port (1-65535)
   - Tag: `Feature: multi-transport-support, Property 3: Network transport port requirement`

4. **Property 4: Host Validation**
   - Generate random host values including empty strings
   - Verify empty host rejected with proper error
   - Tag: `Feature: multi-transport-support, Property 4: Host validation`

5. **Property 5: STDIO Transport Configuration Independence**
   - Generate random port and host values
   - Verify STDIO transport validates successfully regardless of port/host
   - Tag: `Feature: multi-transport-support, Property 5: STDIO transport configuration independence`

6. **Property 6: Transport Initialization Correctness**
   - Generate random valid transport types
   - Verify correct transport implementation created for each type
   - Tag: `Feature: multi-transport-support, Property 6: Transport initialization correctness`

7. **Property 7: Validation Error Descriptiveness**
   - Generate random invalid configurations
   - Verify error messages contain specific validation failure information
   - Tag: `Feature: multi-transport-support, Property 7: Validation error descriptiveness`

### Unit Test Coverage

**Configuration Tests:**
- Test default values (stdio transport, localhost host)
- Test loading from config file
- Test loading from environment variables
- Test CLI flag overrides
- Test GetTransportAddress() for each transport type

**Validation Tests:**
- Test ValidateTransport() with valid configurations
- Test ValidateTransport() with invalid configurations
- Test error message content for each validation failure

**Transport Factory Tests:**
- Test NewTransport() creates StdioTransport for "stdio"
- Test NewTransport() creates SSETransport for "sse" with valid config
- Test NewTransport() creates StreamableHTTPTransport for "streamablehttp" with valid config
- Test NewTransport() returns error for invalid transport type
- Test NewTransport() returns error for network transport without port

**Transport Implementation Tests:**
- Test StdioTransport.Start() calls server.ServeStdio()
- Test SSETransport.Start() creates and starts SSE server
- Test StreamableHTTPTransport.Start() creates and starts HTTP server
- Test each transport's Shutdown() method

**Integration Tests:**
- Test full server startup with each transport type
- Test configuration → validation → transport creation flow
- Test graceful shutdown for each transport type

### Test Organization

```
internal/
├── config/
│   ├── config_test.go              # Unit tests for configuration
│   └── config_property_test.go     # Property tests for configuration
├── server/
│   ├── transport_test.go           # Unit tests for transport factory
│   ├── transport_property_test.go  # Property tests for transport
│   └── server_integration_test.go  # Integration tests
```

### Testing Checkpoints

1. After implementing configuration changes (Config struct, validation)
2. After implementing transport factory (NewTransport, transport implementations)
3. After integrating with server (Start, Shutdown methods)
4. Before final review (run all tests including integration tests)

## Implementation Notes

### Backward Compatibility

**Critical:** The default transport type MUST be "stdio" to maintain backward compatibility with existing deployments. Any configuration without an explicit transport type should behave exactly as the current implementation.

**Validation:** Existing configuration files and command-line invocations without transport settings must continue to work without modification.

### Security Considerations

**Default Host:** Network transports default to "localhost" to prevent accidental exposure to external networks. Users must explicitly configure "0.0.0.0" or a specific interface to accept external connections.

**Port Binding:** Network transports should fail fast if the port is already in use, with a clear error message.

**STDIO Security:** STDIO transport remains the most secure option for local deployments as it requires no network exposure.

### Performance Considerations

**Transport Overhead:**
- STDIO: Minimal overhead (direct stdin/stdout)
- SSE: HTTP overhead + SSE streaming
- StreamableHTTP: HTTP overhead + connection management

**Memory Usage:**
- STDIO: No additional memory for transport
- SSE: Memory per active client connection
- StreamableHTTP: Memory per active client connection + HTTP server overhead

**Startup Time:**
- STDIO: No additional startup time
- SSE: Network binding time (~milliseconds)
- StreamableHTTP: Network binding time (~milliseconds)

### Logging Strategy

**Startup Logging:**
```
INFO: Starting NATS Documentation MCP Server
INFO: Transport: stdio
INFO: Initializing server (fetching and indexing documentation)
INFO: Server initialized successfully, starting MCP server
```

**Network Transport Startup:**
```
INFO: Starting NATS Documentation MCP Server
INFO: Transport: sse, Address: localhost:8080
INFO: Initializing server (fetching and indexing documentation)
INFO: Server initialized successfully, starting MCP server
INFO: SSE server listening on localhost:8080
```

**Shutdown Logging:**
```
INFO: Received shutdown signal: interrupt
INFO: Shutting down server, transport: stdio
INFO: Server shutdown complete
```

### Configuration Examples

**STDIO (Default):**
```yaml
# config.yaml
log_level: info
docs_base_url: https://docs.nats.io
# No transport configuration needed - defaults to stdio
```

**SSE Transport:**
```yaml
# config.yaml
log_level: info
docs_base_url: https://docs.nats.io
transport_type: sse
host: localhost
port: 8080
```

**StreamableHTTP Transport:**
```yaml
# config.yaml
log_level: info
docs_base_url: https://docs.nats.io
transport_type: streamablehttp
host: 0.0.0.0  # Accept external connections
port: 8080
```

**Environment Variables:**
```bash
export TRANSPORT_TYPE=sse
export HOST=localhost
export PORT=8080
./nats-docs-mcp-server
```

**CLI Flags:**
```bash
./nats-docs-mcp-server --transport sse --host localhost --port 8080
```

## References

- [MCP-Go Transport Documentation](https://mcp-go.dev/transports) - Official documentation for mcp-go transport types
- [MCP Specification](https://spec.modelcontextprotocol.io/) - Model Context Protocol specification
- [mcp-go SDK](https://github.com/mark3labs/mcp-go) - Go implementation of MCP protocol
- [12-Factor App Configuration](https://12factor.net/config) - Configuration best practices
