package issue_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/issue"
)

func TestCollectContext_BasicFields(t *testing.T) {
	t.Parallel()

	ctx := issue.CollectContext("some error", "auto plan", 1)

	assert.Equal(t, "some error", ctx.ErrorMessage)
	assert.Equal(t, "auto plan", ctx.Command)
	assert.Equal(t, 1, ctx.ExitCode)
	assert.NotEmpty(t, ctx.OS)
	assert.NotEmpty(t, ctx.GoVersion)
	assert.NotEmpty(t, ctx.AutoVersion)
}

func TestCollectContext_ReadsConfig(t *testing.T) {
	// No t.Parallel(): os.Chdir is process-wide and conflicts with parallel tests.

	dir := t.TempDir()
	cfgContent := "project_name: test-project\nmode: full\n"
	err := os.WriteFile(filepath.Join(dir, "autopus.yaml"), []byte(cfgContent), 0o644)
	require.NoError(t, err)

	orig, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(orig) })
	require.NoError(t, os.Chdir(dir))

	ctx := issue.CollectContext("err", "cmd", 0)

	assert.Contains(t, ctx.ConfigYAML, "test-project")
}

func TestCollectContext_NoConfigFile(t *testing.T) {
	// No t.Parallel(): os.Chdir is process-wide.

	dir := t.TempDir()
	orig, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(orig) })
	require.NoError(t, os.Chdir(dir))

	// Should not panic when autopus.yaml is absent.
	ctx := issue.CollectContext("err", "cmd", 0)
	assert.Empty(t, ctx.ConfigYAML)
}

func TestCollectContext_SanitizesHomePath(t *testing.T) {
	// No t.Parallel(): os.Chdir is process-wide.

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home dir:", err)
	}

	dir := filepath.Join(home, "tmp-autopus-test-"+t.Name())
	require.NoError(t, os.MkdirAll(dir, 0o755))
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	cfgContent := "project_name: " + home + "/myproject\nmode: full\n"
	err = os.WriteFile(filepath.Join(dir, "autopus.yaml"), []byte(cfgContent), 0o644)
	require.NoError(t, err)

	orig, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(orig) })
	require.NoError(t, os.Chdir(dir))

	ctx := issue.CollectContext("err", "cmd", 0)

	assert.NotContains(t, ctx.ConfigYAML, home)
	assert.Contains(t, ctx.ConfigYAML, "$HOME")
}

func TestCollectContext_ReadsTelemetry(t *testing.T) {
	// No t.Parallel(): os.Chdir is process-wide.

	dir := t.TempDir()
	telDir := filepath.Join(dir, ".autopus", "telemetry")
	require.NoError(t, os.MkdirAll(telDir, 0o755))

	telContent := `{"spec_id":"SPEC-001","status":"PASS"}`
	err := os.WriteFile(filepath.Join(telDir, "run-001.jsonl"), []byte(telContent+"\n"), 0o644)
	require.NoError(t, err)

	orig, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(orig) })
	require.NoError(t, os.Chdir(dir))

	ctx := issue.CollectContext("err", "cmd", 0)
	assert.NotEmpty(t, ctx.Telemetry)
}

func TestCollectContext_DetectsPlatformCodex(t *testing.T) {
	// No t.Parallel(): t.Setenv is incompatible with parallel tests.
	t.Setenv("CODEX", "1")
	t.Setenv("CLAUDE_CODE", "")

	ctx := issue.CollectContext("err", "cmd", 0)
	assert.Equal(t, "codex", ctx.Platform)
}

func TestCollectContext_DetectsPlatformClaudeCode(t *testing.T) {
	// No t.Parallel(): t.Setenv is incompatible with parallel tests.
	t.Setenv("CLAUDE_CODE", "1")
	t.Setenv("CODEX", "")

	ctx := issue.CollectContext("err", "cmd", 0)
	assert.Equal(t, "claude-code", ctx.Platform)
}

func TestCollectContext_TelemetryFewLines(t *testing.T) {
	// No t.Parallel(): os.Chdir is process-wide.

	dir := t.TempDir()
	telDir := filepath.Join(dir, ".autopus", "telemetry")
	require.NoError(t, os.MkdirAll(telDir, 0o755))

	// Write fewer than maxTelemetryLines lines.
	lines := "line1\nline2\nline3\n"
	err := os.WriteFile(filepath.Join(telDir, "run-001.jsonl"), []byte(lines), 0o644)
	require.NoError(t, err)

	orig, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(orig) })
	require.NoError(t, os.Chdir(dir))

	ctx := issue.CollectContext("err", "cmd", 0)
	assert.Contains(t, ctx.Telemetry, "line1")
	assert.Contains(t, ctx.Telemetry, "line3")
}

func TestCollectContext_NoTelemetryDir(t *testing.T) {
	// No t.Parallel(): os.Chdir is process-wide.

	dir := t.TempDir()
	// Do NOT create the telemetry dir.

	orig, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(orig) })
	require.NoError(t, os.Chdir(dir))

	ctx := issue.CollectContext("err", "cmd", 0)
	assert.Empty(t, ctx.Telemetry)
}
