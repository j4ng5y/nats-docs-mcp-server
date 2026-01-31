// Package config provides configuration management for the NATS Documentation MCP Server.
// It supports loading configuration from multiple sources: command-line flags, config files,
// and environment variables, with proper precedence handling.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/j4ng5y/nats-docs-mcp-server/internal/classifier"
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

	// Transport settings
	TransportType string // Transport type: stdio, sse, streamablehttp (default: stdio)
	Host          string // Host to bind for network transports (default: localhost)
	Port          int    // Port to bind for network transports (default: 0)

	// Syncp documentation settings
	SyncpEnabled      bool     // Enable Syncp documentation support (default: false)
	SyncpBaseURL      string   // Base URL for Syncp documentation (default: https://docs.synadia.com/control-plane)
	SyncpFetchTimeout int      // Timeout for fetching Syncp documentation in seconds (default: 30)
	SyncpKeywords     []string // Keywords that classify queries as Syncp-specific
	NATSKeywords      []string // Keywords that classify queries as NATS-specific
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

		// Transport defaults
		TransportType: "stdio",
		Host:          "localhost",
		Port:          0,

		// Syncp defaults
		SyncpEnabled:      false, // Disabled by default for backward compatibility
		SyncpBaseURL:      "https://docs.synadia.com/control-plane",
		SyncpFetchTimeout: 30,
		SyncpKeywords:     classifier.DefaultSyncpKeywords(),
		NATSKeywords:      classifier.DefaultNATSKeywords(),
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
	// Transport settings
	if v.IsSet("transport_type") {
		cfg.TransportType = v.GetString("transport_type")
	}
	if v.IsSet("host") {
		cfg.Host = v.GetString("host")
	}
	if v.IsSet("port") {
		cfg.Port = v.GetInt("port")
	}

	// Syncp settings
	if v.IsSet("syncp.enabled") {
		cfg.SyncpEnabled = v.GetBool("syncp.enabled")
	}
	if v.IsSet("syncp.base_url") {
		cfg.SyncpBaseURL = v.GetString("syncp.base_url")
	}
	if v.IsSet("syncp.fetch_timeout") {
		cfg.SyncpFetchTimeout = v.GetInt("syncp.fetch_timeout")
	}
	if v.IsSet("classification.syncp_keywords") {
		cfg.SyncpKeywords = v.GetStringSlice("classification.syncp_keywords")
	}
	if v.IsSet("classification.nats_keywords") {
		cfg.NATSKeywords = v.GetStringSlice("classification.nats_keywords")
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
		// Transport settings
		if v.IsSet("transport_type") {
			cfg.TransportType = v.GetString("transport_type")
		}
		if v.IsSet("host") {
			cfg.Host = v.GetString("host")
		}
		if v.IsSet("port") {
			cfg.Port = v.GetInt("port")
		}
		// Syncp settings
		if v.IsSet("syncp.enabled") {
			cfg.SyncpEnabled = v.GetBool("syncp.enabled")
		}
		if v.IsSet("syncp.base_url") {
			cfg.SyncpBaseURL = v.GetString("syncp.base_url")
		}
		if v.IsSet("syncp.fetch_timeout") {
			cfg.SyncpFetchTimeout = v.GetInt("syncp.fetch_timeout")
		}
		if v.IsSet("classification.syncp_keywords") {
			cfg.SyncpKeywords = v.GetStringSlice("classification.syncp_keywords")
		}
		if v.IsSet("classification.nats_keywords") {
			cfg.NATSKeywords = v.GetStringSlice("classification.nats_keywords")
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
	// Transport settings
	if val, ok := flags["transport_type"]; ok && val != nil {
		if strVal, ok := val.(string); ok {
			cfg.TransportType = strVal
		}
	}
	if val, ok := flags["host"]; ok && val != nil {
		if strVal, ok := val.(string); ok {
			cfg.Host = strVal
		}
	}
	if val, ok := flags["port"]; ok && val != nil {
		if intVal, ok := val.(int); ok {
			cfg.Port = intVal
		}
	}
	// Syncp settings
	if val, ok := flags["syncp_enabled"]; ok && val != nil {
		if boolVal, ok := val.(bool); ok {
			cfg.SyncpEnabled = boolVal
		}
	}
	if val, ok := flags["syncp_base_url"]; ok && val != nil {
		if strVal, ok := val.(string); ok {
			cfg.SyncpBaseURL = strVal
		}
	}
	if val, ok := flags["syncp_fetch_timeout"]; ok && val != nil {
		if intVal, ok := val.(int); ok {
			cfg.SyncpFetchTimeout = intVal
		}
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// loadFromEnv loads configuration from environment variables into the provided Config
// loadFromEnv loads configuration from environment variables with NATS_DOCS_ prefix.
// This implements 12-factor app principles (III. Store config in environment).
// Supports both NATS_DOCS_ prefixed and bare names for backward compatibility.
func loadFromEnv(cfg *Config) {
	// Helper function to get env var with prefix, falling back to bare name
	getEnv := func(name string) string {
		// Try with NATS_DOCS_ prefix first (12-factor standard)
		if val := os.Getenv("NATS_DOCS_" + name); val != "" {
			return val
		}
		// Fall back to bare name for backward compatibility
		return os.Getenv(name)
	}

	// Server settings
	if val := getEnv("LOG_LEVEL"); val != "" {
		cfg.LogLevel = val
	}

	// Documentation settings
	if val := getEnv("DOCS_BASE_URL"); val != "" {
		cfg.DocsBaseURL = val
	}
	if val := getEnv("FETCH_TIMEOUT"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			cfg.FetchTimeout = intVal
		}
	}
	if val := getEnv("MAX_CONCURRENT"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			cfg.MaxConcurrent = intVal
		}
	}
	if val := getEnv("CACHE_DIR"); val != "" {
		cfg.CacheDir = val
	}
	if val := getEnv("MAX_SEARCH_RESULTS"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			cfg.MaxSearchResults = intVal
		}
	}

	// Transport settings
	if val := getEnv("TRANSPORT_TYPE"); val != "" {
		cfg.TransportType = val
	}
	if val := getEnv("HOST"); val != "" {
		cfg.Host = val
	}
	if val := getEnv("PORT"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			cfg.Port = intVal
		}
	}

	// Syncp settings
	if val := getEnv("SYNCP_ENABLED"); val != "" {
		cfg.SyncpEnabled = val == "true" || val == "1" || val == "yes"
	}
	if val := getEnv("SYNCP_BASE_URL"); val != "" {
		cfg.SyncpBaseURL = val
	}
	if val := getEnv("SYNCP_FETCH_TIMEOUT"); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			cfg.SyncpFetchTimeout = intVal
		}
	}

	// Classification keywords - comma-separated lists
	if val := getEnv("SYNCP_KEYWORDS"); val != "" {
		cfg.SyncpKeywords = strings.Split(val, ",")
		// Trim whitespace from each keyword
		for i := range cfg.SyncpKeywords {
			cfg.SyncpKeywords[i] = strings.TrimSpace(cfg.SyncpKeywords[i])
		}
	}
	if val := getEnv("NATS_KEYWORDS"); val != "" {
		cfg.NATSKeywords = strings.Split(val, ",")
		// Trim whitespace from each keyword
		for i := range cfg.NATSKeywords {
			cfg.NATSKeywords[i] = strings.TrimSpace(cfg.NATSKeywords[i])
		}
	}
}

