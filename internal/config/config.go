// Package config provides configuration management for the NATS Documentation MCP Server.
// It supports loading configuration from multiple sources: command-line flags, config files,
// and environment variables, with proper precedence handling.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration settings for the NATS Documentation MCP Server.
// It includes server settings, documentation fetching settings, and search settings.
type Config struct {
	// Server settings
	LogLevel string // Log level: debug, info, warn, error (default: info)

	// Documentation settings
	DocsBaseURL   string // Base URL for NATS documentation (default: https://docs.nats.io)
	FetchTimeout  int    // Timeout for fetching documentation in seconds (default: 30)
	MaxConcurrent int    // Maximum concurrent fetches (default: 5)
	CacheDir      string // Directory for caching fetched documentation (default: empty, no caching)

	// Search settings
	MaxSearchResults int // Maximum number of search results to return (default: 50)
}

// NewConfig creates a new Config with default values for all optional parameters.
// This ensures that the server can run with sensible defaults without requiring
// explicit configuration.
func NewConfig() *Config {
	return &Config{
		// Server defaults
		LogLevel: "info",

		// Documentation defaults
		DocsBaseURL:   "https://docs.nats.io",
		FetchTimeout:  30,
		MaxConcurrent: 5,
		CacheDir:      "",

		// Search defaults
		MaxSearchResults: 50,
	}
}

