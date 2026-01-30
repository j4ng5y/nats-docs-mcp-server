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

// TestDefaultConfig verifies that NewConfig returns a Config with all default values set
func TestDefaultConfig(t *testing.T) {
	cfg := NewConfig()

	// Test server settings
	if cfg.LogLevel != "info" {
		t.Errorf("Expected default LogLevel to be 'info', got '%s'", cfg.LogLevel)
	}

	// Test documentation settings
	if cfg.DocsBaseURL != "https://docs.nats.io" {
		t.Errorf("Expected default DocsBaseURL to be 'https://docs.nats.io', got '%s'", cfg.DocsBaseURL)
	}

	if cfg.FetchTimeout != 30 {
		t.Errorf("Expected default FetchTimeout to be 30, got %d", cfg.FetchTimeout)
	}

	if cfg.MaxConcurrent != 5 {
		t.Errorf("Expected default MaxConcurrent to be 5, got %d", cfg.MaxConcurrent)
	}

	if cfg.CacheDir != "" {
		t.Errorf("Expected default CacheDir to be empty, got '%s'", cfg.CacheDir)
	}

	// Test search settings
	if cfg.MaxSearchResults != 50 {
		t.Errorf("Expected default MaxSearchResults to be 50, got %d", cfg.MaxSearchResults)
	}
}

// TestConfigStructFields verifies that Config struct has all required fields
func TestConfigStructFields(t *testing.T) {
	cfg := Config{
		LogLevel:         "debug",
		DocsBaseURL:      "https://example.com",
		FetchTimeout:     60,
		MaxConcurrent:    10,
		CacheDir:         "/tmp/cache",
		MaxSearchResults: 100,
	}

	// Verify all fields can be set
	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel to be 'debug', got '%s'", cfg.LogLevel)
	}

	if cfg.DocsBaseURL != "https://example.com" {
		t.Errorf("Expected DocsBaseURL to be 'https://example.com', got '%s'", cfg.DocsBaseURL)
	}

	if cfg.FetchTimeout != 60 {
		t.Errorf("Expected FetchTimeout to be 60, got %d", cfg.FetchTimeout)
	}

	if cfg.MaxConcurrent != 10 {
		t.Errorf("Expected MaxConcurrent to be 10, got %d", cfg.MaxConcurrent)
	}

	if cfg.CacheDir != "/tmp/cache" {
		t.Errorf("Expected CacheDir to be '/tmp/cache', got '%s'", cfg.CacheDir)
	}

	if cfg.MaxSearchResults != 100 {
		t.Errorf("Expected MaxSearchResults to be 100, got %d", cfg.MaxSearchResults)
	}
}

// TestConfigZeroValues verifies that Config can be created with zero values
func TestConfigZeroValues(t *testing.T) {
	cfg := Config{}

	// Zero values should be valid (empty strings, zero ints)
	if cfg.LogLevel != "" {
		t.Errorf("Expected zero value LogLevel to be empty, got '%s'", cfg.LogLevel)
	}

	if cfg.DocsBaseURL != "" {
		t.Errorf("Expected zero value DocsBaseURL to be empty, got '%s'", cfg.DocsBaseURL)
	}

	if cfg.FetchTimeout != 0 {
		t.Errorf("Expected zero value FetchTimeout to be 0, got %d", cfg.FetchTimeout)
	}

	if cfg.MaxConcurrent != 0 {
		t.Errorf("Expected zero value MaxConcurrent to be 0, got %d", cfg.MaxConcurrent)
	}

	if cfg.CacheDir != "" {
		t.Errorf("Expected zero value CacheDir to be empty, got '%s'", cfg.CacheDir)
	}

	if cfg.MaxSearchResults != 0 {
		t.Errorf("Expected zero value MaxSearchResults to be 0, got %d", cfg.MaxSearchResults)
	}
}

// TestNewConfigReturnsPointer verifies that NewConfig returns a pointer
func TestNewConfigReturnsPointer(t *testing.T) {
	cfg := NewConfig()

	if cfg == nil {
		t.Fatal("Expected NewConfig to return non-nil pointer")
	}

	// Verify we can modify the returned config
	cfg.LogLevel = "debug"
	if cfg.LogLevel != "debug" {
		t.Errorf("Expected to be able to modify returned config")
	}
}

// TestMultipleNewConfigCalls verifies that each call returns a new instance
func TestMultipleNewConfigCalls(t *testing.T) {
	cfg1 := NewConfig()
	cfg2 := NewConfig()

	// Modify first config
	cfg1.LogLevel = "debug"

	// Second config should still have default value
	if cfg2.LogLevel != "info" {
		t.Errorf("Expected cfg2.LogLevel to be 'info', got '%s'", cfg2.LogLevel)
	}

	// Verify they are different instances
	if cfg1 == cfg2 {
		t.Error("Expected NewConfig to return different instances")
	}
}

// TestLoadFromEnvironmentVariables verifies that configuration can be loaded from environment variables
func TestLoadFromEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("DOCS_BASE_URL", "https://test.example.com")
	os.Setenv("FETCH_TIMEOUT", "60")
	os.Setenv("MAX_CONCURRENT", "10")
	os.Setenv("CACHE_DIR", "/tmp/test-cache")
	os.Setenv("MAX_SEARCH_RESULTS", "100")
	defer func() {
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("DOCS_BASE_URL")
		os.Unsetenv("FETCH_TIMEOUT")
		os.Unsetenv("MAX_CONCURRENT")
		os.Unsetenv("CACHE_DIR")
		os.Unsetenv("MAX_SEARCH_RESULTS")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected Load to succeed, got error: %v", err)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel to be 'debug', got '%s'", cfg.LogLevel)
	}

	if cfg.DocsBaseURL != "https://test.example.com" {
		t.Errorf("Expected DocsBaseURL to be 'https://test.example.com', got '%s'", cfg.DocsBaseURL)
	}

	if cfg.FetchTimeout != 60 {
		t.Errorf("Expected FetchTimeout to be 60, got %d", cfg.FetchTimeout)
	}

	if cfg.MaxConcurrent != 10 {
		t.Errorf("Expected MaxConcurrent to be 10, got %d", cfg.MaxConcurrent)
	}

	if cfg.CacheDir != "/tmp/test-cache" {
		t.Errorf("Expected CacheDir to be '/tmp/test-cache', got '%s'", cfg.CacheDir)
	}

	if cfg.MaxSearchResults != 100 {
		t.Errorf("Expected MaxSearchResults to be 100, got %d", cfg.MaxSearchResults)
	}
}

