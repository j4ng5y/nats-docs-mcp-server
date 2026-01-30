// Package fetcher provides HTTP client functionality for fetching documentation
// from docs.nats.io with retry logic, rate limiting, and error handling.
package fetcher

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
)

// HTTPClient provides HTTP client functionality with timeout, retry logic, and rate limiting
type HTTPClient struct {
	client      *http.Client
	maxRetries  int
	rateLimiter *rate.Limiter
}

// NewHTTPClient creates a new HTTP client with the specified timeout, max retries, and max concurrent requests.
// The client implements exponential backoff retry mechanism and rate limiting for concurrent requests.
//
// Parameters:
//   - timeout: HTTP request timeout duration
//   - maxRetries: Maximum number of retry attempts (not including the initial request)
//   - maxConcurrent: Maximum number of concurrent requests allowed
//
// Returns a configured HTTPClient ready for use.
func NewHTTPClient(timeout time.Duration, maxRetries int, maxConcurrent int) *HTTPClient {
	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: timeout,
	}

	// Create rate limiter for concurrent requests
	// The limiter allows maxConcurrent tokens with a burst of maxConcurrent
	rateLimiter := rate.NewLimiter(rate.Limit(maxConcurrent), maxConcurrent)

	return &HTTPClient{
		client:      httpClient,
		maxRetries:  maxRetries,
		rateLimiter: rateLimiter,
	}
}

// Fetch retrieves content from the specified URL with retry logic and rate limiting.
// It implements exponential backoff for retries, starting with 1 second and doubling on each retry,
// with a maximum delay of 60 seconds between retries.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - url: The URL to fetch
//
// Returns the response body as bytes and any error encountered.
// Retries on 5xx errors and network errors, but not on 4xx client errors.
func (c *HTTPClient) Fetch(ctx context.Context, url string) ([]byte, error) {
	// Wait for rate limiter token
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter wait failed: %w", err)
	}

	var lastErr error
	initialDelay := 1 * time.Second
	maxDelay := 60 * time.Second

	// Initial attempt + retries
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// If this is a retry, apply exponential backoff
		if attempt > 0 {
			delay := time.Duration(math.Pow(2, float64(attempt-1))) * initialDelay

			// Cap delay at maxDelay (60 seconds)
			if delay > maxDelay {
				delay = maxDelay
			}

			// Wait for backoff delay or context cancellation
			select {
			case <-time.After(delay):
				// Continue with retry
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		// Create request
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set User-Agent header
		req.Header.Set("User-Agent", "nats-docs-mcp-server/1.0")

		// Execute request
		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			// Retry on network errors
			continue
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		// Check status code
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Success
			return body, nil
		}

		// 4xx errors are client errors - don't retry
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return nil, fmt.Errorf("client error: HTTP %d", resp.StatusCode)
		}

		// 5xx errors are server errors - retry
		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error: HTTP %d", resp.StatusCode)
			continue
		}

		// Other status codes - don't retry
		return nil, fmt.Errorf("unexpected status code: HTTP %d", resp.StatusCode)
	}

	// All retries exhausted
	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// DocumentationFetcher provides functionality for fetching documentation pages
// from a base URL with logging and error handling.
type DocumentationFetcher struct {
	client  *HTTPClient
	baseURL string
	logger  zerolog.Logger
}

// NewDocumentationFetcher creates a new documentation fetcher with the specified HTTP client and base URL.
//
// Parameters:
//   - client: The HTTP client to use for fetching pages
//   - baseURL: The base URL of the documentation site (e.g., "https://docs.nats.io")
//   - logger: The zerolog logger for structured logging
//
// Returns a configured DocumentationFetcher ready for use.
func NewDocumentationFetcher(client *HTTPClient, baseURL string, logger zerolog.Logger) *DocumentationFetcher {
	return &DocumentationFetcher{
		client:  client,
		baseURL: baseURL,
		logger:  logger,
	}
}

// FetchPage fetches a single documentation page by URL path.
// It handles HTTP errors, timeouts, and logs fetch attempts and results.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - path: The URL path to fetch (e.g., "/nats-concepts/overview")
//
// Returns the page content as bytes and any error encountered.
func (df *DocumentationFetcher) FetchPage(ctx context.Context, path string) ([]byte, error) {
	// Construct full URL
	url := df.baseURL + path

	// Log fetch attempt
	df.logger.Debug().
		Str("url", url).
		Str("path", path).
		Msg("Fetching documentation page")

	// Fetch the page using the HTTP client
	content, err := df.client.Fetch(ctx, url)
	if err != nil {
		// Log fetch failure
		df.logger.Error().
			Err(err).
			Str("url", url).
			Str("path", path).
			Msg("Failed to fetch documentation page")
		return nil, fmt.Errorf("failed to fetch page %s: %w", path, err)
	}

	// Log fetch success
	df.logger.Info().
		Str("url", url).
		Str("path", path).
		Int("content_size", len(content)).
		Msg("Successfully fetched documentation page")

	return content, nil
}

