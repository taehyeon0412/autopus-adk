package learn

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSummary_TopPatterns(t *testing.T) {
	t.Parallel()

	// Given: store with entries having varying reuse counts
	dir := t.TempDir()
	store, err := NewStore(dir)
	require.NoError(t, err)

	entries := []LearningEntry{
		{ID: "L-001", Type: EntryTypeFixPattern, Pattern: "nil check missing", ReuseCount: 5, Timestamp: time.Now()},
		{ID: "L-002", Type: EntryTypeFixPattern, Pattern: "error wrapping", ReuseCount: 10, Timestamp: time.Now()},
		{ID: "L-003", Type: EntryTypeFixPattern, Pattern: "race condition", ReuseCount: 3, Timestamp: time.Now()},
	}
	for _, e := range entries {
		require.NoError(t, store.Append(e))
	}

	// When: generating summary with top 2
	summary, err := GenerateSummary(store, 2)

	// Then: top patterns sorted by reuse_count descending
	require.NoError(t, err)
	assert.Len(t, summary.TopPatterns, 2)
	assert.Equal(t, "error wrapping", summary.TopPatterns[0].Pattern)
	assert.Equal(t, "nil check missing", summary.TopPatterns[1].Pattern)
}

func TestGenerateSummary_NewEntryStats(t *testing.T) {
	t.Parallel()

	// Given: store with entries of different types
	dir := t.TempDir()
	store, err := NewStore(dir)
	require.NoError(t, err)

	entries := []LearningEntry{
		{ID: "L-001", Type: EntryTypeGateFail, Timestamp: time.Now()},
		{ID: "L-002", Type: EntryTypeGateFail, Timestamp: time.Now()},
		{ID: "L-003", Type: EntryTypeCoverageGap, Timestamp: time.Now()},
		{ID: "L-004", Type: EntryTypeExecutorError, Timestamp: time.Now()},
	}
	for _, e := range entries {
		require.NoError(t, store.Append(e))
	}

	// When: generating summary
	summary, err := GenerateSummary(store, 5)

	// Then: stats reflect entry type counts
	require.NoError(t, err)
	assert.Equal(t, 4, summary.TotalEntries)
	assert.Equal(t, 2, summary.TypeCounts[EntryTypeGateFail])
	assert.Equal(t, 1, summary.TypeCounts[EntryTypeCoverageGap])
	assert.Equal(t, 1, summary.TypeCounts[EntryTypeExecutorError])
}

func TestGenerateSummary_ImprovementAreas(t *testing.T) {
	t.Parallel()

	// Given: store with recurring patterns in the same package
	dir := t.TempDir()
	store, err := NewStore(dir)
	require.NoError(t, err)

	entries := []LearningEntry{
		{ID: "L-001", Type: EntryTypeGateFail, Packages: []string{"auth"}, Timestamp: time.Now()},
		{ID: "L-002", Type: EntryTypeCoverageGap, Packages: []string{"auth"}, Timestamp: time.Now()},
		{ID: "L-003", Type: EntryTypeExecutorError, Packages: []string{"auth"}, Timestamp: time.Now()},
		{ID: "L-004", Type: EntryTypeGateFail, Packages: []string{"store"}, Timestamp: time.Now()},
	}
	for _, e := range entries {
		require.NoError(t, store.Append(e))
	}

	// When: generating summary
	summary, err := GenerateSummary(store, 5)

	// Then: "auth" identified as improvement area
	require.NoError(t, err)
	assert.Contains(t, summary.ImprovementAreas, "auth")
}

func TestGenerateSummary_EmptyStore(t *testing.T) {
	t.Parallel()

	// Given: empty store
	dir := t.TempDir()
	store, err := NewStore(dir)
	require.NoError(t, err)

	// When: generating summary
	summary, err := GenerateSummary(store, 5)

	// Then: zero values, no error
	require.NoError(t, err)
	assert.Equal(t, 0, summary.TotalEntries)
	assert.Empty(t, summary.TopPatterns)
	assert.Empty(t, summary.ImprovementAreas)
}

func TestGenerateSummary_TopNExceedsEntries(t *testing.T) {
	t.Parallel()

	// Given: store with fewer entries than requested top N
	dir := t.TempDir()
	store, err := NewStore(dir)
	require.NoError(t, err)

	entry := LearningEntry{
		ID:         "L-001",
		Type:       EntryTypeFixPattern,
		Pattern:    "single pattern",
		ReuseCount: 1,
		Timestamp:  time.Now(),
	}
	require.NoError(t, store.Append(entry))

	// When: requesting top 10
	summary, err := GenerateSummary(store, 10)

	// Then: returns all available (1), no error
	require.NoError(t, err)
	assert.Len(t, summary.TopPatterns, 1)
}
