//go:build property
// +build property

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty1ConfigurationPrecedence tests Property 1: Configuration Precedence
// **Validates: Requirements 1.2, 1.3, 1.4, 5.4, 5.5, 5.6, 5.7, 5.8, 5.9**
//
// Feature: multi-transport-support, Property 1: Configuration Precedence
//
// Property: For any configuration setting (transport type, port, host), when the same setting
// is provided through multiple sources (CLI flag, config file, environment variable), the value
// from the highest precedence source (CLI > config file > env var > default) should be used
// in the final configuration.
func TestProperty1ConfigurationPrecedence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for valid transport types
	genTransportType := gen.OneConstOf("stdio", "sse", "streamablehttp")

	// Generator for valid hosts
	genHost := gen.OneConstOf("localhost", "127.0.0.1", "0.0.0.0", "example.com")

	// Generator for valid ports (1-65535)
	genPort := gen.IntRange(1, 65535)

	properties.Property("CLI flags take precedence over config file and environment variables for transport settings", prop.ForAll(
		func(flagTransport, fileTransport, envTransport string, flagHost, fileHost, envHost string, flagPort, filePort, envPort int) bool {
			// Clean up environment
			os.Unsetenv("TRANSPORT_TYPE")
			os.Unsetenv("HOST")
			os.Unsetenv("PORT")
			defer func() {
				os.Unsetenv("TRANSPORT_TYPE")
				os.Unsetenv("HOST")
				os.Unsetenv("PORT")
			}()

			// Set environment variables (lowest precedence)
			os.Setenv("TRANSPORT_TYPE", envTransport)
			os.Setenv("HOST", envHost)
			os.Setenv("PORT", fmt.Sprintf("%d", envPort))

			// Create a config file (middle precedence)
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")
			configContent := fmt.Sprintf(`
log_level: info
docs_base_url: https://docs.nats.io
transport_type: %s
host: %s
port: %d
`, fileTransport, fileHost, filePort)

			if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
				t.Logf("Failed to create config file: %v", err)
				return false
			}

			// Provide flags (highest precedence)
			flags := map[string]interface{}{
				"transport_type": flagTransport,
				"host":           flagHost,
				"port":           flagPort,
			}

			cfg, err := LoadWithFlags(configFile, flags)
			if err != nil {
				t.Logf("Failed to load config with flags: %v", err)
				return false
			}

			// Verify flags take precedence over config file and env vars
			if cfg.TransportType != flagTransport {
				t.Logf("Expected TransportType from flags %s, got %s", flagTransport, cfg.TransportType)
				return false
			}
			if cfg.Host != flagHost {
				t.Logf("Expected Host from flags %s, got %s", flagHost, cfg.Host)
				return false
			}
			if cfg.Port != flagPort {
				t.Logf("Expected Port from flags %d, got %d", flagPort, cfg.Port)
				return false
			}

			return true
		},
		genTransportType,
		genTransportType,
		genTransportType,
		genHost,
		genHost,
		genHost,
		genPort,
		genPort,
		genPort,
	))

	properties.Property("config file takes precedence over environment variables for transport settings", prop.ForAll(
		func(fileTransport, envTransport string, fileHost, envHost string, filePort, envPort int) bool {
			// Clean up environment
			os.Unsetenv("TRANSPORT_TYPE")
			os.Unsetenv("HOST")
			os.Unsetenv("PORT")
			defer func() {
				os.Unsetenv("TRANSPORT_TYPE")
				os.Unsetenv("HOST")
				os.Unsetenv("PORT")
			}()

			// Set environment variables with different values
			os.Setenv("TRANSPORT_TYPE", envTransport)
			os.Setenv("HOST", envHost)
			os.Setenv("PORT", fmt.Sprintf("%d", envPort))

			// Create a config file with different values
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")
			configContent := fmt.Sprintf(`
log_level: info
docs_base_url: https://docs.nats.io
transport_type: %s
host: %s
port: %d
`, fileTransport, fileHost, filePort)

			if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
				t.Logf("Failed to create config file: %v", err)
				return false
			}

			// Load configuration from file
			cfg, err := LoadFromFile(configFile)
			if err != nil {
				t.Logf("Failed to load config: %v", err)
				return false
			}

			// Verify config file values take precedence over environment variables
			if cfg.TransportType != fileTransport {
				t.Logf("Expected TransportType from file %s, got %s", fileTransport, cfg.TransportType)
				return false
			}
			if cfg.Host != fileHost {
				t.Logf("Expected Host from file %s, got %s", fileHost, cfg.Host)
				return false
			}
			if cfg.Port != filePort {
				t.Logf("Expected Port from file %d, got %d", filePort, cfg.Port)
				return false
			}

			return true
		},
		genTransportType,
		genTransportType,
		genHost,
		genHost,
		genPort,
		genPort,
	))

	properties.Property("environment variables take precedence over defaults for transport settings", prop.ForAll(
		func(envTransport, envHost string, envPort int) bool {
			// Clean up environment
			os.Unsetenv("TRANSPORT_TYPE")
			os.Unsetenv("HOST")
			os.Unsetenv("PORT")
			defer func() {
				os.Unsetenv("TRANSPORT_TYPE")
				os.Unsetenv("HOST")
				os.Unsetenv("PORT")
			}()

			// Set environment variables
			os.Setenv("TRANSPORT_TYPE", envTransport)
			os.Setenv("HOST", envHost)
			os.Setenv("PORT", fmt.Sprintf("%d", envPort))

			cfg, err := Load()
			if err != nil {
				t.Logf("Failed to load config from env: %v", err)
				return false
			}

			// Verify environment variables override defaults
			if cfg.TransportType != envTransport {
				t.Logf("Expected TransportType from env %s, got %s", envTransport, cfg.TransportType)
				return false
			}
			if cfg.Host != envHost {
				t.Logf("Expected Host from env %s, got %s", envHost, cfg.Host)
				return false
			}
			if cfg.Port != envPort {
				t.Logf("Expected Port from env %d, got %d", envPort, cfg.Port)
				return false
			}

			return true
		},
		genTransportType,
		genHost,
		genPort,
	))

	properties.Property("defaults are used when no transport configuration is provided", prop.ForAll(
		func(seed int) bool {
			// Clean up environment to ensure no interference
			os.Unsetenv("TRANSPORT_TYPE")
			os.Unsetenv("HOST")
			os.Unsetenv("PORT")
			defer func() {
				os.Unsetenv("TRANSPORT_TYPE")
				os.Unsetenv("HOST")
				os.Unsetenv("PORT")
			}()

			cfg, err := Load()
			if err != nil {
				t.Logf("Failed to load config with defaults: %v", err)
				return false
			}

			// Verify defaults are used
			expectedDefaults := NewConfig()
			if cfg.TransportType != expectedDefaults.TransportType {
				t.Logf("Expected default TransportType %s, got %s", expectedDefaults.TransportType, cfg.TransportType)
				return false
			}
			if cfg.Host != expectedDefaults.Host {
				t.Logf("Expected default Host %s, got %s", expectedDefaults.Host, cfg.Host)
				return false
			}
			if cfg.Port != expectedDefaults.Port {
				t.Logf("Expected default Port %d, got %d", expectedDefaults.Port, cfg.Port)
				return false
			}

			return true
		},
		gen.Int(),
	))

	properties.Property("partial transport configuration uses defaults for omitted values", prop.ForAll(
		func(setTransport, setHost, setPort bool, transport, host string, port int) bool {
			// Ensure at least one value is omitted to test defaults
			if setTransport && setHost && setPort {
				setTransport = false
			}

			// Clean up environment
			os.Unsetenv("TRANSPORT_TYPE")
			os.Unsetenv("HOST")
			os.Unsetenv("PORT")
			defer func() {
				os.Unsetenv("TRANSPORT_TYPE")
				os.Unsetenv("HOST")
				os.Unsetenv("PORT")
			}()

			// Create a config file with only some transport parameters
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")
			var configLines []string
			configLines = append(configLines, "log_level: info")
			configLines = append(configLines, "docs_base_url: https://docs.nats.io")

			if setTransport {
				configLines = append(configLines, fmt.Sprintf("transport_type: %s", transport))
			}
			if setHost {
				configLines = append(configLines, fmt.Sprintf("host: %s", host))
			}
			if setPort {
				configLines = append(configLines, fmt.Sprintf("port: %d", port))
			}

			configContent := ""
			for _, line := range configLines {
				configContent += line + "\n"
			}

			if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
				t.Logf("Failed to create config file: %v", err)
				return false
			}

			cfg, err := LoadFromFile(configFile)
			if err != nil {
				t.Logf("Failed to load config from file: %v", err)
				return false
			}

			// Verify omitted parameters use defaults
			expectedDefaults := NewConfig()
			if !setTransport && cfg.TransportType != expectedDefaults.TransportType {
				t.Logf("Expected default TransportType %s when omitted, got %s", expectedDefaults.TransportType, cfg.TransportType)
				return false
			}
			if !setHost && cfg.Host != expectedDefaults.Host {
				t.Logf("Expected default Host %s when omitted, got %s", expectedDefaults.Host, cfg.Host)
				return false
			}
			if !setPort && cfg.Port != expectedDefaults.Port {
				t.Logf("Expected default Port %d when omitted, got %d", expectedDefaults.Port, cfg.Port)
				return false
			}

			// Verify provided parameters are used (not defaults)
			if setTransport && cfg.TransportType != transport {
				t.Logf("Expected provided TransportType %s, got %s", transport, cfg.TransportType)
				return false
			}
			if setHost && cfg.Host != host {
				t.Logf("Expected provided Host %s, got %s", host, cfg.Host)
				return false
			}
			if setPort && cfg.Port != port {
				t.Logf("Expected provided Port %d, got %d", port, cfg.Port)
				return false
			}

			return true
		},
		gen.Bool(),
		gen.Bool(),
		gen.Bool(),
		genTransportType,
		genHost,
		genPort,
	))

	properties.Property("empty config file uses all transport defaults", prop.ForAll(
		func(seed int) bool {
			// Clean up environment variables
			os.Unsetenv("TRANSPORT_TYPE")
			os.Unsetenv("HOST")
			os.Unsetenv("PORT")
			defer func() {
				os.Unsetenv("TRANSPORT_TYPE")
				os.Unsetenv("HOST")
				os.Unsetenv("PORT")
			}()

			// Create an empty config file
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")
			if err := os.WriteFile(configFile, []byte(""), 0644); err != nil {
				t.Logf("Failed to create empty config file: %v", err)
				return false
			}

			cfg, err := LoadFromFile(configFile)
			if err != nil {
				t.Logf("Failed to load empty config file: %v", err)
				return false
			}

			// All transport values should be defaults
			expectedDefaults := NewConfig()
			if cfg.TransportType != expectedDefaults.TransportType {
				t.Logf("Expected default TransportType %s for empty config, got %s", expectedDefaults.TransportType, cfg.TransportType)
				return false
			}
			if cfg.Host != expectedDefaults.Host {
				t.Logf("Expected default Host %s for empty config, got %s", expectedDefaults.Host, cfg.Host)
				return false
			}
			if cfg.Port != expectedDefaults.Port {
				t.Logf("Expected default Port %d for empty config, got %d", expectedDefaults.Port, cfg.Port)
				return false
			}

			return true
		},
		gen.Int(),
	))

	properties.Property("empty flags map uses transport defaults from config file or env vars", prop.ForAll(
		func(fileTransport, fileHost string, filePort int) bool {
			// Clean up environment variables
			os.Unsetenv("TRANSPORT_TYPE")
			os.Unsetenv("HOST")
			os.Unsetenv("PORT")
			defer func() {
				os.Unsetenv("TRANSPORT_TYPE")
				os.Unsetenv("HOST")
				os.Unsetenv("PORT")
			}()

			// Create a config file with transport settings
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")
			configContent := fmt.Sprintf(`
log_level: info
docs_base_url: https://docs.nats.io
transport_type: %s
host: %s
port: %d
`, fileTransport, fileHost, filePort)

			if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
				t.Logf("Failed to create config file: %v", err)
				return false
			}

			// Empty flags map
			flags := make(map[string]interface{})

			cfg, err := LoadWithFlags(configFile, flags)
			if err != nil {
				t.Logf("Failed to load config with empty flags: %v", err)
				return false
			}

			// Transport values should come from config file
			if cfg.TransportType != fileTransport {
				t.Logf("Expected TransportType from file %s for empty flags, got %s", fileTransport, cfg.TransportType)
				return false
			}
			if cfg.Host != fileHost {
				t.Logf("Expected Host from file %s for empty flags, got %s", fileHost, cfg.Host)
				return false
			}
			if cfg.Port != filePort {
				t.Logf("Expected Port from file %d for empty flags, got %d", filePort, cfg.Port)
				return false
			}

			return true
		},
		genTransportType,
		genHost,
		genPort,
	))

	properties.TestingRun(t)
}

