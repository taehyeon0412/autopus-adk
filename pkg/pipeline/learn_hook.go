package pipeline

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/insajin/autopus-adk/pkg/learn"
)

var (
	// coverageRe extracts coverage percentage from go test output.
	coverageRe = regexp.MustCompile(`coverage: (\d+\.?\d*)% of statements`)
	// findingRe extracts individual review findings.
	findingRe = regexp.MustCompile(`FINDING: \[(\w+)\] (.+)`)
)

// learnHookGateFail records a gate failure event. Silently returns when store is nil.
func learnHookGateFail(store *learn.Store, phaseID PhaseID, gate GateType, output string, attempt int) {
	if store == nil {
		return
	}
	severity := learn.SeverityMedium
	if attempt >= defaultMaxRetries {
		severity = learn.SeverityCritical
	}
	pattern := fmt.Sprintf("gate %s failed at phase %s (attempt %d)", gate, phaseID, attempt)
	_ = learn.RecordGateFail(store, learn.RecordOpts{
		Phase:    string(phaseID),
		Pattern:  pattern,
		Severity: severity,
	})
}

// learnHookCoverageGap records a coverage gap when coverage is below threshold.
// Silently returns when store is nil.
func learnHookCoverageGap(store *learn.Store, output string, threshold float64) {
	if store == nil {
		return
	}
	coverage := parseCoverage(output)
	if coverage < 0 || coverage >= threshold {
		return
	}
	pattern := fmt.Sprintf("coverage %.1f%% below threshold %.1f%%", coverage, threshold)
	_ = learn.RecordCoverageGap(store, learn.RecordOpts{
		Pattern:  pattern,
		Severity: learn.SeverityMedium,
	})
}

// learnHookReviewIssue records individual review findings from gate output.
// Silently returns when store is nil.
func learnHookReviewIssue(store *learn.Store, output string, specID string) {
	if store == nil {
		return
	}
	matches := findingRe.FindAllStringSubmatch(output, -1)
	if len(matches) == 0 {
		// Graceful degradation: record entire output as single issue.
		_ = learn.RecordReviewIssue(store, learn.RecordOpts{
			SpecID:   specID,
			Pattern:  output,
			Severity: learn.SeverityMedium,
		})
		return
	}
	for _, m := range matches {
		sev := mapFindingSeverity(m[1])
		_ = learn.RecordReviewIssue(store, learn.RecordOpts{
			SpecID:   specID,
			Pattern:  m[2],
			Severity: sev,
		})
	}
}

// learnHookExecutorError records an executor error. Silently returns when store is nil.
func learnHookExecutorError(store *learn.Store, phaseID PhaseID, err error) {
	if store == nil {
		return
	}
	_ = learn.RecordExecutorError(store, learn.RecordOpts{
		Phase:    string(phaseID),
		Pattern:  err.Error(),
		Severity: learn.SeverityHigh,
	})
}

// parseCoverage extracts coverage percentage from output. Returns -1 if not found.
func parseCoverage(output string) float64 {
	m := coverageRe.FindStringSubmatch(output)
	if len(m) < 2 {
		return -1
	}
	v, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return -1
	}
	return v
}

// mapFindingSeverity maps a finding severity string to learn.Severity.
func mapFindingSeverity(s string) learn.Severity {
	switch strings.ToLower(s) {
	case "critical":
		return learn.SeverityCritical
	case "high":
		return learn.SeverityHigh
	case "medium":
		return learn.SeverityMedium
	default:
		return learn.SeverityLow
	}
}