// Load loads configuration from environment variables with defaults.
// Environment variables should be prefixed with the application name.
// Returns a Config with values from environment variables or defaults.
func Load() (*Config, error) {
	// Start with defaults
	cfg := NewConfig()

	// Load environment variables (override defaults)
	loadFromEnv(cfg)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadFromFile loads configuration from a YAML file, with environment variables
// as fallback, and defaults as final fallback.
// The precedence order is: config file > environment variables > defaults.
func LoadFromFile(configPath string) (*Config, error) {
	// Start with defaults
	cfg := NewConfig()

	// Load environment variables (override defaults)
	loadFromEnv(cfg)

	// Load config file (override env vars and defaults)
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Override with config file values (only if they exist in the file)
	if v.IsSet("log_level") {
		cfg.LogLevel = v.GetString("log_level")
	}
	if v.IsSet("docs_base_url") {
		cfg.DocsBaseURL = v.GetString("docs_base_url")
	}
	if v.IsSet("fetch_timeout") {
		cfg.FetchTimeout = v.GetInt("fetch_timeout")
	}
	if v.IsSet("max_concurrent") {
		cfg.MaxConcurrent = v.GetInt("max_concurrent")
	}
	if v.IsSet("cache_dir") {
		cfg.CacheDir = v.GetString("cache_dir")
	}
	if v.IsSet("max_search_results") {
		cfg.MaxSearchResults = v.GetInt("max_search_results")
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadWithFlags loads configuration from command-line flags, config file,
// environment variables, and defaults.
// The precedence order is: flags > config file > environment variables > defaults.
func LoadWithFlags(configPath string, flags map[string]interface{}) (*Config, error) {
	// Start with defaults
	cfg := NewConfig()

	// Load environment variables (override defaults)
	loadFromEnv(cfg)

	// Load config file if provided (override env vars)
	if configPath != "" {
		v := viper.New()
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		// Override with config file values
		if v.IsSet("log_level") {
			cfg.LogLevel = v.GetString("log_level")
		}
		if v.IsSet("docs_base_url") {
			cfg.DocsBaseURL = v.GetString("docs_base_url")
		}
		if v.IsSet("fetch_timeout") {
			cfg.FetchTimeout = v.GetInt("fetch_timeout")
		}
		if v.IsSet("max_concurrent") {
			cfg.MaxConcurrent = v.GetInt("max_concurrent")
		}
		if v.IsSet("cache_dir") {
			cfg.CacheDir = v.GetString("cache_dir")
		}
		if v.IsSet("max_search_results") {
			cfg.MaxSearchResults = v.GetInt("max_search_results")
		}
	}

	// Override with flags (highest precedence)
	if val, ok := flags["log_level"]; ok && val != nil {
		if strVal, ok := val.(string); ok {
			cfg.LogLevel = strVal
		}
	}
	if val, ok := flags["docs_base_url"]; ok && val != nil {
		if strVal, ok := val.(string); ok {
			cfg.DocsBaseURL = strVal
		}
	}
	if val, ok := flags["fetch_timeout"]; ok && val != nil {
		if intVal, ok := val.(int); ok {
			cfg.FetchTimeout = intVal
		}
	}
	if val, ok := flags["max_concurrent"]; ok && val != nil {
		if intVal, ok := val.(int); ok {
			cfg.MaxConcurrent = intVal
		}
	}
	if val, ok := flags["cache_dir"]; ok && val != nil {
		if strVal, ok := val.(string); ok {
			cfg.CacheDir = strVal
		}
	}
	if val, ok := flags["max_search_results"]; ok && val != nil {
		if intVal, ok := val.(int); ok {
			cfg.MaxSearchResults = intVal
		}
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// loadFromEnv loads configuration from environment variables into the provided Config
func loadFromEnv(cfg *Config) {
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		cfg.LogLevel = val
	}
	if val := os.Getenv("DOCS_BASE_URL"); val != "" {
		cfg.DocsBaseURL = val
	}
	if val := os.Getenv("FETCH_TIMEOUT"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			cfg.FetchTimeout = intVal
		}
	}
	if val := os.Getenv("MAX_CONCURRENT"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			cfg.MaxConcurrent = intVal
		}
	}
	if val := os.Getenv("CACHE_DIR"); val != "" {
		cfg.CacheDir = val
	}
	if val := os.Getenv("MAX_SEARCH_RESULTS"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			cfg.MaxSearchResults = intVal
		}
	}
}

// NormalizeEnvKey converts environment variable names to viper keys
// Example: LOG_LEVEL -> log_level
func NormalizeEnvKey(key string) string {
	return strings.ToLower(strings.ReplaceAll(key, "_", "_"))
}

// Validate validates all configuration values and returns descriptive errors
// for any invalid settings. This should be called after loading configuration
// to ensure the server doesn't start with invalid configuration.
func (c *Config) Validate() error {
	var errors []string

	// Validate log level
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.LogLevel] {
		errors = append(errors, fmt.Sprintf("invalid log level: %s (must be one of: debug, info, warn, error)", c.LogLevel))
	}

	// Validate fetch timeout (must be positive)
	if c.FetchTimeout <= 0 {
		errors = append(errors, fmt.Sprintf("fetch_timeout must be positive, got: %d", c.FetchTimeout))
	}

	// Validate max concurrent (must be positive)
	if c.MaxConcurrent <= 0 {
		errors = append(errors, fmt.Sprintf("max_concurrent must be positive, got: %d", c.MaxConcurrent))
	}

	// Validate max search results (must be positive)
	if c.MaxSearchResults <= 0 {
		errors = append(errors, fmt.Sprintf("max_search_results must be positive, got: %d", c.MaxSearchResults))
	}

	// Validate docs base URL
	if c.DocsBaseURL == "" {
		errors = append(errors, "docs_base_url cannot be empty")
	} else {
		// Check if URL has valid scheme (http or https)
		if !strings.HasPrefix(c.DocsBaseURL, "http://") && !strings.HasPrefix(c.DocsBaseURL, "https://") {
			errors = append(errors, fmt.Sprintf("docs_base_url must start with http:// or https://, got: %s", c.DocsBaseURL))
		}
		// Basic URL validation - check for scheme and host
		if strings.HasPrefix(c.DocsBaseURL, "http://") && len(c.DocsBaseURL) <= 7 {
			errors = append(errors, fmt.Sprintf("docs_base_url is incomplete: %s", c.DocsBaseURL))
		}
		if strings.HasPrefix(c.DocsBaseURL, "https://") && len(c.DocsBaseURL) <= 8 {
			errors = append(errors, fmt.Sprintf("docs_base_url is incomplete: %s", c.DocsBaseURL))
		}
	}

	// If there are validation errors, return them all
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}
