package worker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/insajin/autopus-adk/pkg/worker/adapter"
	"github.com/insajin/autopus-adk/pkg/worker/budget"
	"github.com/insajin/autopus-adk/pkg/worker/security"
	"github.com/insajin/autopus-adk/pkg/worker/stream"
)

// StdinWriter wraps an io.WriteCloser to keep the stdin pipe open
// after the initial prompt is written. This enables mid-session
// message injection (e.g., budget warnings).
type StdinWriter struct {
	pipe io.WriteCloser
}

// NewStdinWriter creates a StdinWriter wrapping the given pipe.
func NewStdinWriter(pipe io.WriteCloser) *StdinWriter {
	return &StdinWriter{pipe: pipe}
}

// WritePrompt sends the initial prompt to the subprocess stdin.
// Unlike the previous implementation, the pipe is NOT closed after writing.
func (sw *StdinWriter) WritePrompt(prompt string) error {
	_, err := io.Copy(sw.pipe, strings.NewReader(prompt))
	return err
}

// Write implements io.Writer for injecting messages into stdin.
func (sw *StdinWriter) Write(p []byte) (int, error) {
	return sw.pipe.Write(p)
}

// Close closes the underlying pipe.
func (sw *StdinWriter) Close() error {
	return sw.pipe.Close()
}

// BudgetConfig holds optional budget configuration for subprocess execution.
type BudgetConfig struct {
	Budget        budget.IterationBudget
	EmergencyStop *security.EmergencyStop
}

// executeSubprocess spawns the provider CLI, pipes the prompt via stdin,
// and reads stdout line-by-line through StreamParser.
func (wl *WorkerLoop) executeSubprocess(ctx context.Context, taskCfg adapter.TaskConfig) (adapter.TaskResult, error) {
	return wl.executeWithBudget(ctx, taskCfg, nil)
}

// executeWithBudget spawns the provider CLI with optional budget tracking.
// @WARN: When budget is exhausted, EmergencyStop.Stop() is called.
func (wl *WorkerLoop) executeWithBudget(ctx context.Context, taskCfg adapter.TaskConfig, bc *BudgetConfig) (adapter.TaskResult, error) {
	cmd := wl.config.Provider.BuildCommand(ctx, taskCfg)

	// Set up stdin pipe — kept open for budget warning injection (REQ-BUDGET-03).
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return adapter.TaskResult{}, fmt.Errorf("stdin pipe: %w", err)
	}
	sw := NewStdinWriter(stdinPipe)

	// Set up stdout pipe for stream parsing.
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return adapter.TaskResult{}, fmt.Errorf("stdout pipe: %w", err)
	}

	// Capture stderr for diagnostics on non-zero exit (REQ-SUB-H01).
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return adapter.TaskResult{}, fmt.Errorf("start subprocess: %w", err)
	}

	// Register for emergency stop if budget config provided.
	if bc != nil && bc.EmergencyStop != nil {
		bc.EmergencyStop.SetProcess(cmd)
		defer bc.EmergencyStop.ClearProcess()
	}

	// Write prompt — stdin stays open for warning injection.
	taskID := taskCfg.TaskID
	if err := sw.WritePrompt(taskCfg.Prompt); err != nil {
		log.Printf("[worker] task %s: failed to write prompt: %v", taskID, err)
	}

	// Parse stream output with optional budget tracking.
	result, parseErr := wl.parseStreamWithBudget(stdout, taskID, sw, bc)

	// Close stdin after stream parsing completes.
	sw.Close()

	// Wait for process to exit.
	waitErr := cmd.Wait()
	if parseErr != nil {
		return adapter.TaskResult{}, fmt.Errorf("parse stream: %w", parseErr)
	}
	if waitErr != nil {
		if result.Output != "" {
			log.Printf("[worker] task %s: exited with error but produced output: %v", taskID, waitErr)
			return result, nil
		}
		stderrStr := strings.TrimSpace(stderrBuf.String())
		if stderrStr != "" {
			return adapter.TaskResult{}, fmt.Errorf("subprocess exit: %w\nstderr: %s", waitErr, stderrStr)
		}
		return adapter.TaskResult{}, fmt.Errorf("subprocess exit: %w", waitErr)
	}

	return result, nil
}

// parseStream reads subprocess stdout and extracts the final result (no budget).
func (wl *WorkerLoop) parseStream(r io.Reader, taskID string) (adapter.TaskResult, error) {
	return wl.parseStreamWithBudget(r, taskID, nil, nil)
}

// parseStreamWithBudget extends parseStream with tool call counting and warnings.
func (wl *WorkerLoop) parseStreamWithBudget(r io.Reader, taskID string, sw *StdinWriter, bc *BudgetConfig) (adapter.TaskResult, error) {
	parser := stream.NewParser(r)
	var lastResult adapter.TaskResult
	hasResult := false

	// Set up budget counter and warning injector if configured.
	var counter *budget.Counter
	var injector *budget.WarningInjector
	if bc != nil {
		counter = budget.NewCounter(bc.Budget)
		if sw != nil {
			injector = budget.NewWarningInjector(sw)
		}
	}

	for {
		evt, err := parser.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return adapter.TaskResult{}, err
		}

		adapterEvt := adapter.StreamEvent{
			Type:    evt.Type,
			Subtype: evt.Subtype,
			Data:    evt.Raw,
		}

		switch {
		case evt.Type == "system" && evt.Subtype == "init":
			log.Printf("[worker] task %s: subprocess initialized", taskID)

		case evt.Type == "system" && evt.Subtype == "task_started":
			log.Printf("[worker] task %s: subagent started", taskID)

		case evt.Type == "system" && evt.Subtype == "task_progress":
			log.Printf("[worker] task %s: progress update", taskID)

		case evt.Type == "system" && evt.Subtype == "task_notification":
			log.Printf("[worker] task %s: subagent notification", taskID)

		// REQ-BUDGET-01/05: Count tool_call and tool_use events.
		case evt.Type == stream.EventToolCall || evt.Type == "tool_use":
			if counter != nil {
				r := counter.Increment()
				log.Printf("[worker] task %s: tool call %d/%d", taskID, r.Count, r.Budget.Limit)

				// REQ-BUDGET-06/07: Inject warnings on threshold change.
				if injector != nil {
					injector.Inject(r)
				}

				// REQ-BUDGET-08: Hard limit — emergency stop.
				if r.Level == budget.LevelExhausted && bc.EmergencyStop != nil {
					log.Printf("[worker] task %s: budget exhausted, stopping", taskID)
					_ = bc.EmergencyStop.Stop("iteration_budget_exceeded")
					return lastResult, fmt.Errorf("iteration budget exceeded: %d/%d", r.Count, r.Budget.Limit)
				}
			}

		case evt.Type == "result":
			lastResult = wl.config.Provider.ExtractResult(adapterEvt)
			hasResult = true

		case evt.Type == "error":
			log.Printf("[worker] task %s: error event received", taskID)
		}
	}

	if !hasResult {
		return adapter.TaskResult{}, fmt.Errorf("no result event received for task %s", taskID)
	}
	return lastResult, nil
}
