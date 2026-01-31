// Package server provides the MCP server core implementation, handling protocol
// communication, tool registration, and request routing.
package server

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/server"
)

// TransportStarter defines the interface for all transport implementations.
// It provides a common abstraction for starting and stopping different transport types
// (STDIO, SSE, StreamableHTTP) used by the MCP server.
//
// Implementations must handle:
//   - Starting the transport and establishing communication with the MCP server
//   - Gracefully shutting down the transport and cleaning up resources
//   - Reporting the transport type for logging and diagnostics
type TransportStarter interface {
	// Start initializes and starts the transport with the given MCP server.
	// It should block until the transport is ready to accept connections or
	// until an error occurs.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//   - mcpServer: The MCP server instance to bind to this transport
	//
	// Returns an error if the transport fails to start.
	Start(ctx context.Context, mcpServer *server.MCPServer) error

	// Shutdown gracefully shuts down the transport and cleans up resources.
	// It should close all active connections and stop accepting new ones.
	//
	// Parameters:
	//   - ctx: Context for cancellation and timeout control
	//
	// Returns an error if shutdown fails.
	Shutdown(ctx context.Context) error

	// Type returns the transport type name for logging and diagnostics.
	// Valid values are: "stdio", "sse", "streamablehttp"
	//
	// Returns the transport type as a string.
	Type() string
}

// StdioTransport implements the TransportStarter interface for STDIO transport.
// STDIO transport uses standard input/output for MCP protocol communication,
// making it suitable for local process-based integrations.
//
// This transport:
//   - Reads MCP requests from stdin
//   - Writes MCP responses to stdout
//   - Writes logs to stderr (to avoid protocol interference)
//   - Blocks until the client disconnects or the process terminates
//   - Requires no network configuration (host/port)
type StdioTransport struct{}

// Start initializes and starts the STDIO transport by calling server.ServeStdio().
// This method blocks until the client disconnects or the context is cancelled.
//
// Parameters:
//   - ctx: Context for cancellation control (currently unused by ServeStdio)
//   - mcpServer: The MCP server instance to serve via STDIO
//
// Returns an error if ServeStdio fails to start.
//
// Requirements: 2.1, 2.3, 2.4, 2.5
func (s *StdioTransport) Start(ctx context.Context, mcpServer *server.MCPServer) error {
	return server.ServeStdio(mcpServer)
}

// Shutdown gracefully shuts down the STDIO transport.
// For STDIO transport, this is a no-op since stdin/stdout are managed by the OS
// and will be closed when the process terminates.
//
// Parameters:
//   - ctx: Context for cancellation control (unused for STDIO)
//
// Returns nil (STDIO shutdown never fails).
//
// Requirements: 2.1
func (s *StdioTransport) Shutdown(ctx context.Context) error {
	// No-op for STDIO transport - stdin/stdout are managed by the OS
	return nil
}

// Type returns the transport type identifier for STDIO transport.
//
// Returns "stdio" as the transport type string.
//
// Requirements: 2.1
func (s *StdioTransport) Type() string {
	return "stdio"
}

// SSETransport implements the TransportStarter interface for SSE transport.
// SSE (Server-Sent Events) transport uses HTTP with SSE for MCP protocol communication,
// making it suitable for web-based clients and browser integrations.
//
// This transport:
//   - Creates an HTTP server with SSE endpoint
//   - Sends MCP responses as SSE events
//   - Supports multiple concurrent client sessions
//   - Requires network configuration (host and port)
//   - Enables real-time server-to-client notifications
type SSETransport struct {
	address string
	server  *server.SSEServer
}

// Start initializes and starts the SSE transport by creating an SSE server
// and binding it to the configured address.
// This method blocks until the server stops or an error occurs.
//
// Parameters:
//   - ctx: Context for cancellation control
//   - mcpServer: The MCP server instance to serve via SSE
//
// Returns an error if the SSE server fails to start.
//
// Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7
func (s *SSETransport) Start(ctx context.Context, mcpServer *server.MCPServer) error {
	// Create SSE server
	s.server = server.NewSSEServer(mcpServer)

	// Start the SSE server on the configured address
	return s.server.Start(s.address)
}

