package spec

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	verdictRe        = regexp.MustCompile(`(?i)VERDICT:\s*(PASS|REVISE|REJECT)`)
	findingRe        = regexp.MustCompile(`(?i)FINDING:\s*\[(\w+)]\s*(.+)`)
	structFindingRe  = regexp.MustCompile(`(?i)FINDING:\s*\[(\w+)]\s*\[(\w+)]\s*(\S+)\s+(.+)`)
	findingStatusRe  = regexp.MustCompile(`(?i)FINDING_STATUS:\s*F-(\d+)\s*\|\s*(\w+)\s*\|\s*(.+)`)
)

// ParseVerdict extracts a ReviewResult from raw provider output.
// priorFindings: pass nil for discover mode, pass prior findings slice for verify mode.
func ParseVerdict(specID, output, provider string, revision int, priorFindings []ReviewFinding) ReviewResult {
	result := ReviewResult{
		SpecID:    specID,
		Verdict:   VerdictPass,
		Responses: []string{output},
		Revision:  revision,
	}

	// Extract verdict
	if m := verdictRe.FindStringSubmatch(output); len(m) >= 2 {
		switch strings.ToUpper(m[1]) {
		case "PASS":
			result.Verdict = VerdictPass
		case "REVISE":
			result.Verdict = VerdictRevise
		case "REJECT":
			result.Verdict = VerdictReject
		}
	}

	if priorFindings == nil {
		// Discover mode: parse structured FINDING lines
		result.Findings = parseDiscoverFindings(output, provider, revision)
	} else {
		// Verify mode: apply status updates from FINDING_STATUS lines
		result.Findings = parseVerifyFindings(output, provider, revision, priorFindings)
	}

	return result
}

// parseDiscoverFindings parses FINDING lines from discover mode output.
// Tries structured format first: FINDING: [severity] [category] [scope_ref] description
// Falls back to legacy: FINDING: [severity] description
func parseDiscoverFindings(output, provider string, revision int) []ReviewFinding {
	var findings []ReviewFinding
	seq := 1

	for _, m := range structFindingRe.FindAllStringSubmatch(output, -1) {
		if len(m) >= 5 {
			findings = append(findings, ReviewFinding{
				ID:           fmt.Sprintf("F-%03d", seq),
				Provider:     provider,
				Severity:     strings.ToLower(m[1]),
				Category:     FindingCategory(strings.ToLower(m[2])),
				ScopeRef:     m[3],
				Description:  strings.TrimSpace(m[4]),
				Status:       FindingStatusOpen,
				FirstSeenRev: revision,
				LastSeenRev:  revision,
			})
			seq++
		}
	}

	// If no structured findings found, try legacy format
	if len(findings) == 0 {
		for _, m := range findingRe.FindAllStringSubmatch(output, -1) {
			if len(m) >= 3 {
				findings = append(findings, ReviewFinding{
					ID:           fmt.Sprintf("F-%03d", seq),
					Provider:     provider,
					Severity:     strings.ToLower(m[1]),
					Description:  strings.TrimSpace(m[2]),
					Status:       FindingStatusOpen,
					FirstSeenRev: revision,
					LastSeenRev:  revision,
				})
				seq++
			}
		}
	}

	return findings
}

