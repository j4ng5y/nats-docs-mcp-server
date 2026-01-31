// Package server provides the MCP server core implementation, handling protocol
// communication, tool registration, and request routing.
package server

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/server"
)

// TestTransportStarterInterface verifies that the TransportStarter interface
// is properly defined with the required methods.
func TestTransportStarterInterface(t *testing.T) {
	// This test verifies that the TransportStarter interface exists
	// and has the required methods by attempting to use it as a type.

	// Create a mock implementation to verify interface compliance
	var _ TransportStarter = (*mockTransport)(nil)
}

// mockTransport is a mock implementation of TransportStarter for testing
type mockTransport struct {
	startCalled    bool
	shutdownCalled bool
	transportType  string
}

func (m *mockTransport) Start(ctx context.Context, mcpServer *server.MCPServer) error {
	m.startCalled = true
	return nil
}

func (m *mockTransport) Shutdown(ctx context.Context) error {
	m.shutdownCalled = true
	return nil
}

func (m *mockTransport) Type() string {
	return m.transportType
}

// TestMockTransportImplementation verifies that our mock transport
// correctly implements the TransportStarter interface methods.
func TestMockTransportImplementation(t *testing.T) {
	mock := &mockTransport{transportType: "mock"}

	// Test Type() method
	if got := mock.Type(); got != "mock" {
		t.Errorf("Type() = %q, want %q", got, "mock")
	}

	// Test Start() method
	ctx := context.Background()
	if err := mock.Start(ctx, nil); err != nil {
		t.Errorf("Start() error = %v, want nil", err)
	}
	if !mock.startCalled {
		t.Error("Start() did not set startCalled flag")
	}

	// Test Shutdown() method
	if err := mock.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown() error = %v, want nil", err)
	}
	if !mock.shutdownCalled {
		t.Error("Shutdown() did not set shutdownCalled flag")
	}
}

// TestStdioTransportType verifies that StdioTransport.Type() returns "stdio"
// Requirements: 2.1
func TestStdioTransportType(t *testing.T) {
	transport := &StdioTransport{}

	got := transport.Type()
	want := "stdio"

	if got != want {
		t.Errorf("StdioTransport.Type() = %q, want %q", got, want)
	}
}

// TestStdioTransportShutdown verifies that StdioTransport.Shutdown() is a no-op
// and returns no error.
// Requirements: 2.1
func TestStdioTransportShutdown(t *testing.T) {
	transport := &StdioTransport{}
	ctx := context.Background()

	err := transport.Shutdown(ctx)

	if err != nil {
		t.Errorf("StdioTransport.Shutdown() error = %v, want nil", err)
	}
}

// TestStdioTransportImplementsInterface verifies that StdioTransport
// implements the TransportStarter interface.
func TestStdioTransportImplementsInterface(t *testing.T) {
	var _ TransportStarter = (*StdioTransport)(nil)
}

// TestSSETransportType verifies that SSETransport.Type() returns "sse"
// Requirements: 3.1
func TestSSETransportType(t *testing.T) {
	transport := &SSETransport{
		address: "localhost:8080",
	}

	got := transport.Type()
	want := "sse"

	if got != want {
		t.Errorf("SSETransport.Type() = %q, want %q", got, want)
	}
}

// TestSSETransportImplementsInterface verifies that SSETransport
// implements the TransportStarter interface.
func TestSSETransportImplementsInterface(t *testing.T) {
	var _ TransportStarter = (*SSETransport)(nil)
}

// TestSSETransportShutdown verifies that SSETransport.Shutdown() handles
// nil server gracefully and returns no error.
// Requirements: 3.7, 8.3, 8.4
func TestSSETransportShutdown(t *testing.T) {
	transport := &SSETransport{
		address: "localhost:8080",
		server:  nil, // No server initialized yet
	}
	ctx := context.Background()

	err := transport.Shutdown(ctx)

	if err != nil {
		t.Errorf("SSETransport.Shutdown() with nil server error = %v, want nil", err)
	}
}

// TestStreamableHTTPTransportType verifies that StreamableHTTPTransport.Type() returns "streamablehttp"
// Requirements: 4.1
func TestStreamableHTTPTransportType(t *testing.T) {
	transport := &StreamableHTTPTransport{
		address: "localhost:8080",
	}

	got := transport.Type()
	want := "streamablehttp"

	if got != want {
		t.Errorf("StreamableHTTPTransport.Type() = %q, want %q", got, want)
	}
}

// TestStreamableHTTPTransportImplementsInterface verifies that StreamableHTTPTransport
// implements the TransportStarter interface.
func TestStreamableHTTPTransportImplementsInterface(t *testing.T) {
	var _ TransportStarter = (*StreamableHTTPTransport)(nil)
}

// TestStreamableHTTPTransportShutdown verifies that StreamableHTTPTransport.Shutdown() handles
// nil server gracefully and returns no error.
// Requirements: 4.7, 8.3, 8.4
func TestStreamableHTTPTransportShutdown(t *testing.T) {
	transport := &StreamableHTTPTransport{
		address: "localhost:8080",
		server:  nil, // No server initialized yet
	}
	ctx := context.Background()

	err := transport.Shutdown(ctx)

	if err != nil {
		t.Errorf("StreamableHTTPTransport.Shutdown() with nil server error = %v, want nil", err)
	}
}

