package fetcher

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

// FetchConfig holds configuration for fetching a documentation source
type FetchConfig struct {
	BaseURL      string        // Base URL for documentation (e.g., "https://docs.nats.io")
	MaxRetries   int           // Maximum number of retry attempts
	FetchTimeout time.Duration // Timeout per HTTP request
	MaxConcurrent int          // Maximum concurrent fetches
}

// GitHubFetchConfig holds configuration for fetching from GitHub repositories
type GitHubFetchConfig struct {
	Token         string        // GitHub Personal Access Token
	Repositories  []GitHubRepo  // List of repositories to fetch from
	MaxRetries    int           // Maximum number of retry attempts
	FetchTimeout  time.Duration // Timeout per HTTP request
	MaxConcurrent int           // Maximum concurrent fetches
}

// FetchResult holds the result of fetching a documentation source
type FetchResult struct {
	Source string         // "NATS" or "Synadia"
	Pages  []DocumentPage // Successfully fetched pages
	Error  error          // Non-fatal error during fetching (partial success ok)
}

// MultiSourceFetcher fetches documentation from multiple sources (NATS, Synadia, and GitHub)
// It uses a shared HTTPClient with consistent retry and rate limiting behavior
type MultiSourceFetcher struct {
	natsConfig    FetchConfig
	syadiaConfig   FetchConfig
	githubConfig  GitHubFetchConfig
	natsFetcher   *DocumentationFetcher
	syadiaFetcher  *DocumentationFetcher
	githubFetcher *GitHubFetcher
	logger        zerolog.Logger
}

// NewMultiSourceFetcher creates a fetcher for multiple documentation sources
func NewMultiSourceFetcher(
	natsConfig FetchConfig,
	syadiaConfig FetchConfig,
	githubConfig GitHubFetchConfig,
	logger zerolog.Logger,
) *MultiSourceFetcher {
	// Create HTTP client with reasonable defaults
	// Use NATS config as primary for client settings
	httpClient := NewHTTPClient(
		natsConfig.FetchTimeout,
		natsConfig.MaxRetries,
		natsConfig.MaxConcurrent,
	)

	natsFetcher := NewDocumentationFetcher(httpClient, natsConfig.BaseURL, logger)
	syadiaFetcher := NewDocumentationFetcher(httpClient, syadiaConfig.BaseURL, logger)

	var githubFetcher *GitHubFetcher
	// Create GitHub fetcher if repositories are configured (token is optional - will use unauthenticated requests)
	// Unauthenticated requests have lower rate limits but should work for reasonable use
	if len(githubConfig.Repositories) > 0 {
		githubFetcher = NewGitHubFetcher(httpClient, githubConfig.Token, githubConfig.Repositories, logger)
	}

	return &MultiSourceFetcher{
		natsConfig:    natsConfig,
		syadiaConfig:   syadiaConfig,
		githubConfig:  githubConfig,
		natsFetcher:   natsFetcher,
		syadiaFetcher:  syadiaFetcher,
		githubFetcher: githubFetcher,
		logger:        logger,
	}
}

// FetchNATS retrieves all NATS documentation pages
// Returns pages and any error encountered. Non-fatal errors are returned with partial results.
func (msf *MultiSourceFetcher) FetchNATS(ctx context.Context) ([]DocumentPage, error) {
	msf.logger.Info().
		Str("source", "NATS").
		Str("base_url", msf.natsConfig.BaseURL).
		Msg("Starting NATS documentation fetch")

	pages, err := msf.natsFetcher.FetchAllPages(ctx)

	if err != nil {
		msf.logger.Error().
			Err(err).
			Str("source", "NATS").
			Int("pages_fetched", len(pages)).
			Msg("Error during NATS documentation fetch (partial results)")
		return pages, err
	}

	msf.logger.Info().
		Str("source", "NATS").
		Int("pages_fetched", len(pages)).
		Msg("Successfully completed NATS documentation fetch")

	return pages, nil
}

