package learn

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchRelevance_ExactFilePath(t *testing.T) {
	t.Parallel()

	// Given: an entry with specific file paths
	entry := LearningEntry{
		ID:        "L-001",
		Type:      EntryTypeGateFail,
		Files:     []string{"pkg/learn/store.go"},
		Packages:  []string{"learn"},
		Pattern:   "test failure",
		Timestamp: time.Now(),
	}
	query := RelevanceQuery{
		Files:    []string{"pkg/learn/store.go"},
		Packages: []string{"learn"},
	}

	// When: matching relevance
	score := MatchRelevance(entry, query)

	// Then: exact file path scores highest
	assert.Greater(t, score, 0.0)
}

func TestMatchRelevance_PackagePrefixMatch(t *testing.T) {
	t.Parallel()

	// Given: entry with package "learn"
	entry := LearningEntry{
		ID:       "L-001",
		Type:     EntryTypeGateFail,
		Files:    []string{"pkg/learn/types.go"},
		Packages: []string{"learn"},
		Pattern:  "type error",
	}
	query := RelevanceQuery{
		Packages: []string{"learn"},
	}

	// When: matching by package
	score := MatchRelevance(entry, query)

	// Then: package match scores medium
	assert.Greater(t, score, 0.0)
}

func TestMatchRelevance_KeywordMatch(t *testing.T) {
	t.Parallel()

	// Given: entry with pattern containing keywords
	entry := LearningEntry{
		ID:      "L-001",
		Type:    EntryTypeCoverageGap,
		Pattern: "coverage below 85% threshold",
	}
	query := RelevanceQuery{
		Keywords: []string{"coverage", "threshold"},
	}

	// When: matching by keywords
	score := MatchRelevance(entry, query)

	// Then: keyword match scores lowest but positive
	assert.Greater(t, score, 0.0)
}

func TestMatchRelevance_RecencyBonus(t *testing.T) {
	t.Parallel()

	// Given: two entries — one recent, one old
	recent := LearningEntry{
		ID:        "L-001",
		Type:      EntryTypeGateFail,
		Files:     []string{"main.go"},
		Timestamp: time.Now().Add(-7 * 24 * time.Hour), // 7 days ago
	}
	old := LearningEntry{
		ID:        "L-002",
		Type:      EntryTypeGateFail,
		Files:     []string{"main.go"},
		Timestamp: time.Now().Add(-60 * 24 * time.Hour), // 60 days ago
	}
	query := RelevanceQuery{
		Files: []string{"main.go"},
	}

	// When: matching both
	recentScore := MatchRelevance(recent, query)
	oldScore := MatchRelevance(old, query)

	// Then: recent entry scores higher due to recency bonus (within 30 days)
	assert.Greater(t, recentScore, oldScore)
}

func TestMatchRelevance_BelowThreshold(t *testing.T) {
	t.Parallel()

	// Given: entry with no matching fields
	entry := LearningEntry{
		ID:       "L-001",
		Type:     EntryTypeGateFail,
		Files:    []string{"cmd/root.go"},
		Packages: []string{"cmd"},
		Pattern:  "command error",
	}
	query := RelevanceQuery{
		Files:    []string{"pkg/learn/store.go"},
		Packages: []string{"learn"},
		Keywords: []string{"store"},
	}

	// When: matching unrelated entry
	score := MatchRelevance(entry, query)

	// Then: score is zero or below threshold
	assert.Equal(t, 0.0, score)
}

func TestMatchRelevance_EmptyQuery(t *testing.T) {
	t.Parallel()

	// Given: empty query
	entry := LearningEntry{
		ID:   "L-001",
		Type: EntryTypeGateFail,
	}
	query := RelevanceQuery{}

	// When: matching with empty query
	score := MatchRelevance(entry, query)

	// Then: score is zero
	assert.Equal(t, 0.0, score)
}

func TestQueryRelevant_FiltersAndSorts(t *testing.T) {
	t.Parallel()

	// Given: store with mixed relevance entries
	dir := t.TempDir()
	store, err := NewStore(dir)
	require.NoError(t, err)

	entries := []LearningEntry{
		{ID: "L-001", Type: EntryTypeGateFail, Files: []string{"pkg/learn/store.go"}, Packages: []string{"learn"}, Timestamp: time.Now()},
		{ID: "L-002", Type: EntryTypeGateFail, Files: []string{"cmd/root.go"}, Packages: []string{"cmd"}, Timestamp: time.Now()},
		{ID: "L-003", Type: EntryTypeCoverageGap, Files: []string{"pkg/learn/query.go"}, Packages: []string{"learn"}, Timestamp: time.Now()},
	}
	for _, e := range entries {
		require.NoError(t, store.Append(e))
	}

	query := RelevanceQuery{
		Files:    []string{"pkg/learn/store.go"},
		Packages: []string{"learn"},
	}

	// When: querying relevant entries
	results, err := QueryRelevant(store, query, 0.1)

	// Then: only relevant entries returned, sorted by score descending
	require.NoError(t, err)
	assert.NotEmpty(t, results)
	// First result should be the exact file match
	assert.Equal(t, "L-001", results[0].ID)
}

func TestQueryRelevant_EmptyStore(t *testing.T) {
	t.Parallel()

	// Given: empty store
	dir := t.TempDir()
	store, err := NewStore(dir)
	require.NoError(t, err)

	query := RelevanceQuery{
		Files: []string{"main.go"},
	}

	// When: querying
	results, err := QueryRelevant(store, query, 0.1)

	// Then: empty results
	require.NoError(t, err)
	assert.Empty(t, results)
}
