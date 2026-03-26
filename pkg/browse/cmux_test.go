package browse

import (
	"context"
	"os/exec"
	"strings"
	"testing"
)

// capturedArgs holds the arguments passed to the mocked exec command.
var capturedArgs []string

// mockExecCommand replaces execCommand in tests to capture CLI invocations.
func mockExecCommand(name string, args ...string) *exec.Cmd {
	capturedArgs = append([]string{name}, args...)
	// Return a no-op command that exits successfully.
	return exec.Command("true")
}

// mockExecCommandWithOutput returns a command that prints the given output to stdout.
func mockExecCommandWithOutput(output string) func(string, ...string) *exec.Cmd {
	return func(name string, args ...string) *exec.Cmd {
		capturedArgs = append([]string{name}, args...)
		return exec.Command("echo", output)
	}
}

// TestCmuxBackend_Open_ExecutesCmuxBrowserOpen verifies that Open calls
// `cmux browser open <url>` and parses the returned surface ref.
func TestCmuxBackend_Open_ExecutesCmuxBrowserOpen(t *testing.T) {
	execCommand = mockExecCommandWithOutput("surf-abc123")
	defer func() { execCommand = exec.Command }()
	capturedArgs = nil

	b := &CmuxBrowserBackend{}
	sid, err := b.Open(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if string(sid) == "" {
		t.Error("expected non-empty SessionID")
	}
	if len(capturedArgs) < 3 {
		t.Fatalf("expected at least 3 args, got %v", capturedArgs)
	}
	if capturedArgs[0] != "cmux" {
		t.Errorf("expected cmux binary, got %q", capturedArgs[0])
	}
	if capturedArgs[1] != "browser" || capturedArgs[2] != "open" {
		t.Errorf("expected 'browser open' subcommand, got %v", capturedArgs[1:])
	}
	found := false
	for _, a := range capturedArgs {
		if a == "https://example.com" {
			found = true
		}
	}
	if !found {
		t.Errorf("URL not found in args: %v", capturedArgs)
	}
}

// TestCmuxBackend_Snapshot_ExecutesCmuxBrowserSnapshot verifies that Snapshot calls
// `cmux browser --surface <ref> snapshot`.
func TestCmuxBackend_Snapshot_ExecutesCmuxBrowserSnapshot(t *testing.T) {
	execCommand = mockExecCommandWithOutput("page content")
	defer func() { execCommand = exec.Command }()
	capturedArgs = nil

	b := &CmuxBrowserBackend{surfaceRef: "surf-abc123"}
	content, err := b.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot returned error: %v", err)
	}
	if content == "" {
		t.Error("expected non-empty snapshot content")
	}
	if !strings.Contains(strings.Join(capturedArgs, " "), "--surface surf-abc123") {
		t.Errorf("expected --surface flag in args: %v", capturedArgs)
	}
	if !containsArg(capturedArgs, "snapshot") {
		t.Errorf("expected 'snapshot' subcommand in args: %v", capturedArgs)
	}
}

// TestCmuxBackend_Click_ExecutesCmuxBrowserClick verifies that Click calls
// `cmux browser --surface <ref> click <selector>`.
func TestCmuxBackend_Click_ExecutesCmuxBrowserClick(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()
	capturedArgs = nil

	b := &CmuxBrowserBackend{surfaceRef: "surf-abc123"}
	err := b.Click(context.Background(), "#submit-btn")
	if err != nil {
		t.Fatalf("Click returned error: %v", err)
	}
	if !containsArg(capturedArgs, "click") {
		t.Errorf("expected 'click' subcommand in args: %v", capturedArgs)
	}
	if !containsArg(capturedArgs, "#submit-btn") {
		t.Errorf("expected selector in args: %v", capturedArgs)
	}
	if !strings.Contains(strings.Join(capturedArgs, " "), "--surface surf-abc123") {
		t.Errorf("expected --surface flag in args: %v", capturedArgs)
	}
}

