package config

import (
	"os"
	"testing"
)

// TestLoadTransportFromEnvironmentVariables verifies that transport configuration
// can be loaded from environment variables
func TestLoadTransportFromEnvironmentVariables(t *testing.T) {
	// Set transport environment variables
	_ = os.Setenv("TRANSPORT_TYPE", "sse")
	_ = os.Setenv("HOST", "0.0.0.0")
	_ = os.Setenv("PORT", "8080")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
		_ = os.Unsetenv("HOST")
		_ = os.Unsetenv("PORT")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected Load to succeed, got error: %v", err)
	}

	if cfg.TransportType != "sse" {
		t.Errorf("Expected TransportType to be 'sse', got '%s'", cfg.TransportType)
	}

	if cfg.Host != "0.0.0.0" {
		t.Errorf("Expected Host to be '0.0.0.0', got '%s'", cfg.Host)
	}

	if cfg.Port != 8080 {
		t.Errorf("Expected Port to be 8080, got %d", cfg.Port)
	}
}

// TestLoadTransportFromEnvironmentVariablesStreamableHTTP verifies that
// streamablehttp transport can be loaded from environment variables
func TestLoadTransportFromEnvironmentVariablesStreamableHTTP(t *testing.T) {
	// Set transport environment variables
	_ = os.Setenv("TRANSPORT_TYPE", "streamablehttp")
	_ = os.Setenv("HOST", "localhost")
	_ = os.Setenv("PORT", "9090")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
		_ = os.Unsetenv("HOST")
		_ = os.Unsetenv("PORT")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected Load to succeed, got error: %v", err)
	}

	if cfg.TransportType != "streamablehttp" {
		t.Errorf("Expected TransportType to be 'streamablehttp', got '%s'", cfg.TransportType)
	}

	if cfg.Host != "localhost" {
		t.Errorf("Expected Host to be 'localhost', got '%s'", cfg.Host)
	}

	if cfg.Port != 9090 {
		t.Errorf("Expected Port to be 9090, got %d", cfg.Port)
	}
}

// TestLoadTransportFromEnvironmentVariablesStdio verifies that
// stdio transport can be loaded from environment variables
func TestLoadTransportFromEnvironmentVariablesStdio(t *testing.T) {
	// Set transport environment variables
	_ = os.Setenv("TRANSPORT_TYPE", "stdio")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected Load to succeed, got error: %v", err)
	}

	if cfg.TransportType != "stdio" {
		t.Errorf("Expected TransportType to be 'stdio', got '%s'", cfg.TransportType)
	}

	// For stdio, host and port should have default values
	if cfg.Host != "localhost" {
		t.Errorf("Expected default Host to be 'localhost', got '%s'", cfg.Host)
	}

	if cfg.Port != 0 {
		t.Errorf("Expected default Port to be 0, got %d", cfg.Port)
	}
}

// TestLoadTransportFromEnvironmentVariablesInvalidPort verifies that
// invalid port values in environment variables are ignored
func TestLoadTransportFromEnvironmentVariablesInvalidPort(t *testing.T) {
	// Set transport environment variables with invalid port
	_ = os.Setenv("TRANSPORT_TYPE", "sse")
	_ = os.Setenv("PORT", "not-a-number")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
		_ = os.Unsetenv("PORT")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected Load to succeed, got error: %v", err)
	}

	if cfg.TransportType != "sse" {
		t.Errorf("Expected TransportType to be 'sse', got '%s'", cfg.TransportType)
	}

	// Invalid port should fall back to default (0)
	if cfg.Port != 0 {
		t.Errorf("Expected default Port for invalid env var, got %d", cfg.Port)
	}
}

// TestLoadTransportFromEnvironmentVariablesPartial verifies that
// partial transport configuration from environment variables works
func TestLoadTransportFromEnvironmentVariablesPartial(t *testing.T) {
	// Set only transport type, not host or port
	_ = os.Setenv("TRANSPORT_TYPE", "sse")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected Load to succeed, got error: %v", err)
	}

	if cfg.TransportType != "sse" {
		t.Errorf("Expected TransportType to be 'sse', got '%s'", cfg.TransportType)
	}

	// Host and port should have default values
	if cfg.Host != "localhost" {
		t.Errorf("Expected default Host to be 'localhost', got '%s'", cfg.Host)
	}

	if cfg.Port != 0 {
		t.Errorf("Expected default Port to be 0, got %d", cfg.Port)
	}
}

// TestLoadTransportFromEnvironmentVariablesEmptyValues verifies that
// empty environment variable values are ignored
func TestLoadTransportFromEnvironmentVariablesEmptyValues(t *testing.T) {
	// Set empty transport environment variables
	_ = os.Setenv("TRANSPORT_TYPE", "")
	_ = os.Setenv("HOST", "")
	_ = os.Setenv("PORT", "")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
		_ = os.Unsetenv("HOST")
		_ = os.Unsetenv("PORT")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected Load to succeed, got error: %v", err)
	}

	// Empty values should be ignored, defaults should be used
	if cfg.TransportType != "stdio" {
		t.Errorf("Expected default TransportType to be 'stdio', got '%s'", cfg.TransportType)
	}

	if cfg.Host != "localhost" {
		t.Errorf("Expected default Host to be 'localhost', got '%s'", cfg.Host)
	}

	if cfg.Port != 0 {
		t.Errorf("Expected default Port to be 0, got %d", cfg.Port)
	}
}