// NormalizeEnvKey converts environment variable names to viper keys
// Example: LOG_LEVEL -> log_level
func NormalizeEnvKey(key string) string {
	return strings.ToLower(strings.ReplaceAll(key, "_", "_"))
}

// GetTransportAddress returns the network address for network transports.
// For STDIO transport, it returns an empty string.
// For SSE and StreamableHTTP transports, it returns "host:port".
func (c *Config) GetTransportAddress() string {
	// STDIO transport doesn't use network address
	if c.TransportType == "stdio" {
		return ""
	}

	// Network transports return "host:port"
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetTransportType returns the configured transport type.
// This method is part of the transportConfig interface used by NewTransport.
func (c *Config) GetTransportType() string {
	return c.TransportType
}

// GetPort returns the configured port for network transports.
// This method is part of the transportConfig interface used by NewTransport.
func (c *Config) GetPort() int {
	return c.Port
}

// ValidateTransport validates transport-specific configuration settings.
// It checks that the transport type is valid and that network transports
// have the required port and host configuration.
func (c *Config) ValidateTransport() error {
	var errors []string

	// Validate transport type
	validTransportTypes := map[string]bool{
		"stdio":          true,
		"sse":            true,
		"streamablehttp": true,
	}
	if !validTransportTypes[c.TransportType] {
		errors = append(errors, fmt.Sprintf("invalid transport type: %s (must be one of: stdio, sse, streamablehttp)", c.TransportType))
	}

	// Validate network transport requirements (sse and streamablehttp)
	if c.TransportType == "sse" || c.TransportType == "streamablehttp" {
		// Validate port is in valid range (1-65535)
		if c.Port < 1 || c.Port > 65535 {
			errors = append(errors, fmt.Sprintf("port must be between 1 and 65535 for network transports, got: %d", c.Port))
		}

		// Validate host is non-empty
		if c.Host == "" {
			errors = append(errors, "host cannot be empty for network transports")
		}
	}

	// If there are validation errors, return them all
	if len(errors) > 0 {
		return fmt.Errorf("transport validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
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

	// Validate syncp configuration (only if enabled)
	if c.SyncpEnabled {
		// Validate syncp base URL
		if c.SyncpBaseURL == "" {
			errors = append(errors, "syncp.base_url cannot be empty when syncp is enabled")
		} else {
			// Check if URL has valid scheme (http or https)
			if !strings.HasPrefix(c.SyncpBaseURL, "http://") && !strings.HasPrefix(c.SyncpBaseURL, "https://") {
				errors = append(errors, fmt.Sprintf("syncp.base_url must start with http:// or https://, got: %s", c.SyncpBaseURL))
			}
			// Basic URL validation - check for scheme and host
			if strings.HasPrefix(c.SyncpBaseURL, "http://") && len(c.SyncpBaseURL) <= 7 {
				errors = append(errors, fmt.Sprintf("syncp.base_url is incomplete: %s", c.SyncpBaseURL))
			}
			if strings.HasPrefix(c.SyncpBaseURL, "https://") && len(c.SyncpBaseURL) <= 8 {
				errors = append(errors, fmt.Sprintf("syncp.base_url is incomplete: %s", c.SyncpBaseURL))
			}
		}

		// Validate syncp fetch timeout
		if c.SyncpFetchTimeout <= 0 {
			errors = append(errors, fmt.Sprintf("syncp.fetch_timeout must be positive, got: %d", c.SyncpFetchTimeout))
		}

		// Validate keyword lists are not empty
		if len(c.SyncpKeywords) == 0 {
			errors = append(errors, "classification.syncp_keywords cannot be empty when syncp is enabled")
		}
		if len(c.NATSKeywords) == 0 {
			errors = append(errors, "classification.nats_keywords cannot be empty when syncp is enabled")
		}
	}

	// If there are validation errors, return them all
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}
