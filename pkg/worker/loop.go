package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/insajin/autopus-adk/pkg/worker/a2a"
	"github.com/insajin/autopus-adk/pkg/worker/adapter"
)

// LoopConfig holds configuration for the WorkerLoop.
type LoopConfig struct {
	BackendURL string
	WorkerName string
	Skills     []string
	Provider   adapter.ProviderAdapter
	MCPConfig  string // path to worker-mcp.json
	WorkDir    string // working directory for subprocesses
	AuthToken  string // bearer token for backend auth
}

// WorkerLoop integrates A2A Server, ProviderAdapter, ContextBuilder, and StreamParser.
// It receives tasks via A2A, builds prompts, spawns CLI subprocesses, and reports results.
type WorkerLoop struct {
	config  LoopConfig
	server  *a2a.Server
	builder ContextBuilder
}

// NewWorkerLoop creates a WorkerLoop with the given configuration.
func NewWorkerLoop(config LoopConfig) *WorkerLoop {
	wl := &WorkerLoop{
		config: config,
	}

	serverCfg := a2a.ServerConfig{
		BackendURL: config.BackendURL,
		WorkerName: config.WorkerName,
		Skills:     config.Skills,
		Handler:    wl.handleTask,
		AuthToken:  config.AuthToken,
	}
	wl.server = a2a.NewServer(serverCfg)

	return wl
}

// Start connects to the backend and begins processing tasks.
func (wl *WorkerLoop) Start(ctx context.Context) error {
	log.Printf("[worker] starting loop: provider=%s backend=%s", wl.config.Provider.Name(), wl.config.BackendURL)
	return wl.server.Start(ctx)
}

// Close shuts down the worker loop and its A2A server.
func (wl *WorkerLoop) Close() error {
	return wl.server.Close()
}

// taskPayloadMessage is the JSON structure received from the A2A backend.
type taskPayloadMessage struct {
	Description   string `json:"description"`
	PMNotes       string `json:"pm_notes,omitempty"`
	PolicySummary string `json:"policy_summary,omitempty"`
	KnowledgeCtx  string `json:"knowledge_ctx,omitempty"`
	SpecID        string `json:"spec_id,omitempty"`
	SessionID     string `json:"session_id,omitempty"`
}

// handleTask is the A2A TaskHandler callback invoked when a task is received.
func (wl *WorkerLoop) handleTask(ctx context.Context, taskID string, payload json.RawMessage) (*a2a.TaskResult, error) {
	log.Printf("[worker] received task: %s", taskID)

	// Clean up cached SecurityPolicy file on task completion (success or failure).
	defer cleanupPolicy(taskID)

	var msg taskPayloadMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		return nil, fmt.Errorf("parse task payload: %w", err)
	}

	// Build Layer 4 prompt via ContextBuilder.
	prompt := wl.builder.Build(TaskPayload{
		TaskID:        taskID,
		Description:   msg.Description,
		PMNotes:       msg.PMNotes,
		PolicySummary: msg.PolicySummary,
		KnowledgeCtx:  msg.KnowledgeCtx,
		SpecID:        msg.SpecID,
	})

	// Configure the subprocess task.
	taskCfg := adapter.TaskConfig{
		TaskID:    taskID,
		SessionID: msg.SessionID,
		Prompt:    prompt,
		MCPConfig: wl.config.MCPConfig,
		WorkDir:   wl.config.WorkDir,
	}

	// Execute subprocess and parse stream output.
	result, err := wl.executeSubprocess(ctx, taskCfg)
	if err != nil {
		log.Printf("[worker] task %s failed: %v", taskID, err)
		return nil, err
	}

	log.Printf("[worker] task %s completed: cost=$%.4f duration=%dms", taskID, result.CostUSD, result.DurationMS)

	return &a2a.TaskResult{
		Status:    a2a.StatusCompleted,
		Artifacts: convertArtifacts(result.Artifacts),
	}, nil
}

// cleanupPolicy removes the cached SecurityPolicy file for the given task.
func cleanupPolicy(taskID string) {
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("autopus-%d", os.Getuid()))
	path := filepath.Join(dir, fmt.Sprintf("autopus-policy-%s.json", taskID))
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		log.Printf("[worker] cleanup policy file: %v", err)
	}
}

// convertArtifacts converts adapter artifacts to A2A artifacts.
func convertArtifacts(src []adapter.Artifact) []a2a.Artifact {
	if len(src) == 0 {
		return nil
	}
	out := make([]a2a.Artifact, len(src))
	for i, a := range src {
		out[i] = a2a.Artifact{
			Name:     a.Name,
			MimeType: a.MimeType,
			Data:     a.Data,
		}
	}
	return out
}
