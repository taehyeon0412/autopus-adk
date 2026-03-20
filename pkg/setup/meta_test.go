package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeta_SaveAndLoad(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	meta := NewMeta(dir)
	meta.SetFileMeta("index.md", "# Test", []string{}, dir)

	err := SaveMeta(dir, meta)
	require.NoError(t, err)

	loaded, err := LoadMeta(dir)
	require.NoError(t, err)

	assert.Equal(t, meta.AutopusVersion, loaded.AutopusVersion)
	assert.Contains(t, loaded.Files, "index.md")
}

func TestMeta_HasContentChanged(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	meta := NewMeta(dir)
	meta.SetFileMeta("test.md", "original content", nil, dir)

	assert.False(t, meta.HasContentChanged("test.md", "original content"))
	assert.True(t, meta.HasContentChanged("test.md", "modified content"))
	assert.True(t, meta.HasContentChanged("nonexistent.md", "any"))
}

func TestMeta_HasSourceChanged(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// Create a source file
	writeFile(t, dir, "go.mod", "module test\n\ngo 1.23\n")

	meta := NewMeta(dir)
	meta.SetFileMeta("index.md", "# Test", []string{"go.mod"}, dir)

	// No change
	assert.False(t, meta.HasSourceChanged("index.md", dir))

	// Modify source file
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.24\n"), 0644))
	assert.True(t, meta.HasSourceChanged("index.md", dir))
}

func TestMeta_MissingSourceFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	meta := NewMeta(dir)
	meta.Files["test.md"] = FileMeta{
		ContentHash:  "abc",
		SourceHashes: []string{"nonexistent.txt:sha256:def"},
	}

	assert.True(t, meta.HasSourceChanged("test.md", dir))
}

func TestMeta_LoadCorruptedMeta(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, ".meta.yaml"), []byte("generated_at: [invalid"), 0644))

	_, err := LoadMeta(dir)
	assert.Error(t, err)
}

func TestSplitSourceHash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  []string
	}{
		{"go.mod:sha256:abc123", []string{"go.mod", "sha256", "abc123"}},
		{"path/to/file:sha256:def456", []string{"path/to/file", "sha256", "def456"}},
		{"nocolon", []string{"nocolon"}},
	}

	for _, tt := range tests {
		got := splitSourceHash(tt.input)
		assert.Equal(t, tt.want, got)
	}
}
