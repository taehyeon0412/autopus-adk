// Package cli_test contains tests for learn record, prune, and summary subcommands.
package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- learn record ---

func TestLearnRecord_RequiredFlags_Missing_ReturnsError(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"missing both", []string{"learn", "record"}},
		{"missing pattern", []string{"learn", "record", "--type", "gate_fail"}},
		{"missing type", []string{"learn", "record", "--pattern", "test pattern"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := setupLearnDir(t)
			chdir(t, dir)

			var out bytes.Buffer
			cmd := newTestRootCmd()
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs(tt.args)
			err := cmd.Execute()
			assert.Error(t, err, "missing required flags should error")
		})
	}
}

func TestLearnRecord_InvalidType_ReturnsError(t *testing.T) {
	dir := setupLearnDir(t)
	chdir(t, dir)

	var out bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{
		"learn", "record",
		"--type", "invalid_type",
		"--pattern", "some pattern",
	})
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown type")
}

func TestLearnRecord_ValidTypes_Success(t *testing.T) {
	validTypes := []string{
		"gate_fail", "coverage_gap", "review_issue",
		"executor_error", "fix_pattern",
	}
	for _, entryType := range validTypes {
		t.Run(entryType, func(t *testing.T) {
			dir := setupLearnDir(t)
			chdir(t, dir)

			var out bytes.Buffer
			cmd := newTestRootCmd()
			cmd.SetOut(&out)
			cmd.SetErr(&out)
			cmd.SetArgs([]string{
				"learn", "record",
				"--type", entryType,
				"--pattern", "test pattern for " + entryType,
			})
			err := cmd.Execute()
			require.NoError(t, err)
			assert.Contains(t, out.String(), "Recorded "+entryType+" entry")
			assert.Contains(t, out.String(), "test pattern for "+entryType)

			// Verify the entry was actually written to the store
			jsonlPath := filepath.Join(dir, ".autopus", "learnings", "pipeline.jsonl")
			data, err := os.ReadFile(jsonlPath)
			require.NoError(t, err)
			assert.NotEmpty(t, data, "pipeline.jsonl should contain the recorded entry")
		})
	}
}

func TestLearnRecord_OptionalFlags_Accepted(t *testing.T) {
	dir := setupLearnDir(t)
	chdir(t, dir)

	var out bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{
		"learn", "record",
		"--type", "gate_fail",
		"--pattern", "lint failure",
		"--phase", "gate",
		"--spec-id", "SPEC-TEST-001",
		"--files", "a.go,b.go",
		"--packages", "pkg/core",
		"--resolution", "fixed lint config",
		"--severity", "high",
	})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Recorded gate_fail entry")
}

// --- learn prune ---

func TestLearnPrune_RequiredFlag_Missing_ReturnsError(t *testing.T) {
	dir := setupLearnDir(t)
	chdir(t, dir)

	var out bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"learn", "prune"})
	err := cmd.Execute()
	assert.Error(t, err, "--days is required")
}

func TestLearnPrune_EmptyStore_RemovesZero(t *testing.T) {
	dir := setupLearnDir(t)
	chdir(t, dir)

	var out bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"learn", "prune", "--days", "30"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Removed 0 entries older than 30 days")
}

func TestLearnPrune_FlagParsing_IntValue(t *testing.T) {
	dir := setupLearnDir(t)
	chdir(t, dir)

	var out bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"learn", "prune", "--days", "not_a_number"})
	err := cmd.Execute()
	assert.Error(t, err, "non-integer --days should error")
}

// --- learn summary ---

func TestLearnSummary_EmptyStore_ShowsZeroTotal(t *testing.T) {
	dir := setupLearnDir(t)
	chdir(t, dir)

	var out bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"learn", "summary"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Total entries: 0")
}

func TestLearnSummary_TopFlagDefault(t *testing.T) {
	dir := setupLearnDir(t)
	chdir(t, dir)

	var out bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"learn", "summary"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Total entries:")
}

func TestLearnSummary_CustomTopFlag(t *testing.T) {
	dir := setupLearnDir(t)
	chdir(t, dir)

	var out bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"learn", "summary", "--top", "3"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "Total entries:")
}

func TestLearnSummary_UnknownFlag_ReturnsError(t *testing.T) {
	dir := setupLearnDir(t)
	chdir(t, dir)

	var out bytes.Buffer
	cmd := newTestRootCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs([]string{"learn", "summary", "--bogus"})
	err := cmd.Execute()
	assert.Error(t, err, "unknown flag should produce an error")
}