// TestNewTransportStdio verifies that NewTransport() creates a StdioTransport
// when transport type is "stdio".
// Requirements: 7.2
func TestNewTransportStdio(t *testing.T) {
	cfg := &mockConfig{
		transportType: "stdio",
		host:          "localhost",
		port:          0,
	}

	transport, err := NewTransport(cfg, nil)

	if err != nil {
		t.Fatalf("NewTransport() error = %v, want nil", err)
	}

	if transport == nil {
		t.Fatal("NewTransport() returned nil transport")
	}

	if transport.Type() != "stdio" {
		t.Errorf("NewTransport() created transport type %q, want %q", transport.Type(), "stdio")
	}

	// Verify it's actually a StdioTransport
	if _, ok := transport.(*StdioTransport); !ok {
		t.Errorf("NewTransport() created %T, want *StdioTransport", transport)
	}
}

// TestNewTransportSSE verifies that NewTransport() creates an SSETransport
// when transport type is "sse" and port is configured.
// Requirements: 7.2
func TestNewTransportSSE(t *testing.T) {
	cfg := &mockConfig{
		transportType: "sse",
		host:          "localhost",
		port:          8080,
	}

	transport, err := NewTransport(cfg, nil)

	if err != nil {
		t.Fatalf("NewTransport() error = %v, want nil", err)
	}

	if transport == nil {
		t.Fatal("NewTransport() returned nil transport")
	}

	if transport.Type() != "sse" {
		t.Errorf("NewTransport() created transport type %q, want %q", transport.Type(), "sse")
	}

	// Verify it's actually an SSETransport
	if _, ok := transport.(*SSETransport); !ok {
		t.Errorf("NewTransport() created %T, want *SSETransport", transport)
	}
}

// TestNewTransportStreamableHTTP verifies that NewTransport() creates a StreamableHTTPTransport
// when transport type is "streamablehttp" and port is configured.
// Requirements: 7.2
func TestNewTransportStreamableHTTP(t *testing.T) {
	cfg := &mockConfig{
		transportType: "streamablehttp",
		host:          "localhost",
		port:          8080,
	}

	transport, err := NewTransport(cfg, nil)

	if err != nil {
		t.Fatalf("NewTransport() error = %v, want nil", err)
	}

	if transport == nil {
		t.Fatal("NewTransport() returned nil transport")
	}

	if transport.Type() != "streamablehttp" {
		t.Errorf("NewTransport() created transport type %q, want %q", transport.Type(), "streamablehttp")
	}

	// Verify it's actually a StreamableHTTPTransport
	if _, ok := transport.(*StreamableHTTPTransport); !ok {
		t.Errorf("NewTransport() created %T, want *StreamableHTTPTransport", transport)
	}
}

// TestNewTransportInvalidType verifies that NewTransport() returns an error
// when an invalid transport type is provided.
// Requirements: 7.3
func TestNewTransportInvalidType(t *testing.T) {
	cfg := &mockConfig{
		transportType: "invalid",
		host:          "localhost",
		port:          8080,
	}

	transport, err := NewTransport(cfg, nil)

	if err == nil {
		t.Fatal("NewTransport() with invalid type error = nil, want error")
	}

	if transport != nil {
		t.Errorf("NewTransport() with invalid type returned transport %v, want nil", transport)
	}

	// Verify error message mentions the invalid transport type
	wantSubstring := "invalid"
	if !containsSubstring(err.Error(), wantSubstring) {
		t.Errorf("NewTransport() error = %q, want error containing %q", err.Error(), wantSubstring)
	}
}

// TestNewTransportSSEWithoutPort verifies that NewTransport() returns an error
// when transport type is "sse" but port is not configured.
// Requirements: 7.3
func TestNewTransportSSEWithoutPort(t *testing.T) {
	cfg := &mockConfig{
		transportType: "sse",
		host:          "localhost",
		port:          0, // No port configured
	}

	transport, err := NewTransport(cfg, nil)

	if err == nil {
		t.Fatal("NewTransport() with SSE and no port error = nil, want error")
	}

	if transport != nil {
		t.Errorf("NewTransport() with SSE and no port returned transport %v, want nil", transport)
	}

	// Verify error message mentions port requirement
	wantSubstring := "port"
	if !containsSubstring(err.Error(), wantSubstring) {
		t.Errorf("NewTransport() error = %q, want error containing %q", err.Error(), wantSubstring)
	}
}

// TestNewTransportStreamableHTTPWithoutPort verifies that NewTransport() returns an error
// when transport type is "streamablehttp" but port is not configured.
// Requirements: 7.3
func TestNewTransportStreamableHTTPWithoutPort(t *testing.T) {
	cfg := &mockConfig{
		transportType: "streamablehttp",
		host:          "localhost",
		port:          0, // No port configured
	}

	transport, err := NewTransport(cfg, nil)

	if err == nil {
		t.Fatal("NewTransport() with StreamableHTTP and no port error = nil, want error")
	}

	if transport != nil {
		t.Errorf("NewTransport() with StreamableHTTP and no port returned transport %v, want nil", transport)
	}

	// Verify error message mentions port requirement
	wantSubstring := "port"
	if !containsSubstring(err.Error(), wantSubstring) {
		t.Errorf("NewTransport() error = %q, want error containing %q", err.Error(), wantSubstring)
	}
}

// mockConfig is a mock implementation of the transportConfig interface for testing
type mockConfig struct {
	transportType string
	host          string
	port          int
}

func (m *mockConfig) GetTransportType() string {
	return m.transportType
}

func (m *mockConfig) GetPort() int {
	return m.port
}

func (m *mockConfig) GetTransportAddress() string {
	if m.transportType == "stdio" {
		return ""
	}
	return fmt.Sprintf("%s:%d", m.host, m.port)
}

// containsSubstring checks if a string contains a substring (case-insensitive)
func containsSubstring(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
