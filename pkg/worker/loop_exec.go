package worker

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/insajin/autopus-adk/pkg/worker/adapter"
	"github.com/insajin/autopus-adk/pkg/worker/stream"
)

// executeSubprocess spawns the provider CLI, pipes the prompt via stdin,
// and reads stdout line-by-line through StreamParser.
func (wl *WorkerLoop) executeSubprocess(ctx context.Context, taskCfg adapter.TaskConfig) (adapter.TaskResult, error) {
	cmd := wl.config.Provider.BuildCommand(ctx, taskCfg)

	// Set up stdin pipe for prompt injection (Layer 4).
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return adapter.TaskResult{}, fmt.Errorf("stdin pipe: %w", err)
	}

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

	// Write prompt to stdin and close to signal EOF.
	taskID := taskCfg.TaskID
	go func() {
		defer stdin.Close()
		if _, err := io.Copy(stdin, strings.NewReader(taskCfg.Prompt)); err != nil {
			log.Printf("[worker] task %s: failed to write prompt to stdin: %v", taskID, err)
		}
	}()

	// Parse stream output.
	result, parseErr := wl.parseStream(stdout, taskCfg.TaskID)

	// Wait for process to exit.
	waitErr := cmd.Wait()
	if parseErr != nil {
		return adapter.TaskResult{}, fmt.Errorf("parse stream: %w", parseErr)
	}
	if waitErr != nil {
		// Non-zero exit may still have produced a result.
		if result.Output != "" {
			log.Printf("[worker] task %s: subprocess exited with error but produced output: %v", taskCfg.TaskID, waitErr)
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

// parseStream reads the subprocess stdout line-by-line and extracts the final result.
func (wl *WorkerLoop) parseStream(r io.Reader, taskID string) (adapter.TaskResult, error) {
	parser := stream.NewParser(r)
	var lastResult adapter.TaskResult
	hasResult := false

	for {
		evt, err := parser.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return adapter.TaskResult{}, err
		}

		// Convert stream.Event to adapter.StreamEvent for provider parsing.
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
			log.Printf("[worker] task %s: subagent notification received", taskID)

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
