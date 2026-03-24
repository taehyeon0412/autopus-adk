// Package e2e provides user-facing scenario-based E2E test infrastructure.
package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSync_NewCommands_AddedToScenarios verifies that SyncScenarios adds new
// scenarios for Cobra commands added since the last sync.
// S6: new command detection during sync.
func TestSync_NewCommands_AddedToScenarios(t *testing.T) {
	t.Parallel()

	// Given: an existing ScenarioSet with one scenario
	existing := &ScenarioSet{
		Scenarios: []Scenario{
			{ID: "init", Command: "auto init", Status: "active"},
		},
	}
	// And: two new commands detected in the codebase
	newCommands := []Scenario{
		{ID: "init", Command: "auto init", Status: "active"},
		{ID: "doctor", Command: "auto doctor", Status: "active"},
		{ID: "setup", Command: "auto setup", Status: "active"},
	}

	// When: SyncScenarios is called
	updated, err := SyncScenarios(existing, newCommands)

	// Then: updated set contains all three scenarios
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Len(t, updated.Scenarios, 3)
	ids := make(map[string]bool)
	for _, s := range updated.Scenarios {
		ids[s.ID] = true
	}
	assert.True(t, ids["doctor"], "doctor scenario should be added")
	assert.True(t, ids["setup"], "setup scenario should be added")
}

// TestSync_DeletedCommands_MarkedDeprecated verifies that SyncScenarios marks
// scenarios deprecated when their corresponding commands are removed.
// S7: deleted command handling during sync.
func TestSync_DeletedCommands_MarkedDeprecated(t *testing.T) {
	t.Parallel()

	// Given: an existing set with "old-cmd" scenario
	existing := &ScenarioSet{
		Scenarios: []Scenario{
			{ID: "init", Command: "auto init", Status: "active"},
			{ID: "old-cmd", Command: "auto old-cmd", Status: "active"},
		},
	}
	// And: current commands no longer include "old-cmd"
	currentCommands := []Scenario{
		{ID: "init", Command: "auto init", Status: "active"},
	}

	// When: SyncScenarios is called
	updated, err := SyncScenarios(existing, currentCommands)

	// Then: "old-cmd" scenario is marked deprecated
	require.NoError(t, err)
	require.NotNil(t, updated)
	var oldCmdStatus string
	for _, s := range updated.Scenarios {
		if s.ID == "old-cmd" {
			oldCmdStatus = s.Status
		}
	}
	assert.Equal(t, "deprecated", oldCmdStatus)
}

// TestSync_ManualEdits_Preserved verifies that scenarios with "custom-" prefix
// in their ID are retained unchanged during sync, even if not in extracted commands.
// R3: manual edit preservation.
func TestSync_ManualEdits_Preserved(t *testing.T) {
	t.Parallel()

	// Given: existing set with a manually added "custom-" scenario
	existing := &ScenarioSet{
		Scenarios: []Scenario{
			{ID: "init", Command: "auto init", Status: "active"},
			{ID: "custom-smoke", Command: "auto version", Status: "active"},
		},
	}
	// And: current extracted commands do not include custom-smoke
	currentCommands := []Scenario{
		{ID: "init", Command: "auto init", Status: "active"},
	}

	// When: SyncScenarios is called
	updated, err := SyncScenarios(existing, currentCommands)

	// Then: custom-smoke is still present and unchanged
	require.NoError(t, err)
	require.NotNil(t, updated)
	var customFound bool
	for _, s := range updated.Scenarios {
		if s.ID == "custom-smoke" {
			customFound = true
			assert.Equal(t, "active", s.Status, "custom scenario should retain active status")
		}
	}
	assert.True(t, customFound, "custom-smoke scenario should be preserved")
}

// TestSync_NoChanges_NoModification verifies that SyncScenarios returns an
// identical ScenarioSet when extracted commands match existing scenarios.
func TestSync_NoChanges_NoModification(t *testing.T) {
	t.Parallel()

	// Given: existing and current commands are identical
	scenarios := []Scenario{
		{ID: "init", Command: "auto init", Status: "active"},
		{ID: "doctor", Command: "auto doctor", Status: "active"},
	}
	existing := &ScenarioSet{Scenarios: scenarios}
	currentCommands := []Scenario{
		{ID: "init", Command: "auto init", Status: "active"},
		{ID: "doctor", Command: "auto doctor", Status: "active"},
	}

	// When: SyncScenarios is called
	updated, err := SyncScenarios(existing, currentCommands)

	// Then: the set is unchanged (no additions, no deprecations)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Len(t, updated.Scenarios, 2)
	for _, s := range updated.Scenarios {
		assert.Equal(t, "active", s.Status)
	}
}
