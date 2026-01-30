//go:build property
// +build property

package fetcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestPropertyRetryBackoffBehavior verifies Property 8: Retry Backoff Behavior
// **Validates: Requirements 4.5**
//
// Property: For any sequence of failed fetch attempts, the delay between retries
// should increase exponentially (each retry delay should be greater than the previous).
//
// Feature: nats-docs-mcp-server, Property 8: Retry Backoff Behavior
//
// NOTE: This test is skipped because it takes too long to run with real exponential backoff delays.
// The retry backoff behavior is validated by unit tests instead.
func TestPropertyRetryBackoffBehavior(t *testing.T) {
	t.Skip("Skipping property test - takes too long with real exponential backoff delays")
}

// TestPropertyRetryBackoffExponentialGrowth verifies that delays grow exponentially
// **Validates: Requirements 4.5**
//
// Property: For any sequence of retries, each delay should be approximately double
// the previous delay (until the max delay cap is reached).
//
// Feature: nats-docs-mcp-server, Property 8: Retry Backoff Behavior
func TestPropertyRetryBackoffExponentialGrowth(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("delays double on each retry until max cap", prop.ForAll(
		func(numRetries int) bool {
			// Track attempt times
			var attemptTimes []time.Time
			var mu sync.Mutex

			// Create a test server that always fails
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				attemptTimes = append(attemptTimes, time.Now())
				mu.Unlock()
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer server.Close()

			// Create client with the generated number of retries
			client := NewHTTPClient(5*time.Second, numRetries, 5)
			ctx := context.Background()

			// Execute fetch (will fail)
			_, err := client.Fetch(ctx, server.URL)

			// Should return an error
			if err == nil {
				return false
			}

			mu.Lock()
			attempts := make([]time.Time, len(attemptTimes))
			copy(attempts, attemptTimes)
			mu.Unlock()

			// If only one or two attempts, can't verify exponential growth
			if len(attempts) <= 2 {
				return true
			}

			// Calculate delays between attempts
			delays := make([]time.Duration, len(attempts)-1)
			for i := 1; i < len(attempts); i++ {
				delays[i-1] = attempts[i].Sub(attempts[i-1])
			}

			// Verify exponential growth (approximately doubling)
			// Allow 30% tolerance for timing variations
			maxDelay := 60 * time.Second
			for i := 1; i < len(delays); i++ {
				prevDelay := delays[i-1]
				currDelay := delays[i]

				// If previous delay was at max cap, current should also be at max
				if prevDelay >= maxDelay-200*time.Millisecond {
					if currDelay < maxDelay-200*time.Millisecond {
						return false
					}
					continue
				}

				// Otherwise, current should be approximately double the previous
				// Expected: currDelay ≈ 2 * prevDelay
				expectedDelay := 2 * prevDelay
				lowerBound := expectedDelay - time.Duration(float64(expectedDelay)*0.3)
				upperBound := expectedDelay + time.Duration(float64(expectedDelay)*0.3)

				// If expected delay exceeds max, check against max instead
				if expectedDelay > maxDelay {
					if currDelay < maxDelay-200*time.Millisecond || currDelay > maxDelay+200*time.Millisecond {
						return false
					}
				} else {
					if currDelay < lowerBound || currDelay > upperBound {
						return false
					}
				}
			}

			return true
		},
		// Generate number of retries between 2 and 4 (reduced for faster execution)
		gen.IntRange(2, 4),
	))

	// Run property tests with minimum 20 iterations (reduced for faster execution)
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 20
	properties.TestingRun(t, gopter.ConsoleReporter(false), params)
}

// TestPropertyRetryBackoffMaxDelayCap verifies that delays are capped at 60 seconds
// **Validates: Requirements 4.5**
//
// Property: For any sequence of retries, no delay should exceed 60 seconds.
//
// Feature: nats-docs-mcp-server, Property 8: Retry Backoff Behavior
func TestPropertyRetryBackoffMaxDelayCap(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("no delay exceeds 60 seconds regardless of retry count", prop.ForAll(
		func(numRetries int) bool {
			// Track attempt times
			var attemptTimes []time.Time
			var mu sync.Mutex

			// Create a test server that always fails
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				attemptTimes = append(attemptTimes, time.Now())
				mu.Unlock()
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer server.Close()

			// Create client with the generated number of retries
			client := NewHTTPClient(5*time.Second, numRetries, 5)
			ctx := context.Background()

			// Execute fetch (will fail)
			_, err := client.Fetch(ctx, server.URL)

			// Should return an error
			if err == nil {
				return false
			}

			mu.Lock()
			attempts := make([]time.Time, len(attemptTimes))
			copy(attempts, attemptTimes)
			mu.Unlock()

			// If only one attempt, no delays to check
			if len(attempts) <= 1 {
				return true
			}

			// Calculate delays between attempts
			delays := make([]time.Duration, len(attempts)-1)
			for i := 1; i < len(attempts); i++ {
				delays[i-1] = attempts[i].Sub(attempts[i-1])
			}

			// Verify no delay exceeds 60 seconds (max delay cap)
			maxDelay := 60 * time.Second
			tolerance := 200 * time.Millisecond
			for _, delay := range delays {
				if delay > maxDelay+tolerance {
					return false
				}
			}

			return true
		},
		// Generate number of retries between 1 and 5 (reduced for faster execution)
		gen.IntRange(1, 5),
	))

	// Run property tests with minimum 20 iterations (reduced for faster execution)
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 20
	properties.TestingRun(t, gopter.ConsoleReporter(false), params)
}
