package config

import (
	"testing"
)

// TestGetTransportAddressStdio verifies that GetTransportAddress returns empty string for STDIO transport
func TestGetTransportAddressStdio(t *testing.T) {
	cfg := NewConfig()
	cfg.TransportType = "stdio"
	cfg.Host = "localhost"
	cfg.Port = 8080

	address := cfg.GetTransportAddress()

	if address != "" {
		t.Errorf("Expected GetTransportAddress to return empty string for STDIO transport, got '%s'", address)
	}
}

// TestGetTransportAddressSSE verifies that GetTransportAddress returns "host:port" for SSE transport
func TestGetTransportAddressSSE(t *testing.T) {
	cfg := NewConfig()
	cfg.TransportType = "sse"
	cfg.Host = "localhost"
	cfg.Port = 8080

	address := cfg.GetTransportAddress()

	expected := "localhost:8080"
	if address != expected {
		t.Errorf("Expected GetTransportAddress to return '%s' for SSE transport, got '%s'", expected, address)
	}
}

// TestGetTransportAddressStreamableHTTP verifies that GetTransportAddress returns "host:port" for StreamableHTTP transport
func TestGetTransportAddressStreamableHTTP(t *testing.T) {
	cfg := NewConfig()
	cfg.TransportType = "streamablehttp"
	cfg.Host = "0.0.0.0"
	cfg.Port = 9090

	address := cfg.GetTransportAddress()

	expected := "0.0.0.0:9090"
	if address != expected {
		t.Errorf("Expected GetTransportAddress to return '%s' for StreamableHTTP transport, got '%s'", expected, address)
	}
}

// TestGetTransportAddressVariousHosts verifies that GetTransportAddress works with different host values
func TestGetTransportAddressVariousHosts(t *testing.T) {
	testCases := []struct {
		name          string
		transportType string
		host          string
		port          int
		expected      string
	}{
		{
			name:          "SSE with localhost",
			transportType: "sse",
			host:          "localhost",
			port:          8080,
			expected:      "localhost:8080",
		},
		{
			name:          "SSE with 0.0.0.0",
			transportType: "sse",
			host:          "0.0.0.0",
			port:          3000,
			expected:      "0.0.0.0:3000",
		},
		{
			name:          "StreamableHTTP with IP address",
			transportType: "streamablehttp",
			host:          "192.168.1.100",
			port:          8888,
			expected:      "192.168.1.100:8888",
		},
		{
			name:          "StreamableHTTP with domain",
			transportType: "streamablehttp",
			host:          "example.com",
			port:          443,
			expected:      "example.com:443",
		},
		{
			name:          "STDIO ignores host and port",
			transportType: "stdio",
			host:          "localhost",
			port:          8080,
			expected:      "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewConfig()
			cfg.TransportType = tc.transportType
			cfg.Host = tc.host
			cfg.Port = tc.port

			address := cfg.GetTransportAddress()

			if address != tc.expected {
				t.Errorf("Expected GetTransportAddress to return '%s', got '%s'", tc.expected, address)
			}
		})
	}
}

// TestGetTransportAddressWithZeroPort verifies behavior with port 0
func TestGetTransportAddressWithZeroPort(t *testing.T) {
	cfg := NewConfig()
	cfg.TransportType = "sse"
	cfg.Host = "localhost"
	cfg.Port = 0

	address := cfg.GetTransportAddress()

	expected := "localhost:0"
	if address != expected {
		t.Errorf("Expected GetTransportAddress to return '%s' with port 0, got '%s'", expected, address)
	}
}

// TestGetTransportAddressWithEmptyHost verifies behavior with empty host
func TestGetTransportAddressWithEmptyHost(t *testing.T) {
	cfg := NewConfig()
	cfg.TransportType = "sse"
	cfg.Host = ""
	cfg.Port = 8080

	address := cfg.GetTransportAddress()

	expected := ":8080"
	if address != expected {
		t.Errorf("Expected GetTransportAddress to return '%s' with empty host, got '%s'", expected, address)
	}
}
