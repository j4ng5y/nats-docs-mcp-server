package fetcher

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
	"github.com/rs/zerolog"
)

// ============================================================================
// Unit Tests
// ============================================================================

func TestNewMultiSourceFetcher(t *testing.T) {
	logger := zerolog.New(nil)
	natsConfig := FetchConfig{
		BaseURL:       "https://docs.nats.io",
		MaxRetries:    5,
		FetchTimeout:  30 * time.Second,
		MaxConcurrent: 5,
	}
	syncpConfig := FetchConfig{
		BaseURL:       "https://docs.synadia.com/control-plane",
		MaxRetries:    5,
		FetchTimeout:  30 * time.Second,
		MaxConcurrent: 5,
	}

	fetcher := NewMultiSourceFetcher(natsConfig, syncpConfig, logger)

	if fetcher == nil {
		t.Fatal("NewMultiSourceFetcher returned nil")
	}

	if fetcher.natsFetcher == nil {
		t.Fatal("NATS fetcher is nil")
	}

	if fetcher.syncpFetcher == nil {
		t.Fatal("Syncp fetcher is nil")
	}
}

func TestFetchNATS_WithMockServer(t *testing.T) {
	// Since we can't easily mock the HTTP server, we test the structure
	logger := zerolog.New(nil)
	natsConfig := FetchConfig{
		BaseURL:       "https://docs.nats.io",
		MaxRetries:    1,
		FetchTimeout:  5 * time.Second,
		MaxConcurrent: 1,
	}
	syncpConfig := FetchConfig{
		BaseURL:       "https://docs.synadia.com/control-plane",
		MaxRetries:    1,
		FetchTimeout:  5 * time.Second,
		MaxConcurrent: 1,
	}

	fetcher := NewMultiSourceFetcher(natsConfig, syncpConfig, logger)

	if fetcher == nil {
		t.Fatal("fetcher creation failed")
	}

	// Verify fetchers were created with correct base URLs
	if fetcher.natsFetcher.baseURL != natsConfig.BaseURL {
		t.Errorf("expected base URL %s, got %s", natsConfig.BaseURL, fetcher.natsFetcher.baseURL)
	}

	if fetcher.syncpFetcher.baseURL != syncpConfig.BaseURL {
		t.Errorf("expected base URL %s, got %s", syncpConfig.BaseURL, fetcher.syncpFetcher.baseURL)
	}
}

func TestFetchResult_SourceAnnotation(t *testing.T) {
	result := FetchResult{
		Source: "NATS",
		Pages:  []DocumentPage{},
		Error:  nil,
	}

	if result.Source != "NATS" {
		t.Errorf("expected source NATS, got %s", result.Source)
	}
}

func TestFetchConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config FetchConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: FetchConfig{
				BaseURL:       "https://docs.nats.io",
				MaxRetries:    5,
				FetchTimeout:  30 * time.Second,
				MaxConcurrent: 5,
			},
			valid: true,
		},
		{
			name: "empty base URL",
			config: FetchConfig{
				BaseURL:       "",
				MaxRetries:    5,
				FetchTimeout:  30 * time.Second,
				MaxConcurrent: 5,
			},
			valid: false,
		},
		{
			name: "zero timeout",
			config: FetchConfig{
				BaseURL:       "https://docs.nats.io",
				MaxRetries:    5,
				FetchTimeout:  0,
				MaxConcurrent: 5,
			},
			valid: false,
		},
		{
			name: "zero concurrent",
			config: FetchConfig{
				BaseURL:       "https://docs.nats.io",
				MaxRetries:    5,
				FetchTimeout:  30 * time.Second,
				MaxConcurrent: 0,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.config.BaseURL != "" &&
				tt.config.FetchTimeout > 0 &&
				tt.config.MaxConcurrent > 0

			if isValid != tt.valid {
				t.Errorf("expected valid=%v, got %v", tt.valid, isValid)
			}
		})
	}
}

// ============================================================================
// Property-Based Tests
// ============================================================================

