package adapter

import (
	"context"
	"os/exec"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubAdapter is a minimal ProviderAdapter for testing.
type stubAdapter struct {
	name string
}

func (s *stubAdapter) Name() string { return s.name }

func (s *stubAdapter) BuildCommand(_ context.Context, _ TaskConfig) *exec.Cmd {
	return nil
}

func (s *stubAdapter) ParseEvent(_ []byte) (StreamEvent, error) {
	return StreamEvent{}, nil
}

func (s *stubAdapter) ExtractResult(_ StreamEvent) TaskResult {
	return TaskResult{}
}

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubAdapter{name: "claude"})

	a, err := r.Get("claude")
	require.NoError(t, err)
	assert.Equal(t, "claude", a.Name())
}

func TestRegistryGetNotFound(t *testing.T) {
	r := NewRegistry()
	_, err := r.Get("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "adapter not found: nonexistent")
}

func TestRegistryList(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubAdapter{name: "gemini"})
	r.Register(&stubAdapter{name: "claude"})
	r.Register(&stubAdapter{name: "codex"})

	names := r.List()
	assert.Equal(t, []string{"claude", "codex", "gemini"}, names)
}

func TestRegistryListEmpty(t *testing.T) {
	r := NewRegistry()
	assert.Empty(t, r.List())
}

func TestRegistryOverwrite(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubAdapter{name: "claude"})
	r.Register(&stubAdapter{name: "claude"})

	names := r.List()
	assert.Equal(t, []string{"claude"}, names)
}

func TestRegistryConcurrentAccess(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup

	// Concurrent writes
	for range 10 {
		wg.Go(func() {
			r.Register(&stubAdapter{name: "adapter"})
		})
	}

	// Concurrent reads
	for range 10 {
		wg.Go(func() {
			r.List()
			_, _ = r.Get("adapter")
		})
	}

	wg.Wait()
	assert.Len(t, r.List(), 1)
}
