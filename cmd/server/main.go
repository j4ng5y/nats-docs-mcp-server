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
	configFile  string
	logLevel    string
	showVersion bool
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
and indexes it in memory for fast search and retrieval.`,
		RunE: runServer,
	}

	// Add flags
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file (optional)")
	rootCmd.Flags().StringVarP(&logLevel, "log-level", "l", "", "Log level (debug, info, warn, error)")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Show version information")

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

	// Load configuration
	var cfg *config.Config
	var err error

	if configFile != "" {
		// Load from config file
		cfg, err = config.LoadFromFile(configFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration from file: %w", err)
		}
	} else {
		// Load from environment variables and defaults
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
	}

	// Override log level from command line flag if provided
	if logLevel != "" {
		cfg.LogLevel = logLevel
	}

	// Create logger
	log, err := logger.NewLogger(cfg.LogLevel, os.Stderr)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}

	log.Info("Starting NATS Documentation MCP Server",
		"version", version,
		"commit", commit,
		"date", date)

	// Create server
	srv, err := server.NewServer(cfg, log)
	if err != nil {
		log.Error("Failed to create server", "error", err)
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
