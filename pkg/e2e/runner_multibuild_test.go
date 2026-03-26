package e2e

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRun_MultiBuild_MatchingScenario verifies that a multi-build runner
// selects the correct build entry based on scenario section and runs it
// in the correct directory.
func TestRun_MultiBuild_MatchingScenario(t *testing.T) {
	t.Parallel()

	// Given: a project with two submodules, each with its own build
	projectDir := t.TempDir()
	adkDir := filepath.Join(projectDir, "autopus-adk")
	require.NoError(t, os.MkdirAll(adkDir, 0o755))

	// Write a marker file via the build command to prove the build ran
	// in the correct directory.
	scenario := Scenario{
		ID:      "adk-test",
		Command: "echo ok",
		Section: "ADK CLI Scenarios",
		Verify:  []string{"exit_code(0)"},
		Status:  "active",
	}

	runner := NewRunner(RunnerOptions{
		ProjectDir: projectDir,
		AutoBuild:  true,
		Builds: []BuildEntry{
			{Command: "touch adk-built.marker", Label: "ADK", SubmodulePath: "autopus-adk"},
			{Command: "touch backend-built.marker", Label: "Backend", SubmodulePath: "Autopus"},
		},
		Timeout: 10 * time.Second,
	})

	// When: Run is called with an ADK scenario
	result, err := runner.Run(scenario)

	// Then: build occurred and the ADK marker exists
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.BuildOccurred)

	// Verify marker was created in ADK subdir
	_, statErr := os.Stat(filepath.Join(adkDir, "adk-built.marker"))
	assert.NoError(t, statErr, "ADK build marker should exist in submodule dir")

	// Backend marker should NOT exist
	_, statErr = os.Stat(filepath.Join(projectDir, "Autopus", "backend-built.marker"))
	assert.True(t, os.IsNotExist(statErr), "Backend build should not have run")
}

// TestRun_MultiBuild_NonMatchingScenario verifies that when no build entry
// matches the scenario section, the build step is skipped entirely (R5).
func TestRun_MultiBuild_NonMatchingScenario(t *testing.T) {
	t.Parallel()

	// Given: builds for ADK and Backend, but scenario is in an unknown section
	projectDir := t.TempDir()
	scenario := Scenario{
		ID:      "unknown-section-test",
		Command: "echo ok",
		Section: "Unknown Scenarios",
		Verify:  []string{"exit_code(0)"},
		Status:  "active",
	}

	runner := NewRunner(RunnerOptions{
		ProjectDir: projectDir,
		AutoBuild:  true,
		Builds: []BuildEntry{
			{Command: "touch should-not-exist.marker", Label: "ADK", SubmodulePath: "autopus-adk"},
		},
		Timeout: 10 * time.Second,
	})

	// When: Run is called
	result, err := runner.Run(scenario)

	// Then: no build occurred, command still runs successfully
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.BuildOccurred, "build should be skipped for non-matching section")
	assert.True(t, result.Pass)
}

// TestRun_MultiBuild_SingleUnlabeled_MatchesAny verifies backward
// compatibility: a single unlabeled BuildEntry matches any scenario.
func TestRun_MultiBuild_SingleUnlabeled_MatchesAny(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	scenario := Scenario{
		ID:      "unlabeled-test",
		Command: "echo ok",
		Section: "Whatever Section",
		Verify:  []string{"exit_code(0)"},
		Status:  "active",
	}

	runner := NewRunner(RunnerOptions{
		ProjectDir: projectDir,
		AutoBuild:  true,
		Builds: []BuildEntry{
			{Command: "touch universal.marker", Label: "", SubmodulePath: ""},
		},
		Timeout: 10 * time.Second,
	})

	// When: Run is called
	result, err := runner.Run(scenario)

	// Then: build occurred (single unlabeled matches everything)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.BuildOccurred)

	_, statErr := os.Stat(filepath.Join(projectDir, "universal.marker"))
	assert.NoError(t, statErr, "universal marker should exist in project dir")
}

// TestRun_MultiBuild_EmptyBuilds_LegacyFallback verifies that when Builds
// is empty but BuildCommand is set, the legacy single-build path is used (R4).
func TestRun_MultiBuild_EmptyBuilds_LegacyFallback(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	scenario := Scenario{
		ID:      "legacy-fallback-test",
		Command: "echo ok",
		Verify:  []string{"exit_code(0)"},
		Status:  "active",
	}

	runner := NewRunner(RunnerOptions{
		ProjectDir:   projectDir,
		AutoBuild:    true,
		BuildCommand: "touch legacy.marker",
		Timeout:      10 * time.Second,
	})

	// When: Run is called with no Builds but a BuildCommand
	result, err := runner.Run(scenario)

	// Then: legacy build path is used
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.BuildOccurred)

	_, statErr := os.Stat(filepath.Join(projectDir, "legacy.marker"))
	assert.NoError(t, statErr, "legacy marker should exist")
}

