package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveProvider_PrefersAuthenticatedConfiguredProvider(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	codexDir := filepath.Join(tmpHome, ".codex")
	requireNoError(t, os.MkdirAll(codexDir, 0o755))

	got := resolveProvider([]string{"claude", "codex"})
	assert.Equal(t, "codex", got)
}

func TestResolveProvider_FallsBackToFirstConfiguredWhenNoneAuthenticated(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	got := resolveProvider([]string{"claude", "codex"})
	assert.Equal(t, "claude", got)
}

func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