// TestProperty4HostValidation tests Property 4: Host Validation
// **Validates: Requirements 6.4**
//
// Feature: multi-transport-support, Property 4: Host Validation
//
// Property: For any configuration where host is explicitly set, validation should fail
// if the host value is an empty string.
func TestProperty4HostValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for network transport types
	genNetworkTransportType := gen.OneConstOf("sse", "streamablehttp")

	// Generator for valid ports (1-65535)
	genValidPort := gen.IntRange(1, 65535)

	// Generator for non-empty host strings
	genNonEmptyHost := gen.AnyString().SuchThat(func(s string) bool {
		return s != ""
	})

	properties.Property("network transports reject empty host string", prop.ForAll(
		func(transportType string, port int) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = port
			cfg.Host = "" // Explicitly set to empty string

			err := cfg.ValidateTransport()
			if err == nil {
				t.Logf("Expected network transport '%s' with empty host to be rejected", transportType)
				return false
			}

			// Verify error message mentions host
			errMsg := err.Error()
			if !strings.Contains(errMsg, "host") {
				t.Logf("Expected error message to mention 'host', got: %v", err)
				return false
			}

			// Verify error message mentions "empty"
			if !strings.Contains(errMsg, "empty") {
				t.Logf("Expected error message to mention 'empty', got: %v", err)
				return false
			}

			return true
		},
		genNetworkTransportType,
		genValidPort,
	))

	properties.Property("network transports accept non-empty host strings", prop.ForAll(
		func(transportType string, port int, host string) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = port
			cfg.Host = host

			err := cfg.ValidateTransport()
			if err != nil {
				t.Logf("Expected network transport '%s' with non-empty host '%s' to be accepted, got error: %v",
					transportType, host, err)
				return false
			}

			return true
		},
		genNetworkTransportType,
		genValidPort,
		genNonEmptyHost,
	))

	properties.Property("stdio transport accepts empty host string", prop.ForAll(
		func(port int) bool {
			cfg := NewConfig()
			cfg.TransportType = "stdio"
			cfg.Port = port
			cfg.Host = "" // Explicitly set to empty string

			err := cfg.ValidateTransport()
			if err != nil {
				t.Logf("Expected stdio transport with empty host to be accepted, got error: %v", err)
				return false
			}

			return true
		},
		gen.IntRange(-1000, 100000), // Any port value (stdio ignores port)
	))

	properties.Property("stdio transport accepts any host string", prop.ForAll(
		func(port int, host string) bool {
			cfg := NewConfig()
			cfg.TransportType = "stdio"
			cfg.Port = port
			cfg.Host = host

			err := cfg.ValidateTransport()
			if err != nil {
				t.Logf("Expected stdio transport with any host '%s' to be accepted, got error: %v", host, err)
				return false
			}

			return true
		},
		gen.IntRange(-1000, 100000), // Any port value
		gen.AnyString(),             // Any host string including empty
	))

	properties.Property("host validation is independent of port validation for network transports", prop.ForAll(
		func(transportType string, port int) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = port
			cfg.Host = "" // Empty host

			err := cfg.ValidateTransport()

			// Should always fail due to empty host, regardless of port value
			if err == nil {
				t.Logf("Expected network transport '%s' with empty host to be rejected regardless of port %d",
					transportType, port)
				return false
			}

			// Error message should mention host
			errMsg := err.Error()
			if !strings.Contains(errMsg, "host") {
				t.Logf("Expected error message to mention 'host', got: %v", err)
				return false
			}

			return true
		},
		genNetworkTransportType,
		gen.IntRange(-1000, 100000), // Wide range including invalid ports
	))

	properties.Property("whitespace-only host strings are accepted (not considered empty)", prop.ForAll(
		func(transportType string, port int, whitespaceCount int) bool {
			// Generate whitespace-only strings (spaces, tabs, newlines)
			whitespaceChars := []string{" ", "\t", "\n"}
			host := ""
			for i := 0; i < (whitespaceCount%10)+1; i++ { // 1-10 whitespace characters
				host += whitespaceChars[i%len(whitespaceChars)]
			}

			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = port
			cfg.Host = host

			err := cfg.ValidateTransport()
			// Whitespace-only strings are not empty strings, so they should be accepted
			if err != nil {
				t.Logf("Expected network transport '%s' with whitespace-only host to be accepted, got error: %v",
					transportType, err)
				return false
			}

			return true
		},
		genNetworkTransportType,
		genValidPort,
		gen.IntRange(0, 100),
	))

	properties.Property("host validation error is descriptive", prop.ForAll(
		func(transportType string, port int) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = port
			cfg.Host = ""

			err := cfg.ValidateTransport()
			if err == nil {
				t.Logf("Expected network transport '%s' with empty host to be rejected", transportType)
				return false
			}

			// Error message should be descriptive and mention both "host" and "empty"
			errMsg := err.Error()
			if !strings.Contains(errMsg, "host") || !strings.Contains(errMsg, "empty") {
				t.Logf("Expected error message to mention 'host' and 'empty', got: %v", err)
				return false
			}

			return true
		},
		genNetworkTransportType,
		genValidPort,
	))

	properties.Property("multiple validation errors include host error when host is empty", prop.ForAll(
		func(transportType string, invalidPort int) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = invalidPort // Invalid port
			cfg.Host = ""          // Empty host

			err := cfg.ValidateTransport()
			if err == nil {
				t.Logf("Expected network transport '%s' with invalid port %d and empty host to be rejected",
					transportType, invalidPort)
				return false
			}

			// Error message should mention both port and host issues
			errMsg := err.Error()
			if !strings.Contains(errMsg, "host") {
				t.Logf("Expected error message to mention 'host' when both port and host are invalid, got: %v", err)
				return false
			}
			if !strings.Contains(errMsg, "port") {
				t.Logf("Expected error message to mention 'port' when both port and host are invalid, got: %v", err)
				return false
			}

			return true
		},
		genNetworkTransportType,
		gen.OneGenOf(
			gen.IntRange(-1000, 0),      // Invalid ports (< 1)
			gen.IntRange(65536, 100000), // Invalid ports (> 65535)
		),
	))

	properties.TestingRun(t)
}

