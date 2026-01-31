// Package server provides the MCP server core implementation, handling protocol
// communication, tool registration, and request routing.
package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/j4ng5y/nats-docs-mcp-server/internal/classifier"
	"github.com/j4ng5y/nats-docs-mcp-server/internal/config"
	"github.com/j4ng5y/nats-docs-mcp-server/internal/fetcher"
	"github.com/j4ng5y/nats-docs-mcp-server/internal/index"
	"github.com/j4ng5y/nats-docs-mcp-server/internal/parser"
	"github.com/j4ng5y/nats-docs-mcp-server/internal/search"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog"
)

// Server represents the MCP server instance with all its dependencies.
// It coordinates the MCP protocol handling, documentation indexing, and tool execution.
// Supports dual documentation sources (NATS and Syncp) with classification-based routing.
type Server struct {
	config       *config.Config
	indexManager *index.Manager           // Dual-index manager for NATS and Syncp
	orchestrator *search.Orchestrator     // Search orchestrator for multi-source search
	classifier   classifier.Classifier    // Query classifier for routing
	logger       *slog.Logger
	mcpServer    *server.MCPServer
	multiFetcher *fetcher.MultiSourceFetcher // Fetcher for both NATS and Syncp
	transport    TransportStarter
	initialized  bool
}

// NewServer creates a new MCP server instance with the provided configuration and logger.
// The server is not started until Start() is called.
//
// Parameters:
//   - cfg: Configuration for the server
//   - logger: Structured logger for logging
//
// Returns a configured Server instance ready to be started.
// Returns an error if transport creation fails.
func NewServer(cfg *config.Config, logger *slog.Logger) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// Validate transport configuration
	if err := cfg.ValidateTransport(); err != nil {
		return nil, fmt.Errorf("invalid transport configuration: %w", err)
	}

	// Create MCP server instance
	mcpServer := server.NewMCPServer(
		"nats-docs-mcp-server",
		"1.0.0",
	)

	// Create index manager for dual indices (NATS and Syncp)
	indexManager := index.NewManager()

	// Create classifier with configured keywords
	queryClassifier := classifier.NewKeywordClassifier(
		cfg.SyncpKeywords,
		cfg.NATSKeywords,
	)

	// Create search orchestrator
	searchOrchestrator := search.NewOrchestrator(
		indexManager.GetNATSIndex(),
		indexManager.GetSyncpIndex(),
		queryClassifier,
	)

	// Create zerolog logger for fetcher (use os.Stderr for structured logging)
	zerologLogger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	// Create multi-source fetcher for both NATS and Syncp
	natsConfig := fetcher.FetchConfig{
		BaseURL:       cfg.DocsBaseURL,
		MaxRetries:    5,
		FetchTimeout:  time.Duration(cfg.FetchTimeout) * time.Second,
		MaxConcurrent: cfg.MaxConcurrent,
	}

	syncpConfig := fetcher.FetchConfig{
		BaseURL:       cfg.SyncpBaseURL,
		MaxRetries:    5,
		FetchTimeout:  time.Duration(cfg.SyncpFetchTimeout) * time.Second,
		MaxConcurrent: cfg.MaxConcurrent,
	}

	multiFetcher := fetcher.NewMultiSourceFetcher(natsConfig, syncpConfig, zerologLogger)

	// Create transport based on configuration
	transport, err := NewTransport(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	return &Server{
		config:       cfg,
		indexManager: indexManager,
		orchestrator: searchOrchestrator,
		classifier:   queryClassifier,
		logger:       logger,
		mcpServer:    mcpServer,
		multiFetcher: multiFetcher,
		transport:    transport,
		initialized:  false,
	}, nil
}