// TestRun_MultiBuild_EmptyBuilds_NoBuildCommand verifies that when both
// Builds and BuildCommand are empty, no build occurs.
func TestRun_MultiBuild_EmptyBuilds_NoBuildCommand(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	scenario := Scenario{
		ID:      "no-build-test",
		Command: "echo ok",
		Verify:  []string{"exit_code(0)"},
		Status:  "active",
	}

	runner := NewRunner(RunnerOptions{
		ProjectDir: projectDir,
		AutoBuild:  true,
		Timeout:    10 * time.Second,
	})

	// When: Run is called with no build configuration
	result, err := runner.Run(scenario)

	// Then: no build occurred
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.BuildOccurred)
}

// TestRun_MultiBuild_LabelOnce_CachesPerLabel verifies that builds are
// cached per label — running two scenarios with the same label only builds once.
func TestRun_MultiBuild_LabelOnce_CachesPerLabel(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	adkDir := filepath.Join(projectDir, "autopus-adk")
	require.NoError(t, os.MkdirAll(adkDir, 0o755))

	builds := []BuildEntry{
		{Command: "touch adk-built.marker", Label: "ADK", SubmodulePath: "autopus-adk"},
	}

	runner := NewRunner(RunnerOptions{
		ProjectDir: projectDir,
		AutoBuild:  true,
		Builds:     builds,
		Timeout:    10 * time.Second,
	})

	s1 := Scenario{ID: "adk-1", Command: "echo first", Section: "ADK CLI Scenarios", Verify: []string{"exit_code(0)"}, Status: "active"}
	s2 := Scenario{ID: "adk-2", Command: "echo second", Section: "ADK CLI Scenarios", Verify: []string{"exit_code(0)"}, Status: "active"}

	// When: two scenarios with the same label run
	r1, err1 := runner.Run(s1)
	r2, err2 := runner.Run(s2)

	// Then: first builds, second uses cache
	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.True(t, r1.BuildOccurred, "first run should build")
	assert.False(t, r2.BuildOccurred, "second run should use cached build")
}

// TestRun_MultiBuild_BuildFailure_ReturnsError verifies that a failing build
// command in multi-build mode returns an error with the label name.
func TestRun_MultiBuild_BuildFailure_ReturnsError(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	adkDir := filepath.Join(projectDir, "autopus-adk")
	require.NoError(t, os.MkdirAll(adkDir, 0o755))

	scenario := Scenario{
		ID:      "build-fail-multi",
		Command: "echo should-not-run",
		Section: "ADK CLI Scenarios",
		Verify:  []string{"exit_code(0)"},
		Status:  "active",
	}

	runner := NewRunner(RunnerOptions{
		ProjectDir: projectDir,
		AutoBuild:  true,
		Builds: []BuildEntry{
			{Command: "false", Label: "ADK", SubmodulePath: "autopus-adk"},
		},
		Timeout: 10 * time.Second,
	})

	// When: Run is called with a failing build
	_, err := runner.Run(scenario)

	// Then: error is returned with label info
	require.Error(t, err)
	assert.Contains(t, err.Error(), "auto-build failed")
	assert.Contains(t, err.Error(), "ADK")
}

// TestRun_MultiBuild_BuildFailure_CachedForSubsequentRuns verifies that a
// failed build is cached — subsequent runs with the same label also fail
// without re-executing the build command.
func TestRun_MultiBuild_BuildFailure_CachedForSubsequentRuns(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	adkDir := filepath.Join(projectDir, "autopus-adk")
	require.NoError(t, os.MkdirAll(adkDir, 0o755))

	builds := []BuildEntry{
		{Command: "false", Label: "ADK", SubmodulePath: "autopus-adk"},
	}

	runner := NewRunner(RunnerOptions{
		ProjectDir: projectDir,
		AutoBuild:  true,
		Builds:     builds,
		Timeout:    10 * time.Second,
	})

	s1 := Scenario{ID: "fail-1", Command: "echo a", Section: "ADK CLI Scenarios", Verify: []string{"exit_code(0)"}, Status: "active"}
	s2 := Scenario{ID: "fail-2", Command: "echo b", Section: "ADK CLI Scenarios", Verify: []string{"exit_code(0)"}, Status: "active"}

	// When: two runs with the same failing label
	_, err1 := runner.Run(s1)
	_, err2 := runner.Run(s2)

	// Then: both fail with build error
	require.Error(t, err1)
	require.Error(t, err2)
	assert.Contains(t, err1.Error(), "auto-build failed")
	assert.Contains(t, err2.Error(), "auto-build failed")
}