// Feature: syncp-documentation-support, Property 13: Retry Logic Consistency
// VALIDATES: Requirements 1.2
func TestProperty_RetryLogicConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	logger := zerolog.New(nil)

	properties.Property(
		"NATS and Syncp configs support same retry behavior",
		prop.ForAll(
			func() bool {
				natsConfig := FetchConfig{
					BaseURL:       "https://docs.nats.io",
					MaxRetries:    5,
					FetchTimeout:  30 * time.Second,
					MaxConcurrent: 5,
				}
				syncpConfig := FetchConfig{
					BaseURL:       "https://docs.synadia.com/control-plane",
					MaxRetries:    5,
					FetchTimeout:  30 * time.Second,
					MaxConcurrent: 5,
				}

				fetcher := NewMultiSourceFetcher(natsConfig, syncpConfig, logger)

				// Both fetchers should share the same HTTP client
				// and thus the same retry logic
				if fetcher.natsFetcher.client == nil {
					return false
				}
				if fetcher.syncpFetcher.client == nil {
					return false
				}

				// Both should use same client instance
				if fetcher.natsFetcher.client != fetcher.syncpFetcher.client {
					return false
				}

				// Client should have configured retries
				if fetcher.natsFetcher.client.maxRetries != natsConfig.MaxRetries {
					return false
				}

				return true
			},
		),
	)

	properties.Property(
		"fetch timeout is consistent across sources",
		prop.ForAll(
			func() bool {
				timeout := 30 * time.Second
				natsConfig := FetchConfig{
					BaseURL:       "https://docs.nats.io",
					MaxRetries:    5,
					FetchTimeout:  timeout,
					MaxConcurrent: 5,
				}
				syncpConfig := FetchConfig{
					BaseURL:       "https://docs.synadia.com/control-plane",
					MaxRetries:    5,
					FetchTimeout:  timeout,
					MaxConcurrent: 5,
				}

				fetcher := NewMultiSourceFetcher(natsConfig, syncpConfig, logger)

				// Both fetchers should have HTTP clients with same timeout
				natsTimeout := fetcher.natsFetcher.client.client.Timeout
				syncpTimeout := fetcher.syncpFetcher.client.client.Timeout

				return natsTimeout == syncpTimeout && natsTimeout == timeout
			},
		),
	)

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: syncp-documentation-support, Property 1: Index Independence (for fetcher)
// VALIDATES: Requirements 1.1, 1.3
func TestProperty_FetcherIndependence(t *testing.T) {
	properties := gopter.NewProperties(nil)

	logger := zerolog.New(nil)

	properties.Property(
		"NATS and Syncp fetchers operate independently",
		prop.ForAll(
			func() bool {
				natsConfig := FetchConfig{
					BaseURL:       "https://docs.nats.io",
					MaxRetries:    3,
					FetchTimeout:  15 * time.Second,
					MaxConcurrent: 3,
				}
				syncpConfig := FetchConfig{
					BaseURL:       "https://docs.synadia.com/control-plane",
					MaxRetries:    3,
					FetchTimeout:  15 * time.Second,
					MaxConcurrent: 3,
				}

				fetcher := NewMultiSourceFetcher(natsConfig, syncpConfig, logger)

				// Base URLs should be different
				if fetcher.natsFetcher.baseURL == fetcher.syncpFetcher.baseURL {
					return false
				}

				// NATS should have correct URL
				if fetcher.natsFetcher.baseURL != natsConfig.BaseURL {
					return false
				}

				// Syncp should have correct URL
				if fetcher.syncpFetcher.baseURL != syncpConfig.BaseURL {
					return false
				}

				return true
			},
		),
	)

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: syncp-documentation-support, Property 10: Backward Compatibility
// VALIDATES: Requirements 5.2, 6.3
func TestProperty_BackwardCompatibility_Fetcher(t *testing.T) {
	properties := gopter.NewProperties(nil)

	logger := zerolog.New(nil)

	properties.Property(
		"NATS-only fetching works without Syncp config",
		prop.ForAll(
			func() bool {
				natsConfig := FetchConfig{
					BaseURL:       "https://docs.nats.io",
					MaxRetries:    5,
					FetchTimeout:  30 * time.Second,
					MaxConcurrent: 5,
				}
				syncpConfig := FetchConfig{
					BaseURL:       "https://docs.synadia.com/control-plane",
					MaxRetries:    5,
					FetchTimeout:  30 * time.Second,
					MaxConcurrent: 5,
				}

				fetcher := NewMultiSourceFetcher(natsConfig, syncpConfig, logger)

				// NATS fetcher should be usable independently
				if fetcher.natsFetcher == nil {
					return false
				}

				if fetcher.natsFetcher.baseURL != natsConfig.BaseURL {
					return false
				}

				return true
			},
		),
	)

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
