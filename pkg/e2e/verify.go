// Package e2e provides user-facing scenario-based E2E test infrastructure.
package e2e

import (
	"fmt"
	"os"
	"strings"
)

// VerifyResult represents the outcome of a single verification check.
type VerifyResult struct {
	Primitive string // e.g., "exit_code(0)"
	Pass      bool
	Message   string // human-readable result (e.g., "expected exit_code(0), got 1")
}

// RunOutput holds captured command output for verification.
type RunOutput struct {
	Stdout   string // captured standard output
	Stderr   string // captured standard error
	ExitCode int    // command exit code
	WorkDir  string // temp directory where command ran
}

// CheckExitCode verifies that the actual exit code matches the expected value.
// Returns PASS if codes match, FAIL with descriptive message if they differ.
func CheckExitCode(expected int, actual int) VerifyResult {
	if expected == actual {
		return VerifyResult{
			Primitive: fmt.Sprintf("exit_code(%d)", expected),
			Pass:      true,
		}
	}
	return VerifyResult{
		Primitive: fmt.Sprintf("exit_code(%d)", expected),
		Pass:      false,
		Message:   fmt.Sprintf("expected exit_code(%d), got %d", expected, actual),
	}
}

// CheckStdoutContains verifies that the given substring is present in stdout.
// Returns PASS if found, FAIL with a descriptive message if not found.
func CheckStdoutContains(substr string, stdout string) VerifyResult {
	if strings.Contains(stdout, substr) {
		return VerifyResult{
			Primitive: fmt.Sprintf("stdout_contains(%q)", substr),
			Pass:      true,
		}
	}
	return VerifyResult{
		Primitive: fmt.Sprintf("stdout_contains(%q)", substr),
		Pass:      false,
		Message:   fmt.Sprintf("expected stdout to contain %q", substr),
	}
}

// CheckStderrEmpty verifies that stderr output is empty.
// Returns PASS if stderr is empty, FAIL with message if it contains data.
func CheckStderrEmpty(stderr string) VerifyResult {
	if stderr == "" {
		return VerifyResult{
			Primitive: "stderr_empty()",
			Pass:      true,
		}
	}
	return VerifyResult{
		Primitive: "stderr_empty()",
		Pass:      false,
		Message:   fmt.Sprintf("expected stderr to be empty, got: %q", stderr),
	}
}

// CheckFileExists verifies that a file exists at the given path.
// Returns PASS if the file exists, FAIL with message if it does not.
func CheckFileExists(path string) VerifyResult {
	_, err := os.Stat(path)
	if err == nil {
		return VerifyResult{
			Primitive: fmt.Sprintf("file_exists(%q)", path),
			Pass:      true,
		}
	}
	return VerifyResult{
		Primitive: fmt.Sprintf("file_exists(%q)", path),
		Pass:      false,
		Message:   fmt.Sprintf("expected file to exist at %q", path),
	}
}

// CheckFileContains verifies that a file at the given path contains the expected substring.
// Returns PASS if the substring is found, FAIL with message if not found or file doesn't exist.
func CheckFileContains(path string, substr string) VerifyResult {
	content, err := os.ReadFile(path)
	if err != nil {
		return VerifyResult{
			Primitive: fmt.Sprintf("file_contains(%q, %q)", path, substr),
			Pass:      false,
			Message:   fmt.Sprintf("failed to read file %q: %v", path, err),
		}
	}

	if strings.Contains(string(content), substr) {
		return VerifyResult{
			Primitive: fmt.Sprintf("file_contains(%q, %q)", path, substr),
			Pass:      true,
		}
	}
	return VerifyResult{
		Primitive: fmt.Sprintf("file_contains(%q, %q)", path, substr),
		Pass:      false,
		Message:   fmt.Sprintf("expected file %q to contain %q", path, substr),
	}
}
