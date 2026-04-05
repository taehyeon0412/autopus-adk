package spec

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// GolangciIssue represents a single issue from golangci-lint JSON output.
type GolangciIssue struct {
	FromLinter string `json:"FromLinter"`
	Text       string `json:"Text"`
	Pos        struct {
		Filename string `json:"Filename"`
		Line     int    `json:"Line"`
		Column   int    `json:"Column"`
	} `json:"Pos"`
	Severity string `json:"Severity"`
}

// GolangciOutput is the top-level golangci-lint JSON output structure.
type GolangciOutput struct {
	Issues []GolangciIssue `json:"Issues"`
}

// ParseGolangciOutput parses golangci-lint --out-format json output into ReviewFindings.
func ParseGolangciOutput(data []byte) ([]ReviewFinding, error) {
	var output GolangciOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, fmt.Errorf("parse golangci output: %w", err)
	}

	findings := make([]ReviewFinding, 0, len(output.Issues))
	for _, issue := range output.Issues {
		scopeRef := fmt.Sprintf("%s:%d", issue.Pos.Filename, issue.Pos.Line)
		description := fmt.Sprintf("%s: %s", issue.FromLinter, issue.Text)
		findings = append(findings, ReviewFinding{
			Category:    FindingCategoryStyle,
			ScopeRef:    scopeRef,
			Description: description,
			Severity:    issue.Severity,
			Provider:    "golangci-lint",
		})
	}
	return findings, nil
}

// RunStaticAnalysis runs a static analysis tool and returns findings.
// If the binary is not installed, returns empty slice with no error (graceful skip).
func RunStaticAnalysis(dir, binary string) ([]ReviewFinding, error) {
	cmd := exec.Command(binary, "run", "--out-format", "json", "./...")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		// Graceful skip when binary is not found.
		var execErr *exec.Error
		if errors.As(err, &execErr) && errors.Is(execErr.Err, exec.ErrNotFound) {
			return nil, nil
		}
		// Also handle PathError (binary not found on some systems).
		if isNotFoundError(err) {
			return nil, nil
		}
		// golangci-lint exits non-zero when it finds issues — parse stdout anyway.
		if len(out) > 0 {
			return ParseGolangciOutput(out)
		}
		return nil, nil
	}

	if len(out) == 0 {
		return nil, nil
	}
	return ParseGolangciOutput(out)
}

// isNotFoundError checks if the error indicates a binary was not found.
func isNotFoundError(err error) bool {
	var execErr *exec.Error
	if errors.As(err, &execErr) {
		return errors.Is(execErr.Err, exec.ErrNotFound)
	}
	return strings.Contains(err.Error(), "executable file not found")
}

// MergeStaticWithLLMFindings deduplicates static analysis findings with LLM findings.
// When LLM reports a style finding with same normalized ScopeRef as a static finding,
// keep the static finding and discard the LLM duplicate.
func MergeStaticWithLLMFindings(static, llm []ReviewFinding) []ReviewFinding {
	// Build set of normalized ScopeRef from static style findings.
	staticStyleRefs := make(map[string]bool)
	for _, f := range static {
		if f.Category == FindingCategoryStyle {
			staticStyleRefs[NormalizeScopeRef(f.ScopeRef, "")] = true
		}
	}

	// Filter LLM findings — skip style findings that overlap with static.
	filtered := make([]ReviewFinding, 0, len(llm))
	for _, f := range llm {
		if f.Category == FindingCategoryStyle && staticStyleRefs[NormalizeScopeRef(f.ScopeRef, "")] {
			continue
		}
		filtered = append(filtered, f)
	}

	result := make([]ReviewFinding, 0, len(static)+len(filtered))
	result = append(result, static...)
	result = append(result, filtered...)
	return result
}
