package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadWithFlagsTransportSSE verifies that transport flags can be loaded
// for SSE transport
func TestLoadWithFlagsTransportSSE(t *testing.T) {
	// Clean up environment
	_ = os.Unsetenv("TRANSPORT_TYPE")
	_ = os.Unsetenv("HOST")
	_ = os.Unsetenv("PORT")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
		_ = os.Unsetenv("HOST")
		_ = os.Unsetenv("PORT")
	}()

	// Provide transport flags
	flags := map[string]interface{}{
		"transport_type": "sse",
		"host":           "0.0.0.0",
		"port":           8080,
	}

	cfg, err := LoadWithFlags("", flags)
	if err != nil {
		t.Fatalf("Expected LoadWithFlags to succeed, got error: %v", err)
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

// TestLoadWithFlagsTransportStreamableHTTP verifies that transport flags can be loaded
// for StreamableHTTP transport
func TestLoadWithFlagsTransportStreamableHTTP(t *testing.T) {
	// Clean up environment
	_ = os.Unsetenv("TRANSPORT_TYPE")
	_ = os.Unsetenv("HOST")
	_ = os.Unsetenv("PORT")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
		_ = os.Unsetenv("HOST")
		_ = os.Unsetenv("PORT")
	}()

	// Provide transport flags
	flags := map[string]interface{}{
		"transport_type": "streamablehttp",
		"host":           "localhost",
		"port":           9090,
	}

	cfg, err := LoadWithFlags("", flags)
	if err != nil {
		t.Fatalf("Expected LoadWithFlags to succeed, got error: %v", err)
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

// TestLoadWithFlagsTransportStdio verifies that transport flags can be loaded
// for STDIO transport
func TestLoadWithFlagsTransportStdio(t *testing.T) {
	// Clean up environment
	_ = os.Unsetenv("TRANSPORT_TYPE")
	_ = os.Unsetenv("HOST")
	_ = os.Unsetenv("PORT")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
		_ = os.Unsetenv("HOST")
		_ = os.Unsetenv("PORT")
	}()

	// Provide transport flags
	flags := map[string]interface{}{
		"transport_type": "stdio",
	}

	cfg, err := LoadWithFlags("", flags)
	if err != nil {
		t.Fatalf("Expected LoadWithFlags to succeed, got error: %v", err)
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

// TestLoadWithFlagsTransportOverridesConfigFile verifies that transport flags
// override config file settings
func TestLoadWithFlagsTransportOverridesConfigFile(t *testing.T) {
	// Clean up environment
	_ = os.Unsetenv("TRANSPORT_TYPE")
	_ = os.Unsetenv("HOST")
	_ = os.Unsetenv("PORT")
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
transport_type: stdio
host: file-host
port: 7777
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Provide flags with different values (should override file)
	flags := map[string]interface{}{
		"transport_type": "sse",
		"host":           "flag-host",
		"port":           8888,
	}

	cfg, err := LoadWithFlags(configFile, flags)
	if err != nil {
		t.Fatalf("Expected LoadWithFlags to succeed, got error: %v", err)
	}

	// Flags should override config file
	if cfg.TransportType != "sse" {
		t.Errorf("Expected TransportType from flags 'sse', got '%s'", cfg.TransportType)
	}

	if cfg.Host != "flag-host" {
		t.Errorf("Expected Host from flags 'flag-host', got '%s'", cfg.Host)
	}

	if cfg.Port != 8888 {
		t.Errorf("Expected Port from flags 8888, got %d", cfg.Port)
	}
}

// TestLoadWithFlagsTransportOverridesEnvironment verifies that transport flags
// override environment variables
func TestLoadWithFlagsTransportOverridesEnvironment(t *testing.T) {
	// Set environment variables with different values
	_ = os.Setenv("TRANSPORT_TYPE", "stdio")
	_ = os.Setenv("HOST", "env-host")
	_ = os.Setenv("PORT", "7777")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
		_ = os.Unsetenv("HOST")
		_ = os.Unsetenv("PORT")
	}()

	// Provide flags with different values (should override env)
	flags := map[string]interface{}{
		"transport_type": "streamablehttp",
		"host":           "flag-host",
		"port":           9999,
	}

	cfg, err := LoadWithFlags("", flags)
	if err != nil {
		t.Fatalf("Expected LoadWithFlags to succeed, got error: %v", err)
	}

	// Flags should override environment variables
	if cfg.TransportType != "streamablehttp" {
		t.Errorf("Expected TransportType from flags 'streamablehttp', got '%s'", cfg.TransportType)
	}

	if cfg.Host != "flag-host" {
		t.Errorf("Expected Host from flags 'flag-host', got '%s'", cfg.Host)
	}

	if cfg.Port != 9999 {
		t.Errorf("Expected Port from flags 9999, got %d", cfg.Port)
	}
}

// TestLoadWithFlagsTransportPrecedence verifies the full precedence order:
// flags > config file > environment > defaults
func TestLoadWithFlagsTransportPrecedence(t *testing.T) {
	// Set environment variables (lowest precedence)
	_ = os.Setenv("TRANSPORT_TYPE", "stdio")
	_ = os.Setenv("HOST", "env-host")
	_ = os.Setenv("PORT", "5555")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
		_ = os.Unsetenv("HOST")
		_ = os.Unsetenv("PORT")
	}()

	// Create a config file (middle precedence)
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `
log_level: info
docs_base_url: https://docs.nats.io
transport_type: sse
host: file-host
port: 6666
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Provide flags (highest precedence)
	flags := map[string]interface{}{
		"transport_type": "streamablehttp",
		"host":           "flag-host",
		"port":           7777,
	}

	cfg, err := LoadWithFlags(configFile, flags)
	if err != nil {
		t.Fatalf("Expected LoadWithFlags to succeed, got error: %v", err)
	}

	// Flags should have highest precedence
	if cfg.TransportType != "streamablehttp" {
		t.Errorf("Expected TransportType from flags 'streamablehttp', got '%s'", cfg.TransportType)
	}

	if cfg.Host != "flag-host" {
		t.Errorf("Expected Host from flags 'flag-host', got '%s'", cfg.Host)
	}

	if cfg.Port != 7777 {
		t.Errorf("Expected Port from flags 7777, got %d", cfg.Port)
	}
}

// TestLoadWithFlagsPartialTransportFlags verifies that partial transport flags
// work correctly with other sources
func TestLoadWithFlagsPartialTransportFlags(t *testing.T) {
	// Clean up environment
	_ = os.Unsetenv("TRANSPORT_TYPE")
	_ = os.Unsetenv("HOST")
	_ = os.Unsetenv("PORT")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
		_ = os.Unsetenv("HOST")
		_ = os.Unsetenv("PORT")
	}()

	// Create a config file with some transport settings
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `
log_level: info
docs_base_url: https://docs.nats.io
transport_type: sse
host: file-host
port: 8080
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Provide only transport_type flag (port and host should come from file)
	flags := map[string]interface{}{
		"transport_type": "streamablehttp",
	}

	cfg, err := LoadWithFlags(configFile, flags)
	if err != nil {
		t.Fatalf("Expected LoadWithFlags to succeed, got error: %v", err)
	}

	// Transport type should come from flags
	if cfg.TransportType != "streamablehttp" {
		t.Errorf("Expected TransportType from flags 'streamablehttp', got '%s'", cfg.TransportType)
	}

	// Host and port should come from config file
	if cfg.Host != "file-host" {
		t.Errorf("Expected Host from file 'file-host', got '%s'", cfg.Host)
	}

	if cfg.Port != 8080 {
		t.Errorf("Expected Port from file 8080, got %d", cfg.Port)
	}
}

// TestLoadWithFlagsEmptyTransportFlags verifies that empty flags map
// uses config file or defaults
func TestLoadWithFlagsEmptyTransportFlags(t *testing.T) {
	// Clean up environment
	_ = os.Unsetenv("TRANSPORT_TYPE")
	_ = os.Unsetenv("HOST")
	_ = os.Unsetenv("PORT")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
		_ = os.Unsetenv("HOST")
		_ = os.Unsetenv("PORT")
	}()

	// Create a config file with transport settings
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `
log_level: info
docs_base_url: https://docs.nats.io
transport_type: sse
host: file-host
port: 8080
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Empty flags map
	flags := make(map[string]interface{})

	cfg, err := LoadWithFlags(configFile, flags)
	if err != nil {
		t.Fatalf("Expected LoadWithFlags to succeed, got error: %v", err)
	}

	// Should use config file values
	if cfg.TransportType != "sse" {
		t.Errorf("Expected TransportType from file 'sse', got '%s'", cfg.TransportType)
	}

	if cfg.Host != "file-host" {
		t.Errorf("Expected Host from file 'file-host', got '%s'", cfg.Host)
	}

	if cfg.Port != 8080 {
		t.Errorf("Expected Port from file 8080, got %d", cfg.Port)
	}
}

// TestLoadWithFlagsNoConfigFileUsesDefaults verifies that without config file,
// defaults are used for values not in flags
func TestLoadWithFlagsNoConfigFileUsesDefaults(t *testing.T) {
	// Clean up environment
	_ = os.Unsetenv("TRANSPORT_TYPE")
	_ = os.Unsetenv("HOST")
	_ = os.Unsetenv("PORT")
	defer func() {
		_ = os.Unsetenv("TRANSPORT_TYPE")
		_ = os.Unsetenv("HOST")
		_ = os.Unsetenv("PORT")
	}()

	// Provide only transport_type flag
	flags := map[string]interface{}{
		"transport_type": "sse",
	}

	cfg, err := LoadWithFlags("", flags)
	if err != nil {
		t.Fatalf("Expected LoadWithFlags to succeed, got error: %v", err)
	}

	// Transport type should come from flags
	if cfg.TransportType != "sse" {
		t.Errorf("Expected TransportType from flags 'sse', got '%s'", cfg.TransportType)
	}

	// Host and port should use defaults
	if cfg.Host != "localhost" {
		t.Errorf("Expected default Host 'localhost', got '%s'", cfg.Host)
	}

	if cfg.Port != 0 {
		t.Errorf("Expected default Port 0, got %d", cfg.Port)
	}
}
