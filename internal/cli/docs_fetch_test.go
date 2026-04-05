package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDocsFetchCmd_WithLibrary verifies that `auto docs fetch cobra` fetches docs for cobra.
// Given: a docs fetch command with a library name argument
// When: the command is executed
// Then: it completes without error and outputs content for the specified library
func TestDocsFetchCmd_WithLibrary(t *testing.T) {
	t.Parallel()

	cmd := newDocsFetchCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"cobra"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "cobra")
}

// TestDocsFetchCmd_WithFormat verifies that --format prompt outputs prompt injection format.
// Given: a docs fetch command with --format prompt flag
// When: the command is executed
// Then: the output contains the "## Reference Documentation" prompt section header
func TestDocsFetchCmd_WithFormat(t *testing.T) {
	t.Parallel()

	cmd := newDocsFetchCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"cobra", "--format", "prompt"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "## Reference Documentation")
}

// TestDocsFetchCmd_WithTopic verifies that --topic flag narrows the documentation query.
// Given: a docs fetch command with --topic "commands" flag
// When: the command is executed
// Then: no error is returned and topic is used in the fetch
func TestDocsFetchCmd_WithTopic(t *testing.T) {
	t.Parallel()

	cmd := newDocsFetchCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"cobra", "--topic", "commands"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.NotEmpty(t, buf.String())
}

// TestDocsFetchCmd_AutoDetect verifies that `auto docs fetch` without args auto-detects libraries.
// Given: a docs fetch command with no library argument
// When: the command is executed from a project directory
// Then: it attempts auto-detection and outputs detected library list or docs
func TestDocsFetchCmd_AutoDetect(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AUTO_DOCS_PROJECT_DIR", dir)

	cmd := newDocsFetchCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{})

	// When no manifest file exists in the temp dir, command should still exit cleanly
	// with a message about no detected libraries or empty output.
	err := cmd.Execute()
	require.NoError(t, err)
}
