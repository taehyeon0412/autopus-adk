package docs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Cache stores documentation entries on disk as JSON files with TTL support.
type Cache struct {
	dir string
	ttl time.Duration
}

// NewCache creates a Cache that persists entries to dir with the given TTL.
func NewCache(dir string, ttl time.Duration) *Cache {
	return &Cache{dir: dir, ttl: ttl}
}

// sanitizeKey converts a cache key like "cobra:commands" to a safe filename stem.
func sanitizeKey(key string) string {
	return strings.NewReplacer(":", "_", "/", "_", "\\", "_").Replace(key)
}

// filePath returns the absolute path for the given cache key.
func (c *Cache) filePath(key string) string {
	return filepath.Join(c.dir, sanitizeKey(key)+".json")
}

// Set writes entry to disk under the given key, setting CachedAt to now.
func (c *Cache) Set(key string, entry *CacheEntry) error {
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return err
	}
	entry.CachedAt = time.Now()
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return os.WriteFile(c.filePath(key), data, 0o644)
}

// Get reads the cached entry for key. Returns nil if not found or expired.
func (c *Cache) Get(key string) (*CacheEntry, error) {
	data, err := os.ReadFile(c.filePath(key))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	if time.Since(entry.CachedAt) > c.ttl {
		return nil, nil
	}
	return &entry, nil
}

// Clear removes all .json cache files in the cache directory.
func (c *Cache) Clear() error {
	matches, err := filepath.Glob(filepath.Join(c.dir, "*.json"))
	if err != nil {
		return err
	}
	for _, f := range matches {
		if err := os.Remove(f); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// List returns metadata for all cached entries including their expiry time.
func (c *Cache) List() ([]ListEntry, error) {
	matches, err := filepath.Glob(filepath.Join(c.dir, "*.json"))
	if err != nil {
		return nil, err
	}
	entries := make([]ListEntry, 0, len(matches))
	for _, f := range matches {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var entry CacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			continue
		}
		base := filepath.Base(f)
		stem := strings.TrimSuffix(base, ".json")
		// Reverse sanitization: restore ":" from "_" — best-effort for display
		key := strings.Replace(stem, "_", ":", 1)
		entries = append(entries, ListEntry{
			Key:       key,
			ExpiresAt: entry.CachedAt.Add(c.ttl),
		})
	}
	return entries, nil
}
