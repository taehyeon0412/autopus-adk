package orchestra

import (
	"context"
	"fmt"
	"sync"

	"github.com/insajin/autopus-adk/pkg/terminal"
)

// mockTerminal implements terminal.Terminal for testing pane runner logic.
type mockTerminal struct {
	mu               sync.Mutex
	name             string
	splitPaneErr     error
	sendCommandErr      error
	sendCommandErrAfter int // error only after N successful calls (0 = always error)
	closeErr         error
	splitPaneCalls   []terminal.Direction
	sendCommandCalls []struct {
		PaneID terminal.PaneID
		Cmd    string
	}
	closeCalls         []string
	nextPaneID         int
	createdPanes       []terminal.PaneID
	readScreenOutput   string   // configurable ReadScreen return value
	readScreenCalls    int      // count ReadScreen calls
	readScreenErr      error    // configurable ReadScreen error
	pipePaneStartCalls int      // count PipePaneStart calls
	pipePaneStopCalls  int      // count PipePaneStop calls
	pipePaneStartFiles []string // output files passed to PipePaneStart
	sendLongTextCalls  []struct {
		PaneID terminal.PaneID
		Text   string
	}
}

func (m *mockTerminal) Name() string { return m.name }

func (m *mockTerminal) CreateWorkspace(_ context.Context, _ string) error {
	return nil
}

func (m *mockTerminal) SplitPane(_ context.Context, dir terminal.Direction) (terminal.PaneID, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.splitPaneCalls = append(m.splitPaneCalls, dir)
	if m.splitPaneErr != nil {
		return "", m.splitPaneErr
	}
	m.nextPaneID++
	id := terminal.PaneID(fmt.Sprintf("pane-%d", m.nextPaneID))
	m.createdPanes = append(m.createdPanes, id)
	return id, nil
}

func (m *mockTerminal) SendCommand(_ context.Context, paneID terminal.PaneID, cmd string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendCommandCalls = append(m.sendCommandCalls, struct {
		PaneID terminal.PaneID
		Cmd    string
	}{paneID, cmd})
	// If sendCommandErrAfter is set, only error after that many calls
	if m.sendCommandErrAfter > 0 && len(m.sendCommandCalls) <= m.sendCommandErrAfter {
		return nil
	}
	return m.sendCommandErr
}

func (m *mockTerminal) SendLongText(_ context.Context, paneID terminal.PaneID, text string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sendLongTextCalls = append(m.sendLongTextCalls, struct {
		PaneID terminal.PaneID
		Text   string
	}{paneID, text})
	return nil
}

func (m *mockTerminal) Notify(_ context.Context, _ string) error {
	return nil
}

func (m *mockTerminal) ReadScreen(_ context.Context, _ terminal.PaneID, _ terminal.ReadScreenOpts) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readScreenCalls++
	return m.readScreenOutput, m.readScreenErr
}

func (m *mockTerminal) PipePaneStart(_ context.Context, _ terminal.PaneID, outputFile string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pipePaneStartCalls++
	m.pipePaneStartFiles = append(m.pipePaneStartFiles, outputFile)
	return nil
}

func (m *mockTerminal) PipePaneStop(_ context.Context, _ terminal.PaneID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.pipePaneStopCalls++
	return nil
}

func (m *mockTerminal) Close(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeCalls = append(m.closeCalls, name)
	return m.closeErr
}

// pipePaneErrorMock embeds mockTerminal but overrides PipePaneStart to return an error.
type pipePaneErrorMock struct {
	mockTerminal
}

func (m *pipePaneErrorMock) PipePaneStart(_ context.Context, _ terminal.PaneID, _ string) error {
	return fmt.Errorf("pipe-pane start error")
}

// sendLongTextErrorMock embeds mockTerminal but overrides SendLongText to return an error.
type sendLongTextErrorMock struct {
	mockTerminal
}

func (m *sendLongTextErrorMock) SendLongText(_ context.Context, _ terminal.PaneID, _ string) error {
	return fmt.Errorf("send long text error")
}

// countingScreenMock embeds mockTerminal but alternates ReadScreen output
// based on call count to simulate screen changes between rounds.
type countingScreenMock struct {
	mockTerminal
	callCount int
	outputs   []string
}

func (m *countingScreenMock) ReadScreen(_ context.Context, _ terminal.PaneID, _ terminal.ReadScreenOpts) (string, error) {
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

func newCmuxMock() *mockTerminal {
	return &mockTerminal{name: "cmux"}
}

func newPlainMock() *mockTerminal {
	return &mockTerminal{name: "plain"}
}
