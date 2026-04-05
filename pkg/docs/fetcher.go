package docs

import (
	"fmt"
)

// Fetcher orchestrates documentation retrieval with a fallback chain:
// Context7 → Scraper → Cache.
type Fetcher struct {
	c7      DocFetcher
	scraper DocFetcher
	cache   FetcherCache
}

// NewFetcher creates a Fetcher with the given Context7 client, scraper, and cache.
func NewFetcher(c7 DocFetcher, scraper DocFetcher, cache FetcherCache) *Fetcher {
	return &Fetcher{c7: c7, scraper: scraper, cache: cache}
}

// Fetch retrieves documentation for a library and topic using the fallback chain.
// Order: Context7 → Scraper → Cache. The first successful source is cached and returned.
func (f *Fetcher) Fetch(library, topic string) (*DocResult, error) {
	key := cacheKey(library, topic)

	// Try Context7 first
	if result, err := f.c7.Fetch(library, topic); err == nil && result != nil {
		_ = f.cache.Set(key, result)
		return result, nil
	}

	// Fall back to Scraper
	if result, err := f.scraper.Fetch(library, topic); err == nil && result != nil {
		_ = f.cache.Set(key, result)
		return result, nil
	}

	// Fall back to Cache
	if result, err := f.cache.Get(key); err == nil && result != nil {
		result.Source = "cache"
		return result, nil
	}

	return nil, fmt.Errorf("docs: all sources failed for %q/%q", library, topic)
}

// FetchMultiple retrieves documentation for multiple libraries with adaptive token budgeting.
// Each result is trimmed to the per-library budget (budget * 4 chars ≈ tokens).
func (f *Fetcher) FetchMultiple(libraries []string, topic string) ([]*DocResult, error) {
	budget := CalculateTokenBudget(len(libraries))
	maxChars := budget * 4

	results := make([]*DocResult, 0, len(libraries))
	for _, lib := range libraries {
		result, err := f.Fetch(lib, topic)
		if err != nil {
			continue
		}
		if len(result.Content) > maxChars {
			result.Content = result.Content[:maxChars]
			result.Tokens = budget
		}
		results = append(results, result)
	}
	return results, nil
}

// cacheKey builds the cache key for a library/topic pair.
func cacheKey(library, topic string) string {
	return library + ":" + topic
}
