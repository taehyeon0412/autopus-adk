package daemon

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePlist_ContainsExpectedElements(t *testing.T) {
	t.Parallel()

	cfg := LaunchdConfig{
		BinaryPath: "/usr/local/bin/autopus",
		Args:       []string{"worker", "start"},
		LogDir:     "/tmp/autopus-logs",
	}

	content, err := GeneratePlist(cfg)
	require.NoError(t, err)

	assert.Contains(t, content, "<key>Label</key>")
	assert.Contains(t, content, "<string>co.autopus.worker</string>")
	assert.Contains(t, content, "<key>ProgramArguments</key>")
	assert.Contains(t, content, "<string>/usr/local/bin/autopus</string>")
	assert.Contains(t, content, "<string>worker</string>")
	assert.Contains(t, content, "<string>start</string>")
	assert.Contains(t, content, "<key>KeepAlive</key>")
	assert.Contains(t, content, "<true/>")
	assert.Contains(t, content, "<key>RunAtLoad</key>")
	assert.Contains(t, content, "/tmp/autopus-logs/autopus-worker.out.log")
	assert.Contains(t, content, "/tmp/autopus-logs/autopus-worker.err.log")
}

func TestGeneratePlist_DefaultLogDir(t *testing.T) {
	t.Parallel()

	cfg := LaunchdConfig{
		BinaryPath: "/usr/local/bin/autopus",
	}

	content, err := GeneratePlist(cfg)
	require.NoError(t, err)

	// Should use the default log directory under ~/.config/autopus/logs
	assert.Contains(t, content, "autopus-worker.out.log")
	assert.Contains(t, content, "autopus-worker.err.log")
}

func TestGeneratePlist_NoArgs(t *testing.T) {
	t.Parallel()

	cfg := LaunchdConfig{
		BinaryPath: "/usr/local/bin/autopus",
		LogDir:     "/tmp/logs",
	}

	content, err := GeneratePlist(cfg)
	require.NoError(t, err)

	assert.Contains(t, content, "<string>/usr/local/bin/autopus</string>")
	assert.Contains(t, content, "<?xml version")
	assert.Contains(t, content, "<!DOCTYPE plist")
}

func TestLaunchdPlistPath_Format(t *testing.T) {
	t.Parallel()

	path := launchdPlistPath()
	assert.Contains(t, path, "co.autopus.worker.plist")
	assert.Contains(t, path, "LaunchAgents")
}
