package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadFromFileWithTransportSSE verifies that transport configuration
// can be loaded from a config file for SSE transport
func TestLoadFromFileWithTransportSSE(t *testing.T) {
	// Clean up environment to ensure file takes precedence
	_ = os.Unsetenv("TRANSPORT_TYPE")
	_ = os.Unsetenv("HOST")
	_ = os.Unsetenv("PORT")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
		_ = os.Unsetenv("HOST")
		_ = os.Unsetenv("PORT")
	}()

	// Create a config file with SSE transport
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `
log_level: info
docs_base_url: https://docs.nats.io
transport_type: sse
host: 0.0.0.0
port: 8080
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	cfg, err := LoadFromFile(configFile)
	if err != nil {
		t.Fatalf("Expected LoadFromFile to succeed, got error: %v", err)
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

// TestLoadFromFileWithTransportStreamableHTTP verifies that transport configuration
// can be loaded from a config file for StreamableHTTP transport
func TestLoadFromFileWithTransportStreamableHTTP(t *testing.T) {
	// Clean up environment
	_ = os.Unsetenv("TRANSPORT_TYPE")
	_ = os.Unsetenv("HOST")
	_ = os.Unsetenv("PORT")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
		_ = os.Unsetenv("HOST")
		_ = os.Unsetenv("PORT")
	}()

	// Create a config file with StreamableHTTP transport
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `
log_level: info
docs_base_url: https://docs.nats.io
transport_type: streamablehttp
host: localhost
port: 9090
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	cfg, err := LoadFromFile(configFile)
	if err != nil {
		t.Fatalf("Expected LoadFromFile to succeed, got error: %v", err)
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

// TestLoadFromFileWithTransportStdio verifies that transport configuration
// can be loaded from a config file for STDIO transport
func TestLoadFromFileWithTransportStdio(t *testing.T) {
	// Clean up environment
	_ = os.Unsetenv("TRANSPORT_TYPE")
	_ = os.Unsetenv("HOST")
	_ = os.Unsetenv("PORT")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
		_ = os.Unsetenv("HOST")
		_ = os.Unsetenv("PORT")
	}()

	// Create a config file with STDIO transport
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `
log_level: info
docs_base_url: https://docs.nats.io
transport_type: stdio
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	cfg, err := LoadFromFile(configFile)
	if err != nil {
		t.Fatalf("Expected LoadFromFile to succeed, got error: %v", err)
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

// TestLoadFromFileWithPartialTransportConfig verifies that partial transport
// configuration in config file uses defaults for omitted values
func TestLoadFromFileWithPartialTransportConfig(t *testing.T) {
	// Clean up environment
	_ = os.Unsetenv("TRANSPORT_TYPE")
	_ = os.Unsetenv("HOST")
	_ = os.Unsetenv("PORT")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
		_ = os.Unsetenv("HOST")
		_ = os.Unsetenv("PORT")
	}()

	// Create a config file with only transport type
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `
log_level: info
docs_base_url: https://docs.nats.io
transport_type: sse
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	cfg, err := LoadFromFile(configFile)
	if err != nil {
		t.Fatalf("Expected LoadFromFile to succeed, got error: %v", err)
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

// TestLoadFromFileTransportOverridesEnvironment verifies that config file
// transport settings override environment variables
func TestLoadFromFileTransportOverridesEnvironment(t *testing.T) {
	// Set environment variables with different values
	_ = os.Setenv("TRANSPORT_TYPE", "stdio")
	_ = os.Setenv("HOST", "env-host")
	_ = os.Setenv("PORT", "7777")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
		_ = os.Unsetenv("HOST")
		_ = os.Unsetenv("PORT")
	}()

	// Create a config file with different transport settings
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `
log_level: info
docs_base_url: https://docs.nats.io
transport_type: sse
host: file-host
port: 8888
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	cfg, err := LoadFromFile(configFile)
	if err != nil {
		t.Fatalf("Expected LoadFromFile to succeed, got error: %v", err)
	}

	// Config file values should override environment variables
	if cfg.TransportType != "sse" {
		t.Errorf("Expected TransportType from file 'sse', got '%s'", cfg.TransportType)
	}

	if cfg.Host != "file-host" {
		t.Errorf("Expected Host from file 'file-host', got '%s'", cfg.Host)
	}

	if cfg.Port != 8888 {
		t.Errorf("Expected Port from file 8888, got %d", cfg.Port)
	}
}

// TestLoadFromFileWithNoTransportConfig verifies that config file without
// transport settings uses defaults
func TestLoadFromFileWithNoTransportConfig(t *testing.T) {
	// Clean up environment
	_ = os.Unsetenv("TRANSPORT_TYPE")
	_ = os.Unsetenv("HOST")
	_ = os.Unsetenv("PORT")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
		_ = os.Unsetenv("HOST")
		_ = os.Unsetenv("PORT")
	}()

	// Create a config file without transport settings
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `
log_level: info
docs_base_url: https://docs.nats.io
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	cfg, err := LoadFromFile(configFile)
	if err != nil {
		t.Fatalf("Expected LoadFromFile to succeed, got error: %v", err)
	}

	// Should use default transport values
	if cfg.TransportType != "stdio" {
		t.Errorf("Expected default TransportType 'stdio', got '%s'", cfg.TransportType)
	}

	if cfg.Host != "localhost" {
		t.Errorf("Expected default Host 'localhost', got '%s'", cfg.Host)
	}

	if cfg.Port != 0 {
		t.Errorf("Expected default Port 0, got %d", cfg.Port)
	}
}
