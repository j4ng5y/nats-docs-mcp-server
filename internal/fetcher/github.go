// Package fetcher provides HTTP client functionality for fetching documentation
// including support for GitHub repositories via the GitHub REST API.
package fetcher

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

// GitHubRepo represents a GitHub repository configuration
type GitHubRepo struct {
	Owner  string // Repository owner (e.g., "nats-io")
	Name   string // Repository name (e.g., "nats-server")
	Branch string // Branch to fetch from (e.g., "main")
	ShortName string // Short name for document IDs (e.g., "nats-server")
}

// GitHubFile represents a file fetched from GitHub
type GitHubFile struct {
	Path    string // File path in repository (e.g., "docs/README.md")
	Content []byte // File content
	Repo    string // Short repository name
	SHA     string // Git SHA of the file
}

// GitHubTreeEntry represents an entry in the GitHub tree API response
type gitHubTreeEntry struct {
	Path string `json:"path"`
	Type string `json:"type"`
	SHA  string `json:"sha"`
	URL  string `json:"url"`
	Size int    `json:"size,omitempty"`
}

// GitHubTreeResponse represents the GitHub tree API response
type gitHubTreeResponse struct {
	SHA      string             `json:"sha"`
	URL      string             `json:"url"`
	Tree     []gitHubTreeEntry  `json:"tree"`
	Truncated bool              `json:"truncated"`
}

// GitHubFileResponse represents the GitHub contents API response
type gitHubFileResponse struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	SHA     string `json:"sha"`
	Size    int    `json:"size"`
	Type    string `json:"type"`
	Content string `json:"content"` // Base64-encoded
	URL     string `json:"url"`
}

// GitHubFetcher provides functionality for fetching documentation from GitHub repositories
type GitHubFetcher struct {
	client       *HTTPClient
	token        string
	repositories []GitHubRepo
	logger       zerolog.Logger
}

// NewGitHubFetcher creates a new GitHub fetcher with the specified configuration
func NewGitHubFetcher(client *HTTPClient, token string, repos []GitHubRepo, logger zerolog.Logger) *GitHubFetcher {
	return &GitHubFetcher{
		client:       client,
		token:        token,
		repositories: repos,
		logger:       logger,
	}
}

