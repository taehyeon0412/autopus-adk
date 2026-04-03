package learn

import (
	"regexp"
	"time"
)

// EntryType represents the category of a learning entry.
type EntryType string

const (
	EntryTypeGateFail      EntryType = "gate_fail"
	EntryTypeCoverageGap   EntryType = "coverage_gap"
	EntryTypeReviewIssue   EntryType = "review_issue"
	EntryTypeExecutorError EntryType = "executor_error"
	EntryTypeFixPattern    EntryType = "fix_pattern"
)

// Severity represents the impact level of a learning entry.
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// LearningEntry represents one learning record (R12).
type LearningEntry struct {
	ID         string    `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	Type       EntryType `json:"type"`
	Phase      string    `json:"phase"`
	SpecID     string    `json:"spec_id,omitempty"`
	Files      []string  `json:"files"`
	Packages   []string  `json:"packages"`
	Pattern    string    `json:"pattern"`
	Resolution string    `json:"resolution"`
	Severity   Severity  `json:"severity"`
	ReuseCount int       `json:"reuse_count"`
}

// entryIDRegex matches L-{NNN} format (one or more digits after L-).
var entryIDRegex = regexp.MustCompile(`^L-\d{3,}$`)

// IsValidEntryID checks whether id matches the L-{NNN} format.
func IsValidEntryID(id string) bool {
	return entryIDRegex.MatchString(id)
}

// RelevanceQuery holds parameters for relevance matching (R13).
type RelevanceQuery struct {
	Files    []string
	Packages []string
	Keywords []string
}

// Summary holds learning summary for sync display (R8).
type Summary struct {
	TotalEntries     int
	NewEntries       int
	TypeCounts       map[EntryType]int
	TopPatterns      []PatternStat
	ImprovementAreas []string
	Improvements     []string
}

// RecordOpts holds options for recording a learning entry.
type RecordOpts struct {
	Phase      string
	SpecID     string
	Files      []string
	Packages   []string
	Pattern    string
	Resolution string
	Severity   Severity
}

// PatternStat tracks reuse frequency of a pattern.
type PatternStat struct {
	Pattern    string
	ReuseCount int
}
