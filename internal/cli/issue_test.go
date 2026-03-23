package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)


func TestNewIssueCmd_SubcommandsRegistered(t *testing.T) {
	t.Parallel()

	cmd := newIssueCmd()
	assert.Equal(t, "issue", cmd.Use)

	subNames := make([]string, 0, len(cmd.Commands()))
	for _, sub := range cmd.Commands() {
		subNames = append(subNames, sub.Use)
	}

	// Use strings.HasPrefix to accommodate "search <query>" etc.
	hasReport, hasList, hasSearch := false, false, false
	for _, name := range subNames {
		if strings.HasPrefix(name, "report") {
			hasReport = true
		}
		if strings.HasPrefix(name, "list") {
			hasList = true
		}
		if strings.HasPrefix(name, "search") {
			hasSearch = true
		}
	}
	assert.True(t, hasReport, "report subcommand expected")
	assert.True(t, hasList, "list subcommand expected")
	assert.True(t, hasSearch, "search subcommand expected")
}

func TestIssueCmdHelp(t *testing.T) {
	t.Parallel()

	// Test using the issue command directly without going through root,
	// since root registration happens in T9 (root.go modification).
	cmd := newIssueCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	// Help always returns nil with cobra.
	_ = cmd.Execute()
	assert.Contains(t, buf.String(), "issue")
}

func TestIssueReportCmd_DryRunFlag(t *testing.T) {
	t.Parallel()

	cmd := newIssueReportCmd()
	f := cmd.Flags().Lookup("dry-run")
	require.NotNil(t, f, "dry-run flag must exist")
	assert.Equal(t, "false", f.DefValue)
}

func TestIssueReportCmd_AutoSubmitFlag(t *testing.T) {
	t.Parallel()

	cmd := newIssueReportCmd()
	f := cmd.Flags().Lookup("auto-submit")
	require.NotNil(t, f, "auto-submit flag must exist")
	assert.Equal(t, "false", f.DefValue)
}

func TestIssueListCmd_ExistsWithUse(t *testing.T) {
	t.Parallel()

	cmd := newIssueListCmd()
	assert.True(t, strings.HasPrefix(cmd.Use, "list"), "list cmd use should start with 'list'")
}

func TestIssueSearchCmd_ExistsWithUse(t *testing.T) {
	t.Parallel()

	cmd := newIssueSearchCmd()
	assert.True(t, strings.HasPrefix(cmd.Use, "search"), "search cmd use should start with 'search'")
}