// parseVerifyFindings applies FINDING_STATUS updates from verify mode output.
// New critical/security findings are registered with EscapeHatch=true.
// Other new findings are tagged out_of_scope.
func parseVerifyFindings(output, provider string, revision int, priorFindings []ReviewFinding) []ReviewFinding {
	// Start with copies of prior findings, updating LastSeenRev
	updated := make([]ReviewFinding, len(priorFindings))
	for i, f := range priorFindings {
		updated[i] = f
		updated[i].LastSeenRev = revision
	}

	// Build index by ID for fast lookup
	idxByID := make(map[string]int, len(updated))
	for i, f := range updated {
		idxByID[f.ID] = i
	}

	// Apply FINDING_STATUS updates
	for _, m := range findingStatusRe.FindAllStringSubmatch(output, -1) {
		if len(m) >= 3 {
			id := fmt.Sprintf("F-%s", m[1])
			statusStr := strings.ToLower(strings.TrimSpace(m[2]))
			if idx, ok := idxByID[id]; ok {
				switch statusStr {
				case "resolved":
					updated[idx].Status = FindingStatusResolved
				case "regressed":
					updated[idx].Status = FindingStatusRegressed
				default:
					updated[idx].Status = FindingStatusOpen
				}
			}
		}
	}

	// Parse any new FINDING lines (escape hatch or out_of_scope)
	seq := len(priorFindings) + 1
	for _, m := range structFindingRe.FindAllStringSubmatch(output, -1) {
		if len(m) >= 5 {
			severity := strings.ToLower(m[1])
			category := FindingCategory(strings.ToLower(m[2]))
			f := ReviewFinding{
				ID:           fmt.Sprintf("F-%03d", seq),
				Provider:     provider,
				Severity:     severity,
				Category:     category,
				ScopeRef:     m[3],
				Description:  strings.TrimSpace(m[4]),
				FirstSeenRev: revision,
				LastSeenRev:  revision,
			}
			if severity == "critical" || category == FindingCategorySecurity {
				f.Status = FindingStatusOpen
				f.EscapeHatch = true
			} else {
				f.Status = FindingStatusOutOfScope
			}
			updated = append(updated, f)
			seq++
		}
	}

	return updated
}

// MergeFindingStatuses applies supermajority merge across providers (REQ-011).
// threshold: fraction of providers that must agree (e.g., 0.67 for 2/3).
// resolved requires >= threshold agreement; regressed > open in priority.
func MergeFindingStatuses(providerResults [][]ReviewFinding, threshold float64) []ReviewFinding {
	if len(providerResults) == 0 {
		return nil
	}

	// Flatten and group by finding ID
	byID := make(map[string][]ReviewFinding)
	for _, findings := range providerResults {
		for _, f := range findings {
			byID[f.ID] = append(byID[f.ID], f)
		}
	}

	total := float64(len(providerResults))
	var merged []ReviewFinding

	for id, group := range byID {
		if len(group) == 0 {
			continue
		}
		base := group[0]

		resolvedCount := 0
		regressedCount := 0
		for _, f := range group {
			if f.Status == FindingStatusResolved {
				resolvedCount++
			}
			if f.Status == FindingStatusRegressed {
				regressedCount++
			}
		}

		if float64(resolvedCount)/total >= threshold {
			base.Status = FindingStatusResolved
		} else if regressedCount > 0 {
			base.Status = FindingStatusRegressed
		} else {
			base.Status = FindingStatusOpen
		}

		_ = id
		merged = append(merged, base)
	}

	return merged
}

// ShouldTripCircuitBreaker returns true if the review loop should halt.
// Compares open+regressed counts (excluding escape hatch and out_of_scope/deferred).
// If new escape hatch findings were introduced in curr, the breaker does NOT trip —
// a newly discovered critical/security issue is considered progress.
func ShouldTripCircuitBreaker(prev, curr []ReviewFinding) bool {
	prevCount := countActiveFindings(prev, true)
	currCount := countActiveFindings(curr, true)

	// New escape hatch findings indicate newly discovered critical issues — not stalling.
	if countEscapeHatch(curr) > countEscapeHatch(prev) {
		return false
	}

	return currCount >= prevCount
}

// countActiveFindings counts open+regressed findings, always excluding escape hatch.
func countActiveFindings(findings []ReviewFinding, excludeEscapeHatch bool) int {
	count := 0
	for _, f := range findings {
		if f.Status == FindingStatusOutOfScope || f.Status == FindingStatusDeferred {
			continue
		}
		if excludeEscapeHatch && f.EscapeHatch {
			continue
		}
		if f.Status == FindingStatusOpen || f.Status == FindingStatusRegressed {
			count++
		}
	}
	return count
}

// countEscapeHatch returns the number of escape hatch findings.
func countEscapeHatch(findings []ReviewFinding) int {
	count := 0
	for _, f := range findings {
		if f.EscapeHatch {
			count++
		}
	}
	return count
}

// MergeVerdicts combines multiple review results into a single verdict.
// REJECT wins over REVISE, REVISE wins over PASS.
func MergeVerdicts(results []ReviewResult) ReviewVerdict {
	verdict := VerdictPass
	for _, r := range results {
		switch r.Verdict {
		case VerdictReject:
			return VerdictReject
		case VerdictRevise:
			verdict = VerdictRevise
		}
	}
	return verdict
}
