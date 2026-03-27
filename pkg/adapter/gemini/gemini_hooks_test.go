package gemini

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInjectOrchestraAfterAgentHook_NewSettings(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := NewWithRoot(dir)

	err := a.InjectOrchestraAfterAgentHook("/path/to/collect.sh")
	require.NoError(t, err)

	data, readErr := os.ReadFile(filepath.Join(dir, ".gemini", "settings.json"))
	require.NoError(t, readErr)

	var settings map[string]any
	require.NoError(t, json.Unmarshal(data, &settings))

	hooksMap, ok := settings["hooks"].(map[string]any)
	require.True(t, ok)
	afterAgent, ok := hooksMap["AfterAgent"].([]any)
	require.True(t, ok)
	require.Len(t, afterAgent, 1)

	entry, ok := afterAgent[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "/path/to/collect.sh", entry["command"])
}

func TestInjectOrchestraAfterAgentHook_PreservesExistingHooks(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := NewWithRoot(dir)

	// Write existing settings with a user hook.
	settingsDir := filepath.Join(dir, ".gemini")
	require.NoError(t, os.MkdirAll(settingsDir, 0755))

	existing := map[string]any{
		"hooks": map[string]any{
			"AfterAgent": []any{
				map[string]any{"command": "user-hook.sh"},
			},
			"BeforeAgent": []any{
				map[string]any{"command": "pre-hook.sh"},
			},
		},
		"theme": "dark",
	}
	data, _ := json.Marshal(existing)
	require.NoError(t, os.WriteFile(filepath.Join(settingsDir, "settings.json"), data, 0644))

	err := a.InjectOrchestraAfterAgentHook("/path/to/collect.sh")
	require.NoError(t, err)

	updated, _ := os.ReadFile(filepath.Join(settingsDir, "settings.json"))
	var settings map[string]any
	require.NoError(t, json.Unmarshal(updated, &settings))

	// Existing user fields preserved.
	assert.Equal(t, "dark", settings["theme"])

	hooksMap := settings["hooks"].(map[string]any)
	// BeforeAgent preserved.
	beforeAgent := hooksMap["BeforeAgent"].([]any)
	assert.Len(t, beforeAgent, 1)

	// AfterAgent has both user hook and autopus hook.
	afterAgent := hooksMap["AfterAgent"].([]any)
	assert.Len(t, afterAgent, 2)
}

func TestInjectOrchestraAfterAgentHook_InvalidJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := NewWithRoot(dir)

	// Write invalid JSON — should reset to empty settings.
	settingsDir := filepath.Join(dir, ".gemini")
	require.NoError(t, os.MkdirAll(settingsDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(settingsDir, "settings.json"),
		[]byte("{broken"),
		0644,
	))

	err := a.InjectOrchestraAfterAgentHook("/path/to/collect.sh")
	require.NoError(t, err)

	// Should have written valid settings with the hook.
	data, _ := os.ReadFile(filepath.Join(settingsDir, "settings.json"))
	var settings map[string]any
	require.NoError(t, json.Unmarshal(data, &settings))

	hooksMap := settings["hooks"].(map[string]any)
	afterAgent := hooksMap["AfterAgent"].([]any)
	assert.Len(t, afterAgent, 1)
}

func TestInjectOrchestraAfterAgentHook_NoExistingFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := NewWithRoot(dir)

	// .gemini directory does not exist yet.
	err := a.InjectOrchestraAfterAgentHook("/script.sh")
	require.NoError(t, err)

	// Directory and file should be created.
	info, statErr := os.Stat(filepath.Join(dir, ".gemini", "settings.json"))
	require.NoError(t, statErr)
	assert.False(t, info.IsDir())
}
