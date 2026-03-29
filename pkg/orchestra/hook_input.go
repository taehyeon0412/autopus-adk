package orchestra

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// HookInput represents a prompt payload for file-based IPC input delivery.
type HookInput struct {
	Provider string `json:"provider"`
	Round    int    `json:"round"`
	Prompt   string `json:"prompt"`
}

// atomicWriteJSON writes data as JSON to path using tmp+rename for atomicity.
// File permissions are set to 0o600.
// @AX:NOTE [AUTO] magic constant 0o600 — restrictive perms for IPC files in /tmp; matches session dir 0o700
func atomicWriteJSON(path string, data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	dir := filepath.Dir(path)
	tmp := path + ".tmp"

	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return fmt.Errorf("write tmp file: %w", err)
	}

	if err := os.Chmod(tmp, 0o600); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("chmod tmp file: %w", err)
	}

	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename %s -> %s: %w", tmp, filepath.Base(path), err)
	}

	_ = dir // suppress unused warning
	return nil
}