// Shutdown gracefully shuts down the SSE transport.
// This stops the HTTP server and closes all active client connections.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns an error if shutdown fails.
//
// Requirements: 3.7, 8.3, 8.4
func (s *SSETransport) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// Type returns the transport type identifier for SSE transport.
//
// Returns "sse" as the transport type string.
//
// Requirements: 3.1
func (s *SSETransport) Type() string {
	return "sse"
}

// StreamableHTTPTransport implements the TransportStarter interface for StreamableHTTP transport.
// StreamableHTTP transport uses HTTP with SSE for MCP protocol communication,
// making it suitable for full-featured web integration and enterprise deployments.
//
// This transport:
//   - Creates a full HTTP server with SSE capabilities
//   - Supports HTTP POST requests for client-to-server messages
//   - Sends MCP responses as SSE events
//   - Supports multiple concurrent client sessions
//   - Requires network configuration (host and port)
//   - Provides HTTP health check endpoint
type StreamableHTTPTransport struct {
	address string
	server  *server.StreamableHTTPServer
}

// Start initializes and starts the StreamableHTTP transport by creating a StreamableHTTP server
// and binding it to the configured address.
// This method blocks until the server stops or an error occurs.
//
// Parameters:
//   - ctx: Context for cancellation control
//   - mcpServer: The MCP server instance to serve via StreamableHTTP
//
// Returns an error if the StreamableHTTP server fails to start.
//
// Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7
func (s *StreamableHTTPTransport) Start(ctx context.Context, mcpServer *server.MCPServer) error {
	// Create StreamableHTTP server
	s.server = server.NewStreamableHTTPServer(mcpServer)

	// Start the StreamableHTTP server on the configured address
	return s.server.Start(s.address)
}

// Shutdown gracefully shuts down the StreamableHTTP transport.
// This stops the HTTP server and closes all active client connections.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns an error if shutdown fails.
//
// Requirements: 4.7, 8.3, 8.4
func (s *StreamableHTTPTransport) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	return s.server.Shutdown(ctx)
}

// Type returns the transport type identifier for StreamableHTTP transport.
//
// Returns "streamablehttp" as the transport type string.
//
// Requirements: 4.1
func (s *StreamableHTTPTransport) Type() string {
	return "streamablehttp"
}

// NewTransport creates the appropriate transport based on configuration.
// It validates that network transports (SSE and StreamableHTTP) have a port configured,
// and returns an error for invalid transport types.
//
// Parameters:
//   - cfg: Configuration containing transport type, host, and port settings
//   - logger: Logger instance for logging (currently unused but included for future use)
//
// Returns:
//   - TransportStarter: The created transport implementation
//   - error: An error if the transport type is invalid or required configuration is missing
//
// Requirements: 7.1, 7.2, 7.3
func NewTransport(cfg transportConfig, logger interface{}) (TransportStarter, error) {
	switch cfg.GetTransportType() {
	case "stdio":
		return &StdioTransport{}, nil
	case "sse":
		if cfg.GetPort() == 0 {
			return nil, fmt.Errorf("port must be configured for SSE transport")
		}
		return &SSETransport{
			address: cfg.GetTransportAddress(),
		}, nil
	case "streamablehttp":
		if cfg.GetPort() == 0 {
			return nil, fmt.Errorf("port must be configured for StreamableHTTP transport")
		}
		return &StreamableHTTPTransport{
			address: cfg.GetTransportAddress(),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported transport type: %s (must be one of: stdio, sse, streamablehttp)", cfg.GetTransportType())
	}
}

// transportConfig defines the interface for configuration objects used by NewTransport.
// This allows the function to work with different config implementations (real Config or mocks).
type transportConfig interface {
	GetTransportType() string
	GetPort() int
	GetTransportAddress() string
}
