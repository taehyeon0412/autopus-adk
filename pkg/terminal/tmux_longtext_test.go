package terminal

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTmuxMultiMock replaces execCommand with a mock that records all calls.
func newTmuxMultiMock() (restore func(), captured *capturedCmds) {
	orig := execCommand
	cap := &capturedCmds{}
	execCommand = func(name string, args ...string) *exec.Cmd {
		cap.calls = append(cap.calls, struct {
			name string
			args []string
		}{name, args})
		return exec.Command("true")
	}
	return func() { execCommand = orig }, cap
}

// TestTmuxAdapter_SendLongText_ShortText verifies short text delegates to send-keys.
func TestTmuxAdapter_SendLongText_ShortText(t *testing.T) {
	restore, captured := newTmuxMultiMock()
	defer restore()

	a := &TmuxAdapter{session: "test"}
	err := a.SendLongText(context.Background(), "0", "short prompt")
	require.NoError(t, err)

	// Should issue a single send-keys call (no load-buffer/paste-buffer)
	require.Len(t, captured.calls, 1)
	combined := strings.Join(captured.calls[0].args, " ")
	assert.Contains(t, combined, "send-keys")
	assert.Contains(t, combined, "short prompt")
	assert.NotContains(t, combined, "load-buffer")
}

// TestTmuxAdapter_SendLongText_LongText verifies long text uses load-buffer/paste-buffer.
func TestTmuxAdapter_SendLongText_LongText(t *testing.T) {
	restore, captured := newTmuxMultiMock()
	defer restore()

	// Create a text >= 500 bytes
	longText := strings.Repeat("x", 600)
	a := &TmuxAdapter{session: "test"}
	err := a.SendLongText(context.Background(), "0", longText)
	require.NoError(t, err)

	// Should issue load-buffer then paste-buffer (2 calls)
	require.Len(t, captured.calls, 2)

	loadArgs := strings.Join(captured.calls[0].args, " ")
	assert.Contains(t, loadArgs, "load-buffer")

	pasteArgs := strings.Join(captured.calls[1].args, " ")
	assert.Contains(t, pasteArgs, "paste-buffer")
	assert.Contains(t, pasteArgs, "-t")
}

// TestTmuxAdapter_SendLongText_InvalidPaneID verifies validation error.
func TestTmuxAdapter_SendLongText_InvalidPaneID(t *testing.T) {
	a := &TmuxAdapter{}
	err := a.SendLongText(context.Background(), "", "text")
	assert.Error(t, err)
}

// TestTmuxAdapter_SendLongText_Error verifies error propagation for short text path.
func TestTmuxAdapter_SendLongText_Error(t *testing.T) {
	restore := newTmuxErrorMock()
	defer restore()

	a := &TmuxAdapter{session: "test"}
	err := a.SendLongText(context.Background(), "0", "short")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "send-keys")
}
