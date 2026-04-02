package adapter

import (
	"fmt"
	"sort"
	"sync"
)

// Registry provides thread-safe storage and lookup of ProviderAdapters.
type Registry struct {
	mu       sync.RWMutex
	adapters map[string]ProviderAdapter
}

// NewRegistry creates an empty adapter registry.
func NewRegistry() *Registry {
	return &Registry{
		adapters: make(map[string]ProviderAdapter),
	}
}

// Register adds a provider adapter to the registry, keyed by adapter.Name().
func (r *Registry) Register(adapter ProviderAdapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[adapter.Name()] = adapter
}

// Get returns the adapter registered under the given name.
// Returns an error if no adapter is found.
func (r *Registry) Get(name string) (ProviderAdapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.adapters[name]
	if !ok {
		return nil, fmt.Errorf("adapter not found: %s", name)
	}
	return a, nil
}

// List returns the names of all registered adapters in sorted order.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.adapters))
	for name := range r.adapters {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
