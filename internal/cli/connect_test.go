package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/worker/setup"
)

func TestSaveConnectConfig_ClearsLegacyKnowledgeSourceID(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	legacyConfig := []byte("workspace_id: old-workspace\nknowledge_source_id: legacy-source\nbackend_url: https://old.example.com\n")
	require.NoError(t, os.MkdirAll(filepath.Dir(setup.DefaultWorkerConfigPath()), 0o700))
	require.NoError(t, os.WriteFile(setup.DefaultWorkerConfigPath(), legacyConfig, 0o600))

	err := saveConnectConfig("ws-123", "https://api.autopus.co")
	require.NoError(t, err)

	cfg, err := setup.LoadWorkerConfig()
	require.NoError(t, err)
	assert.Equal(t, "ws-123", cfg.WorkspaceID)
	assert.Equal(t, "https://api.autopus.co", cfg.BackendURL)

	data, err := os.ReadFile(setup.DefaultWorkerConfigPath())
	require.NoError(t, err)
	assert.NotContains(t, string(data), "knowledge_source_id")
}
