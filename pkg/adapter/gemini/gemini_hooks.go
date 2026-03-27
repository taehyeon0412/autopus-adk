package gemini

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// InjectOrchestraAfterAgentHook adds the autopus orchestra result collector
// AfterAgent hook to .gemini/settings.json, preserving existing user hooks.
// This is session-specific and injected separately from harness-managed hooks.
// @AX:WARN [AUTO] appends to AfterAgent slice without dedup — repeated calls create duplicate hook entries
func (a *Adapter) InjectOrchestraAfterAgentHook(scriptPath string) error {
	settingsDir := filepath.Join(a.root, ".gemini")
	if err := os.MkdirAll(settingsDir, 0755); err != nil {
		return fmt.Errorf("create .gemini dir: %w", err)
	}

	settingsPath := filepath.Join(settingsDir, "settings.json")

	var settings map[string]any
	data, err := os.ReadFile(settingsPath)
	if err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			settings = make(map[string]any)
		}
	} else {
		settings = make(map[string]any)
	}

	// Ensure hooks map exists
	hooksMap, _ := settings["hooks"].(map[string]any)
	if hooksMap == nil {
		hooksMap = make(map[string]any)
	}

	entry := map[string]any{
		"command": scriptPath,
	}

	// Append to existing AfterAgent entries (preserve user hooks)
	existing, _ := hooksMap["AfterAgent"].([]any)
	hooksMap["AfterAgent"] = append(existing, entry)
	settings["hooks"] = hooksMap

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	return os.WriteFile(settingsPath, append(out, '\n'), 0644)
}
