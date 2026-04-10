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

func TestBuildIssueTitle_BothFields(t *testing.T) {
	t.Parallel()
	title := buildIssueTitle("something broke", "doctor")
	assert.Equal(t, "[auto] doctor: something broke", title)
}

func TestBuildIssueTitle_LongError(t *testing.T) {
	t.Parallel()
	longMsg := strings.Repeat("x", 80)
	title := buildIssueTitle(longMsg, "init")
	// Should truncate to 60 chars + "..."
	assert.Equal(t, "[auto] init: "+strings.Repeat("x", 60)+"...", title)
}

func TestBuildIssueTitle_ErrOnly(t *testing.T) {
	t.Parallel()
	title := buildIssueTitle("oops", "")
	assert.Equal(t, "[auto] oops", title)
}

func TestBuildIssueTitle_LongErrOnly(t *testing.T) {
	t.Parallel()
	longMsg := strings.Repeat("y", 80)
	title := buildIssueTitle(longMsg, "")
	assert.Equal(t, "[auto] "+strings.Repeat("y", 72)+"...", title)
}

func TestBuildIssueTitle_Empty(t *testing.T) {
	t.Parallel()
	title := buildIssueTitle("", "")
	assert.Equal(t, "[auto] issue report", title)
}

func TestParseGitHubRepo(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "org/repo", parseGitHubRepo("git@github.com:org/repo.git"))
	assert.Equal(t, "org/repo", parseGitHubRepo("https://github.com/org/repo.git"))
	assert.Equal(t, "", parseGitHubRepo("https://example.com/org/repo.git"))
}

func TestResolveIssueRepoInputs_Priority(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "custom/repo", resolveIssueRepoInputs("custom/repo", "auto spec review", "cfg/repo", "git/repo"))
	assert.Equal(t, "cfg/repo", resolveIssueRepoInputs("", "auto spec review", "cfg/repo", "git/repo"))
	assert.Equal(t, defaultIssueRepo, resolveIssueRepoInputs("", "auto spec review", "", "git/repo"))
	assert.Equal(t, "git/repo", resolveIssueRepoInputs("", "make test", "", "git/repo"))
	assert.Equal(t, defaultIssueRepo, resolveIssueRepoInputs("", "make test", "", ""))
}

func TestConfirmIssue_Yes(t *testing.T) {
	t.Parallel()

	cmd := newIssueCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetIn(strings.NewReader("y\n"))

	result := confirmIssue(cmd, "confirm? ")
	assert.True(t, result)
}

func TestConfirmIssue_No(t *testing.T) {
	t.Parallel()

	cmd := newIssueCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetIn(strings.NewReader("n\n"))

	result := confirmIssue(cmd, "confirm? ")
	assert.False(t, result)
}

func TestConfirmIssue_EmptyInput(t *testing.T) {
	t.Parallel()

	cmd := newIssueCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetIn(strings.NewReader(""))

	result := confirmIssue(cmd, "confirm? ")
	assert.False(t, result)
}
