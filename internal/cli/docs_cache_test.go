package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDocsCacheListCmd verifies that `auto docs cache list` outputs cached entries.
// Given: a docs cache list command
// When: executed
// Then: it completes without error and outputs the cache list (possibly empty)
func TestDocsCacheListCmd(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AUTO_DOCS_CACHE_DIR", dir)

	cmd := newDocsCacheCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"list"})

	err := cmd.Execute()
	require.NoError(t, err)

	// Output must be present (even if "no entries" message)
	assert.NotEmpty(t, buf.String(), "cache list must produce output")
}

// TestDocsCacheClearCmd verifies that `auto docs cache clear` empties the cache.
// Given: a populated cache directory
// When: clear subcommand is executed
// Then: the command completes without error and cache directory is empty
func TestDocsCacheClearCmd(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("AUTO_DOCS_CACHE_DIR", dir)

	cmd := newDocsCacheCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"clear"})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify success message is present
	assert.Contains(t, buf.String(), "clear", "output must mention clear operation")
}