// Initialize performs server initialization including documentation fetching and indexing.
// This should be called before Start() to ensure the server is ready to handle requests.
// Fetches both NATS and Syncp documentation (if enabled) and indexes them separately.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns an error if initialization fails.
func (s *Server) Initialize(ctx context.Context) error {
	if s.initialized {
		return fmt.Errorf("server already initialized")
	}

	s.logger.Info("Starting server initialization")

	// Fetch NATS documentation (always required)
	s.logger.Info("Fetching NATS documentation", "base_url", s.config.DocsBaseURL)
	natsPages, err := s.multiFetcher.FetchNATS(ctx)
	if err != nil {
		s.logger.Error("Failed to fetch NATS documentation", "error", err)
		return fmt.Errorf("failed to fetch NATS documentation: %w", err)
	}
	s.logger.Info("Fetched NATS documentation pages", "count", len(natsPages))

	// Fetch Syncp documentation (if enabled, graceful degradation if fails)
	var syncpPages []fetcher.DocumentPage
	if s.config.SyncpEnabled {
		s.logger.Info("Fetching Syncp documentation", "base_url", s.config.SyncpBaseURL)
		syncpPages, err = s.multiFetcher.FetchSyncp(ctx)
		if err != nil {
			s.logger.Warn("Failed to fetch Syncp documentation, continuing with NATS only", "error", err)
			syncpPages = []fetcher.DocumentPage{}
		} else {
			s.logger.Info("Fetched Syncp documentation pages", "count", len(syncpPages))
		}
	}

	// Parse and index NATS documentation
	s.logger.Info("Parsing and indexing NATS documentation")
	natsIndexDocs := make([]*index.Document, 0)
	for _, page := range natsPages {
		doc, err := parser.ParseHTML(strings.NewReader(string(page.Content)))
		if err != nil {
			s.logger.Warn("Failed to parse NATS page", "path", page.Path, "error", err)
			continue
		}

		indexDoc := &index.Document{
			ID:       page.Path,
			Title:    doc.Title,
			URL:      s.config.DocsBaseURL + page.Path,
			Content:  extractContent(doc),
			Sections: convertSections(doc.Sections),
		}
		natsIndexDocs = append(natsIndexDocs, indexDoc)
	}

	// Index NATS documents
	if len(natsIndexDocs) == 0 {
		return fmt.Errorf("failed to parse any NATS documentation pages")
	}

	if err := s.indexManager.IndexNATS(natsIndexDocs); err != nil {
		s.logger.Error("Failed to index NATS documentation", "error", err)
		return fmt.Errorf("failed to index NATS documentation: %w", err)
	}
	s.logger.Info("NATS documentation indexed", "count", len(natsIndexDocs))

	// Parse and index Syncp documentation (if available)
	if len(syncpPages) > 0 {
		s.logger.Info("Parsing and indexing Syncp documentation")
		syncpIndexDocs := make([]*index.Document, 0)
		for _, page := range syncpPages {
			doc, err := parser.ParseHTML(strings.NewReader(string(page.Content)))
			if err != nil {
				s.logger.Warn("Failed to parse Syncp page", "path", page.Path, "error", err)
				continue
			}

			indexDoc := &index.Document{
				ID:       page.Path,
				Title:    doc.Title,
				URL:      s.config.SyncpBaseURL + page.Path,
				Content:  extractContent(doc),
				Sections: convertSections(doc.Sections),
			}
			syncpIndexDocs = append(syncpIndexDocs, indexDoc)
		}

		if len(syncpIndexDocs) > 0 {
			if err := s.indexManager.IndexSyncp(syncpIndexDocs); err != nil {
				s.logger.Warn("Failed to index Syncp documentation, continuing with NATS only", "error", err)
			} else {
				s.logger.Info("Syncp documentation indexed", "count", len(syncpIndexDocs))
			}
		}
	}

	// Report index statistics
	stats := s.indexManager.Stats()
	s.logger.Info("Documentation indexing complete",
		"nats_docs", stats.NATSDocCount,
		"syncp_docs", stats.SyncpDocCount,
		"total_docs", stats.TotalDocCount)

	s.initialized = true
	return nil
}

// RegisterTools registers all MCP tools with the server.
// This should be called after Initialize() and before Start().
//
// Returns an error if tool registration fails.
func (s *Server) RegisterTools() error {
	if !s.initialized {
		return fmt.Errorf("server not initialized, call Initialize() first")
	}

	s.logger.Info("Registering MCP tools")

	// Register search_nats_docs tool
	searchTool := mcp.NewTool(
		"search_nats_docs",
		mcp.WithDescription("Search NATS documentation by keywords or topics. Returns relevant documentation sections with summaries."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query (keywords or topic)"),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results (default: 10)"),
		),
	)

	s.mcpServer.AddTool(searchTool, s.handleSearchTool)

	// Register retrieve_nats_doc tool
	retrieveTool := mcp.NewTool(
		"retrieve_nats_doc",
		mcp.WithDescription("Retrieve complete content of a specific NATS documentation page by ID or URL path."),
		mcp.WithString("doc_id",
			mcp.Required(),
			mcp.Description("Document ID or URL path (e.g., 'nats-concepts/overview')"),
		),
	)

	s.mcpServer.AddTool(retrieveTool, s.handleRetrieveTool)

	s.logger.Info("MCP tools registered successfully")
	return nil
}