// TestLoadFromConfigFile verifies that configuration can be loaded from a YAML file
func TestLoadFromConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `
log_level: warn
docs_base_url: https://config.example.com
fetch_timeout: 45
max_concurrent: 8
cache_dir: /tmp/config-cache
max_search_results: 75
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	cfg, err := LoadFromFile(configFile)
	if err != nil {
		t.Fatalf("Expected LoadFromFile to succeed, got error: %v", err)
	}

	if cfg.LogLevel != "warn" {
		t.Errorf("Expected LogLevel to be 'warn', got '%s'", cfg.LogLevel)
	}

	if cfg.DocsBaseURL != "https://config.example.com" {
		t.Errorf("Expected DocsBaseURL to be 'https://config.example.com', got '%s'", cfg.DocsBaseURL)
	}

	if cfg.FetchTimeout != 45 {
		t.Errorf("Expected FetchTimeout to be 45, got %d", cfg.FetchTimeout)
	}

	if cfg.MaxConcurrent != 8 {
		t.Errorf("Expected MaxConcurrent to be 8, got %d", cfg.MaxConcurrent)
	}

	if cfg.CacheDir != "/tmp/config-cache" {
		t.Errorf("Expected CacheDir to be '/tmp/config-cache', got '%s'", cfg.CacheDir)
	}

	if cfg.MaxSearchResults != 75 {
		t.Errorf("Expected MaxSearchResults to be 75, got %d", cfg.MaxSearchResults)
	}
}

// TestLoadWithDefaults verifies that Load returns defaults when no config is provided
func TestLoadWithDefaults(t *testing.T) {
	// Clear any environment variables that might interfere
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("DOCS_BASE_URL")
	os.Unsetenv("FETCH_TIMEOUT")
	os.Unsetenv("MAX_CONCURRENT")
	os.Unsetenv("CACHE_DIR")
	os.Unsetenv("MAX_SEARCH_RESULTS")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected Load to succeed, got error: %v", err)
	}

	// Should have default values
	if cfg.LogLevel != "info" {
		t.Errorf("Expected default LogLevel to be 'info', got '%s'", cfg.LogLevel)
	}

	if cfg.DocsBaseURL != "https://docs.nats.io" {
		t.Errorf("Expected default DocsBaseURL to be 'https://docs.nats.io', got '%s'", cfg.DocsBaseURL)
	}

	if cfg.FetchTimeout != 30 {
		t.Errorf("Expected default FetchTimeout to be 30, got %d", cfg.FetchTimeout)
	}

	if cfg.MaxConcurrent != 5 {
		t.Errorf("Expected default MaxConcurrent to be 5, got %d", cfg.MaxConcurrent)
	}

	if cfg.CacheDir != "" {
		t.Errorf("Expected default CacheDir to be empty, got '%s'", cfg.CacheDir)
	}

	if cfg.MaxSearchResults != 50 {
		t.Errorf("Expected default MaxSearchResults to be 50, got %d", cfg.MaxSearchResults)
	}
}

// TestConfigPrecedence verifies that config file takes precedence over environment variables
func TestConfigPrecedence(t *testing.T) {
	// Set environment variables
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("DOCS_BASE_URL", "https://env.example.com")
	defer func() {
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("DOCS_BASE_URL")
	}()

	// Create a config file with different values
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `
log_level: error
docs_base_url: https://file.example.com
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	cfg, err := LoadFromFile(configFile)
	if err != nil {
		t.Fatalf("Expected LoadFromFile to succeed, got error: %v", err)
	}

	// Config file should take precedence
	if cfg.LogLevel != "error" {
		t.Errorf("Expected LogLevel from config file to be 'error', got '%s'", cfg.LogLevel)
	}

	if cfg.DocsBaseURL != "https://file.example.com" {
		t.Errorf("Expected DocsBaseURL from config file to be 'https://file.example.com', got '%s'", cfg.DocsBaseURL)
	}
}

// TestLoadWithFlags verifies that command-line flags have highest precedence
func TestLoadWithFlags(t *testing.T) {
	// Set environment variables
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("DOCS_BASE_URL", "https://env.example.com")
	defer func() {
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("DOCS_BASE_URL")
	}()

	// Create a config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `
log_level: warn
docs_base_url: https://file.example.com
fetch_timeout: 45
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Provide flags (should have highest precedence)
	flags := map[string]interface{}{
		"log_level":      "error",
		"docs_base_url":  "https://flag.example.com",
		"max_concurrent": 20,
	}

	cfg, err := LoadWithFlags(configFile, flags)
	if err != nil {
		t.Fatalf("Expected LoadWithFlags to succeed, got error: %v", err)
	}

	// Flags should override everything
	if cfg.LogLevel != "error" {
		t.Errorf("Expected LogLevel from flags to be 'error', got '%s'", cfg.LogLevel)
	}

	if cfg.DocsBaseURL != "https://flag.example.com" {
		t.Errorf("Expected DocsBaseURL from flags to be 'https://flag.example.com', got '%s'", cfg.DocsBaseURL)
	}

	if cfg.MaxConcurrent != 20 {
		t.Errorf("Expected MaxConcurrent from flags to be 20, got %d", cfg.MaxConcurrent)
	}

	// Config file should override env for values not in flags
	if cfg.FetchTimeout != 45 {
		t.Errorf("Expected FetchTimeout from config file to be 45, got %d", cfg.FetchTimeout)
	}
}

// TestLoadWithFlagsNoConfigFile verifies LoadWithFlags works without a config file
func TestLoadWithFlagsNoConfigFile(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("DOCS_BASE_URL")

	flags := map[string]interface{}{
		"log_level":     "debug",
		"docs_base_url": "https://flag.example.com",
	}

	cfg, err := LoadWithFlags("", flags)
	if err != nil {
		t.Fatalf("Expected LoadWithFlags to succeed without config file, got error: %v", err)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel from flags to be 'debug', got '%s'", cfg.LogLevel)
	}

	if cfg.DocsBaseURL != "https://flag.example.com" {
		t.Errorf("Expected DocsBaseURL from flags to be 'https://flag.example.com', got '%s'", cfg.DocsBaseURL)
	}

	// Should have defaults for other values
	if cfg.FetchTimeout != 30 {
		t.Errorf("Expected default FetchTimeout to be 30, got %d", cfg.FetchTimeout)
	}
}

// TestLoadFromFileInvalidPath verifies error handling for invalid config file path
func TestLoadFromFileInvalidPath(t *testing.T) {
	_, err := LoadFromFile("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Expected LoadFromFile to return error for invalid path")
	}
}

