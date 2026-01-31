package config

import (
	"strings"
	"testing"
)

// TestValidateTransportValidTypes verifies that valid transport types are accepted
func TestValidateTransportValidTypes(t *testing.T) {
	testCases := []struct {
		transportType string
		port          int
		host          string
	}{
		{"stdio", 0, "localhost"},
		{"sse", 8080, "localhost"},
		{"streamablehttp", 8080, "localhost"},
	}

	for _, tc := range testCases {
		cfg := NewConfig()
		cfg.TransportType = tc.transportType
		cfg.Port = tc.port
		cfg.Host = tc.host

		err := cfg.ValidateTransport()
		if err != nil {
			t.Errorf("Expected ValidateTransport to accept transport type '%s', got error: %v", tc.transportType, err)
		}
	}
}

// TestValidateTransportInvalidType verifies that invalid transport types are rejected
func TestValidateTransportInvalidType(t *testing.T) {
	invalidTypes := []string{"invalid", "http", "tcp", "websocket", ""}

	for _, transportType := range invalidTypes {
		cfg := NewConfig()
		cfg.TransportType = transportType

		err := cfg.ValidateTransport()
		if err == nil {
			t.Errorf("Expected ValidateTransport to reject invalid transport type '%s'", transportType)
		}

		// Verify error message lists valid options
		if err != nil {
			errMsg := err.Error()
			if !strings.Contains(errMsg, "stdio") || !strings.Contains(errMsg, "sse") || !strings.Contains(errMsg, "streamablehttp") {
				t.Errorf("Expected error message to list valid transport types, got: %v", err)
			}
		}
	}
}

// TestValidateTransportNetworkTransportRequiresPort verifies that network transports require a valid port
func TestValidateTransportNetworkTransportRequiresPort(t *testing.T) {
	networkTransports := []string{"sse", "streamablehttp"}

	for _, transportType := range networkTransports {
		// Test with port = 0 (not configured)
		cfg := NewConfig()
		cfg.TransportType = transportType
		cfg.Port = 0

		err := cfg.ValidateTransport()
		if err == nil {
			t.Errorf("Expected ValidateTransport to reject %s transport with port 0", transportType)
		}

		// Test with negative port
		cfg.Port = -1
		err = cfg.ValidateTransport()
		if err == nil {
			t.Errorf("Expected ValidateTransport to reject %s transport with negative port", transportType)
		}

		// Test with port > 65535
		cfg.Port = 65536
		err = cfg.ValidateTransport()
		if err == nil {
			t.Errorf("Expected ValidateTransport to reject %s transport with port > 65535", transportType)
		}
	}
}

// TestValidateTransportNetworkTransportValidPorts verifies that network transports accept valid ports
func TestValidateTransportNetworkTransportValidPorts(t *testing.T) {
	networkTransports := []string{"sse", "streamablehttp"}
	validPorts := []int{1, 80, 443, 8080, 65535}

	for _, transportType := range networkTransports {
		for _, port := range validPorts {
			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = port

			err := cfg.ValidateTransport()
			if err != nil {
				t.Errorf("Expected ValidateTransport to accept %s transport with port %d, got error: %v", transportType, port, err)
			}
		}
	}
}

// TestValidateTransportStdioIgnoresPort verifies that STDIO transport doesn't require port
func TestValidateTransportStdioIgnoresPort(t *testing.T) {
	// Test with various port values (including invalid ones)
	portValues := []int{0, -1, 65536, 100000}

	for _, port := range portValues {
		cfg := NewConfig()
		cfg.TransportType = "stdio"
		cfg.Port = port

		err := cfg.ValidateTransport()
		if err != nil {
			t.Errorf("Expected ValidateTransport to accept stdio transport with any port value %d, got error: %v", port, err)
		}
	}
}

// TestValidateTransportEmptyHostRejected verifies that empty host is rejected when explicitly set
func TestValidateTransportEmptyHostRejected(t *testing.T) {
	// For network transports, empty host should be rejected
	networkTransports := []string{"sse", "streamablehttp"}

	for _, transportType := range networkTransports {
		cfg := NewConfig()
		cfg.TransportType = transportType
		cfg.Port = 8080
		cfg.Host = ""

		err := cfg.ValidateTransport()
		if err == nil {
			t.Errorf("Expected ValidateTransport to reject %s transport with empty host", transportType)
		}

		if err != nil && !strings.Contains(err.Error(), "host") {
			t.Errorf("Expected error message to mention host, got: %v", err)
		}
	}
}

// TestValidateTransportNonEmptyHostAccepted verifies that non-empty host is accepted
func TestValidateTransportNonEmptyHostAccepted(t *testing.T) {
	networkTransports := []string{"sse", "streamablehttp"}
	validHosts := []string{"localhost", "127.0.0.1", "0.0.0.0", "example.com"}

	for _, transportType := range networkTransports {
		for _, host := range validHosts {
			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = 8080
			cfg.Host = host

			err := cfg.ValidateTransport()
			if err != nil {
				t.Errorf("Expected ValidateTransport to accept %s transport with host '%s', got error: %v", transportType, host, err)
			}
		}
	}
}

// TestValidateTransportErrorMessages verifies that error messages are descriptive
func TestValidateTransportErrorMessages(t *testing.T) {
	testCases := []struct {
		name          string
		transportType string
		port          int
		host          string
		expectError   bool
		errorContains []string
	}{
		{
			name:          "invalid transport type",
			transportType: "invalid",
			port:          8080,
			host:          "localhost",
			expectError:   true,
			errorContains: []string{"invalid", "transport", "stdio", "sse", "streamablehttp"},
		},
		{
			name:          "network transport missing port",
			transportType: "sse",
			port:          0,
			host:          "localhost",
			expectError:   true,
			errorContains: []string{"port", "1", "65535"},
		},
		{
			name:          "network transport port too high",
			transportType: "streamablehttp",
			port:          70000,
			host:          "localhost",
			expectError:   true,
			errorContains: []string{"port", "65535"},
		},
		{
			name:          "network transport empty host",
			transportType: "sse",
			port:          8080,
			host:          "",
			expectError:   true,
			errorContains: []string{"host", "empty"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewConfig()
			cfg.TransportType = tc.transportType
			cfg.Port = tc.port
			cfg.Host = tc.host

			err := cfg.ValidateTransport()

			if tc.expectError && err == nil {
				t.Errorf("Expected error for test case '%s', but got nil", tc.name)
				return
			}

			if !tc.expectError && err != nil {
				t.Errorf("Expected no error for test case '%s', but got: %v", tc.name, err)
				return
			}

			if tc.expectError && err != nil {
				errMsg := err.Error()
				for _, expectedStr := range tc.errorContains {
					if !strings.Contains(errMsg, expectedStr) {
						t.Errorf("Expected error message to contain '%s', got: %v", expectedStr, err)
					}
				}
			}
		})
	}
}

// TestValidateTransportStdioWithEmptyHost verifies that STDIO transport accepts empty host
func TestValidateTransportStdioWithEmptyHost(t *testing.T) {
	cfg := NewConfig()
	cfg.TransportType = "stdio"
	cfg.Host = ""
	cfg.Port = 0

	err := cfg.ValidateTransport()
	if err != nil {
		t.Errorf("Expected ValidateTransport to accept stdio transport with empty host, got error: %v", err)
	}
}
