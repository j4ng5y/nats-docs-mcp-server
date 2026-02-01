// NATS Documentation MCP Server
//
// This is the main entry point for the NATS Documentation MCP Server.
// It provides LLMs with programmatic access to NATS documentation through
// the Model Context Protocol (MCP).
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/j4ng5y/nats-docs-mcp-server/internal/config"
	"github.com/j4ng5y/nats-docs-mcp-server/internal/logger"
	"github.com/j4ng5y/nats-docs-mcp-server/internal/server"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	configFile    string
	logLevel      string
	showVersion   bool
	transportType string
	hostFlag      string
	portFlag      int
	refreshCache  bool
	cacheMaxAge   int
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "nats-docs-mcp-server",
		Short: "NATS Documentation MCP Server",
		Long: `NATS Documentation MCP Server provides LLMs with programmatic access
to NATS documentation through the Model Context Protocol (MCP).

The server exposes two main tools:
  - search_nats_docs: Search NATS documentation by keywords
  - retrieve_nats_doc: Retrieve specific documentation pages

The server fetches documentation from https://docs.nats.io/ at startup
and indexes it in memory for fast search and retrieval.

CONFIGURATION (12-Factor App):
The server works without any configuration file - it uses sensible defaults
and loads configuration from environment variables. Set NATS_DOCS_* variables
to customize behavior:

  NATS_DOCS_LOG_LEVEL           Log level (debug, info, warn, error)
  NATS_DOCS_DOCS_BASE_URL       NATS documentation URL
  NATS_DOCS_FETCH_TIMEOUT       Fetch timeout in seconds
  NATS_DOCS_MAX_CONCURRENT      Maximum concurrent fetches
  NATS_DOCS_MAX_SEARCH_RESULTS  Maximum search results
  NATS_DOCS_TRANSPORT_TYPE      Transport (stdio, sse, streamablehttp)
  NATS_DOCS_HOST                Host for network transports
  NATS_DOCS_PORT                Port for network transports
  NATS_DOCS_SYNCP_ENABLED       Enable Synadia documentation (true/false)
  NATS_DOCS_SYNCP_BASE_URL      Synadia documentation URL
  NATS_DOCS_SYNCP_FETCH_TIMEOUT Synadia fetch timeout in seconds

Command-line flags override environment variables.
Optionally provide a config file with --config for convenience.`,
		RunE: runServer,
	}

	// Add flags
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file (optional)")
	rootCmd.Flags().StringVarP(&logLevel, "log-level", "l", "", "Log level (debug, info, warn, error)")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Show version information")
	rootCmd.Flags().StringVarP(&transportType, "transport", "t", "", "Transport type (stdio, sse, streamablehttp)")
	rootCmd.Flags().StringVar(&hostFlag, "host", "", "Host for network transports (SSE, StreamableHTTP)")
	rootCmd.Flags().IntVarP(&portFlag, "port", "p", 0, "Port for network transports (SSE, StreamableHTTP)")
	rootCmd.Flags().BoolVar(&refreshCache, "refresh-cache", false, "Force refresh documentation cache on startup")
	rootCmd.Flags().IntVar(&cacheMaxAge, "cache-max-age", 0, "Maximum cache age in days (0=use default)")

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) error {
	// Show version if requested
	if showVersion {
		fmt.Printf("NATS Documentation MCP Server\n")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Commit:  %s\n", commit)
		fmt.Printf("Built:   %s\n", date)
		return nil
	}

	// Load configuration with precedence: flags > config file > environment > defaults
	// This follows 12-factor app principles (III. Store config in environment)
	var cfg *config.Config
	var err error

	if configFile != "" {
		// Load from config file if explicitly provided
		cfg, err = config.LoadFromFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration from file: %w", err)
		}
	} else {
		// Load from environment variables and defaults (no config file needed)
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
	}

	// Override log level from command line flag if provided
	if logLevel != "" {
		cfg.LogLevel = logLevel
	}

	// Override transport settings from command line flags if provided
	if transportType != "" {
		cfg.TransportType = transportType
	}
	if hostFlag != "" {
		cfg.Host = hostFlag
	}
	if portFlag != 0 {
		cfg.Port = portFlag
	}

	// Override cache settings from command line flags if provided
	if refreshCache {
		cfg.RefreshCache = true
	}
	if cacheMaxAge > 0 {
		cfg.CacheMaxAge = cacheMaxAge
	}

	// Validate transport configuration
	if err := cfg.ValidateTransport(); err != nil {
		return fmt.Errorf("invalid transport configuration: %w", err)
	}

	// Create logger
	log, err := logger.NewLogger(cfg.LogLevel, os.Stderr)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	log.Info("Starting NATS Documentation MCP Server",
		"version", version,
		"commit", commit,
		"date", date,
		"transport", cfg.GetTransportType())
	if addr := cfg.GetTransportAddress(); addr != "" {
		log.Info("Transport address", "address", addr)
	}

	// Create server
	srv, err := server.NewServer(cfg, log)
	if err != nil {
		log.Error("Failed to create server", "error", err)
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// If refresh-cache flag is set, just refresh the cache and exit
	if refreshCache {
		log.Info("Refreshing documentation cache...")
		docsCount, err := srv.RefreshCache(ctx)
		if err != nil {
			log.Error("Cache refresh failed", "error", err)
			return fmt.Errorf("cache refresh failed: %w", err)
		}
		log.Info("Cache refresh complete", "documents_cached", docsCount)
		return nil
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server initialization in a goroutine
	errChan := make(chan error, 1)
	go func() {
		// Initialize server (fetch and index documentation)
		log.Info("Initializing server (fetching and indexing documentation)")
		if err := srv.Initialize(ctx); err != nil {
			errChan <- fmt.Errorf("server initialization failed: %w", err)
			return
		}

		// Register MCP tools
		log.Info("Registering MCP tools")
		if err := srv.RegisterTools(); err != nil {
			errChan <- fmt.Errorf("tool registration failed: %w", err)
			return
		}

		log.Info("Server initialized successfully, starting MCP server")

		// Start the MCP server (this blocks until shutdown)
		if err := srv.Start(ctx); err != nil {
			errChan <- fmt.Errorf("server error: %w", err)
			return
		}

		errChan <- nil
	}()

	// Wait for either an error or shutdown signal
	select {
	case err := <-errChan:
		if err != nil {
			log.Error("Server error", "error", err)
			return err
		}
		log.Info("Server stopped normally")
		return nil

	case sig := <-sigChan:
		log.Info("Received shutdown signal", "signal", sig)
		cancel()

		// Graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("Error during shutdown", "error", err)
			return fmt.Errorf("shutdown error: %w", err)
		}

		log.Info("Server shutdown complete")
		return nil
	}
}
