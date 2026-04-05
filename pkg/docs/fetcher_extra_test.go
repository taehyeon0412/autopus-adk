package docs

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFetcher_FetchMultiple_TokenTrimming verifies that FetchMultiple trims content to budget.
// Given: a fetcher that returns long content
// When: FetchMultiple is called with multiple libraries
// Then: each result's content is trimmed to the per-library budget
func TestFetcher_FetchMultiple_TokenTrimming(t *testing.T) {
	t.Parallel()

	longContent := strings.Repeat("x", 50000) // far exceeds any budget
	c7 := &stubContext7{result: &DocResult{LibraryName: "cobra", Source: "context7", Content: longContent, Tokens: 10000}}
	scraper := &stubScraper{err: errors.New("not used")}
	cache := newInMemoryCache()

	fetcher := NewFetcher(c7, scraper, cache)
	results, err := fetcher.FetchMultiple([]string{"cobra", "viper", "testify"}, "api")

	require.NoError(t, err)
	require.NotEmpty(t, results)

	budget := CalculateTokenBudget(3)
	maxChars := budget * 4
	for _, r := range results {
		assert.LessOrEqual(t, len(r.Content), maxChars,
			"content must be trimmed to budget * 4 chars")
	}
}

// TestFetcher_FetchMultiple_SkipsErrors verifies that failed fetches are skipped silently.
// Given: fetcher that fails for all sources for "bad-lib" but succeeds for "cobra"
// When: FetchMultiple is called with both libraries
// Then: only "cobra" result is returned
func TestFetcher_FetchMultiple_SkipsErrors(t *testing.T) {
	t.Parallel()

	callCount := 0
	c7 := &countingFetcher{
		fetchFn: func(lib, topic string) (*DocResult, error) {
			callCount++
			if lib == "cobra" {
				return &DocResult{LibraryName: "cobra", Source: "context7", Content: "cobra docs", Tokens: 10}, nil
			}
			return nil, errors.New("not found")
		},
	}
	scraper := &stubScraper{err: errors.New("scraper unavailable")}
	cache := newInMemoryCache()

	fetcher := NewFetcher(c7, scraper, cache)
	results, err := fetcher.FetchMultiple([]string{"cobra", "bad-lib"}, "api")

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "cobra", results[0].LibraryName)
}

// TestFetcher_FetchMultiple_Empty verifies that empty input returns empty results.
// Given: fetcher with no libraries
// When: FetchMultiple is called with empty list
// Then: returns empty slice without error
func TestFetcher_FetchMultiple_Empty(t *testing.T) {
	t.Parallel()

	c7 := &stubContext7{err: errors.New("not called")}
	fetcher := NewFetcher(c7, &stubScraper{}, newInMemoryCache())

	results, err := fetcher.FetchMultiple([]string{}, "api")
	require.NoError(t, err)
	assert.Empty(t, results)
}

// TestFetcher_AllSourcesFail verifies that an error is returned when all fallbacks fail.
// Given: all fetcher sources return errors and cache is empty
// When: Fetch is called
// Then: an error is returned
func TestFetcher_AllSourcesFail(t *testing.T) {
	t.Parallel()

	c7 := &stubContext7{err: errors.New("c7 fail")}
	scraper := &stubScraper{err: errors.New("scraper fail")}
	cache := newInMemoryCache()

	fetcher := NewFetcher(c7, scraper, cache)
	_, err := fetcher.Fetch("nonexistent", "api")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "all sources failed")
}

// countingFetcher is a test double with configurable fetch behavior.
type countingFetcher struct {
	fetchFn func(lib, topic string) (*DocResult, error)
}

func (c *countingFetcher) Fetch(lib, topic string) (*DocResult, error) {
	return c.fetchFn(lib, topic)
}
