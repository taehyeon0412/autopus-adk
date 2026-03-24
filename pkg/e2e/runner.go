// Package e2e provides user-facing scenario-based E2E test infrastructure.
package e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

// RunnerOptions configures the scenario runner.
type RunnerOptions struct {
	ProjectDir   string        // project root directory
	AutoBuild    bool          // auto-build binary before first run
	BuildCommand string        // build command (e.g., "go build -o auto .")
	Timeout      time.Duration // per-scenario timeout (default: 30s)
}

// RunnerResult holds the result of a scenario execution.
type RunnerResult struct {
	ExitCode       int    // command exit code
	Stdout         string // captured stdout
	Stderr         string // captured stderr
	Pass           bool   // true if all verify primitives passed
	FailureDetails string // human-readable failure description
	TimedOut       bool   // true if command exceeded timeout
	BuildOccurred  bool   // true if auto-build was triggered this run
	WorkDir        string // temp directory used for this run
}

// WorkDirExists checks if the work directory still exists.
// Returns an error if the directory has been removed (expected after cleanup).
func (r *RunnerResult) WorkDirExists() (bool, error) {
	_, err := os.Stat(r.WorkDir)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Runner executes E2E scenarios.
type Runner struct {
	opts      RunnerOptions
	buildOnce sync.Once
	buildErr  error
}

// @AX:NOTE [AUTO] @AX:REASON: magic constant — 30s default per-scenario timeout; adjust for slow integration tests or network-dependent scenarios
// NewRunner creates a new Runner with the given options.
func NewRunner(opts RunnerOptions) *Runner {
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}
	return &Runner{opts: opts}
}

// @AX:NOTE [AUTO] @AX:REASON: public API boundary — primary scenario execution entry point; fan_in=1 (internal/cli/test.go); sync.Once ensures build runs exactly once per Runner instance
// Run executes a single scenario and returns the result.
func (r *Runner) Run(scenario Scenario) (*RunnerResult, error) {
	result := &RunnerResult{}

	// @AX:WARN [AUTO] @AX:REASON: sync.Once closure — buildErr assigned inside Do() without mutex; safe only because buildOnce guarantees single execution; do not add concurrent access to buildErr outside Do()
	// Auto-build if enabled (only once per Runner instance).
	if r.opts.AutoBuild && r.opts.BuildCommand != "" {
		builtThisCall := false
		r.buildOnce.Do(func() {
			cmd := exec.Command("sh", "-c", r.opts.BuildCommand)
			cmd.Dir = r.opts.ProjectDir
			if err := cmd.Run(); err != nil {
				r.buildErr = err
			}
			builtThisCall = true
		})
		if r.buildErr != nil {
			return nil, fmt.Errorf("auto-build failed: %w", r.buildErr)
		}
		result.BuildOccurred = builtThisCall
	}

	// Create isolated temp directory for this run.
	workDir, err := os.MkdirTemp("", "e2e-run-*")
	if err != nil {
		return nil, fmt.Errorf("create work dir: %w", err)
	}
	result.WorkDir = workDir
	defer func() {
		_ = os.RemoveAll(workDir)
	}()

	// Execute command with timeout.
	ctx, cancel := context.WithTimeout(context.Background(), r.opts.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", scenario.Command)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.Pass = false
		result.ExitCode = -1
		return result, nil
	}

	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
	} else {
		result.ExitCode = 0
	}

	// Evaluate verify primitives.
	allPass := true
	for _, primitive := range scenario.Verify {
		vr := evaluatePrimitive(primitive, result)
		if !vr.Pass {
			allPass = false
			result.FailureDetails += vr.Message + "\n"
		}
	}
	result.Pass = allPass

	return result, nil
}

// evaluatePrimitive evaluates a single verification primitive string against the result.
func evaluatePrimitive(primitive string, result *RunnerResult) VerifyResult {
	switch {
	case primitive == "exit_code(0)":
		return CheckExitCode(0, result.ExitCode)
	case len(primitive) > 12 && primitive[:11] == "exit_code(":
		var code int
		_, _ = fmt.Sscanf(primitive, "exit_code(%d)", &code)
		return CheckExitCode(code, result.ExitCode)
	case primitive == "stderr_empty()":
		return CheckStderrEmpty(result.Stderr)
	default:
		// Return pass for unknown primitives (other agents will implement them).
		return VerifyResult{Primitive: primitive, Pass: true}
	}
}
