// Package server provides the MCP server core implementation, handling protocol
// communication, tool registration, and request routing.
package server

import (
	"testing"

	"github.com/j4ng5y/nats-docs-mcp-server/internal/config"
)

// TestNewTransportWithRealConfig verifies that NewTransport works with the real Config struct
// from the config package. This integration test ensures the transportConfig interface
// is properly implemented by config.Config.
func TestNewTransportWithRealConfig(t *testing.T) {
	tests := []struct {
		name         string
		cfg          *config.Config
		wantType     string
		wantErr      bool
		errSubstring string
	}{
		{
			name: "stdio transport with real config",
			cfg: &config.Config{
				TransportType: "stdio",
				Host:          "localhost",
				Port:          0,
			},
			wantType: "stdio",
			wantErr:  false,
		},
		{
			name: "sse transport with real config",
			cfg: &config.Config{
				TransportType: "sse",
				Host:          "localhost",
				Port:          8080,
			},
			wantType: "sse",
			wantErr:  false,
		},
		{
			name: "streamablehttp transport with real config",
			cfg: &config.Config{
				TransportType: "streamablehttp",
				Host:          "localhost",
				Port:          8080,
			},
			wantType: "streamablehttp",
			wantErr:  false,
		},
		{
			name: "sse transport without port",
			cfg: &config.Config{
				TransportType: "sse",
				Host:          "localhost",
				Port:          0,
			},
			wantErr:      true,
			errSubstring: "port",
		},
		{
			name: "invalid transport type",
			cfg: &config.Config{
				TransportType: "invalid",
				Host:          "localhost",
				Port:          8080,
			},
			wantErr:      true,
			errSubstring: "unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport, err := NewTransport(tt.cfg, nil)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewTransport() error = nil, want error containing %q", tt.errSubstring)
					return
				}
				if !containsSubstring(err.Error(), tt.errSubstring) {
					t.Errorf("NewTransport() error = %q, want error containing %q", err.Error(), tt.errSubstring)
				}
				return
			}

			if err != nil {
				t.Errorf("NewTransport() unexpected error = %v", err)
				return
			}

			if transport == nil {
				t.Error("NewTransport() returned nil transport")
				return
			}

			if got := transport.Type(); got != tt.wantType {
				t.Errorf("NewTransport() created transport type %q, want %q", got, tt.wantType)
			}
		})
	}
}

// TestConfigImplementsTransportConfig verifies that config.Config implements
// the transportConfig interface required by NewTransport.
func TestConfigImplementsTransportConfig(t *testing.T) {
	cfg := &config.Config{
		TransportType: "stdio",
		Host:          "localhost",
		Port:          8080,
	}

	// Verify GetTransportType method
	if got := cfg.GetTransportType(); got != "stdio" {
		t.Errorf("Config.GetTransportType() = %q, want %q", got, "stdio")
	}

	// Verify GetPort method
	if got := cfg.GetPort(); got != 8080 {
		t.Errorf("Config.GetPort() = %d, want %d", got, 8080)
	}

	// Verify GetTransportAddress method
	cfg.TransportType = "sse"
	wantAddr := "localhost:8080"
	if got := cfg.GetTransportAddress(); got != wantAddr {
		t.Errorf("Config.GetTransportAddress() = %q, want %q", got, wantAddr)
	}

	// Verify GetTransportAddress returns empty for stdio
	cfg.TransportType = "stdio"
	if got := cfg.GetTransportAddress(); got != "" {
		t.Errorf("Config.GetTransportAddress() for stdio = %q, want empty string", got)
	}
}
