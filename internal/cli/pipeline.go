package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/insajin/autopus-adk/pkg/pipeline"
)

// @AX:NOTE [AUTO] @AX:REASON: magic constant for checkpoint storage path
// pipelineStateDir is the directory where per-SPEC checkpoint files are stored.
const pipelineStateDir = ".autopus/pipeline-state"

// specCheckpointPath returns the flat YAML path for a given specID:
// .autopus/pipeline-state/{specID}.yaml
func specCheckpointPath(specID string) string {
	return filepath.Join(pipelineStateDir, specID+".yaml")
}

// LoadCheckpointIfContinue returns a Checkpoint when continueFlag is true,
// loading from .autopus/pipeline-state/{specID}.yaml.
// Returns nil without error when continueFlag is false (fresh start).
func LoadCheckpointIfContinue(specID string, continueFlag bool) (*pipeline.Checkpoint, error) {
	return loadCheckpointIfContinue(specID, continueFlag, os.Stderr)
}

// loadCheckpointIfContinue is the testable implementation that accepts a writer
// for stale-checkpoint warnings.
func loadCheckpointIfContinue(specID string, continueFlag bool, warn io.Writer) (*pipeline.Checkpoint, error) {
	if !continueFlag {
		return nil, nil
	}

	path := specCheckpointPath(specID)

	hash, err := getCurrentGitHash()
	if err != nil {
		// Git not available: fall back to load without hash comparison.
		cp, loadErr := loadFlatCheckpoint(path)
		if loadErr != nil {
			return nil, checkpointNotFoundErr(specID)
		}
		return cp, nil
	}

	cp, err := loadFlatCheckpointWithHash(path, hash)
	if err != nil {
		return nil, checkpointNotFoundErr(specID)
	}

	if cp.Stale {
		fmt.Fprintf(warn,
			"Warning: checkpoint for SPEC-%s is stale (saved hash %s differs from HEAD %s). Proceeding anyway.\n",
			specID, cp.GitCommitHash, hash,
		)
	}

	return cp, nil
}

// loadFlatCheckpoint reads a checkpoint from an explicit file path
// (e.g. .autopus/pipeline-state/SPEC-001.yaml).
func loadFlatCheckpoint(path string) (*pipeline.Checkpoint, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("checkpoint: file not found: %s", path)
		}
		return nil, fmt.Errorf("checkpoint: read %s: %w", path, err)
	}
	var cp pipeline.Checkpoint
	if err := yaml.Unmarshal(data, &cp); err != nil {
		return nil, fmt.Errorf("checkpoint: unmarshal: %w", err)
	}
	return &cp, nil
}

// loadFlatCheckpointWithHash loads a flat checkpoint and sets Stale when hashes differ.
func loadFlatCheckpointWithHash(path, currentHash string) (*pipeline.Checkpoint, error) {
	cp, err := loadFlatCheckpoint(path)
	if err != nil {
		return nil, err
	}
	if cp.GitCommitHash != currentHash {
		cp.Stale = true
	}
	return cp, nil
}

// checkpointNotFoundErr returns the standard error when no checkpoint exists for specID.
func checkpointNotFoundErr(specID string) error {
	return fmt.Errorf(
		"No checkpoint found for SPEC-%s. Start a new pipeline with: auto go %s",
		specID, specID,
	)
}

// getCurrentGitHash returns the current git HEAD commit hash.
func getCurrentGitHash() (string, error) {
	out, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