// TestProperty3NetworkTransportPortRequirement tests Property 3: Network Transport Port Requirement
// **Validates: Requirements 3.2, 4.2, 5.1, 6.1, 6.2, 6.3**
//
// Feature: multi-transport-support, Property 3: Network Transport Port Requirement
//
// Property: For any configuration where transport type is "sse" or "streamablehttp",
// validation should fail if port is not configured, or if port is less than 1,
// or if port is greater than 65535.
func TestProperty3NetworkTransportPortRequirement(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for network transport types
	genNetworkTransportType := gen.OneConstOf("sse", "streamablehttp")

	// Generator for valid ports (1-65535)
	genValidPort := gen.IntRange(1, 65535)

	// Generator for invalid ports (outside valid range)
	genInvalidPort := gen.OneGenOf(
		gen.IntRange(-1000, 0),      // Negative and zero ports
		gen.IntRange(65536, 100000), // Ports above valid range
	)

	// Generator for valid hosts
	genValidHost := gen.OneConstOf("localhost", "127.0.0.1", "0.0.0.0", "example.com")

	properties.Property("network transports require port in valid range (1-65535)", prop.ForAll(
		func(transportType string, port int, host string) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = port
			cfg.Host = host

			err := cfg.ValidateTransport()
			if err != nil {
				t.Logf("Expected network transport '%s' with valid port %d to be accepted, got error: %v",
					transportType, port, err)
				return false
			}

			return true
		},
		genNetworkTransportType,
		genValidPort,
		genValidHost,
	))

	properties.Property("network transports reject port less than 1", prop.ForAll(
		func(transportType string, port int, host string) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = port
			cfg.Host = host

			err := cfg.ValidateTransport()
			if err == nil {
				t.Logf("Expected network transport '%s' with port %d (< 1) to be rejected",
					transportType, port)
				return false
			}

			// Verify error message mentions port and valid range
			errMsg := err.Error()
			if !strings.Contains(errMsg, "port") {
				t.Logf("Expected error message to mention 'port', got: %v", err)
				return false
			}
			if !strings.Contains(errMsg, "1") || !strings.Contains(errMsg, "65535") {
				t.Logf("Expected error message to mention valid port range (1-65535), got: %v", err)
				return false
			}

			return true
		},
		genNetworkTransportType,
		gen.IntRange(-1000, 0), // Ports less than 1
		genValidHost,
	))

	properties.Property("network transports reject port greater than 65535", prop.ForAll(
		func(transportType string, port int, host string) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = port
			cfg.Host = host

			err := cfg.ValidateTransport()
			if err == nil {
				t.Logf("Expected network transport '%s' with port %d (> 65535) to be rejected",
					transportType, port)
				return false
			}

			// Verify error message mentions port and valid range
			errMsg := err.Error()
			if !strings.Contains(errMsg, "port") {
				t.Logf("Expected error message to mention 'port', got: %v", err)
				return false
			}
			if !strings.Contains(errMsg, "65535") {
				t.Logf("Expected error message to mention valid port range upper bound (65535), got: %v", err)
				return false
			}

			return true
		},
		genNetworkTransportType,
		gen.IntRange(65536, 100000), // Ports greater than 65535
		genValidHost,
	))

	properties.Property("network transports reject any invalid port value", prop.ForAll(
		func(transportType string, port int, host string) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = port
			cfg.Host = host

			err := cfg.ValidateTransport()
			if err == nil {
				t.Logf("Expected network transport '%s' with invalid port %d to be rejected",
					transportType, port)
				return false
			}

			// Verify error message is descriptive
			errMsg := err.Error()
			if !strings.Contains(errMsg, "port") {
				t.Logf("Expected error message to mention 'port', got: %v", err)
				return false
			}

			return true
		},
		genNetworkTransportType,
		genInvalidPort,
		genValidHost,
	))

	properties.Property("network transports accept all valid port values", prop.ForAll(
		func(transportType string, port int, host string) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = port
			cfg.Host = host

			err := cfg.ValidateTransport()
			if err != nil {
				t.Logf("Expected network transport '%s' with valid port %d to be accepted, got error: %v",
					transportType, port, err)
				return false
			}

			return true
		},
		genNetworkTransportType,
		genValidPort,
		genValidHost,
	))

	properties.Property("network transports require port regardless of host value", prop.ForAll(
		func(transportType string, port int, host string) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = port
			cfg.Host = host

			err := cfg.ValidateTransport()

			// If port is invalid, validation should fail regardless of host
			if port < 1 || port > 65535 {
				if err == nil {
					t.Logf("Expected network transport '%s' with invalid port %d to be rejected regardless of host '%s'",
						transportType, port, host)
					return false
				}
				// Error should mention port
				if !strings.Contains(err.Error(), "port") {
					t.Logf("Expected error message to mention 'port', got: %v", err)
					return false
				}
			} else {
				// If port is valid but host is empty, validation should fail for host reason
				if host == "" {
					if err == nil {
						t.Logf("Expected network transport '%s' with empty host to be rejected", transportType)
						return false
					}
					// Error should mention host
					if !strings.Contains(err.Error(), "host") {
						t.Logf("Expected error message to mention 'host', got: %v", err)
						return false
					}
				} else {
					// Both port and host are valid, should succeed
					if err != nil {
						t.Logf("Expected network transport '%s' with valid port %d and host '%s' to be accepted, got error: %v",
							transportType, port, host, err)
						return false
					}
				}
			}

			return true
		},
		genNetworkTransportType,
		gen.IntRange(-1000, 100000), // Wide range of ports
		gen.AnyString(),             // Any host string including empty
	))

	properties.Property("stdio transport accepts any port value", prop.ForAll(
		func(port int) bool {
			cfg := NewConfig()
			cfg.TransportType = "stdio"
			cfg.Port = port
			cfg.Host = "localhost" // Valid host to isolate port validation

			err := cfg.ValidateTransport()
			if err != nil {
				t.Logf("Expected stdio transport to accept any port value %d, got error: %v", port, err)
				return false
			}

			return true
		},
		gen.IntRange(-1000, 100000), // Wide range including invalid ports
	))

	properties.Property("port validation error messages include actual port value", prop.ForAll(
		func(transportType string, port int) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = port
			cfg.Host = "localhost"

			err := cfg.ValidateTransport()
			if err == nil {
				t.Logf("Expected network transport '%s' with invalid port %d to be rejected", transportType, port)
				return false
			}

			// Error message should include the actual port value
			errMsg := err.Error()
			portStr := fmt.Sprintf("%d", port)
			if !strings.Contains(errMsg, portStr) {
				t.Logf("Expected error message to include actual port value %d, got: %v", port, err)
				return false
			}

			return true
		},
		genNetworkTransportType,
		genInvalidPort,
	))

	properties.Property("boundary port values are handled correctly", prop.ForAll(
		func(transportType string, seed int) bool {
			// Test boundary values: 0, 1, 65535, 65536
			boundaryPorts := []struct {
				port  int
				valid bool
			}{
				{0, false},     // Just below valid range
				{1, true},      // Lower boundary (valid)
				{65535, true},  // Upper boundary (valid)
				{65536, false}, // Just above valid range
			}

			for _, bp := range boundaryPorts {
				cfg := NewConfig()
				cfg.TransportType = transportType
				cfg.Port = bp.port
				cfg.Host = "localhost"

				err := cfg.ValidateTransport()

				if bp.valid && err != nil {
					t.Logf("Expected network transport '%s' with boundary port %d to be accepted, got error: %v",
						transportType, bp.port, err)
					return false
				}

				if !bp.valid && err == nil {
					t.Logf("Expected network transport '%s' with boundary port %d to be rejected",
						transportType, bp.port)
					return false
				}
			}

			return true
		},
		genNetworkTransportType,
		gen.Int(), // Seed for deterministic test
	))

	properties.TestingRun(t)
}

