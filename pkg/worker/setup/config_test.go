package setup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestSaveAndLoadWorkerConfig_RoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "worker.yaml")

	original := WorkerConfig{
		BackendURL:  "https://api.autopus.co",
		WorkspaceID: "ws-123",
		Providers:   []string{"claude", "codex"},
		WorkDir:     "/tmp/autopus-work",
		A2AURL:      "https://a2a.autopus.co",
		Concurrency: 3,
	}

	data, err := yaml.Marshal(original)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0600))

	loaded, err := LoadWorkerConfigFrom(path)
	require.NoError(t, err)

	assert.Equal(t, original.BackendURL, loaded.BackendURL)
	assert.Equal(t, original.WorkspaceID, loaded.WorkspaceID)
	assert.Equal(t, original.Providers, loaded.Providers)
	assert.Equal(t, original.WorkDir, loaded.WorkDir)
	assert.Equal(t, original.A2AURL, loaded.A2AURL)
	assert.Equal(t, original.Concurrency, loaded.Concurrency)
}

func TestLoadWorkerConfigFrom_MissingFile(t *testing.T) {
	t.Parallel()

	_, err := LoadWorkerConfigFrom("/nonexistent/worker.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read worker config")
}

func TestLoadWorkerConfigFrom_InvalidYAML(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "worker.yaml")
	require.NoError(t, os.WriteFile(path, []byte(":::invalid"), 0600))

	_, err := LoadWorkerConfigFrom(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal worker config")
}

func TestDefaultWorkerConfigPath(t *testing.T) {
	t.Parallel()

	path := DefaultWorkerConfigPath()
	assert.Contains(t, path, "worker.yaml")
	assert.Contains(t, path, ".config")
	assert.Contains(t, path, "autopus")
}

func TestLoadWorkerConfigFrom_EmptyProviders(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "worker.yaml")

	cfg := WorkerConfig{
		BackendURL: "https://api.autopus.co",
	}

	data, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0600))

	loaded, err := LoadWorkerConfigFrom(path)
	require.NoError(t, err)
	assert.Empty(t, loaded.Providers)
}
