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

	"github.com/j4ng5y/nats-docs-mcp-server/internal/config"
	"github.com/j4ng5y/nats-docs-mcp-server/internal/fetcher"
	"github.com/j4ng5y/nats-docs-mcp-server/internal/index"
	"github.com/j4ng5y/nats-docs-mcp-server/internal/parser"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rs/zerolog"
)

// Server represents the MCP server instance with all its dependencies.
// It coordinates the MCP protocol handling, documentation indexing, and tool execution.
type Server struct {
	config      *config.Config
	index       *index.DocumentationIndex
	logger      *slog.Logger
	mcpServer   *server.MCPServer
	fetcher     *fetcher.DocumentationFetcher
	initialized bool
}

// NewServer creates a new MCP server instance with the provided configuration and logger.
// The server is not started until Start() is called.
//
// Parameters:
//   - cfg: Configuration for the server
//   - logger: Structured logger for logging
//
// Returns a configured Server instance ready to be started.
func NewServer(cfg *config.Config, logger *slog.Logger) (*Server, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// Create MCP server instance
	mcpServer := server.NewMCPServer(
		"nats-docs-mcp-server",
		"1.0.0",
	)

	// Create documentation index
	docIndex := index.NewDocumentationIndex()

	// Create HTTP client for fetching
	httpClient := fetcher.NewHTTPClient(
		time.Duration(cfg.FetchTimeout)*time.Second, // Convert seconds to Duration
		5, // Max retries
		cfg.MaxConcurrent,
	)

	// Create zerolog logger for fetcher (use os.Stderr for structured logging)
	zerologLogger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	// Create documentation fetcher
	docFetcher := fetcher.NewDocumentationFetcher(
		httpClient,
		cfg.DocsBaseURL,
		zerologLogger,
	)

	return &Server{
		config:      cfg,
		index:       docIndex,
		logger:      logger,
		mcpServer:   mcpServer,
		fetcher:     docFetcher,
		initialized: false,
	}, nil
}

// Initialize performs server initialization including documentation fetching and indexing.
// This should be called before Start() to ensure the server is ready to handle requests.
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

	// Fetch all documentation pages
	s.logger.Info("Fetching documentation from", "base_url", s.config.DocsBaseURL)
	pages, err := s.fetcher.FetchAllPages(ctx)
	if err != nil {
		s.logger.Error("Failed to fetch documentation", "error", err)
		return fmt.Errorf("failed to fetch documentation: %w", err)
	}

	s.logger.Info("Fetched documentation pages", "count", len(pages))

	// Parse and index each page
	s.logger.Info("Parsing and indexing documentation")
	successCount := 0
	failCount := 0

	for _, page := range pages {
		// Parse HTML content
		doc, err := parser.ParseHTML(strings.NewReader(string(page.Content)))
		if err != nil {
			s.logger.Warn("Failed to parse page", "path", page.Path, "error", err)
			failCount++
			continue
		}

		// Convert parser.Document to index.Document
		indexDoc := &index.Document{
			ID:       page.Path,
			Title:    doc.Title,
			URL:      s.config.DocsBaseURL + page.Path,
			Content:  extractContent(doc),
			Sections: convertSections(doc.Sections),
		}

		// Index the document
		if err := s.index.Index(indexDoc); err != nil {
			s.logger.Warn("Failed to index document", "path", page.Path, "error", err)
			failCount++
			continue
		}

		successCount++
	}

	s.logger.Info("Documentation indexing complete",
		"successful", successCount,
		"failed", failCount,
		"total", len(pages))

	if successCount == 0 {
		return fmt.Errorf("failed to index any documentation pages")
	}

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

	s.logger.Info("Starting MCP server")

	// Start the MCP server with stdio transport
	if err := server.ServeStdio(s.mcpServer); err != nil {
		s.logger.Error("MCP server error", "error", err)
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
	s.logger.Info("Shutting down server")

	// The mcp-go server handles shutdown via signal handling in ServeStdio
	// No additional cleanup needed for now

	s.logger.Info("Server shutdown complete")
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
func (s *Server) handleSearchTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract query parameter (required)
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError("query parameter is required and must be a non-empty string"), nil
	}

	// Extract limit parameter (optional, default to 10)
	limit := request.GetInt("limit", 10)

	// Perform search
	results, err := s.index.Search(query, limit)
	if err != nil {
		s.logger.Error("Search failed", "query", query, "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
	}

	// Format results
	var content strings.Builder
	content.WriteString(fmt.Sprintf("Found %d results for query: %s\n\n", len(results), query))

	for i, result := range results {
		content.WriteString(fmt.Sprintf("%d. %s\n", i+1, result.Title))
		content.WriteString(fmt.Sprintf("   URL: %s\n", result.URL))
		content.WriteString(fmt.Sprintf("   Relevance: %.2f\n", result.Relevance))
		content.WriteString(fmt.Sprintf("   Summary: %s\n\n", result.Summary))
	}

	s.logger.Info("Search completed", "query", query, "results", len(results))

	return mcp.NewToolResultText(content.String()), nil
}

// handleRetrieveTool handles the retrieve_nats_doc tool invocation
func (s *Server) handleRetrieveTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract doc_id parameter (required)
	docID, err := request.RequireString("doc_id")
	if err != nil {
		return mcp.NewToolResultError("doc_id parameter is required and must be a non-empty string"), nil
	}

	// Retrieve document
	doc, err := s.index.Get(docID)
	if err != nil {
		s.logger.Warn("Document not found", "doc_id", docID, "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("document not found: %s", docID)), nil
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
