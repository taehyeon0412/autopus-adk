package pipeline

import (
	"context"
	"testing"

	"github.com/insajin/autopus-adk/pkg/terminal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// S10: 3-person team layout
func TestPlanLayout_ThreeMembers(t *testing.T) {
	t.Parallel()
	teammates := []string{"lead", "builder", "guardian"}
	plan := planLayout(teammates)

	assert.Equal(t, 3, plan.SplitCount)
	assert.Equal(t, teammates, plan.Roles)
}

// S11: 5-person team layout
func TestPlanLayout_FiveMembers(t *testing.T) {
	t.Parallel()
	teammates := []string{"lead", "builder-1", "builder-2", "builder-3", "guardian"}
	plan := planLayout(teammates)

	assert.Equal(t, 5, plan.SplitCount)
	assert.Equal(t, teammates, plan.Roles)
}

func TestPlanLayout_Empty(t *testing.T) {
	t.Parallel()
	plan := planLayout(nil)
	assert.Equal(t, 0, plan.SplitCount)
	assert.Empty(t, plan.Roles)
}

func TestApplyLayout_CallsSplitPaneVertical(t *testing.T) {
	t.Parallel()
	term := newTeamMock("cmux")
	plan := planLayout([]string{"lead", "builder", "guardian"})

	result, err := applyLayout(context.Background(), term, plan)
	require.NoError(t, err)

	assert.Equal(t, 3, term.splitCount)
	assert.Len(t, result.TeammatePaneIDs, 3)
	// All splits must be Vertical.
	for _, dir := range term.splitDirs {
		assert.Equal(t, terminal.Vertical, dir)
	}
}

func TestApplyLayout_ZeroSplits(t *testing.T) {
	t.Parallel()
	term := newTeamMock("cmux")
	plan := planLayout(nil)

	result, err := applyLayout(context.Background(), term, plan)
	require.NoError(t, err)
	assert.Empty(t, result.TeammatePaneIDs)
	assert.Equal(t, 0, term.splitCount)
}

// S9: SplitPane failure triggers cleanup of already-created panes
func TestApplyLayout_SplitPaneFailure_Cleanup(t *testing.T) {
	t.Parallel()
	term := newTeamMock("cmux")
	term.failSplitAfter = 2 // fail on 3rd split

	plan := planLayout([]string{"lead", "builder", "guardian"})
	result, err := applyLayout(context.Background(), term, plan)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "guardian")
	// Cleanup: Close should have been called for the 2 successfully created panes.
	assert.Len(t, term.closedSessions, 2)
}

func TestApplyLayout_FirstSplitFailure(t *testing.T) {
	t.Parallel()
	term := newTeamMock("cmux")
	term.failSplitAfter = 0

	plan := planLayout([]string{"lead"})
	_, err := applyLayout(context.Background(), term, plan)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "lead")
	// No panes were created so nothing to clean up.
	assert.Empty(t, term.closedSessions)
}