// TestProperty2TransportTypeValidation tests Property 2: Transport Type Validation
// **Validates: Requirements 1.6, 1.7**
//
// Feature: multi-transport-support, Property 2: Transport Type Validation
//
// Property: For any transport type value, the configuration validation should accept
// exactly the three valid values ("stdio", "sse", "streamablehttp") and reject all
// other values with a descriptive error message listing the valid options.
func TestProperty2TransportTypeValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for valid transport types
	genValidTransportType := gen.OneConstOf("stdio", "sse", "streamablehttp")

	// Generator for invalid transport types (arbitrary strings that are not valid)
	genInvalidTransportType := gen.AnyString().SuchThat(func(s string) bool {
		return s != "stdio" && s != "sse" && s != "streamablehttp"
	})

	properties.Property("valid transport types are accepted by validation", prop.ForAll(
		func(transportType string) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType

			// For network transports, set valid port to avoid port validation errors
			if transportType == "sse" || transportType == "streamablehttp" {
				cfg.Port = 8080
				cfg.Host = "localhost"
			}

			err := cfg.ValidateTransport()
			if err != nil {
				t.Logf("Expected valid transport type '%s' to be accepted, got error: %v", transportType, err)
				return false
			}

			return true
		},
		genValidTransportType,
	))

	properties.Property("invalid transport types are rejected with descriptive error", prop.ForAll(
		func(transportType string) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType

			err := cfg.ValidateTransport()
			if err == nil {
				t.Logf("Expected invalid transport type '%s' to be rejected", transportType)
				return false
			}

			// Verify error message is descriptive and lists valid options
			errMsg := err.Error()
			if !strings.Contains(errMsg, "stdio") {
				t.Logf("Expected error message to list 'stdio' as valid option, got: %v", err)
				return false
			}
			if !strings.Contains(errMsg, "sse") {
				t.Logf("Expected error message to list 'sse' as valid option, got: %v", err)
				return false
			}
			if !strings.Contains(errMsg, "streamablehttp") {
				t.Logf("Expected error message to list 'streamablehttp' as valid option, got: %v", err)
				return false
			}

			return true
		},
		genInvalidTransportType,
	))

	properties.Property("validation accepts valid transport types regardless of other config values", prop.ForAll(
		func(transportType string, port int, host string) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType

			// For network transports, ensure port is valid to isolate transport type validation
			if transportType == "sse" || transportType == "streamablehttp" {
				// Use a valid port
				cfg.Port = 8080
				cfg.Host = "localhost"
			} else {
				// For stdio, any port/host values should be fine
				cfg.Port = port
				cfg.Host = host
			}

			err := cfg.ValidateTransport()
			if err != nil {
				t.Logf("Expected valid transport type '%s' to be accepted with port=%d, host='%s', got error: %v",
					transportType, cfg.Port, cfg.Host, err)
				return false
			}

			return true
		},
		genValidTransportType,
		gen.IntRange(-1000, 100000), // Wide range including invalid ports
		gen.AnyString(),             // Any host string
	))

	properties.Property("validation rejects invalid transport types regardless of other config values", prop.ForAll(
		func(transportType string, port int, host string) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = port
			cfg.Host = host

			err := cfg.ValidateTransport()
			if err == nil {
				t.Logf("Expected invalid transport type '%s' to be rejected", transportType)
				return false
			}

			// Error message should mention the invalid transport type
			errMsg := err.Error()
			if !strings.Contains(errMsg, "transport") {
				t.Logf("Expected error message to mention 'transport', got: %v", err)
				return false
			}

			return true
		},
		genInvalidTransportType,
		gen.IntRange(-1000, 100000),
		gen.AnyString(),
	))

	properties.Property("empty string transport type is rejected", prop.ForAll(
		func(seed int) bool {
			cfg := NewConfig()
			cfg.TransportType = ""

			err := cfg.ValidateTransport()
			if err == nil {
				t.Logf("Expected empty transport type to be rejected")
				return false
			}

			// Verify error message lists valid options
			errMsg := err.Error()
			if !strings.Contains(errMsg, "stdio") || !strings.Contains(errMsg, "sse") || !strings.Contains(errMsg, "streamablehttp") {
				t.Logf("Expected error message to list valid transport types, got: %v", err)
				return false
			}

			return true
		},
		gen.Int(),
	))

	properties.Property("transport type validation is case-sensitive", prop.ForAll(
		func(seed int) bool {
			// Test uppercase versions of valid transport types
			invalidCases := []string{"STDIO", "SSE", "STREAMABLEHTTP", "Stdio", "Sse", "StreamableHTTP"}

			for _, transportType := range invalidCases {
				cfg := NewConfig()
				cfg.TransportType = transportType

				err := cfg.ValidateTransport()
				if err == nil {
					t.Logf("Expected case-variant transport type '%s' to be rejected (validation is case-sensitive)", transportType)
					return false
				}
			}

			return true
		},
		gen.Int(),
	))

	properties.Property("whitespace-padded transport types are rejected", prop.ForAll(
		func(transportType string) bool {
			// Add whitespace padding to valid transport types
			paddedType := " " + transportType + " "

			cfg := NewConfig()
			cfg.TransportType = paddedType

			err := cfg.ValidateTransport()
			if err == nil {
				t.Logf("Expected whitespace-padded transport type '%s' to be rejected", paddedType)
				return false
			}

			return true
		},
		genValidTransportType,
	))

	properties.TestingRun(t)
}

