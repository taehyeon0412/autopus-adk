// Package terminal provides exec abstraction for testability.
package terminal

import (
	"context"
	"os/exec"
)

// execCommand is a mockable function variable for creating exec.Cmd instances.
// Tests can replace this variable to intercept terminal commands.
// @AX:WARN [AUTO] global state mutation — execCommand is a mutable package-level variable replaced by tests
// @AX:REASON: concurrent test execution may cause data races when multiple tests replace this variable simultaneously; use t.Parallel() guards or per-instance injection
var execCommand = func(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

// execCommandContext is a mockable function variable for creating context-aware exec.Cmd instances.
// Used by WaitForSignal to respect context cancellation and timeouts.
var execCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}
