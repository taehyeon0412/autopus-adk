package orchestra

import (
	"context"
	"fmt"

	"github.com/insajin/autopus-adk/pkg/terminal"
)

// mockTerminal implements terminal.Terminal for testing pane runner logic.
type mockTerminal struct {
	name             string
	splitPaneErr     error
	sendCommandErr   error
	closeErr         error
	splitPaneCalls   []terminal.Direction
	sendCommandCalls []struct {
		PaneID terminal.PaneID
		Cmd    string
	}
	closeCalls   []string
	nextPaneID   int
	createdPanes []terminal.PaneID
}

func (m *mockTerminal) Name() string { return m.name }

func (m *mockTerminal) CreateWorkspace(_ context.Context, _ string) error {
	return nil
}

func (m *mockTerminal) SplitPane(_ context.Context, dir terminal.Direction) (terminal.PaneID, error) {
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
	m.sendCommandCalls = append(m.sendCommandCalls, struct {
		PaneID terminal.PaneID
		Cmd    string
	}{paneID, cmd})
	return m.sendCommandErr
}

func (m *mockTerminal) Notify(_ context.Context, _ string) error {
	return nil
}

func (m *mockTerminal) Close(_ context.Context, name string) error {
	m.closeCalls = append(m.closeCalls, name)
	return m.closeErr
}

func newCmuxMock() *mockTerminal {
	return &mockTerminal{name: "cmux"}
}

func newPlainMock() *mockTerminal {
	return &mockTerminal{name: "plain"}
}