// FetchSynadia retrieves all Synadia documentation pages
// Returns pages and any error encountered. Non-fatal errors are returned with partial results.
// If syncp fetching fails completely, this is treated as graceful degradation - the server
// continues with NATS documentation only.
func (msf *MultiSourceFetcher) FetchSynadia(ctx context.Context) ([]DocumentPage, error) {
	msf.logger.Info().
		Str("source", "Synadia").
		Str("base_url", msf.syadiaConfig.BaseURL).
		Msg("Starting Synadia documentation fetch")

	pages, err := msf.syadiaFetcher.FetchAllPages(ctx)

	if err != nil {
		msf.logger.Error().
			Err(err).
			Str("source", "Synadia").
			Int("pages_fetched", len(pages)).
			Msg("Error during Synadia documentation fetch (graceful degradation to NATS-only mode)")
		return pages, err
	}

	msf.logger.Info().
		Str("source", "Synadia").
		Int("pages_fetched", len(pages)).
		Msg("Successfully completed Synadia documentation fetch")

	return pages, nil
}

// FetchBoth retrieves documentation from both NATS and Synadia sources concurrently
// It uses goroutines to fetch both sources in parallel for efficiency.
// Returns results for both sources, with graceful degradation if one source fails.
func (msf *MultiSourceFetcher) FetchBoth(ctx context.Context) (natsResult, syncpResult FetchResult) {
	// Use channels to collect results from concurrent fetches
	natsChan := make(chan FetchResult, 1)
	syncpChan := make(chan FetchResult, 1)

	// Fetch NATS documentation
	go func() {
		pages, err := msf.FetchNATS(ctx)
		natsChan <- FetchResult{
			Source: "NATS",
			Pages:  pages,
			Error:  err,
		}
	}()

	// Fetch Synadia documentation
	go func() {
		pages, err := msf.FetchSynadia(ctx)
		syncpChan <- FetchResult{
			Source: "Synadia",
			Pages:  pages,
			Error:  err,
		}
	}()

	// Wait for both to complete
	natsResult = <-natsChan
	syncpResult = <-syncpChan

	return natsResult, syncpResult
}

// FetchGitHub retrieves all markdown files from configured GitHub repositories
// Returns files and any error encountered. Non-fatal errors are returned with partial results.
func (msf *MultiSourceFetcher) FetchGitHub(ctx context.Context) ([]GitHubFile, error) {
	if msf.githubFetcher == nil {
		return nil, fmt.Errorf("GitHub fetcher not configured")
	}

	msf.logger.Info().
		Msg("Starting GitHub documentation fetch")

	files, err := msf.githubFetcher.FetchAllFiles(ctx)

	if err != nil {
		msf.logger.Error().
			Err(err).
			Int("files_fetched", len(files)).
			Msg("Error during GitHub documentation fetch (graceful degradation)")
		return files, err
	}

	msf.logger.Info().
		Int("files_fetched", len(files)).
		Msg("Successfully completed GitHub documentation fetch")

	return files, nil
}

// FetchAllWithFallback fetches documentation from both sources with fallback behavior
// If Synadia fetching fails completely, it continues with NATS-only results
// This is useful for graceful degradation in production environments
func (msf *MultiSourceFetcher) FetchAllWithFallback(ctx context.Context) (natsPages, syncpPages []DocumentPage, err error) {
	natsResult, syncpResult := msf.FetchBoth(ctx)

	natsPages = natsResult.Pages
	syncpPages = syncpResult.Pages

	// NATS must succeed (primary documentation source)
	if natsResult.Error != nil {
		return natsPages, syncpPages, fmt.Errorf("NATS documentation fetch failed: %w", natsResult.Error)
	}

	// Synadia fetch is optional - log error but don't fail overall
	if syncpResult.Error != nil {
		msf.logger.Warn().
			Err(syncpResult.Error).
			Msg("Synadia documentation fetch failed, continuing with NATS-only mode")
		// Don't return error here - this is graceful degradation
		return natsPages, syncpPages, nil
	}

	return natsPages, syncpPages, nil
}
