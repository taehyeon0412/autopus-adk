package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SetupProgress tracks the current step in the setup wizard.
type SetupProgress struct {
	Step      int       `json:"step"`
	Timestamp time.Time `json:"timestamp"`
}

// progressPath returns the path to the progress file.
func progressPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".worker-progress.json")
	}
	return filepath.Join(home, ".config", "autopus", ".worker-progress.json")
}

// LoadProgress reads the setup progress from disk.
// Returns nil without error if the file does not exist.
func LoadProgress() (*SetupProgress, error) {
	path := progressPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read progress: %w", err)
	}

	var p SetupProgress
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("unmarshal progress: %w", err)
	}
	return &p, nil
}

// SaveProgress writes the current step to disk with the current timestamp.
func SaveProgress(step int) error {
	p := SetupProgress{
		Step:      step,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal progress: %w", err)
	}

	path := progressPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write progress: %w", err)
	}
	return nil
}

// IsExpired returns true if the progress is older than 1 hour.
func (p *SetupProgress) IsExpired() bool {
	return time.Since(p.Timestamp) > time.Hour
}

// ClearProgress removes the progress file.
func ClearProgress() error {
	path := progressPath()
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clear progress: %w", err)
	}
	return nil
}
