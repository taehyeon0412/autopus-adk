package security

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateHookConfig(t *testing.T) {
	t.Parallel()

	config := GenerateHookConfig("/tmp/policy.json")

	hooks, ok := config["hooks"].(map[string]any)
	require.True(t, ok, "hooks key should be a map")

	preToolUse, ok := hooks["PreToolUse"]
	require.True(t, ok, "PreToolUse key should exist")

	entries, ok := preToolUse.([]hookEntry)
	require.True(t, ok, "PreToolUse should be []hookEntry")
	require.Len(t, entries, 1)

	assert.Equal(t, "Bash|Write|Edit", entries[0].Matcher)
	require.Len(t, entries[0].Hooks, 1)
	assert.Contains(t, entries[0].Hooks[0], "/tmp/policy.json")
	assert.Contains(t, entries[0].Hooks[0], "auto worker validate")
}

func TestWriteHookConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	err := WriteHookConfig(dir, "/tmp/policy.json")
	require.NoError(t, err)

	settingsPath := filepath.Join(dir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	require.NoError(t, err)

	var settings map[string]any
	require.NoError(t, json.Unmarshal(data, &settings))

	_, ok := settings["hooks"]
	assert.True(t, ok, "settings should contain hooks key")
}

func TestWriteHookConfigMergesExisting(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	settingsDir := filepath.Join(dir, ".claude")
	require.NoError(t, os.MkdirAll(settingsDir, 0755))

	// Write existing settings.
	existing := map[string]any{
		"theme":   "dark",
		"verbose": true,
	}
	data, err := json.MarshalIndent(existing, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(settingsDir, "settings.json"), data, 0644))

	// Write hook config — should merge.
	err = WriteHookConfig(dir, "/tmp/policy.json")
	require.NoError(t, err)

	merged, err := os.ReadFile(filepath.Join(settingsDir, "settings.json"))
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(merged, &result))

	assert.Equal(t, "dark", result["theme"], "existing keys should be preserved")
	assert.Equal(t, true, result["verbose"], "existing keys should be preserved")
	_, ok := result["hooks"]
	assert.True(t, ok, "hooks should be added")
}

func TestRemoveHookConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	// First write hooks.
	require.NoError(t, WriteHookConfig(dir, "/tmp/policy.json"))

	// Then remove.
	err := RemoveHookConfig(dir)
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	require.NoError(t, err)

	var settings map[string]any
	require.NoError(t, json.Unmarshal(data, &settings))

	_, ok := settings["hooks"]
	assert.False(t, ok, "hooks key should be removed")
}

func TestRemoveHookConfigPreservesOtherKeys(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	settingsDir := filepath.Join(dir, ".claude")
	require.NoError(t, os.MkdirAll(settingsDir, 0755))

	// Write settings with hooks and other keys.
	settings := map[string]any{
		"theme": "dark",
		"hooks": map[string]any{"PreToolUse": []any{}},
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(settingsDir, "settings.json"), data, 0644))

	err = RemoveHookConfig(dir)
	require.NoError(t, err)

	result, err := os.ReadFile(filepath.Join(settingsDir, "settings.json"))
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(result, &got))

	assert.Equal(t, "dark", got["theme"])
	_, hasHooks := got["hooks"]
	assert.False(t, hasHooks)
}

func TestWriteHookConfigInvalidExistingJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	settingsDir := filepath.Join(dir, ".claude")
	require.NoError(t, os.MkdirAll(settingsDir, 0755))

	// Write invalid JSON.
	require.NoError(t, os.WriteFile(filepath.Join(settingsDir, "settings.json"), []byte("{bad"), 0644))

	err := WriteHookConfig(dir, "/tmp/policy.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse existing settings.json")
}

func TestWriteHookConfigDirCreationFailure(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// Block .claude dir creation by placing a file there.
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".claude"), []byte("block"), 0600))

	err := WriteHookConfig(dir, "/tmp/policy.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create .claude directory")
}

func TestRemoveHookConfigInvalidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	settingsDir := filepath.Join(dir, ".claude")
	require.NoError(t, os.MkdirAll(settingsDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(settingsDir, "settings.json"), []byte("{bad"), 0644))

	err := RemoveHookConfig(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parse settings.json")
}

func TestRemoveHookConfigMissingFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	err := RemoveHookConfig(dir)
	assert.NoError(t, err, "removing hooks from non-existent file should not error")
}