// TestProperty5StdioTransportConfigurationIndependence tests Property 5: STDIO Transport Configuration Independence
// **Validates: Requirements 2.2, 5.10**
//
// Feature: multi-transport-support, Property 5: STDIO Transport Configuration Independence
//
// Property: For any configuration where transport type is "stdio", validation should succeed
// regardless of whether port or host are configured (these settings should be ignored for
// STDIO transport).
func TestProperty5StdioTransportConfigurationIndependence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for any port value (including invalid ones for network transports)
	genAnyPort := gen.IntRange(-10000, 100000)

	// Generator for any host string (including empty strings)
	genAnyHost := gen.AnyString()

	properties.Property("stdio transport validates successfully with any port value", prop.ForAll(
		func(port int) bool {
			cfg := NewConfig()
			cfg.TransportType = "stdio"
			cfg.Port = port
			cfg.Host = "localhost" // Valid host to isolate port independence

			err := cfg.ValidateTransport()
			if err != nil {
				t.Logf("Expected stdio transport to accept any port value %d, got error: %v", port, err)
				return false
			}

			return true
		},
		genAnyPort,
	))

	properties.Property("stdio transport validates successfully with any host value", prop.ForAll(
		func(host string) bool {
			cfg := NewConfig()
			cfg.TransportType = "stdio"
			cfg.Port = 8080 // Valid port to isolate host independence
			cfg.Host = host

			err := cfg.ValidateTransport()
			if err != nil {
				t.Logf("Expected stdio transport to accept any host value '%s', got error: %v", host, err)
				return false
			}

			return true
		},
		genAnyHost,
	))

	properties.Property("stdio transport validates successfully with empty host", prop.ForAll(
		func(port int) bool {
			cfg := NewConfig()
			cfg.TransportType = "stdio"
			cfg.Port = port
			cfg.Host = "" // Empty host (would be invalid for network transports)

			err := cfg.ValidateTransport()
			if err != nil {
				t.Logf("Expected stdio transport to accept empty host with port %d, got error: %v", port, err)
				return false
			}

			return true
		},
		genAnyPort,
	))

	properties.Property("stdio transport validates successfully with invalid port for network transports", prop.ForAll(
		func(port int, host string) bool {
			// Only test ports that would be invalid for network transports
			if port >= 1 && port <= 65535 {
				return true // Skip valid ports
			}

			cfg := NewConfig()
			cfg.TransportType = "stdio"
			cfg.Port = port
			cfg.Host = host

			err := cfg.ValidateTransport()
			if err != nil {
				t.Logf("Expected stdio transport to accept invalid network port %d with host '%s', got error: %v",
					port, host, err)
				return false
			}

			return true
		},
		genAnyPort,
		genAnyHost,
	))

	properties.Property("stdio transport validates successfully with zero port", prop.ForAll(
		func(host string) bool {
			cfg := NewConfig()
			cfg.TransportType = "stdio"
			cfg.Port = 0 // Zero port (would be invalid for network transports)
			cfg.Host = host

			err := cfg.ValidateTransport()
			if err != nil {
				t.Logf("Expected stdio transport to accept zero port with host '%s', got error: %v", host, err)
				return false
			}

			return true
		},
		genAnyHost,
	))

	properties.Property("stdio transport validates successfully with negative port", prop.ForAll(
		func(port int, host string) bool {
			// Only test negative ports
			if port >= 0 {
				return true // Skip non-negative ports
			}

			cfg := NewConfig()
			cfg.TransportType = "stdio"
			cfg.Port = port
			cfg.Host = host

			err := cfg.ValidateTransport()
			if err != nil {
				t.Logf("Expected stdio transport to accept negative port %d with host '%s', got error: %v",
					port, host, err)
				return false
			}

			return true
		},
		genAnyPort,
		genAnyHost,
	))

	properties.Property("stdio transport validates successfully with port above 65535", prop.ForAll(
		func(port int, host string) bool {
			// Only test ports above valid range
			if port <= 65535 {
				return true // Skip valid range ports
			}

			cfg := NewConfig()
			cfg.TransportType = "stdio"
			cfg.Port = port
			cfg.Host = host

			err := cfg.ValidateTransport()
			if err != nil {
				t.Logf("Expected stdio transport to accept port %d (> 65535) with host '%s', got error: %v",
					port, host, err)
				return false
			}

			return true
		},
		genAnyPort,
		genAnyHost,
	))

	properties.Property("stdio transport validates successfully with any combination of port and host", prop.ForAll(
		func(port int, host string) bool {
			cfg := NewConfig()
			cfg.TransportType = "stdio"
			cfg.Port = port
			cfg.Host = host

			err := cfg.ValidateTransport()
			if err != nil {
				t.Logf("Expected stdio transport to accept any port %d and host '%s', got error: %v",
					port, host, err)
				return false
			}

			return true
		},
		genAnyPort,
		genAnyHost,
	))

	properties.Property("stdio transport ignores port configuration that would fail for network transports", prop.ForAll(
		func(seed int) bool {
			// Test specific invalid port values for network transports
			invalidPorts := []int{-1, 0, -100, 65536, 70000, 100000}

			for _, port := range invalidPorts {
				cfg := NewConfig()
				cfg.TransportType = "stdio"
				cfg.Port = port
				cfg.Host = "localhost"

				err := cfg.ValidateTransport()
				if err != nil {
					t.Logf("Expected stdio transport to ignore invalid network port %d, got error: %v", port, err)
					return false
				}
			}

			return true
		},
		gen.Int(),
	))

	properties.Property("stdio transport ignores host configuration that would fail for network transports", prop.ForAll(
		func(seed int) bool {
			cfg := NewConfig()
			cfg.TransportType = "stdio"
			cfg.Port = 8080
			cfg.Host = "" // Empty host (would fail for network transports)

			err := cfg.ValidateTransport()
			if err != nil {
				t.Logf("Expected stdio transport to ignore empty host, got error: %v", err)
				return false
			}

			return true
		},
		gen.Int(),
	))

	properties.Property("stdio transport validation is independent of network transport validation rules", prop.ForAll(
		func(port int, host string) bool {
			cfg := NewConfig()
			cfg.TransportType = "stdio"
			cfg.Port = port
			cfg.Host = host

			err := cfg.ValidateTransport()

			// STDIO should always validate successfully regardless of port/host values
			if err != nil {
				t.Logf("Expected stdio transport to validate successfully with port %d and host '%s', got error: %v",
					port, host, err)
				return false
			}

			// Now test that the same port/host combination would fail for network transports
			// if they are invalid for network transports
			if (port < 1 || port > 65535) || host == "" {
				// Test with SSE transport
				cfgSSE := NewConfig()
				cfgSSE.TransportType = "sse"
				cfgSSE.Port = port
				cfgSSE.Host = host

				errSSE := cfgSSE.ValidateTransport()
				// Should fail for network transport with invalid port or empty host
				if errSSE == nil && (port < 1 || port > 65535 || host == "") {
					t.Logf("Expected network transport to reject invalid port %d or empty host '%s', but it was accepted",
						port, host)
					return false
				}
			}

			return true
		},
		genAnyPort,
		genAnyHost,
	))

	properties.Property("stdio transport with default port and host values validates successfully", prop.ForAll(
		func(seed int) bool {
			cfg := NewConfig()
			cfg.TransportType = "stdio"
			// Use default port (0) and host ("localhost") from NewConfig()

			err := cfg.ValidateTransport()
			if err != nil {
				t.Logf("Expected stdio transport with default port and host to validate successfully, got error: %v", err)
				return false
			}

			return true
		},
		gen.Int(),
	))

	properties.Property("stdio transport validation does not check port range", prop.ForAll(
		func(port int) bool {
			cfg := NewConfig()
			cfg.TransportType = "stdio"
			cfg.Port = port
			cfg.Host = "localhost"

			err := cfg.ValidateTransport()

			// STDIO should never fail due to port range
			if err != nil {
				errMsg := err.Error()
				if strings.Contains(errMsg, "port") && strings.Contains(errMsg, "range") {
					t.Logf("Expected stdio transport to not check port range, but got port range error: %v", err)
					return false
				}
				// If error is not about port range, it's still unexpected for stdio
				t.Logf("Expected stdio transport to validate successfully, got unexpected error: %v", err)
				return false
			}

			return true
		},
		genAnyPort,
	))

	properties.Property("stdio transport validation does not check host emptiness", prop.ForAll(
		func(host string) bool {
			cfg := NewConfig()
			cfg.TransportType = "stdio"
			cfg.Port = 8080
			cfg.Host = host

			err := cfg.ValidateTransport()

			// STDIO should never fail due to empty host
			if err != nil {
				errMsg := err.Error()
				if strings.Contains(errMsg, "host") && strings.Contains(errMsg, "empty") {
					t.Logf("Expected stdio transport to not check host emptiness, but got host empty error: %v", err)
					return false
				}
				// If error is not about host emptiness, it's still unexpected for stdio
				t.Logf("Expected stdio transport to validate successfully, got unexpected error: %v", err)
				return false
			}

			return true
		},
		genAnyHost,
	))

	properties.TestingRun(t)
}

