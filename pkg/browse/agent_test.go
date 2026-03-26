package browse

import (
	"context"
	"os/exec"
	"testing"
)

// TestAgentBackend_Open_ExecutesAgentBrowserOpen verifies that Open calls
// `agent-browser open <url>` and returns a SessionID.
func TestAgentBackend_Open_ExecutesAgentBrowserOpen(t *testing.T) {
	execCommand = mockExecCommandWithOutput("session-001")
	defer func() { execCommand = exec.Command }()
	capturedArgs = nil

	b := &AgentBrowserBackend{}
	sid, err := b.Open(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if string(sid) == "" {
		t.Error("expected non-empty SessionID")
	}
	if capturedArgs[0] != "agent-browser" {
		t.Errorf("expected agent-browser binary, got %q", capturedArgs[0])
	}
	if !containsArg(capturedArgs, "open") {
		t.Errorf("expected 'open' subcommand in args: %v", capturedArgs)
	}
	if !containsArg(capturedArgs, "https://example.com") {
		t.Errorf("expected URL in args: %v", capturedArgs)
	}
}

// TestAgentBackend_Snapshot_ExecutesAgentBrowserSnapshot verifies that Snapshot calls
// `agent-browser snapshot` and returns content.
func TestAgentBackend_Snapshot_ExecutesAgentBrowserSnapshot(t *testing.T) {
	execCommand = mockExecCommandWithOutput("page html content")
	defer func() { execCommand = exec.Command }()
	capturedArgs = nil

	b := &AgentBrowserBackend{}
	content, err := b.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot returned error: %v", err)
	}
	if content == "" {
		t.Error("expected non-empty snapshot content")
	}
	if capturedArgs[0] != "agent-browser" {
		t.Errorf("expected agent-browser binary, got %q", capturedArgs[0])
	}
	if !containsArg(capturedArgs, "snapshot") {
		t.Errorf("expected 'snapshot' subcommand in args: %v", capturedArgs)
	}
}

// TestAgentBackend_Click_ExecutesAgentBrowserClick verifies that Click calls
// `agent-browser click <selector>`.
func TestAgentBackend_Click_ExecutesAgentBrowserClick(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()
	capturedArgs = nil

	b := &AgentBrowserBackend{}
	err := b.Click(context.Background(), "#submit-btn")
	if err != nil {
		t.Fatalf("Click returned error: %v", err)
	}
	if capturedArgs[0] != "agent-browser" {
		t.Errorf("expected agent-browser binary, got %q", capturedArgs[0])
	}
	if !containsArg(capturedArgs, "click") {
		t.Errorf("expected 'click' subcommand in args: %v", capturedArgs)
	}
	if !containsArg(capturedArgs, "#submit-btn") {
		t.Errorf("expected selector in args: %v", capturedArgs)
	}
}

// TestAgentBackend_Fill_ExecutesAgentBrowserFill verifies that Fill calls
// `agent-browser fill <selector> <text>`.
func TestAgentBackend_Fill_ExecutesAgentBrowserFill(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()
	capturedArgs = nil

	b := &AgentBrowserBackend{}
	err := b.Fill(context.Background(), "#email", "user@example.com")
	if err != nil {
		t.Fatalf("Fill returned error: %v", err)
	}
	if capturedArgs[0] != "agent-browser" {
		t.Errorf("expected agent-browser binary, got %q", capturedArgs[0])
	}
	if !containsArg(capturedArgs, "fill") {
		t.Errorf("expected 'fill' subcommand in args: %v", capturedArgs)
	}
	if !containsArg(capturedArgs, "#email") {
		t.Errorf("expected selector in args: %v", capturedArgs)
	}
	if !containsArg(capturedArgs, "user@example.com") {
		t.Errorf("expected text in args: %v", capturedArgs)
	}
}

// TestAgentBackend_Screenshot_ExecutesAgentBrowserScreenshot verifies that Screenshot calls
// `agent-browser screenshot <path>`.
func TestAgentBackend_Screenshot_ExecutesAgentBrowserScreenshot(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()
	capturedArgs = nil

	b := &AgentBrowserBackend{}
	err := b.Screenshot(context.Background(), "/tmp/agent-shot.png")
	if err != nil {
		t.Fatalf("Screenshot returned error: %v", err)
	}
	if capturedArgs[0] != "agent-browser" {
		t.Errorf("expected agent-browser binary, got %q", capturedArgs[0])
	}
	if !containsArg(capturedArgs, "screenshot") {
		t.Errorf("expected 'screenshot' subcommand in args: %v", capturedArgs)
	}
	if !containsArg(capturedArgs, "/tmp/agent-shot.png") {
		t.Errorf("expected output path in args: %v", capturedArgs)
	}
}

// TestAgentBackend_Close verifies that Close is a no-op and returns nil.
func TestAgentBackend_Close(t *testing.T) {
	b := &AgentBrowserBackend{}
	err := b.Close(context.Background())
	if err != nil {
		t.Errorf("expected Close to return nil, got %v", err)
	}
}

// TestAgentBackend_Name_ReturnsAgentBrowser verifies that Name() returns "agent-browser".
func TestAgentBackend_Name_ReturnsAgentBrowser(t *testing.T) {
	b := &AgentBrowserBackend{}
	if b.Name() != "agent-browser" {
		t.Errorf("expected 'agent-browser', got %q", b.Name())
	}
}
