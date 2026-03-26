package pipeline

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/insajin/autopus-adk/pkg/terminal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamShellEscape(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"simple", "hello", "'hello'"},
		{"with spaces", "hello world", "'hello world'"},
		{"with single quote", "it's", `'it'\''s'`},
		{"empty", "", "''"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expect, teamShellEscape(tt.input))
		})
	}
}

func TestSanitizeRole(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"simple", "builder", "builder"},
		{"with hyphen", "builder-1", "builder-1"},
		{"with underscore", "build_er", "build_er"},
		{"special chars", "ro!@#le", "role"},
		{"all special", "!@#$", "unknown"},
		{"empty", "", "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expect, sanitizeRole(tt.input))
		})
	}
}

// S12: Log file naming uniqueness
func TestLogFileNaming_ContainsSpecIDAndRole(t *testing.T) {
	t.Parallel()
	term := newTeamMock("cmux")
	paneIDs := []terminal.PaneID{"pane-1"}
	roles := []string{"builder-1"}

	panes, err := createTeammatePanes(context.Background(), term, "SPEC-TEAMPANE-001", paneIDs, roles)
	require.NoError(t, err)
	require.Len(t, panes, 1)

	assert.Contains(t, panes[0].LogPath, "autopus-team-SPEC-TEAMPANE-001-builder-1-")
	// Cleanup.
	cleanupTeammatePanes(panes)
}

func TestCreateTeammatePanes_InvalidSpecID(t *testing.T) {
	t.Parallel()
	term := newTeamMock("cmux")
	_, err := createTeammatePanes(context.Background(), term, "../evil", nil, nil)
	require.Error(t, err)
}

func TestStreamToPane_SendsTailCommand(t *testing.T) {
	t.Parallel()
	term := newTeamMock("cmux")
	streamToPane(context.Background(), term, "pane-1", "/tmp/test.log")

	require.Len(t, term.sentCommands, 1)
	assert.Contains(t, term.sentCommands[0].cmd, "tail -f")
	assert.Contains(t, term.sentCommands[0].cmd, "/tmp/test.log")
	assert.Equal(t, terminal.PaneID("pane-1"), term.sentCommands[0].paneID)
}

// S6: Close() removes all temp log files
func TestCleanupTeammatePanes_RemovesLogFiles(t *testing.T) {
	t.Parallel()
	// Create actual temp files.
	f1, err := os.CreateTemp("", "autopus-test-cleanup-")
	require.NoError(t, err)
	f1.Close()
	f2, err := os.CreateTemp("", "autopus-test-cleanup-")
	require.NoError(t, err)
	f2.Close()

	panes := []TeammatePaneInfo{
		{Role: "lead", LogPath: f1.Name()},
		{Role: "builder", LogPath: f2.Name()},
	}

	cleanupTeammatePanes(panes)

	_, err = os.Stat(f1.Name())
	assert.True(t, os.IsNotExist(err), "log file 1 should be removed")
	_, err = os.Stat(f2.Name())
	assert.True(t, os.IsNotExist(err), "log file 2 should be removed")
}

func TestCleanupTeammatePanes_EmptyLogPath(t *testing.T) {
	t.Parallel()
	panes := []TeammatePaneInfo{{Role: "test", LogPath: ""}}
	// Should not panic.
	cleanupTeammatePanes(panes)
}

func TestSendFailureMessage(t *testing.T) {
	t.Parallel()
	term := newTeamMock("cmux")
	sendFailureMessage(context.Background(), term, "pane-1", "builder", "timeout")

	require.Len(t, term.sentCommands, 1)
	assert.Contains(t, term.sentCommands[0].cmd, "echo")
	assert.Contains(t, term.sentCommands[0].cmd, "[FAILED] builder: timeout")
}

func TestCreateTeammatePanes_MultipleRoles(t *testing.T) {
	t.Parallel()
	term := newTeamMock("cmux")
	paneIDs := []terminal.PaneID{"pane-1", "pane-2", "pane-3"}
	roles := []string{"lead", "builder", "guardian"}

	panes, err := createTeammatePanes(context.Background(), term, "SPEC-TEST-001", paneIDs, roles)
	require.NoError(t, err)
	require.Len(t, panes, 3)

	for i, p := range panes {
		assert.Equal(t, roles[i], p.Role)
		assert.Equal(t, paneIDs[i], p.PaneID)
		assert.NotEmpty(t, p.LogPath)
		assert.True(t, strings.Contains(p.LogPath, roles[i]))
	}
	// Tail commands should have been sent for each pane.
	assert.Len(t, term.sentCommands, 3)

	cleanupTeammatePanes(panes)
}