// TestCmuxBackend_Fill_ExecutesCmuxBrowserFill verifies that Fill calls
// `cmux browser --surface <ref> fill <selector> <text>`.
func TestCmuxBackend_Fill_ExecutesCmuxBrowserFill(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()
	capturedArgs = nil

	b := &CmuxBrowserBackend{surfaceRef: "surf-abc123"}
	err := b.Fill(context.Background(), "#username", "testuser")
	if err != nil {
		t.Fatalf("Fill returned error: %v", err)
	}
	if !containsArg(capturedArgs, "fill") {
		t.Errorf("expected 'fill' subcommand in args: %v", capturedArgs)
	}
	if !containsArg(capturedArgs, "#username") {
		t.Errorf("expected selector in args: %v", capturedArgs)
	}
	if !containsArg(capturedArgs, "testuser") {
		t.Errorf("expected text in args: %v", capturedArgs)
	}
}

// TestCmuxBackend_Screenshot_ExecutesCmuxBrowserScreenshot verifies that Screenshot calls
// `cmux browser --surface <ref> screenshot --out <path>`.
func TestCmuxBackend_Screenshot_ExecutesCmuxBrowserScreenshot(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()
	capturedArgs = nil

	b := &CmuxBrowserBackend{surfaceRef: "surf-abc123"}
	err := b.Screenshot(context.Background(), "/tmp/shot.png")
	if err != nil {
		t.Fatalf("Screenshot returned error: %v", err)
	}
	if !containsArg(capturedArgs, "screenshot") {
		t.Errorf("expected 'screenshot' subcommand in args: %v", capturedArgs)
	}
	if !containsArg(capturedArgs, "--out") {
		t.Errorf("expected --out flag in args: %v", capturedArgs)
	}
	if !containsArg(capturedArgs, "/tmp/shot.png") {
		t.Errorf("expected output path in args: %v", capturedArgs)
	}
}

// TestCmuxBackend_Close_ExecutesCmuxCloseSurface verifies that Close calls
// `cmux close-surface --surface <ref>`.
func TestCmuxBackend_Close_ExecutesCmuxCloseSurface(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()
	capturedArgs = nil

	b := &CmuxBrowserBackend{surfaceRef: "surf-abc123"}
	err := b.Close(context.Background())
	if err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	if capturedArgs[0] != "cmux" {
		t.Errorf("expected cmux binary, got %q", capturedArgs[0])
	}
	if !containsArg(capturedArgs, "close-surface") {
		t.Errorf("expected 'close-surface' subcommand in args: %v", capturedArgs)
	}
	if !strings.Contains(strings.Join(capturedArgs, " "), "--surface surf-abc123") {
		t.Errorf("expected --surface flag in args: %v", capturedArgs)
	}
}

// TestCmuxBackend_Name_ReturnsCmux verifies that Name() returns "cmux".
func TestCmuxBackend_Name_ReturnsCmux(t *testing.T) {
	b := &CmuxBrowserBackend{}
	if b.Name() != "cmux" {
		t.Errorf("expected 'cmux', got %q", b.Name())
	}
}

// TestCmuxBackend_Open_ShellEscapesURL verifies that URLs with special characters
// are safely passed without shell injection.
func TestCmuxBackend_Open_ShellEscapesURL(t *testing.T) {
	execCommand = mockExecCommandWithOutput("surf-xyz")
	defer func() { execCommand = exec.Command }()
	capturedArgs = nil

	b := &CmuxBrowserBackend{}
	specialURL := "https://example.com/path?q=hello world&foo=bar"
	_, err := b.Open(context.Background(), specialURL)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	// The URL must appear as a single argument (not split by shell).
	found := false
	for _, a := range capturedArgs {
		if a == specialURL {
			found = true
		}
	}
	if !found {
		t.Errorf("special URL not passed as single arg; capturedArgs: %v", capturedArgs)
	}
}

// containsArg returns true if needle is present in the args slice.
func containsArg(args []string, needle string) bool {
	for _, a := range args {
		if a == needle {
			return true
		}
	}
	return false
}
