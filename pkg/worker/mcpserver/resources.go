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
	Title       string `json:"title,omitempty"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

// ResourceTemplate describes a parameterized MCP resource.
type ResourceTemplate struct {
	URITemplate string `json:"uriTemplate"`
	Name        string `json:"name"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
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
	backendURL  string
	authToken   string
	cache       map[string]*cacheEntry
	mu          sync.RWMutex
	fetchers    map[string]resourceFetcher
	descriptors []ResourceDescriptor
	templates   []ResourceTemplate
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

// ListTemplates returns descriptors for parameterized resources.
func (r *ResourceRegistry) ListTemplates() []ResourceTemplate {
	return r.templates
}

func (r *ResourceRegistry) registerDefaults() {
	r.descriptors = []ResourceDescriptor{
		{URI: "autopus://status", Name: "system_status", Title: "System Status", Description: "Current Autopus backend health and status.", MimeType: "application/json"},
		{URI: "autopus://workspaces", Name: "workspaces", Title: "Workspaces", Description: "Available workspaces visible to this worker.", MimeType: "application/json"},
		{URI: "autopus://agents", Name: "agents", Title: "Agents", Description: "Available Autopus agents.", MimeType: "application/json"},
		{URI: "autopus://executions/", Name: "execution_collection", Title: "Execution Collection", Description: "Execution details namespace; use the execution template for a concrete ID.", MimeType: "application/json"},
	}
	r.templates = []ResourceTemplate{
		{
			URITemplate: "autopus://executions/{id}",
			Name:        "execution_by_id",
			Title:       "Execution By ID",
			Description: "Execution details for a specific execution identifier.",
			MimeType:    "application/json",
		},
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
