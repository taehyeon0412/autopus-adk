// Package pipeline provides pipeline state management types and persistence.
package pipeline

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const checkpointFilename = ".autopus-checkpoint.yaml"

// Save writes the checkpoint as YAML to {dir}/.autopus-checkpoint.yaml.
func (c *Checkpoint) Save(dir string) error {
	data, err := c.MarshalYAML()
	if err != nil {
		return fmt.Errorf("checkpoint: marshal: %w", err)
	}
	path := filepath.Join(dir, checkpointFilename)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("checkpoint: write %s: %w", path, err)
	}
	return nil
}

// Load reads the checkpoint file from {dir}/.autopus-checkpoint.yaml.
// Returns an error if the file does not exist.
func Load(dir string) (*Checkpoint, error) {
	path := filepath.Join(dir, checkpointFilename)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("checkpoint: file not found: %s", path)
		}
		return nil, fmt.Errorf("checkpoint: read %s: %w", path, err)
	}

	var cp Checkpoint
	if err := yaml.Unmarshal(data, &cp); err != nil {
		return nil, fmt.Errorf("checkpoint: unmarshal: %w", err)
	}
	return &cp, nil
}

// LoadWithHash loads the checkpoint and sets Stale=true when the saved
// GitCommitHash differs from currentHash.
func LoadWithHash(dir string, currentHash string) (*Checkpoint, error) {
	cp, err := Load(dir)
	if err != nil {
		return nil, err
	}
	if cp.GitCommitHash != currentHash {
		cp.Stale = true
	}
	return cp, nil
}