// FetchAllFiles discovers and fetches all markdown files from configured repositories
func (gf *GitHubFetcher) FetchAllFiles(ctx context.Context) ([]GitHubFile, error) {
	gf.logger.Info().Msg("Starting GitHub documentation fetch")

	var allFiles []GitHubFile
	var mu sync.Mutex
	var wg sync.WaitGroup
	var fetchErrors []error
	var rateLimited bool

	for _, repo := range gf.repositories {
		wg.Add(1)

		// Capture repo for goroutine
		currentRepo := repo

		go func() {
			defer wg.Done()

			gf.logger.Info().
				Str("owner", currentRepo.Owner).
				Str("name", currentRepo.Name).
				Str("branch", currentRepo.Branch).
				Msg("Fetching repository")

			// Discover markdown files
			files, err := gf.discoverMarkdownFiles(ctx, currentRepo)
			if err != nil {
				if isRateLimitError(err) {
					mu.Lock()
					rateLimited = true
					mu.Unlock()
				}
				mu.Lock()
				fetchErrors = append(fetchErrors, fmt.Errorf("failed to discover files in %s/%s: %w", currentRepo.Owner, currentRepo.Name, err))
				mu.Unlock()
				return
			}

			gf.logger.Info().
				Str("repo", currentRepo.ShortName).
				Int("files", len(files)).
				Msg("Discovered markdown files")

			// Fetch content for each file
			for _, file := range files {
				content, err := gf.fetchFileContent(ctx, currentRepo, file.Path)
				if err != nil {
					if isRateLimitError(err) {
						mu.Lock()
						rateLimited = true
						mu.Unlock()
					}
					gf.logger.Warn().
						Str("repo", currentRepo.ShortName).
						Str("path", file.Path).
						Err(err).
						Msg("Failed to fetch file content")
					continue
				}

				mu.Lock()
				allFiles = append(allFiles, GitHubFile{
					Path:    file.Path,
					Content: content,
					Repo:    currentRepo.ShortName,
					SHA:     file.SHA,
				})
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	gf.logger.Info().
		Int("total_files", len(allFiles)).
		Int("errors", len(fetchErrors)).
		Bool("rate_limited", rateLimited).
		Msg("Completed GitHub documentation fetch")

	// If there were errors, return them along with successfully fetched files
	if len(fetchErrors) > 0 && len(allFiles) == 0 {
		// All repos failed
		var errMsg strings.Builder
		errMsg.WriteString(fmt.Sprintf("failed to fetch from %d repositories: ", len(fetchErrors)))
		for i, err := range fetchErrors {
			if i > 0 {
				errMsg.WriteString("; ")
			}
			errMsg.WriteString(err.Error())
		}
		// Add rate limit hint if we detected rate limiting
		if rateLimited {
			errMsg.WriteString(". Rate limited by GitHub API. To increase limits, provide a GitHub Personal Access Token via NATS_DOCS_GITHUB_TOKEN")
		}
		return nil, fmt.Errorf("%s", errMsg.String())
	}

	// Return partial results if some repos failed
	if len(fetchErrors) > 0 {
		if rateLimited {
			gf.logger.Warn().
				Int("error_count", len(fetchErrors)).
				Msg("GitHub API rate limit reached. Provide a Personal Access Token via NATS_DOCS_GITHUB_TOKEN to increase limits")
		} else {
			gf.logger.Warn().
				Int("error_count", len(fetchErrors)).
				Msg("Some repositories failed to fetch, continuing with partial results")
		}
	}

	return allFiles, nil
}

// isRateLimitError checks if an error indicates GitHub API rate limiting
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// Check for rate limit indicators
	return strings.Contains(errStr, "403") ||
		strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "API rate limit exceeded")
}

// discoverMarkdownFiles discovers all markdown files in a repository
func (gf *GitHubFetcher) discoverMarkdownFiles(ctx context.Context, repo GitHubRepo) ([]gitHubTreeEntry, error) {
	// Use GitHub Tree API to get all files recursively
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1", repo.Owner, repo.Name, repo.Branch)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("token %s", gf.token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "nats-docs-mcp-server/1.0")

	// Use HTTP client with rate limiting
	content, err := gf.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tree: %w", err)
	}

	// Parse response
	var tree gitHubTreeResponse
	if err := json.Unmarshal(content, &tree); err != nil {
		return nil, fmt.Errorf("failed to parse tree response: %w", err)
	}

	if tree.Truncated {
		gf.logger.Warn().
			Str("repo", repo.ShortName).
			Msg("GitHub tree response was truncated, some files may be missing")
	}

	// Filter for markdown files
	var markdownFiles []gitHubTreeEntry
	for _, entry := range tree.Tree {
		// Only include files (not directories)
		if entry.Type != "blob" {
			continue
		}
		// Only include markdown files
		if !strings.HasSuffix(strings.ToLower(entry.Path), ".md") {
			continue
		}
		// Skip vendor and node_modules directories
		if strings.Contains(entry.Path, "/vendor/") || strings.Contains(entry.Path, "/node_modules/") {
			continue
		}

		markdownFiles = append(markdownFiles, entry)
	}

	return markdownFiles, nil
}

// fetchFileContent fetches the content of a file from GitHub
func (gf *GitHubFetcher) fetchFileContent(ctx context.Context, repo GitHubRepo, path string) ([]byte, error) {
	// Use GitHub Contents API to fetch file
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", repo.Owner, repo.Name, path, repo.Branch)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("token %s", gf.token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "nats-docs-mcp-server/1.0")

	// Use HTTP client with rate limiting
	content, err := gf.client.Fetch(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch file: %w", err)
	}

	// Parse response
	var fileResp gitHubFileResponse
	if err := json.Unmarshal(content, &fileResp); err != nil {
		return nil, fmt.Errorf("failed to parse file response: %w", err)
	}

	// For markdown files, the API returns base64-encoded content
	// However, we need to check if it's actually base64 or raw text
	decodedContent, err := decodeBase64IfNeeded(fileResp.Content)
	if err != nil {
		// If decoding fails, try to use as-is
		decodedContent = []byte(fileResp.Content)
	}

	return decodedContent, nil
}

// decodeBase64IfNeeded attempts to decode base64 content, falls back to raw content
func decodeBase64IfNeeded(content string) ([]byte, error) {
	// GitHub API returns base64-encoded content
	return base64.StdEncoding.DecodeString(content)
}