// Start starts the MCP server and begins listening for client connections.
// This is a blocking call that runs until the context is cancelled or an error occurs.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns an error if the server fails to start or encounters an error during operation.
func (s *Server) Start(ctx context.Context) error {
	if !s.initialized {
		return fmt.Errorf("server not initialized, call Initialize() first")
	}

	s.logger.Info("Starting MCP server", "transport", s.transport.Type())
	if addr := s.config.GetTransportAddress(); addr != "" {
		s.logger.Info("Transport address", "address", addr)
	}

	// Start the transport with the MCP server
	if err := s.transport.Start(ctx, s.mcpServer); err != nil {
		s.logger.Error("MCP server error", "error", err, "transport", s.transport.Type())
		return fmt.Errorf("MCP server error: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server and cleans up resources.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns an error if shutdown fails.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server", "transport", s.transport.Type())

	// Shutdown the transport
	if err := s.transport.Shutdown(ctx); err != nil {
		s.logger.Error("Error during transport shutdown", "error", err, "transport", s.transport.Type())
		return fmt.Errorf("transport shutdown error: %w", err)
	}

	s.logger.Info("Server shutdown complete", "transport", s.transport.Type())
	return nil
}

// extractContent extracts all text content from a parsed document
func extractContent(doc *parser.Document) string {
	var content strings.Builder
	for _, section := range doc.Sections {
		content.WriteString(section.Content)
		content.WriteString("\n")
	}
	return content.String()
}

// convertSections converts parser.Section to index.Section
func convertSections(sections []parser.Section) []index.Section {
	result := make([]index.Section, len(sections))
	for i, s := range sections {
		result[i] = index.Section{
			Heading: s.Heading,
			Content: s.Content,
			Level:   s.Level,
		}
	}
	return result
}

// handleSearchTool handles the search_nats_docs tool invocation
// Uses the Search Orchestrator to route queries to appropriate documentation source(s)
func (s *Server) handleSearchTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract query parameter (required)
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query parameter is required and must be a non-empty string"), nil
	}

	// Extract limit parameter (optional, default to 10)
	limit := request.GetInt("limit", 10)

	// Perform multi-source search using orchestrator
	results, err := s.orchestrator.Search(query, limit)
	if err != nil {
		s.logger.Error("Search failed", "query", query, "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
	}

	// Format results with source information
	var content strings.Builder
	content.WriteString(fmt.Sprintf("Found %d results for query: %s\n\n", len(results), query))

	for i, result := range results {
		content.WriteString(fmt.Sprintf("%d. %s [%s]\n", i+1, result.Title, result.Source))
		content.WriteString(fmt.Sprintf("   URL: %s\n", result.URL))
		content.WriteString(fmt.Sprintf("   Relevance: %.2f\n", result.Score))
		content.WriteString(fmt.Sprintf("   Summary: %s\n\n", result.Snippet))
	}

	s.logger.Info("Search completed", "query", query, "results", len(results))

	return mcp.NewToolResultText(content.String()), nil
}

// handleRetrieveTool handles the retrieve_nats_doc tool invocation
// Retrieves documents from both NATS and Syncp indices
func (s *Server) handleRetrieveTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract doc_id parameter (required)
	docID, err := request.RequireString("doc_id")
	if err != nil {
		return mcp.NewToolResultError("doc_id parameter is required and must be a non-empty string"), nil
	}

	// Try to retrieve from NATS index first
	doc, err := s.indexManager.GetNATSIndex().Get(docID)
	if err != nil {
		// Try Syncp index if NATS fails
		doc, err = s.indexManager.GetSyncpIndex().Get(docID)
		if err != nil {
			s.logger.Warn("Document not found in any index", "doc_id", docID)
			return mcp.NewToolResultError(fmt.Sprintf("document not found: %s", docID)), nil
		}
	}

	// Format document content
	var content strings.Builder
	content.WriteString(fmt.Sprintf("# %s\n\n", doc.Title))
	content.WriteString(fmt.Sprintf("URL: %s\n\n", doc.URL))

	// Add sections
	for _, section := range doc.Sections {
		content.WriteString(fmt.Sprintf("%s %s\n\n", strings.Repeat("#", section.Level+1), section.Heading))
		content.WriteString(fmt.Sprintf("%s\n\n", section.Content))
	}

	s.logger.Info("Document retrieved", "doc_id", docID, "title", doc.Title)

	return mcp.NewToolResultText(content.String()), nil
}