// Sitemap XML structures for parsing sitemap-pages.xml
type urlset struct {
	XMLName xml.Name   `xml:"urlset"`
	URLs    []urlEntry `xml:"url"`
}

type urlEntry struct {
	Loc      string `xml:"loc"`
	Priority string `xml:"priority"`
	LastMod  string `xml:"lastmod"`
}

// DocumentPage represents a fetched documentation page
type DocumentPage struct {
	Path    string
	Content []byte
}

// DiscoverPages fetches the sitemap and extracts all documentation page URLs.
// It returns a slice of URL paths relative to the base URL.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns a slice of URL paths and any error encountered.
func (df *DocumentationFetcher) DiscoverPages(ctx context.Context) ([]string, error) {
	// Fetch the sitemap
	sitemapPath := "/sitemap-pages.xml"

	df.logger.Debug().
		Str("path", sitemapPath).
		Msg("Fetching sitemap")

	content, err := df.FetchPage(ctx, sitemapPath)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sitemap: %w", err)
	}

	// Parse the XML
	var sitemap urlset
	if err := xml.Unmarshal(content, &sitemap); err != nil {
		df.logger.Error().
			Err(err).
			Msg("Failed to parse sitemap XML")
		return nil, fmt.Errorf("failed to parse sitemap XML: %w", err)
	}

	// Extract paths from URLs
	var paths []string
	for _, entry := range sitemap.URLs {
		// Parse the URL to extract the path
		parsedURL, err := url.Parse(entry.Loc)
		if err != nil {
			df.logger.Warn().
				Err(err).
				Str("url", entry.Loc).
				Msg("Failed to parse URL from sitemap, skipping")
			continue
		}

		// Get the path (relative to base URL)
		path := parsedURL.Path
		if path == "" {
			path = "/"
		}

		paths = append(paths, path)
	}

	df.logger.Info().
		Int("count", len(paths)).
		Msg("Discovered documentation pages")

	return paths, nil
}

// FetchAllPages discovers all documentation pages and fetches them concurrently.
// It uses the configured rate limiting to control concurrent fetches.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//
// Returns a slice of DocumentPage structs and any error encountered.
// If some pages fail to fetch, it returns the successfully fetched pages along with an error.
func (df *DocumentationFetcher) FetchAllPages(ctx context.Context) ([]DocumentPage, error) {
	// Discover all pages
	paths, err := df.DiscoverPages(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover pages: %w", err)
	}

	df.logger.Info().
		Int("total_pages", len(paths)).
		Msg("Starting concurrent page fetching")

	// Fetch all pages concurrently
	var wg sync.WaitGroup
	var mu sync.Mutex
	var pages []DocumentPage
	var fetchErrors []error

	for _, path := range paths {
		wg.Add(1)

		// Capture path for goroutine
		currentPath := path

		go func() {
			defer wg.Done()

			// Fetch the page
			content, err := df.FetchPage(ctx, currentPath)
			if err != nil {
				mu.Lock()
				fetchErrors = append(fetchErrors, fmt.Errorf("failed to fetch %s: %w", currentPath, err))
				mu.Unlock()
				return
			}

			// Add to results
			mu.Lock()
			pages = append(pages, DocumentPage{
				Path:    currentPath,
				Content: content,
			})
			mu.Unlock()
		}()
	}

	// Wait for all fetches to complete
	wg.Wait()

	df.logger.Info().
		Int("successful", len(pages)).
		Int("failed", len(fetchErrors)).
		Int("total", len(paths)).
		Msg("Completed page fetching")

	// If there were any errors, return them along with successful pages
	if len(fetchErrors) > 0 {
		// Combine all errors into one
		var errMsg strings.Builder
		errMsg.WriteString(fmt.Sprintf("failed to fetch %d pages: ", len(fetchErrors)))
		for i, err := range fetchErrors {
			if i > 0 {
				errMsg.WriteString("; ")
			}
			errMsg.WriteString(err.Error())
		}
		return pages, fmt.Errorf("%s", errMsg.String())
	}

	return pages, nil
}
