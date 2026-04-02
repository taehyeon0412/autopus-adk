package mcpserver

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

const defaultTTL = 30 * time.Second

// ResourceDescriptor describes an MCP resource for the resources/list response.
type ResourceDescriptor struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

// cacheEntry stores a cached resource value with TTL.
type cacheEntry struct {
	data      any
	fetchedAt time.Time
	ttl       time.Duration
}

func (e *cacheEntry) isExpired() bool {
	return time.Since(e.fetchedAt) > e.ttl
}

// resourceFetcher fetches a resource from the backend.
type resourceFetcher func(ctx context.Context, uri string) (any, error)

// ResourceRegistry manages MCP resources with TTL-based caching.
type ResourceRegistry struct {
	backendURL string
	authToken  string
	cache      map[string]*cacheEntry
	mu         sync.RWMutex
	fetchers   map[string]resourceFetcher
	descriptors []ResourceDescriptor
}

// NewResourceRegistry creates a registry with the 4 standard resources.
func NewResourceRegistry(backendURL, authToken string) *ResourceRegistry {
	r := &ResourceRegistry{
		backendURL: backendURL,
		authToken:  authToken,
		cache:      make(map[string]*cacheEntry),
		fetchers:   make(map[string]resourceFetcher),
	}
	r.registerDefaults()
	return r
}

// Get retrieves a resource by URI, using cache when valid.
func (r *ResourceRegistry) Get(ctx context.Context, uri string) (any, error) {
	// Check cache first.
	r.mu.RLock()
	entry, cached := r.cache[uri]
	r.mu.RUnlock()

	if cached && !entry.isExpired() {
		return entry.data, nil
	}

	// Find the matching fetcher.
	fetcher := r.matchFetcher(uri)
	if fetcher == nil {
		return nil, fmt.Errorf("unknown resource: %s", uri)
	}

	data, err := fetcher(ctx, uri)
	if err != nil {
		// On error, serve stale cache if available.
		if cached {
			return entry.data, nil
		}
		return nil, err
	}

	// Update cache.
	r.mu.Lock()
	r.cache[uri] = &cacheEntry{data: data, fetchedAt: time.Now(), ttl: defaultTTL}
	r.mu.Unlock()

	return data, nil
}

// ListResources returns descriptors for all registered resources.
func (r *ResourceRegistry) ListResources() []ResourceDescriptor {
	return r.descriptors
}

func (r *ResourceRegistry) registerDefaults() {
	r.descriptors = []ResourceDescriptor{
		{URI: "autopus://status", Name: "System Status", Description: "Current system status", MimeType: "application/json"},
		{URI: "autopus://workspaces", Name: "Workspaces", Description: "Available workspaces", MimeType: "application/json"},
		{URI: "autopus://agents", Name: "Agents", Description: "Available agents", MimeType: "application/json"},
		{URI: "autopus://executions/", Name: "Execution", Description: "Execution details by ID", MimeType: "application/json"},
	}

	// Static URI fetchers.
	r.fetchers["autopus://status"] = r.fetchFromAPI("/api/v1/status")
	r.fetchers["autopus://workspaces"] = r.fetchFromAPI("/api/v1/workspaces")
	r.fetchers["autopus://agents"] = r.fetchFromAPI("/api/v1/agents")
	// Dynamic URI prefix — matched separately.
	r.fetchers["autopus://executions/"] = r.fetchExecution
}

// matchFetcher finds the appropriate fetcher for a URI.
func (r *ResourceRegistry) matchFetcher(uri string) resourceFetcher {
	// Exact match first.
	if f, ok := r.fetchers[uri]; ok {
		return f
	}
	// Prefix match for dynamic resources.
	if strings.HasPrefix(uri, "autopus://executions/") {
		return r.fetchers["autopus://executions/"]
	}
	return nil
}

// fetchFromAPI returns a fetcher that GETs the given API path.
func (r *ResourceRegistry) fetchFromAPI(path string) resourceFetcher {
	return func(ctx context.Context, _ string) (any, error) {
		s := &MCPServer{backendURL: r.backendURL, authToken: r.authToken}
		return s.doGet(ctx, path)
	}
}

// fetchExecution extracts the ID from the URI and fetches the execution.
func (r *ResourceRegistry) fetchExecution(ctx context.Context, uri string) (any, error) {
	id := strings.TrimPrefix(uri, "autopus://executions/")
	if id == "" {
		return nil, fmt.Errorf("execution id is required")
	}
	s := &MCPServer{backendURL: r.backendURL, authToken: r.authToken}
	return s.doGet(ctx, fmt.Sprintf("/api/v1/executions/%s", id))
}