// TestLoadFromFileEmptyValues verifies that empty values in config file are handled
func TestLoadFromFileEmptyValues(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `
cache_dir: ""
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	cfg, err := LoadFromFile(configFile)
	if err != nil {
		t.Fatalf("Expected LoadFromFile to succeed, got error: %v", err)
	}

	// Empty string values should be set
	if cfg.CacheDir != "" {
		t.Errorf("Expected CacheDir to be empty string, got '%s'", cfg.CacheDir)
	}

	// Other values should have defaults
	if cfg.DocsBaseURL != "https://docs.nats.io" {
		t.Errorf("Expected default DocsBaseURL, got '%s'", cfg.DocsBaseURL)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("Expected default LogLevel, got '%s'", cfg.LogLevel)
	}
}

// TestLoadWithNilFlags verifies that nil flags don't cause issues
func TestLoadWithNilFlags(t *testing.T) {
	flags := map[string]interface{}{
		"log_level":     nil,
		"docs_base_url": "https://flag.example.com",
	}

	cfg, err := LoadWithFlags("", flags)
	if err != nil {
		t.Fatalf("Expected LoadWithFlags to succeed with nil flag values, got error: %v", err)
	}

	// Nil flag should not override default
	if cfg.LogLevel != "info" {
		t.Errorf("Expected default LogLevel to be 'info', got '%s'", cfg.LogLevel)
	}

	// Non-nil flag should be set
	if cfg.DocsBaseURL != "https://flag.example.com" {
		t.Errorf("Expected DocsBaseURL from flags to be 'https://flag.example.com', got '%s'", cfg.DocsBaseURL)
	}
}

// TestLoadFromEnvWithInvalidIntegers verifies that invalid integer env vars are ignored
func TestLoadFromEnvWithInvalidIntegers(t *testing.T) {
	os.Setenv("FETCH_TIMEOUT", "not-a-number")
	os.Setenv("MAX_CONCURRENT", "invalid")
	defer func() {
		os.Unsetenv("FETCH_TIMEOUT")
		os.Unsetenv("MAX_CONCURRENT")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Expected Load to succeed, got error: %v", err)
	}

	// Should fall back to defaults for invalid integers
	if cfg.FetchTimeout != 30 {
		t.Errorf("Expected default FetchTimeout for invalid env var, got %d", cfg.FetchTimeout)
	}

	if cfg.MaxConcurrent != 5 {
		t.Errorf("Expected default MaxConcurrent for invalid env var, got %d", cfg.MaxConcurrent)
	}
}

// TestValidateInvalidLogLevel verifies that invalid log levels are rejected
func TestValidateInvalidLogLevel(t *testing.T) {
	cfg := NewConfig()
	cfg.LogLevel = "invalid"

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected Validate to return error for invalid log level")
	}

	if err != nil && !strings.Contains(err.Error(), "invalid log level: invalid") {
		t.Errorf("Expected error message to contain 'invalid log level: invalid', got: %v", err)
	}
}

// TestValidateValidLogLevels verifies that valid log levels are accepted
func TestValidateValidLogLevels(t *testing.T) {
	validLevels := []string{"debug", "info", "warn", "error"}

	for _, level := range validLevels {
		cfg := NewConfig()
		cfg.LogLevel = level

		err := cfg.Validate()
		if err != nil {
			t.Errorf("Expected Validate to accept log level '%s', got error: %v", level, err)
		}
	}
}

// TestValidateNegativeTimeout verifies that negative timeouts are rejected
func TestValidateNegativeTimeout(t *testing.T) {
	cfg := NewConfig()
	cfg.FetchTimeout = -1

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected Validate to return error for negative timeout")
	}

	if err != nil && !strings.Contains(err.Error(), "fetch_timeout must be positive") {
		t.Errorf("Expected error message to contain 'fetch_timeout must be positive', got: %v", err)
	}
}

// TestValidateZeroTimeout verifies that zero timeout is rejected
func TestValidateZeroTimeout(t *testing.T) {
	cfg := NewConfig()
	cfg.FetchTimeout = 0

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected Validate to return error for zero timeout")
	}

	if err != nil && !strings.Contains(err.Error(), "fetch_timeout must be positive") {
		t.Errorf("Expected error message to contain 'fetch_timeout must be positive', got: %v", err)
	}
}

// TestValidateNegativeMaxConcurrent verifies that negative max concurrent is rejected
func TestValidateNegativeMaxConcurrent(t *testing.T) {
	cfg := NewConfig()
	cfg.MaxConcurrent = -1

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected Validate to return error for negative max concurrent")
	}

	if err != nil && !strings.Contains(err.Error(), "max_concurrent must be positive") {
		t.Errorf("Expected error message to contain 'max_concurrent must be positive', got: %v", err)
	}
}

// TestValidateZeroMaxConcurrent verifies that zero max concurrent is rejected
func TestValidateZeroMaxConcurrent(t *testing.T) {
	cfg := NewConfig()
	cfg.MaxConcurrent = 0

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected Validate to return error for zero max concurrent")
	}

	if err != nil && !strings.Contains(err.Error(), "max_concurrent must be positive") {
		t.Errorf("Expected error message to contain 'max_concurrent must be positive', got: %v", err)
	}
}

// TestValidateNegativeMaxSearchResults verifies that negative max search results is rejected
func TestValidateNegativeMaxSearchResults(t *testing.T) {
	cfg := NewConfig()
	cfg.MaxSearchResults = -1

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected Validate to return error for negative max search results")
	}

	if err != nil && !strings.Contains(err.Error(), "max_search_results must be positive") {
		t.Errorf("Expected error message to contain 'max_search_results must be positive', got: %v", err)
	}
}

// TestValidateZeroMaxSearchResults verifies that zero max search results is rejected
func TestValidateZeroMaxSearchResults(t *testing.T) {
	cfg := NewConfig()
	cfg.MaxSearchResults = 0

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected Validate to return error for zero max search results")
	}

	if err != nil && !strings.Contains(err.Error(), "max_search_results must be positive") {
		t.Errorf("Expected error message to contain 'max_search_results must be positive', got: %v", err)
	}
}

// TestValidateInvalidURL verifies that invalid URLs are rejected
func TestValidateInvalidURL(t *testing.T) {
	invalidURLs := []string{
		"not-a-url",
		"ftp://invalid-scheme.com",
		"://missing-scheme.com",
		"http://",
		"",
	}

	for _, url := range invalidURLs {
		cfg := NewConfig()
		cfg.DocsBaseURL = url

		err := cfg.Validate()
		if err == nil {
			t.Errorf("Expected Validate to return error for invalid URL: %s", url)
		}
	}
}

// TestValidateValidURLs verifies that valid URLs are accepted
func TestValidateValidURLs(t *testing.T) {
	validURLs := []string{
		"https://docs.nats.io",
		"http://localhost:8080",
		"https://example.com/path",
		"http://192.168.1.1",
	}

	for _, url := range validURLs {
		cfg := NewConfig()
		cfg.DocsBaseURL = url

		err := cfg.Validate()
		if err != nil {
			t.Errorf("Expected Validate to accept valid URL '%s', got error: %v", url, err)
		}
	}
}

// TestValidateValidConfig verifies that a valid config passes validation
func TestValidateValidConfig(t *testing.T) {
	cfg := NewConfig()

	err := cfg.Validate()
	if err != nil {
		t.Errorf("Expected default config to be valid, got error: %v", err)
	}
}

// TestValidateMultipleErrors verifies that multiple validation errors are reported
func TestValidateMultipleErrors(t *testing.T) {
	cfg := NewConfig()
	cfg.LogLevel = "invalid"
	cfg.FetchTimeout = -1
	cfg.MaxConcurrent = 0
	cfg.DocsBaseURL = "not-a-url"

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected Validate to return error for multiple invalid values")
	}

	// Error message should mention multiple issues
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Expected non-empty error message")
	}
}

// TestLoadValidatesConfiguration verifies that Load validates the configuration
func TestLoadValidatesConfiguration(t *testing.T) {
	os.Setenv("LOG_LEVEL", "invalid")
	defer os.Unsetenv("LOG_LEVEL")

	_, err := Load()
	if err == nil {
		t.Error("Expected Load to return validation error for invalid log level")
	}
}

// TestLoadFromFileValidatesConfiguration verifies that LoadFromFile validates the configuration
func TestLoadFromFileValidatesConfiguration(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `
log_level: invalid
fetch_timeout: -1
`
	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	_, err := LoadFromFile(configFile)
	if err == nil {
		t.Error("Expected LoadFromFile to return validation error")
	}
}

// TestLoadWithFlagsValidatesConfiguration verifies that LoadWithFlags validates the configuration
func TestLoadWithFlagsValidatesConfiguration(t *testing.T) {
	flags := map[string]interface{}{
		"log_level":      "invalid",
		"fetch_timeout":  -1,
		"max_concurrent": 0,
	}

	_, err := LoadWithFlags("", flags)
	if err == nil {
		t.Error("Expected LoadWithFlags to return validation error")
	}
}

// Property-Based Tests

// TestProperty9ConfigurationLoading tests Property 9: Configuration Loading
// **Validates: Requirements 8.1, 8.2, 8.3, 5.5**
//
// Feature: nats-docs-mcp-server, Property 9: Configuration Loading
//
// Property: For any valid configuration parameter provided via command-line flag,
// environment variable, or config file, the server should load and use that value,
// with the correct precedence: flags > config file > env vars > defaults.
func TestProperty9ConfigurationLoading(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for valid log levels
	genLogLevel := gen.OneConstOf("debug", "info", "warn", "error")

	// Generator for valid URLs
	genURL := gen.OneConstOf(
		"https://docs.nats.io",
		"http://localhost:8080",
		"https://example.com",
		"http://test.example.com/path",
	)

	// Generator for positive integers
	genPositiveInt := gen.IntRange(1, 1000)

	// Generator for cache directory (can be empty or a path)
	genCacheDir := gen.OneConstOf("", "/tmp/cache", "/var/cache/nats", "/home/user/.cache")

	properties.Property("config file takes precedence over environment variables", prop.ForAll(
		func(logLevel, url, cacheDir string, timeout, concurrent, maxResults int) bool {
			// Clean up environment
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("DOCS_BASE_URL")
			os.Unsetenv("FETCH_TIMEOUT")
			os.Unsetenv("MAX_CONCURRENT")
			os.Unsetenv("CACHE_DIR")
			os.Unsetenv("MAX_SEARCH_RESULTS")

			// Set environment variables with different values
			os.Setenv("LOG_LEVEL", "debug")
			os.Setenv("DOCS_BASE_URL", "https://env.example.com")
			os.Setenv("FETCH_TIMEOUT", "99")
			os.Setenv("MAX_CONCURRENT", "99")
			os.Setenv("CACHE_DIR", "/env/cache")
			os.Setenv("MAX_SEARCH_RESULTS", "99")
			defer func() {
				os.Unsetenv("LOG_LEVEL")
				os.Unsetenv("DOCS_BASE_URL")
				os.Unsetenv("FETCH_TIMEOUT")
				os.Unsetenv("MAX_CONCURRENT")
				os.Unsetenv("CACHE_DIR")
				os.Unsetenv("MAX_SEARCH_RESULTS")
			}()

			// Create a config file with generated values
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")

			// Properly quote cache_dir if it's empty to ensure YAML treats it as empty string, not null
			cacheDirYAML := cacheDir
			if cacheDir == "" {
				cacheDirYAML = `""`
			}

			configContent := fmt.Sprintf(`
log_level: %s
docs_base_url: %s
fetch_timeout: %d
max_concurrent: %d
cache_dir: %s
max_search_results: %d
`, logLevel, url, timeout, concurrent, cacheDirYAML, maxResults)

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
			if cfg.LogLevel != logLevel {
				t.Logf("Expected LogLevel %s, got %s", logLevel, cfg.LogLevel)
				return false
			}
			if cfg.DocsBaseURL != url {
				t.Logf("Expected DocsBaseURL %s, got %s", url, cfg.DocsBaseURL)
				return false
			}
			if cfg.FetchTimeout != timeout {
				t.Logf("Expected FetchTimeout %d, got %d", timeout, cfg.FetchTimeout)
				return false
			}
			if cfg.MaxConcurrent != concurrent {
				t.Logf("Expected MaxConcurrent %d, got %d", concurrent, cfg.MaxConcurrent)
				return false
			}
			if cfg.CacheDir != cacheDir {
				t.Logf("Expected CacheDir %s, got %s", cacheDir, cfg.CacheDir)
				return false
			}
			if cfg.MaxSearchResults != maxResults {
				t.Logf("Expected MaxSearchResults %d, got %d", maxResults, cfg.MaxSearchResults)
				return false
			}

			return true
		},
		genLogLevel,
		genURL,
		genCacheDir,
		genPositiveInt,
		genPositiveInt,
		genPositiveInt,
	))

	properties.Property("flags take precedence over config file and environment variables", prop.ForAll(
		func(logLevel, url, cacheDir string, timeout, concurrent, maxResults int) bool {
			// Clean up environment
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("DOCS_BASE_URL")
			os.Unsetenv("FETCH_TIMEOUT")
			os.Unsetenv("MAX_CONCURRENT")
			os.Unsetenv("CACHE_DIR")
			os.Unsetenv("MAX_SEARCH_RESULTS")

			// Set environment variables
			os.Setenv("LOG_LEVEL", "debug")
			os.Setenv("DOCS_BASE_URL", "https://env.example.com")
			defer func() {
				os.Unsetenv("LOG_LEVEL")
				os.Unsetenv("DOCS_BASE_URL")
			}()

			// Create a config file
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")
			configContent := `
log_level: warn
docs_base_url: https://file.example.com
fetch_timeout: 45
`
			if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
				t.Logf("Failed to create config file: %v", err)
				return false
			}

			// Provide flags (should have highest precedence)
			flags := map[string]interface{}{
				"log_level":          logLevel,
				"docs_base_url":      url,
				"fetch_timeout":      timeout,
				"max_concurrent":     concurrent,
				"cache_dir":          cacheDir,
				"max_search_results": maxResults,
			}

			cfg, err := LoadWithFlags(configFile, flags)
			if err != nil {
				t.Logf("Failed to load config with flags: %v", err)
				return false
			}

			// Verify flags take precedence
			if cfg.LogLevel != logLevel {
				t.Logf("Expected LogLevel from flags %s, got %s", logLevel, cfg.LogLevel)
				return false
			}
			if cfg.DocsBaseURL != url {
				t.Logf("Expected DocsBaseURL from flags %s, got %s", url, cfg.DocsBaseURL)
				return false
			}
			if cfg.FetchTimeout != timeout {
				t.Logf("Expected FetchTimeout from flags %d, got %d", timeout, cfg.FetchTimeout)
				return false
			}
			if cfg.MaxConcurrent != concurrent {
				t.Logf("Expected MaxConcurrent from flags %d, got %d", concurrent, cfg.MaxConcurrent)
				return false
			}
			if cfg.CacheDir != cacheDir {
				t.Logf("Expected CacheDir from flags %s, got %s", cacheDir, cfg.CacheDir)
				return false
			}
			if cfg.MaxSearchResults != maxResults {
				t.Logf("Expected MaxSearchResults from flags %d, got %d", maxResults, cfg.MaxSearchResults)
				return false
			}

			return true
		},
		genLogLevel,
		genURL,
		genCacheDir,
		genPositiveInt,
		genPositiveInt,
		genPositiveInt,
	))

	properties.Property("environment variables take precedence over defaults", prop.ForAll(
		func(logLevel, url, cacheDir string, timeout, concurrent, maxResults int) bool {
			// Clean up environment
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("DOCS_BASE_URL")
			os.Unsetenv("FETCH_TIMEOUT")
			os.Unsetenv("MAX_CONCURRENT")
			os.Unsetenv("CACHE_DIR")
			os.Unsetenv("MAX_SEARCH_RESULTS")

			// Set environment variables
			os.Setenv("LOG_LEVEL", logLevel)
			os.Setenv("DOCS_BASE_URL", url)
			os.Setenv("FETCH_TIMEOUT", fmt.Sprintf("%d", timeout))
			os.Setenv("MAX_CONCURRENT", fmt.Sprintf("%d", concurrent))
			os.Setenv("CACHE_DIR", cacheDir)
			os.Setenv("MAX_SEARCH_RESULTS", fmt.Sprintf("%d", maxResults))
			defer func() {
				os.Unsetenv("LOG_LEVEL")
				os.Unsetenv("DOCS_BASE_URL")
				os.Unsetenv("FETCH_TIMEOUT")
				os.Unsetenv("MAX_CONCURRENT")
				os.Unsetenv("CACHE_DIR")
				os.Unsetenv("MAX_SEARCH_RESULTS")
			}()

			cfg, err := Load()
			if err != nil {
				t.Logf("Failed to load config from env: %v", err)
				return false
			}

			// Verify environment variables override defaults
			if cfg.LogLevel != logLevel {
				t.Logf("Expected LogLevel from env %s, got %s", logLevel, cfg.LogLevel)
				return false
			}
			if cfg.DocsBaseURL != url {
				t.Logf("Expected DocsBaseURL from env %s, got %s", url, cfg.DocsBaseURL)
				return false
			}
			if cfg.FetchTimeout != timeout {
				t.Logf("Expected FetchTimeout from env %d, got %d", timeout, cfg.FetchTimeout)
				return false
			}
			if cfg.MaxConcurrent != concurrent {
				t.Logf("Expected MaxConcurrent from env %d, got %d", concurrent, cfg.MaxConcurrent)
				return false
			}
			if cfg.CacheDir != cacheDir {
				t.Logf("Expected CacheDir from env %s, got %s", cacheDir, cfg.CacheDir)
				return false
			}
			if cfg.MaxSearchResults != maxResults {
				t.Logf("Expected MaxSearchResults from env %d, got %d", maxResults, cfg.MaxSearchResults)
				return false
			}

			return true
		},
		genLogLevel,
		genURL,
		genCacheDir,
		genPositiveInt,
		genPositiveInt,
		genPositiveInt,
	))

	properties.Property("defaults are used when no configuration is provided", prop.ForAll(
		func(seed int) bool {
			// Clean up environment to ensure no interference
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("DOCS_BASE_URL")
			os.Unsetenv("FETCH_TIMEOUT")
			os.Unsetenv("MAX_CONCURRENT")
			os.Unsetenv("CACHE_DIR")
			os.Unsetenv("MAX_SEARCH_RESULTS")

			cfg, err := Load()
			if err != nil {
				t.Logf("Failed to load config with defaults: %v", err)
				return false
			}

			// Verify defaults are used
			expectedDefaults := NewConfig()
			if cfg.LogLevel != expectedDefaults.LogLevel {
				t.Logf("Expected default LogLevel %s, got %s", expectedDefaults.LogLevel, cfg.LogLevel)
				return false
			}
			if cfg.DocsBaseURL != expectedDefaults.DocsBaseURL {
				t.Logf("Expected default DocsBaseURL %s, got %s", expectedDefaults.DocsBaseURL, cfg.DocsBaseURL)
				return false
			}
			if cfg.FetchTimeout != expectedDefaults.FetchTimeout {
				t.Logf("Expected default FetchTimeout %d, got %d", expectedDefaults.FetchTimeout, cfg.FetchTimeout)
				return false
			}
			if cfg.MaxConcurrent != expectedDefaults.MaxConcurrent {
				t.Logf("Expected default MaxConcurrent %d, got %d", expectedDefaults.MaxConcurrent, cfg.MaxConcurrent)
				return false
			}
			if cfg.CacheDir != expectedDefaults.CacheDir {
				t.Logf("Expected default CacheDir %s, got %s", expectedDefaults.CacheDir, cfg.CacheDir)
				return false
			}
			if cfg.MaxSearchResults != expectedDefaults.MaxSearchResults {
				t.Logf("Expected default MaxSearchResults %d, got %d", expectedDefaults.MaxSearchResults, cfg.MaxSearchResults)
				return false
			}

			return true
		},
		gen.Int(),
	))

	properties.TestingRun(t)
}

// TestProperty10ConfigurationValidation tests Property 10: Configuration Validation
// **Validates: Requirements 8.4**
//
// Feature: nats-docs-mcp-server, Property 10: Configuration Validation
//
// Property: For any invalid configuration value (e.g., negative timeout, invalid log level),
// the server should report a validation error during startup and not proceed with invalid configuration.
func TestProperty10ConfigurationValidation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for invalid log levels (anything not in the valid set)
	genInvalidLogLevel := gen.OneConstOf(
		"invalid",
		"INVALID",
		"trace",
		"fatal",
		"",
		"Info",  // case-sensitive, should be lowercase
		"DEBUG", // case-sensitive, should be lowercase
		"warning",
		"critical",
		"123",
		"log",
	)

	// Generator for invalid (non-positive) integers
	genNonPositiveInt := gen.IntRange(-1000, 0)

	// Generator for invalid URLs
	genInvalidURL := gen.OneConstOf(
		"",
		"not-a-url",
		"ftp://invalid-scheme.com",
		"://missing-scheme.com",
		"http://",
		"https://",
		"invalid",
		"www.example.com", // missing scheme
		"htp://typo.com",  // typo in scheme
	)

	properties.Property("invalid log level causes validation error", prop.ForAll(
		func(invalidLogLevel string) bool {
			cfg := NewConfig()
			cfg.LogLevel = invalidLogLevel

			err := cfg.Validate()
			if err == nil {
				t.Logf("Expected validation error for invalid log level: %s", invalidLogLevel)
				return false
			}

			// Error message should mention the invalid log level
			if !strings.Contains(err.Error(), "invalid log level") {
				t.Logf("Expected error message to contain 'invalid log level', got: %v", err)
				return false
			}

			return true
		},
		genInvalidLogLevel,
	))

	properties.Property("negative or zero fetch timeout causes validation error", prop.ForAll(
		func(invalidTimeout int) bool {
			cfg := NewConfig()
			cfg.FetchTimeout = invalidTimeout

			err := cfg.Validate()
			if err == nil {
				t.Logf("Expected validation error for invalid fetch timeout: %d", invalidTimeout)
				return false
			}

			// Error message should mention fetch_timeout
			if !strings.Contains(err.Error(), "fetch_timeout must be positive") {
				t.Logf("Expected error message to contain 'fetch_timeout must be positive', got: %v", err)
				return false
			}

			return true
		},
		genNonPositiveInt,
	))

	properties.Property("negative or zero max concurrent causes validation error", prop.ForAll(
		func(invalidMaxConcurrent int) bool {
			cfg := NewConfig()
			cfg.MaxConcurrent = invalidMaxConcurrent

			err := cfg.Validate()
			if err == nil {
				t.Logf("Expected validation error for invalid max concurrent: %d", invalidMaxConcurrent)
				return false
			}

			// Error message should mention max_concurrent
			if !strings.Contains(err.Error(), "max_concurrent must be positive") {
				t.Logf("Expected error message to contain 'max_concurrent must be positive', got: %v", err)
				return false
			}

			return true
		},
		genNonPositiveInt,
	))

	properties.Property("negative or zero max search results causes validation error", prop.ForAll(
		func(invalidMaxResults int) bool {
			cfg := NewConfig()
			cfg.MaxSearchResults = invalidMaxResults

			err := cfg.Validate()
			if err == nil {
				t.Logf("Expected validation error for invalid max search results: %d", invalidMaxResults)
				return false
			}

			// Error message should mention max_search_results
			if !strings.Contains(err.Error(), "max_search_results must be positive") {
				t.Logf("Expected error message to contain 'max_search_results must be positive', got: %v", err)
				return false
			}

			return true
		},
		genNonPositiveInt,
	))

	properties.Property("invalid docs base URL causes validation error", prop.ForAll(
		func(invalidURL string) bool {
			cfg := NewConfig()
			cfg.DocsBaseURL = invalidURL

			err := cfg.Validate()
			if err == nil {
				t.Logf("Expected validation error for invalid docs base URL: %s", invalidURL)
				return false
			}

			// Error message should mention docs_base_url or URL-related issue
			errMsg := err.Error()
			if !strings.Contains(errMsg, "docs_base_url") {
				t.Logf("Expected error message to contain 'docs_base_url', got: %v", err)
				return false
			}

			return true
		},
		genInvalidURL,
	))

	properties.Property("multiple invalid values cause validation error with all issues", prop.ForAll(
		func(invalidLogLevel string, invalidTimeout, invalidMaxConcurrent int) bool {
			cfg := NewConfig()
			cfg.LogLevel = invalidLogLevel
			cfg.FetchTimeout = invalidTimeout
			cfg.MaxConcurrent = invalidMaxConcurrent

			err := cfg.Validate()
			if err == nil {
				t.Logf("Expected validation error for multiple invalid values")
				return false
			}

			// Error message should mention multiple issues
			errMsg := err.Error()

			// Should mention log level issue
			if !strings.Contains(errMsg, "invalid log level") {
				t.Logf("Expected error message to contain 'invalid log level', got: %v", err)
				return false
			}

			// Should mention timeout issue
			if !strings.Contains(errMsg, "fetch_timeout must be positive") {
				t.Logf("Expected error message to contain 'fetch_timeout must be positive', got: %v", err)
				return false
			}

			// Should mention max concurrent issue
			if !strings.Contains(errMsg, "max_concurrent must be positive") {
				t.Logf("Expected error message to contain 'max_concurrent must be positive', got: %v", err)
				return false
			}

			return true
		},
		genInvalidLogLevel,
		genNonPositiveInt,
		genNonPositiveInt,
	))

	properties.Property("Load rejects invalid configuration from environment", prop.ForAll(
		func(invalidLogLevel string) bool {
			// Clean up environment
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("DOCS_BASE_URL")
			os.Unsetenv("FETCH_TIMEOUT")
			os.Unsetenv("MAX_CONCURRENT")
			os.Unsetenv("CACHE_DIR")
			os.Unsetenv("MAX_SEARCH_RESULTS")

			// Set invalid log level
			os.Setenv("LOG_LEVEL", invalidLogLevel)
			defer os.Unsetenv("LOG_LEVEL")

			_, err := Load()

			// Empty string is not set by loadFromEnv (it checks val != "")
			// so it will use the default value which is valid
			if invalidLogLevel == "" {
				// Should succeed with default value
				if err != nil {
					t.Logf("Expected Load to succeed with empty env var (uses default), got error: %v", err)
					return false
				}
				return true
			}

			// Non-empty invalid values should cause validation error
			if err == nil {
				t.Logf("Expected Load to reject invalid log level from environment: %s", invalidLogLevel)
				return false
			}

			// Error should mention validation failure
			if !strings.Contains(err.Error(), "invalid log level") {
				t.Logf("Expected error to mention invalid log level, got: %v", err)
				return false
			}

			return true
		},
		genInvalidLogLevel,
	))

	properties.Property("LoadFromFile rejects invalid configuration from file", prop.ForAll(
		func(invalidTimeout int) bool {
			// Create a config file with invalid timeout
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")
			configContent := fmt.Sprintf(`
log_level: info
docs_base_url: https://docs.nats.io
fetch_timeout: %d
`, invalidTimeout)

			if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
				t.Logf("Failed to create config file: %v", err)
				return false
			}

			_, err := LoadFromFile(configFile)
			if err == nil {
				t.Logf("Expected LoadFromFile to reject invalid timeout from file: %d", invalidTimeout)
				return false
			}

			// Error should mention validation failure
			if !strings.Contains(err.Error(), "fetch_timeout must be positive") {
				t.Logf("Expected error to mention fetch_timeout validation, got: %v", err)
				return false
			}

			return true
		},
		genNonPositiveInt,
	))

	properties.Property("LoadWithFlags rejects invalid configuration from flags", prop.ForAll(
		func(invalidMaxConcurrent int) bool {
			flags := map[string]interface{}{
				"log_level":      "info",
				"docs_base_url":  "https://docs.nats.io",
				"max_concurrent": invalidMaxConcurrent,
			}

			_, err := LoadWithFlags("", flags)
			if err == nil {
				t.Logf("Expected LoadWithFlags to reject invalid max_concurrent from flags: %d", invalidMaxConcurrent)
				return false
			}

			// Error should mention validation failure
			if !strings.Contains(err.Error(), "max_concurrent must be positive") {
				t.Logf("Expected error to mention max_concurrent validation, got: %v", err)
				return false
			}

			return true
		},
		genNonPositiveInt,
	))

	properties.Property("valid configuration passes validation", prop.ForAll(
		func(seed int) bool {
			// Create a valid configuration
			cfg := NewConfig()

			// Validation should pass
			err := cfg.Validate()
			if err != nil {
				t.Logf("Expected valid default config to pass validation, got error: %v", err)
				return false
			}

			return true
		},
		gen.Int(),
	))

	properties.TestingRun(t)
}

// TestProperty11ConfigurationDefaults tests Property 11: Configuration Defaults
// **Validates: Requirements 8.5**
//
// Feature: nats-docs-mcp-server, Property 11: Configuration Defaults
//
// Property: For any optional configuration parameter that is not provided,
// the server should use a documented default value.
func TestProperty11ConfigurationDefaults(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Generator for subsets of configuration parameters to omit
	// We'll test various combinations of missing parameters
	genOmitLogLevel := gen.Bool()
	genOmitDocsBaseURL := gen.Bool()
	genOmitFetchTimeout := gen.Bool()
	genOmitMaxConcurrent := gen.Bool()
	genOmitCacheDir := gen.Bool()
	genOmitMaxSearchResults := gen.Bool()

	// Expected default values (documented in NewConfig)
	expectedDefaults := NewConfig()

	properties.Property("NewConfig always returns documented defaults", prop.ForAll(
		func(seed int) bool {
			cfg := NewConfig()

			// Verify all defaults match documented values
			if cfg.LogLevel != "info" {
				t.Logf("Expected default LogLevel 'info', got '%s'", cfg.LogLevel)
				return false
			}
			if cfg.DocsBaseURL != "https://docs.nats.io" {
				t.Logf("Expected default DocsBaseURL 'https://docs.nats.io', got '%s'", cfg.DocsBaseURL)
				return false
			}
			if cfg.FetchTimeout != 30 {
				t.Logf("Expected default FetchTimeout 30, got %d", cfg.FetchTimeout)
				return false
			}
			if cfg.MaxConcurrent != 5 {
				t.Logf("Expected default MaxConcurrent 5, got %d", cfg.MaxConcurrent)
				return false
			}
			if cfg.CacheDir != "" {
				t.Logf("Expected default CacheDir empty string, got '%s'", cfg.CacheDir)
				return false
			}
			if cfg.MaxSearchResults != 50 {
				t.Logf("Expected default MaxSearchResults 50, got %d", cfg.MaxSearchResults)
				return false
			}

			return true
		},
		gen.Int(),
	))

	properties.Property("Load uses defaults when environment variables are not set", prop.ForAll(
		func(seed int) bool {
			// Clean up all environment variables
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("DOCS_BASE_URL")
			os.Unsetenv("FETCH_TIMEOUT")
			os.Unsetenv("MAX_CONCURRENT")
			os.Unsetenv("CACHE_DIR")
			os.Unsetenv("MAX_SEARCH_RESULTS")

			cfg, err := Load()
			if err != nil {
				t.Logf("Expected Load to succeed with defaults, got error: %v", err)
				return false
			}

			// Verify all values match defaults
			if cfg.LogLevel != expectedDefaults.LogLevel {
				t.Logf("Expected default LogLevel '%s', got '%s'", expectedDefaults.LogLevel, cfg.LogLevel)
				return false
			}
			if cfg.DocsBaseURL != expectedDefaults.DocsBaseURL {
				t.Logf("Expected default DocsBaseURL '%s', got '%s'", expectedDefaults.DocsBaseURL, cfg.DocsBaseURL)
				return false
			}
			if cfg.FetchTimeout != expectedDefaults.FetchTimeout {
				t.Logf("Expected default FetchTimeout %d, got %d", expectedDefaults.FetchTimeout, cfg.FetchTimeout)
				return false
			}
			if cfg.MaxConcurrent != expectedDefaults.MaxConcurrent {
				t.Logf("Expected default MaxConcurrent %d, got %d", expectedDefaults.MaxConcurrent, cfg.MaxConcurrent)
				return false
			}
			if cfg.CacheDir != expectedDefaults.CacheDir {
				t.Logf("Expected default CacheDir '%s', got '%s'", expectedDefaults.CacheDir, cfg.CacheDir)
				return false
			}
			if cfg.MaxSearchResults != expectedDefaults.MaxSearchResults {
				t.Logf("Expected default MaxSearchResults %d, got %d", expectedDefaults.MaxSearchResults, cfg.MaxSearchResults)
				return false
			}

			return true
		},
		gen.Int(),
	))

	properties.Property("LoadFromFile uses defaults for parameters not in config file", prop.ForAll(
		func(omitLogLevel, omitDocsBaseURL, omitFetchTimeout, omitMaxConcurrent, omitCacheDir, omitMaxSearchResults bool) bool {
			// Ensure at least one parameter is omitted to test defaults
			if !omitLogLevel && !omitDocsBaseURL && !omitFetchTimeout && !omitMaxConcurrent && !omitCacheDir && !omitMaxSearchResults {
				// If all are false, force at least one to be omitted
				omitLogLevel = true
			}

			// Clean up environment variables to ensure they don't interfere
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("DOCS_BASE_URL")
			os.Unsetenv("FETCH_TIMEOUT")
			os.Unsetenv("MAX_CONCURRENT")
			os.Unsetenv("CACHE_DIR")
			os.Unsetenv("MAX_SEARCH_RESULTS")

			// Create a config file with only some parameters
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")
			var configLines []string

			if !omitLogLevel {
				configLines = append(configLines, "log_level: debug")
			}
			if !omitDocsBaseURL {
				configLines = append(configLines, "docs_base_url: https://custom.example.com")
			}
			if !omitFetchTimeout {
				configLines = append(configLines, "fetch_timeout: 60")
			}
			if !omitMaxConcurrent {
				configLines = append(configLines, "max_concurrent: 10")
			}
			if !omitCacheDir {
				configLines = append(configLines, "cache_dir: /tmp/custom")
			}
			if !omitMaxSearchResults {
				configLines = append(configLines, "max_search_results: 100")
			}

			configContent := strings.Join(configLines, "\n") + "\n"
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
			if omitLogLevel && cfg.LogLevel != expectedDefaults.LogLevel {
				t.Logf("Expected default LogLevel '%s' when omitted, got '%s'", expectedDefaults.LogLevel, cfg.LogLevel)
				return false
			}
			if omitDocsBaseURL && cfg.DocsBaseURL != expectedDefaults.DocsBaseURL {
				t.Logf("Expected default DocsBaseURL '%s' when omitted, got '%s'", expectedDefaults.DocsBaseURL, cfg.DocsBaseURL)
				return false
			}
			if omitFetchTimeout && cfg.FetchTimeout != expectedDefaults.FetchTimeout {
				t.Logf("Expected default FetchTimeout %d when omitted, got %d", expectedDefaults.FetchTimeout, cfg.FetchTimeout)
				return false
			}
			if omitMaxConcurrent && cfg.MaxConcurrent != expectedDefaults.MaxConcurrent {
				t.Logf("Expected default MaxConcurrent %d when omitted, got %d", expectedDefaults.MaxConcurrent, cfg.MaxConcurrent)
				return false
			}
			if omitCacheDir && cfg.CacheDir != expectedDefaults.CacheDir {
				t.Logf("Expected default CacheDir '%s' when omitted, got '%s'", expectedDefaults.CacheDir, cfg.CacheDir)
				return false
			}
			if omitMaxSearchResults && cfg.MaxSearchResults != expectedDefaults.MaxSearchResults {
				t.Logf("Expected default MaxSearchResults %d when omitted, got %d", expectedDefaults.MaxSearchResults, cfg.MaxSearchResults)
				return false
			}

			// Verify provided parameters are used (not defaults)
			if !omitLogLevel && cfg.LogLevel != "debug" {
				t.Logf("Expected provided LogLevel 'debug', got '%s'", cfg.LogLevel)
				return false
			}
			if !omitDocsBaseURL && cfg.DocsBaseURL != "https://custom.example.com" {
				t.Logf("Expected provided DocsBaseURL 'https://custom.example.com', got '%s'", cfg.DocsBaseURL)
				return false
			}
			if !omitFetchTimeout && cfg.FetchTimeout != 60 {
				t.Logf("Expected provided FetchTimeout 60, got %d", cfg.FetchTimeout)
				return false
			}
			if !omitMaxConcurrent && cfg.MaxConcurrent != 10 {
				t.Logf("Expected provided MaxConcurrent 10, got %d", cfg.MaxConcurrent)
				return false
			}
			if !omitCacheDir && cfg.CacheDir != "/tmp/custom" {
				t.Logf("Expected provided CacheDir '/tmp/custom', got '%s'", cfg.CacheDir)
				return false
			}
			if !omitMaxSearchResults && cfg.MaxSearchResults != 100 {
				t.Logf("Expected provided MaxSearchResults 100, got %d", cfg.MaxSearchResults)
				return false
			}

			return true
		},
		genOmitLogLevel,
		genOmitDocsBaseURL,
		genOmitFetchTimeout,
		genOmitMaxConcurrent,
		genOmitCacheDir,
		genOmitMaxSearchResults,
	))

	properties.Property("LoadWithFlags uses defaults for parameters not in flags", prop.ForAll(
		func(omitLogLevel, omitDocsBaseURL, omitFetchTimeout, omitMaxConcurrent, omitCacheDir, omitMaxSearchResults bool) bool {
			// Ensure at least one parameter is omitted to test defaults
			if !omitLogLevel && !omitDocsBaseURL && !omitFetchTimeout && !omitMaxConcurrent && !omitCacheDir && !omitMaxSearchResults {
				omitLogLevel = true
			}

			// Clean up environment variables
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("DOCS_BASE_URL")
			os.Unsetenv("FETCH_TIMEOUT")
			os.Unsetenv("MAX_CONCURRENT")
			os.Unsetenv("CACHE_DIR")
			os.Unsetenv("MAX_SEARCH_RESULTS")

			// Create flags with only some parameters
			flags := make(map[string]interface{})

			if !omitLogLevel {
				flags["log_level"] = "warn"
			}
			if !omitDocsBaseURL {
				flags["docs_base_url"] = "https://flags.example.com"
			}
			if !omitFetchTimeout {
				flags["fetch_timeout"] = 90
			}
			if !omitMaxConcurrent {
				flags["max_concurrent"] = 15
			}
			if !omitCacheDir {
				flags["cache_dir"] = "/tmp/flags"
			}
			if !omitMaxSearchResults {
				flags["max_search_results"] = 200
			}

			cfg, err := LoadWithFlags("", flags)
			if err != nil {
				t.Logf("Failed to load config with flags: %v", err)
				return false
			}

			// Verify omitted parameters use defaults
			if omitLogLevel && cfg.LogLevel != expectedDefaults.LogLevel {
				t.Logf("Expected default LogLevel '%s' when omitted from flags, got '%s'", expectedDefaults.LogLevel, cfg.LogLevel)
				return false
			}
			if omitDocsBaseURL && cfg.DocsBaseURL != expectedDefaults.DocsBaseURL {
				t.Logf("Expected default DocsBaseURL '%s' when omitted from flags, got '%s'", expectedDefaults.DocsBaseURL, cfg.DocsBaseURL)
				return false
			}
			if omitFetchTimeout && cfg.FetchTimeout != expectedDefaults.FetchTimeout {
				t.Logf("Expected default FetchTimeout %d when omitted from flags, got %d", expectedDefaults.FetchTimeout, cfg.FetchTimeout)
				return false
			}
			if omitMaxConcurrent && cfg.MaxConcurrent != expectedDefaults.MaxConcurrent {
				t.Logf("Expected default MaxConcurrent %d when omitted from flags, got %d", expectedDefaults.MaxConcurrent, cfg.MaxConcurrent)
				return false
			}
			if omitCacheDir && cfg.CacheDir != expectedDefaults.CacheDir {
				t.Logf("Expected default CacheDir '%s' when omitted from flags, got '%s'", expectedDefaults.CacheDir, cfg.CacheDir)
				return false
			}
			if omitMaxSearchResults && cfg.MaxSearchResults != expectedDefaults.MaxSearchResults {
				t.Logf("Expected default MaxSearchResults %d when omitted from flags, got %d", expectedDefaults.MaxSearchResults, cfg.MaxSearchResults)
				return false
			}

			// Verify provided parameters are used (not defaults)
			if !omitLogLevel && cfg.LogLevel != "warn" {
				t.Logf("Expected provided LogLevel 'warn', got '%s'", cfg.LogLevel)
				return false
			}
			if !omitDocsBaseURL && cfg.DocsBaseURL != "https://flags.example.com" {
				t.Logf("Expected provided DocsBaseURL 'https://flags.example.com', got '%s'", cfg.DocsBaseURL)
				return false
			}
			if !omitFetchTimeout && cfg.FetchTimeout != 90 {
				t.Logf("Expected provided FetchTimeout 90, got %d", cfg.FetchTimeout)
				return false
			}
			if !omitMaxConcurrent && cfg.MaxConcurrent != 15 {
				t.Logf("Expected provided MaxConcurrent 15, got %d", cfg.MaxConcurrent)
				return false
			}
			if !omitCacheDir && cfg.CacheDir != "/tmp/flags" {
				t.Logf("Expected provided CacheDir '/tmp/flags', got '%s'", cfg.CacheDir)
				return false
			}
			if !omitMaxSearchResults && cfg.MaxSearchResults != 200 {
				t.Logf("Expected provided MaxSearchResults 200, got %d", cfg.MaxSearchResults)
				return false
			}

			return true
		},
		genOmitLogLevel,
		genOmitDocsBaseURL,
		genOmitFetchTimeout,
		genOmitMaxConcurrent,
		genOmitCacheDir,
		genOmitMaxSearchResults,
	))

	properties.Property("empty config file uses all defaults", prop.ForAll(
		func(seed int) bool {
			// Clean up environment variables
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("DOCS_BASE_URL")
			os.Unsetenv("FETCH_TIMEOUT")
			os.Unsetenv("MAX_CONCURRENT")
			os.Unsetenv("CACHE_DIR")
			os.Unsetenv("MAX_SEARCH_RESULTS")

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

			// All values should be defaults
			if cfg.LogLevel != expectedDefaults.LogLevel {
				t.Logf("Expected default LogLevel '%s' for empty config, got '%s'", expectedDefaults.LogLevel, cfg.LogLevel)
				return false
			}
			if cfg.DocsBaseURL != expectedDefaults.DocsBaseURL {
				t.Logf("Expected default DocsBaseURL '%s' for empty config, got '%s'", expectedDefaults.DocsBaseURL, cfg.DocsBaseURL)
				return false
			}
			if cfg.FetchTimeout != expectedDefaults.FetchTimeout {
				t.Logf("Expected default FetchTimeout %d for empty config, got %d", expectedDefaults.FetchTimeout, cfg.FetchTimeout)
				return false
			}
			if cfg.MaxConcurrent != expectedDefaults.MaxConcurrent {
				t.Logf("Expected default MaxConcurrent %d for empty config, got %d", expectedDefaults.MaxConcurrent, cfg.MaxConcurrent)
				return false
			}
			if cfg.CacheDir != expectedDefaults.CacheDir {
				t.Logf("Expected default CacheDir '%s' for empty config, got '%s'", expectedDefaults.CacheDir, cfg.CacheDir)
				return false
			}
			if cfg.MaxSearchResults != expectedDefaults.MaxSearchResults {
				t.Logf("Expected default MaxSearchResults %d for empty config, got %d", expectedDefaults.MaxSearchResults, cfg.MaxSearchResults)
				return false
			}

			return true
		},
		gen.Int(),
	))

	properties.Property("empty flags map uses all defaults", prop.ForAll(
		func(seed int) bool {
			// Clean up environment variables
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("DOCS_BASE_URL")
			os.Unsetenv("FETCH_TIMEOUT")
			os.Unsetenv("MAX_CONCURRENT")
			os.Unsetenv("CACHE_DIR")
			os.Unsetenv("MAX_SEARCH_RESULTS")

			// Empty flags map
			flags := make(map[string]interface{})

			cfg, err := LoadWithFlags("", flags)
			if err != nil {
				t.Logf("Failed to load config with empty flags: %v", err)
				return false
			}

			// All values should be defaults
			if cfg.LogLevel != expectedDefaults.LogLevel {
				t.Logf("Expected default LogLevel '%s' for empty flags, got '%s'", expectedDefaults.LogLevel, cfg.LogLevel)
				return false
			}
			if cfg.DocsBaseURL != expectedDefaults.DocsBaseURL {
				t.Logf("Expected default DocsBaseURL '%s' for empty flags, got '%s'", expectedDefaults.DocsBaseURL, cfg.DocsBaseURL)
				return false
			}
			if cfg.FetchTimeout != expectedDefaults.FetchTimeout {
				t.Logf("Expected default FetchTimeout %d for empty flags, got %d", expectedDefaults.FetchTimeout, cfg.FetchTimeout)
				return false
			}
			if cfg.MaxConcurrent != expectedDefaults.MaxConcurrent {
				t.Logf("Expected default MaxConcurrent %d for empty flags, got %d", expectedDefaults.MaxConcurrent, cfg.MaxConcurrent)
				return false
			}
			if cfg.CacheDir != expectedDefaults.CacheDir {
				t.Logf("Expected default CacheDir '%s' for empty flags, got '%s'", expectedDefaults.CacheDir, cfg.CacheDir)
				return false
			}
			if cfg.MaxSearchResults != expectedDefaults.MaxSearchResults {
				t.Logf("Expected default MaxSearchResults %d for empty flags, got %d", expectedDefaults.MaxSearchResults, cfg.MaxSearchResults)
				return false
			}

			return true
		},
		gen.Int(),
	))

	properties.TestingRun(t)
}