// TestProperty7ValidationErrorDescriptiveness tests Property 7: Validation Error Descriptiveness
// **Validates: Requirements 6.5**
//
// Feature: multi-transport-support, Property 7: Validation Error Descriptiveness
//
// Property: For any configuration that fails validation, the returned error message should
// contain specific information about which validation rule failed and what the valid values
// or constraints are.
func TestProperty7ValidationErrorDescriptiveness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for invalid transport types (any string that's not valid)
	genInvalidTransportType := gen.AnyString().SuchThat(func(s string) bool {
		return s != "stdio" && s != "sse" && s != "streamablehttp"
	})

	// Generator for invalid ports (outside valid range)
	genInvalidPort := gen.OneGenOf(
		gen.IntRange(-1000, 0),      // Negative and zero ports
		gen.IntRange(65536, 100000), // Ports above valid range
	)

	// Generator for network transport types
	genNetworkTransportType := gen.OneConstOf("sse", "streamablehttp")

	properties.Property("invalid transport type error includes the invalid value and lists valid options", prop.ForAll(
		func(invalidType string) bool {
			cfg := NewConfig()
			cfg.TransportType = invalidType

			err := cfg.ValidateTransport()
			if err == nil {
				t.Logf("Expected invalid transport type '%s' to be rejected", invalidType)
				return false
			}

			errMsg := err.Error()

			// Error should mention "transport"
			if !strings.Contains(errMsg, "transport") {
				t.Logf("Expected error message to mention 'transport', got: %v", err)
				return false
			}

			// Error should include the invalid value
			if !strings.Contains(errMsg, invalidType) {
				t.Logf("Expected error message to include invalid transport type '%s', got: %v", invalidType, err)
				return false
			}

			// Error should list all valid transport types
			if !strings.Contains(errMsg, "stdio") {
				t.Logf("Expected error message to list 'stdio' as valid option, got: %v", err)
				return false
			}
			if !strings.Contains(errMsg, "sse") {
				t.Logf("Expected error message to list 'sse' as valid option, got: %v", err)
				return false
			}
			if !strings.Contains(errMsg, "streamablehttp") {
				t.Logf("Expected error message to list 'streamablehttp' as valid option, got: %v", err)
				return false
			}

			return true
		},
		genInvalidTransportType,
	))

	properties.Property("invalid port error includes the invalid value and valid range constraints", prop.ForAll(
		func(transportType string, invalidPort int) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = invalidPort
			cfg.Host = "localhost"

			err := cfg.ValidateTransport()
			if err == nil {
				t.Logf("Expected network transport '%s' with invalid port %d to be rejected", transportType, invalidPort)
				return false
			}

			errMsg := err.Error()

			// Error should mention "port"
			if !strings.Contains(errMsg, "port") {
				t.Logf("Expected error message to mention 'port', got: %v", err)
				return false
			}

			// Error should include the invalid port value
			portStr := fmt.Sprintf("%d", invalidPort)
			if !strings.Contains(errMsg, portStr) {
				t.Logf("Expected error message to include invalid port value %d, got: %v", invalidPort, err)
				return false
			}

			// Error should mention the valid range (1-65535)
			if !strings.Contains(errMsg, "1") {
				t.Logf("Expected error message to mention lower bound '1', got: %v", err)
				return false
			}
			if !strings.Contains(errMsg, "65535") {
				t.Logf("Expected error message to mention upper bound '65535', got: %v", err)
				return false
			}

			return true
		},
		genNetworkTransportType,
		genInvalidPort,
	))

	properties.Property("empty host error mentions host and empty constraint", prop.ForAll(
		func(transportType string, port int) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = port
			cfg.Host = ""

			err := cfg.ValidateTransport()
			if err == nil {
				t.Logf("Expected network transport '%s' with empty host to be rejected", transportType)
				return false
			}

			errMsg := err.Error()

			// Error should mention "host"
			if !strings.Contains(errMsg, "host") {
				t.Logf("Expected error message to mention 'host', got: %v", err)
				return false
			}

			// Error should mention "empty"
			if !strings.Contains(errMsg, "empty") {
				t.Logf("Expected error message to mention 'empty', got: %v", err)
				return false
			}

			return true
		},
		genNetworkTransportType,
		gen.IntRange(-1000, 100000), // Any port value
	))

	properties.Property("multiple validation errors include all specific error details", prop.ForAll(
		func(invalidType string, invalidPort int) bool {
			cfg := NewConfig()
			cfg.TransportType = invalidType
			cfg.Port = invalidPort
			cfg.Host = ""

			err := cfg.ValidateTransport()
			if err == nil {
				t.Logf("Expected configuration with multiple invalid values to be rejected")
				return false
			}

			errMsg := err.Error()

			// Error should mention transport type issue
			if !strings.Contains(errMsg, "transport") {
				t.Logf("Expected error message to mention 'transport' when transport type is invalid, got: %v", err)
				return false
			}

			// Error should include the invalid transport type value
			if !strings.Contains(errMsg, invalidType) {
				t.Logf("Expected error message to include invalid transport type '%s', got: %v", invalidType, err)
				return false
			}

			// Error should list valid transport types
			if !strings.Contains(errMsg, "stdio") || !strings.Contains(errMsg, "sse") || !strings.Contains(errMsg, "streamablehttp") {
				t.Logf("Expected error message to list valid transport types, got: %v", err)
				return false
			}

			return true
		},
		genInvalidTransportType,
		genInvalidPort,
	))

	properties.Property("validation errors are descriptive for all error types", prop.ForAll(
		func(invalidType string, invalidPort int) bool {
			// Skip empty string because it's a special case
			if invalidType == "" {
				return true
			}

			// Test invalid transport type error descriptiveness
			cfg1 := NewConfig()
			cfg1.TransportType = invalidType
			err1 := cfg1.ValidateTransport()

			if err1 != nil {
				errMsg := err1.Error()
				// Should mention transport and include the invalid value
				if !strings.Contains(errMsg, "transport") || !strings.Contains(errMsg, invalidType) {
					t.Logf("Expected transport type error to be descriptive, got: %v", err1)
					return false
				}
			}

			// Test invalid port error descriptiveness
			cfg2 := NewConfig()
			cfg2.TransportType = "sse"
			cfg2.Port = invalidPort
			cfg2.Host = "localhost"
			err2 := cfg2.ValidateTransport()

			if err2 != nil {
				errMsg := err2.Error()
				portStr := fmt.Sprintf("%d", invalidPort)
				// Should mention port and include the invalid value
				if !strings.Contains(errMsg, "port") || !strings.Contains(errMsg, portStr) {
					t.Logf("Expected port error to be descriptive, got: %v", err2)
					return false
				}
			}

			// Test empty host error descriptiveness
			cfg3 := NewConfig()
			cfg3.TransportType = "streamablehttp"
			cfg3.Port = 8080
			cfg3.Host = ""
			err3 := cfg3.ValidateTransport()

			if err3 != nil {
				errMsg := err3.Error()
				// Should mention host and empty
				if !strings.Contains(errMsg, "host") || !strings.Contains(errMsg, "empty") {
					t.Logf("Expected host error to be descriptive, got: %v", err3)
					return false
				}
			}

			return true
		},
		genInvalidTransportType,
		genInvalidPort,
	))

	properties.Property("port validation error is descriptive for network transports only", prop.ForAll(
		func(transportType string, invalidPort int) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = invalidPort
			cfg.Host = "localhost"

			err := cfg.ValidateTransport()

			if transportType == "stdio" {
				// STDIO should not produce port validation errors
				if err != nil {
					errMsg := err.Error()
					if strings.Contains(errMsg, "port") && strings.Contains(errMsg, "range") {
						t.Logf("Expected stdio transport to not produce port validation errors, got: %v", err)
						return false
					}
				}
			} else if transportType == "sse" || transportType == "streamablehttp" {
				// Network transports should produce descriptive port errors
				if err == nil {
					t.Logf("Expected network transport '%s' with invalid port %d to be rejected", transportType, invalidPort)
					return false
				}

				errMsg := err.Error()
				if !strings.Contains(errMsg, "port") {
					t.Logf("Expected error message to mention 'port', got: %v", err)
					return false
				}

				// Should include the invalid port value
				portStr := fmt.Sprintf("%d", invalidPort)
				if !strings.Contains(errMsg, portStr) {
					t.Logf("Expected error message to include invalid port value %d, got: %v", invalidPort, err)
					return false
				}
			}

			return true
		},
		gen.OneConstOf("stdio", "sse", "streamablehttp"),
		genInvalidPort,
	))

	properties.Property("host validation error is descriptive for network transports only", prop.ForAll(
		func(transportType string, port int) bool {
			cfg := NewConfig()
			cfg.TransportType = transportType
			cfg.Port = port
			cfg.Host = ""

			err := cfg.ValidateTransport()

			if transportType == "stdio" {
				// STDIO should not produce host validation errors
				if err != nil {
					errMsg := err.Error()
					if strings.Contains(errMsg, "host") && strings.Contains(errMsg, "empty") {
						t.Logf("Expected stdio transport to not produce host validation errors, got: %v", err)
						return false
					}
				}
			} else if transportType == "sse" || transportType == "streamablehttp" {
				// Network transports should produce descriptive host errors
				if err == nil {
					t.Logf("Expected network transport '%s' with empty host to be rejected", transportType)
					return false
				}

				errMsg := err.Error()
				if !strings.Contains(errMsg, "host") {
					t.Logf("Expected error message to mention 'host', got: %v", err)
					return false
				}
				if !strings.Contains(errMsg, "empty") {
					t.Logf("Expected error message to mention 'empty', got: %v", err)
					return false
				}
			}

			return true
		},
		gen.OneConstOf("stdio", "sse", "streamablehttp"),
		gen.IntRange(-1000, 100000),
	))

	properties.Property("error messages use consistent terminology", prop.ForAll(
		func(invalidType string, invalidPort int) bool {
			// Test invalid transport type
			cfg1 := NewConfig()
			cfg1.TransportType = invalidType
			err1 := cfg1.ValidateTransport()

			if err1 != nil {
				errMsg := err1.Error()
				// Should use consistent terminology: "transport type" or "transport"
				if !strings.Contains(errMsg, "transport") {
					t.Logf("Expected error message to use 'transport' terminology, got: %v", err1)
					return false
				}
			}

			// Test invalid port
			cfg2 := NewConfig()
			cfg2.TransportType = "sse"
			cfg2.Port = invalidPort
			cfg2.Host = "localhost"
			err2 := cfg2.ValidateTransport()

			if err2 != nil {
				errMsg := err2.Error()
				// Should use consistent terminology: "port"
				if !strings.Contains(errMsg, "port") {
					t.Logf("Expected error message to use 'port' terminology, got: %v", err2)
					return false
				}
			}

			// Test empty host
			cfg3 := NewConfig()
			cfg3.TransportType = "streamablehttp"
			cfg3.Port = 8080
			cfg3.Host = ""
			err3 := cfg3.ValidateTransport()

			if err3 != nil {
				errMsg := err3.Error()
				// Should use consistent terminology: "host"
				if !strings.Contains(errMsg, "host") {
					t.Logf("Expected error message to use 'host' terminology, got: %v", err3)
					return false
				}
			}

			return true
		},
		genInvalidTransportType,
		genInvalidPort,
	))

	properties.Property("validation errors provide actionable information", prop.ForAll(
		func(invalidType string, invalidPort int) bool {
			// Test that errors tell users what they need to fix

			// Invalid transport type error should list valid options
			cfg1 := NewConfig()
			cfg1.TransportType = invalidType
			err1 := cfg1.ValidateTransport()

			if err1 != nil {
				errMsg := err1.Error()
				// Should provide valid options
				validOptionsCount := 0
				if strings.Contains(errMsg, "stdio") {
					validOptionsCount++
				}
				if strings.Contains(errMsg, "sse") {
					validOptionsCount++
				}
				if strings.Contains(errMsg, "streamablehttp") {
					validOptionsCount++
				}

				if validOptionsCount < 3 {
					t.Logf("Expected error message to list all valid transport types, got: %v", err1)
					return false
				}
			}

			// Invalid port error should specify valid range
			cfg2 := NewConfig()
			cfg2.TransportType = "sse"
			cfg2.Port = invalidPort
			cfg2.Host = "localhost"
			err2 := cfg2.ValidateTransport()

			if err2 != nil {
				errMsg := err2.Error()
				// Should specify valid range
				if !strings.Contains(errMsg, "1") || !strings.Contains(errMsg, "65535") {
					t.Logf("Expected error message to specify valid port range, got: %v", err2)
					return false
				}
			}

			return true
		},
		genInvalidTransportType,
		genInvalidPort,
	))

	properties.Property("error messages are human-readable and not just error codes", prop.ForAll(
		func(invalidType string) bool {
			cfg := NewConfig()
			cfg.TransportType = invalidType

			err := cfg.ValidateTransport()
			if err == nil {
				t.Logf("Expected invalid transport type '%s' to be rejected", invalidType)
				return false
			}

			errMsg := err.Error()

			// Error message should be more than just a code or number
			// It should contain descriptive words
			descriptiveWords := []string{"invalid", "must", "transport", "type"}
			foundWords := 0
			for _, word := range descriptiveWords {
				if strings.Contains(strings.ToLower(errMsg), word) {
					foundWords++
				}
			}

			if foundWords < 2 {
				t.Logf("Expected error message to be human-readable with descriptive words, got: %v", err)
				return false
			}

			return true
		},
		genInvalidTransportType,
	))

	properties.TestingRun(t)
}
