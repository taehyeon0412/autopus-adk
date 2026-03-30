package orchestra

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/insajin/autopus-adk/pkg/terminal"
	"github.com/stretchr/testify/assert"
)

// signalMock embeds mockTerminal and adds SignalCapable support for testing.
type signalMock struct {
	mockTerminal
	waitErr     error
	waitCalled  bool
	waitName    string
	waitTimeout time.Duration
}

func (m *signalMock) SurfaceHealth(_ context.Context, _ terminal.PaneID) (terminal.SurfaceStatus, error) {
	return terminal.SurfaceStatus{Valid: true}, nil
}

func (m *signalMock) WaitForSignal(_ context.Context, name string, timeout time.Duration) error {
	m.waitCalled = true
	m.waitName = name
	m.waitTimeout = timeout
	return m.waitErr
}

func (m *signalMock) SendSignal(_ context.Context, _ string) error {
	return nil
}

// TestSignalDetector_SignalNameRound1 verifies signal name for round 0 and 1.
func TestSignalDetector_SignalNameRound1(t *testing.T) {
	t.Parallel()
	mock := &signalMock{}
	mock.name = "cmux"

	detector := &SignalDetector{term: mock, signal: mock}
	pi := paneInfo{paneID: "pane-1", provider: ProviderConfig{Name: "claude"}}
	patterns := DefaultCompletionPatterns()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ok, err := detector.WaitForCompletion(ctx, pi, patterns, "", 0)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, mock.waitCalled)
	assert.Equal(t, "done-claude", mock.waitName)
}

// TestSignalDetector_SignalNameRoundN verifies signal name includes round for N > 1.
func TestSignalDetector_SignalNameRoundN(t *testing.T) {
	t.Parallel()
	mock := &signalMock{}
	mock.name = "cmux"

	detector := &SignalDetector{term: mock, signal: mock}
	pi := paneInfo{paneID: "pane-1", provider: ProviderConfig{Name: "gemini"}}
	patterns := DefaultCompletionPatterns()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ok, err := detector.WaitForCompletion(ctx, pi, patterns, "", 3)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "done-gemini-round3", mock.waitName)
}

// TestSignalDetector_FallbackOnError verifies fallback to poll when signal fails.
func TestSignalDetector_FallbackOnError(t *testing.T) {
	t.Parallel()
	mock := &signalMock{}
	mock.name = "cmux"
	mock.waitErr = fmt.Errorf("signal timeout")
	// Set up screen to show prompt so poll detector confirms completion
	mock.readScreenOutput = ">\n"

	// Use a countingScreenMock to provide 2 consecutive prompt matches for fallback
	countMock := &signalCountMock{
		signalMock: signalMock{waitErr: fmt.Errorf("signal timeout")},
		outputs:    []string{">\n", ">\n"},
	}
	countMock.name = "cmux"

	detector := &SignalDetector{term: countMock, signal: countMock}
	pi := paneInfo{paneID: "pane-1", provider: ProviderConfig{Name: "codex"}}
	patterns := []CompletionPattern{
		{Provider: "codex", Pattern: DefaultCompletionPatterns()[1].Pattern},
	}

	// Use codex> pattern to match
	countMock.outputs = []string{"codex>\n", "codex>\n"}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ok, err := detector.WaitForCompletion(ctx, pi, patterns, "", 1)
	assert.NoError(t, err)
	assert.True(t, ok, "should complete via fallback poll after signal failure")
}

// TestNewCompletionDetector_SignalCapable verifies factory returns SignalDetector for SignalCapable terminal.
func TestNewCompletionDetector_SignalCapable(t *testing.T) {
	t.Parallel()
	mock := &signalMock{}
	mock.name = "cmux"
	detector := NewCompletionDetector(mock)
	_, ok := detector.(*SignalDetector)
	assert.True(t, ok, "should return SignalDetector for SignalCapable terminal")
}

// TestNewCompletionDetector_PlainTerminal verifies factory returns ScreenPollDetector for plain terminal.
func TestNewCompletionDetector_PlainTerminal(t *testing.T) {
	t.Parallel()
	mock := newPlainMock()
	detector := NewCompletionDetector(mock)
	_, ok := detector.(*ScreenPollDetector)
	assert.True(t, ok, "should return ScreenPollDetector for plain terminal")
}

// signalCountMock extends signalMock with counting ReadScreen like countingScreenMock.
type signalCountMock struct {
	signalMock
	callCount int
	outputs   []string
}

func (m *signalCountMock) ReadScreen(_ context.Context, _ terminal.PaneID, _ terminal.ReadScreenOpts) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readScreenCalls++
	m.callCount++
	if len(m.outputs) == 0 {
		return m.readScreenOutput, m.readScreenErr
	}
	idx := (m.callCount - 1) % len(m.outputs)
	return m.outputs[idx], nil
}
