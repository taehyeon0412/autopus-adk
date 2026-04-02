package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// KnowledgeSearcher queries the backend Knowledge Hub API.
type KnowledgeSearcher struct {
	backendURL string
	authToken  string
	client     *http.Client
}

// SearchResult represents a single knowledge search result.
type SearchResult struct {
	ID      string  `json:"id"`
	Title   string  `json:"title"`
	Content string  `json:"content"`
	Score   float64 `json:"score"`
}

// NewKnowledgeSearcher creates a new KnowledgeSearcher with a 5-second timeout.
func NewKnowledgeSearcher(backendURL, authToken string) *KnowledgeSearcher {
	return &KnowledgeSearcher{
		backendURL: backendURL,
		authToken:  authToken,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// Search queries the Knowledge Hub API and returns matching results.
func (ks *KnowledgeSearcher) Search(ctx context.Context, query string) ([]SearchResult, error) {
	endpoint := fmt.Sprintf("%s/api/v1/knowledge/search?q=%s",
		ks.backendURL, url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("knowledge search: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+ks.authToken)

	resp, err := ks.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("knowledge search: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("knowledge search: unexpected status %d", resp.StatusCode)
	}

	var results []SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("knowledge search: decode response: %w", err)
	}

	return results, nil
}
