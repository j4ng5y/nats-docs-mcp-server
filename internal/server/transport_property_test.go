// Package server provides the MCP server core implementation, handling protocol
// communication, tool registration, and request routing.
package server

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty6_TransportInitializationCorrectness verifies that for any valid
// transport type ("stdio", "sse", "streamablehttp"), the transport factory creates
// an instance of the corresponding transport implementation type.
//
// Feature: multi-transport-support, Property 6: Transport initialization correctness
//
// Validates: Requirements 7.2
func TestProperty6_TransportInitializationCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: For any valid transport type, NewTransport creates the correct implementation
	properties.Property("NewTransport creates correct transport type for valid transport types",
		prop.ForAll(
			func(transportType string, port int) bool {
				// Create mock config with the generated transport type
				cfg := &mockConfig{
					transportType: transportType,
					host:          "localhost",
					port:          port,
				}

				// Call NewTransport
				transport, err := NewTransport(cfg, nil)

				// For valid transport types, we should get a transport without error
				if err != nil {
					t.Logf("NewTransport returned error for transport type %q: %v", transportType, err)
					return false
				}

				if transport == nil {
					t.Logf("NewTransport returned nil transport for transport type %q", transportType)
					return false
				}

				// Verify the transport type matches what we requested
				if transport.Type() != transportType {
					t.Logf("NewTransport created transport with type %q, expected %q", transport.Type(), transportType)
					return false
				}

				// Verify the correct implementation type was created
				switch transportType {
				case "stdio":
					if _, ok := transport.(*StdioTransport); !ok {
						t.Logf("NewTransport created %T for stdio, expected *StdioTransport", transport)
						return false
					}
				case "sse":
					if _, ok := transport.(*SSETransport); !ok {
						t.Logf("NewTransport created %T for sse, expected *SSETransport", transport)
						return false
					}
				case "streamablehttp":
					if _, ok := transport.(*StreamableHTTPTransport); !ok {
						t.Logf("NewTransport created %T for streamablehttp, expected *StreamableHTTPTransport", transport)
						return false
					}
				}

				return true
			},
			// Generate valid transport types
			gen.OneConstOf("stdio", "sse", "streamablehttp"),
			// Generate valid port numbers (1-65535) for network transports
			// For stdio, port is ignored, so any value is fine
			gen.IntRange(1, 65535),
		))

	properties.TestingRun(t)
}
