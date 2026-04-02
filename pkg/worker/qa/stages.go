// Package qa provides QA pipeline stages for build, test, and health checks.
package qa

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// Stage represents a single step in a QA pipeline.
type Stage interface {
	Name() string
	Run(ctx context.Context, workDir string) (*StageResult, error)
}

// StageResult holds the outcome of a single stage execution.
type StageResult struct {
	Name     string        `json:"name"`
	Status   string        `json:"status"` // "pass", "fail", "skip"
	Output   string        `json:"output"`
	Duration time.Duration `json:"duration"`
}

// allowedPrefixes defines the command prefixes permitted for execution.
var allowedPrefixes = []string{"go ", "npm ", "make ", "make\n", "docker "}

// validateCommand checks that a command starts with an allowed prefix.
func validateCommand(cmd string) error {
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(cmd, prefix) {
			return nil
		}
	}
	// Also allow bare "make" with no arguments.
	if cmd == "make" {
		return nil
	}
	return fmt.Errorf("command %q not in allowlist (allowed: go, npm, make, docker)", cmd)
}

// runCommand executes a shell command in the given directory.
func runCommand(ctx context.Context, workDir, cmd string) (string, error) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}
	c := exec.CommandContext(ctx, parts[0], parts[1:]...)
	c.Dir = workDir
	out, err := c.CombinedOutput()
	return string(out), err
}

// BuildStage runs a build command (e.g., "go build ./...").
type BuildStage struct {
	Command string
}

func (s *BuildStage) Name() string { return "build" }

func (s *BuildStage) Run(ctx context.Context, workDir string) (*StageResult, error) {
	if err := validateCommand(s.Command); err != nil {
		return &StageResult{Name: s.Name(), Status: "fail", Output: err.Error()}, err
	}
	start := time.Now()
	out, err := runCommand(ctx, workDir, s.Command)
	dur := time.Since(start)
	if err != nil {
		return &StageResult{Name: s.Name(), Status: "fail", Output: out, Duration: dur}, err
	}
	return &StageResult{Name: s.Name(), Status: "pass", Output: out, Duration: dur}, nil
}

// TestStage runs a test command (e.g., "go test ./...").
type TestStage struct {
	Command string
}

func (s *TestStage) Name() string { return "test" }

func (s *TestStage) Run(ctx context.Context, workDir string) (*StageResult, error) {
	if err := validateCommand(s.Command); err != nil {
		return &StageResult{Name: s.Name(), Status: "fail", Output: err.Error()}, err
	}
	start := time.Now()
	out, err := runCommand(ctx, workDir, s.Command)
	dur := time.Since(start)
	if err != nil {
		return &StageResult{Name: s.Name(), Status: "fail", Output: out, Duration: dur}, err
	}
	return &StageResult{Name: s.Name(), Status: "pass", Output: out, Duration: dur}, nil
}

// ServiceHealthStage polls a health endpoint until it responds 200 OK.
type ServiceHealthStage struct {
	URL      string
	Interval time.Duration // default 500ms
	Timeout  time.Duration // default 30s
}

func (s *ServiceHealthStage) Name() string { return "service-health" }

func (s *ServiceHealthStage) Run(ctx context.Context, _ string) (*StageResult, error) {
	interval := s.Interval
	if interval == 0 {
		interval = 500 * time.Millisecond
	}
	timeout := s.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	start := time.Now()
	deadline := start.Add(timeout)
	client := &http.Client{Timeout: 5 * time.Second}

	for {
		if time.Now().After(deadline) {
			dur := time.Since(start)
			return &StageResult{
				Name: s.Name(), Status: "fail",
				Output: fmt.Sprintf("health check timed out after %s", timeout), Duration: dur,
			}, fmt.Errorf("health check timeout: %s", s.URL)
		}

		resp, err := client.Get(s.URL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				dur := time.Since(start)
				return &StageResult{
					Name: s.Name(), Status: "pass",
					Output: "health check passed", Duration: dur,
				}, nil
			}
		}

		select {
		case <-ctx.Done():
			dur := time.Since(start)
			return &StageResult{
				Name: s.Name(), Status: "fail",
				Output: "context cancelled", Duration: dur,
			}, ctx.Err()
		case <-time.After(interval):
		}
	}
}

// CleanupStage runs cleanup commands. It always executes, even after failures.
type CleanupStage struct {
	Commands []string
}

func (s *CleanupStage) Name() string { return "cleanup" }

func (s *CleanupStage) Run(ctx context.Context, workDir string) (*StageResult, error) {
	start := time.Now()
	var outputs []string
	var lastErr error

	for _, cmd := range s.Commands {
		if err := validateCommand(cmd); err != nil {
			outputs = append(outputs, fmt.Sprintf("[%s] blocked: %v", cmd, err))
			lastErr = err
			continue
		}
		out, err := runCommand(ctx, workDir, cmd)
		if err != nil {
			outputs = append(outputs, fmt.Sprintf("[%s] error: %s", cmd, out))
			lastErr = err
		} else {
			outputs = append(outputs, fmt.Sprintf("[%s] ok", cmd))
		}
	}

	dur := time.Since(start)
	status := "pass"
	if lastErr != nil {
		status = "fail"
	}
	return &StageResult{
		Name: s.Name(), Status: status,
		Output: strings.Join(outputs, "\n"), Duration: dur,
	}, lastErr
}
