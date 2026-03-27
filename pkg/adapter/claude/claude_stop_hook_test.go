package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInjectStopHook_NewSettings(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := NewWithRoot(dir)

	err := a.InjectStopHook()
	require.NoError(t, err)

	data, readErr := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	require.NoError(t, readErr)

	var settings map[string]any
	require.NoError(t, json.Unmarshal(data, &settings))

	hooksMap, ok := settings["hooks"].(map[string]any)
	require.True(t, ok)
	stopHooks, ok := hooksMap["Stop"].([]any)
	require.True(t, ok)
	require.Len(t, stopHooks, 1)

	entry, ok := stopHooks[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "", entry["matcher"])

	// Verify the hook command path contains the expected path.
	hooks := entry["hooks"].([]any)
	hookEntry := hooks[0].(map[string]any)
	assert.Contains(t, hookEntry["command"], "hook-claude-stop.sh")
	assert.Equal(t, "command", hookEntry["type"])
	assert.Equal(t, float64(10), hookEntry["timeout"])
}

func TestInjectStopHook_PreservesExistingStopHooks(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := NewWithRoot(dir)

	// Write existing settings with a user Stop hook.
	settingsDir := filepath.Join(dir, ".claude")
	require.NoError(t, os.MkdirAll(settingsDir, 0755))
	existing := map[string]any{
		"hooks": map[string]any{
			"Stop": []any{
				map[string]any{"matcher": "", "hooks": []any{
					map[string]any{"type": "command", "command": "user-stop.sh"},
				}},
			},
		},
	}
	data, _ := json.Marshal(existing)
	require.NoError(t, os.WriteFile(filepath.Join(settingsDir, "settings.json"), data, 0644))

	err := a.InjectStopHook()
	require.NoError(t, err)

	updated, _ := os.ReadFile(filepath.Join(settingsDir, "settings.json"))
	var settings map[string]any
	require.NoError(t, json.Unmarshal(updated, &settings))

	hooksMap := settings["hooks"].(map[string]any)
	stopHooks := hooksMap["Stop"].([]any)
	// User hook + autopus hook = 2 entries.
	assert.Len(t, stopHooks, 2)
}

func TestInjectStopHook_InvalidExistingJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := NewWithRoot(dir)

	settingsDir := filepath.Join(dir, ".claude")
	require.NoError(t, os.MkdirAll(settingsDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(settingsDir, "settings.json"),
		[]byte("{bad json}"),
		0644,
	))

	// Should reset to empty and still write the hook.
	err := a.InjectStopHook()
	require.NoError(t, err)

	data, _ := os.ReadFile(filepath.Join(settingsDir, "settings.json"))
	var settings map[string]any
	require.NoError(t, json.Unmarshal(data, &settings))
	hooksMap := settings["hooks"].(map[string]any)
	stopHooks := hooksMap["Stop"].([]any)
	assert.Len(t, stopHooks, 1)
}

func TestInjectOrchestraStopHook_NewSettings(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := NewWithRoot(dir)

	err := a.InjectOrchestraStopHook("/tmp/autopus/session/hook.sh")
	require.NoError(t, err)

	data, readErr := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	require.NoError(t, readErr)

	var settings map[string]any
	require.NoError(t, json.Unmarshal(data, &settings))

	hooksMap := settings["hooks"].(map[string]any)
	stopHooks := hooksMap["Stop"].([]any)
	require.Len(t, stopHooks, 1)

	entry := stopHooks[0].(map[string]any)
	hooks := entry["hooks"].([]any)
	hookEntry := hooks[0].(map[string]any)
	assert.Equal(t, "/tmp/autopus/session/hook.sh", hookEntry["command"])
	assert.Equal(t, "command", hookEntry["type"])
}

func TestInjectOrchestraStopHook_PreservesExistingHooks(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := NewWithRoot(dir)

	settingsDir := filepath.Join(dir, ".claude")
	require.NoError(t, os.MkdirAll(settingsDir, 0755))
	existing := map[string]any{
		"hooks": map[string]any{
			"Stop": []any{
				map[string]any{"matcher": "", "hooks": []any{
					map[string]any{"type": "command", "command": "existing.sh"},
				}},
			},
			"PreToolUse": []any{
				map[string]any{"matcher": "Bash"},
			},
		},
		"theme": "light",
	}
	data, _ := json.Marshal(existing)
	require.NoError(t, os.WriteFile(filepath.Join(settingsDir, "settings.json"), data, 0644))

	err := a.InjectOrchestraStopHook("/new/hook.sh")
	require.NoError(t, err)

	updated, _ := os.ReadFile(filepath.Join(settingsDir, "settings.json"))
	var settings map[string]any
	require.NoError(t, json.Unmarshal(updated, &settings))

	// User theme preserved.
	assert.Equal(t, "light", settings["theme"])

	hooksMap := settings["hooks"].(map[string]any)
	// PreToolUse preserved.
	_, hasPreTool := hooksMap["PreToolUse"]
	assert.True(t, hasPreTool)

	// Stop has 2 entries (existing + new).
	stopHooks := hooksMap["Stop"].([]any)
	assert.Len(t, stopHooks, 2)
}

func TestInjectOrchestraStopHook_InvalidJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	a := NewWithRoot(dir)

	settingsDir := filepath.Join(dir, ".claude")
	require.NoError(t, os.MkdirAll(settingsDir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(settingsDir, "settings.json"),
		[]byte("not json"),
		0644,
	))

	err := a.InjectOrchestraStopHook("/hook.sh")
	require.NoError(t, err)

	data, _ := os.ReadFile(filepath.Join(settingsDir, "settings.json"))
	var settings map[string]any
	require.NoError(t, json.Unmarshal(data, &settings))
	hooksMap := settings["hooks"].(map[string]any)
	stopHooks := hooksMap["Stop"].([]any)
	assert.Len(t, stopHooks, 1)
}
