package learn

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrune_RemovesOldEntries(t *testing.T) {
	t.Parallel()

	// Given: store with entries older than 90 days
	dir := t.TempDir()
	store, err := NewStore(dir)
	require.NoError(t, err)

	old := LearningEntry{
		ID:        "L-001",
		Type:      EntryTypeGateFail,
		Timestamp: time.Now().Add(-100 * 24 * time.Hour), // 100 days ago
		Severity:  SeverityLow,
	}
	recent := LearningEntry{
		ID:        "L-002",
		Type:      EntryTypeGateFail,
		Timestamp: time.Now().Add(-30 * 24 * time.Hour), // 30 days ago
		Severity:  SeverityLow,
	}
	require.NoError(t, store.Append(old))
	require.NoError(t, store.Append(recent))

	// When: pruning with 90-day threshold
	pruned, err := Prune(store, 90)

	// Then: old entry removed, recent preserved
	require.NoError(t, err)
	assert.Equal(t, 1, pruned)

	entries, err := store.Read()
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "L-002", entries[0].ID)
}

func TestPrune_PreservesRecentEntries(t *testing.T) {
	t.Parallel()

	// Given: store with only recent entries
	dir := t.TempDir()
	store, err := NewStore(dir)
	require.NoError(t, err)

	entry := LearningEntry{
		ID:        "L-001",
		Type:      EntryTypeCoverageGap,
		Timestamp: time.Now().Add(-10 * 24 * time.Hour), // 10 days ago
	}
	require.NoError(t, store.Append(entry))

	// When: pruning
	pruned, err := Prune(store, 90)

	// Then: nothing pruned
	require.NoError(t, err)
	assert.Equal(t, 0, pruned)

	entries, err := store.Read()
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

func TestPrune_EmptyStore(t *testing.T) {
	t.Parallel()

	// Given: empty store
	dir := t.TempDir()
	store, err := NewStore(dir)
	require.NoError(t, err)

	// When: pruning
	pruned, err := Prune(store, 90)

	// Then: zero pruned, no error
	require.NoError(t, err)
	assert.Equal(t, 0, pruned)
}

func TestPrune_AllExpired(t *testing.T) {
	t.Parallel()

	// Given: store with all expired entries
	dir := t.TempDir()
	store, err := NewStore(dir)
	require.NoError(t, err)

	for i := 1; i <= 3; i++ {
		entry := LearningEntry{
			ID:        "L-00" + string(rune('0'+i)),
			Type:      EntryTypeExecutorError,
			Timestamp: time.Now().Add(-200 * 24 * time.Hour),
		}
		require.NoError(t, store.Append(entry))
	}

	// When: pruning
	pruned, err := Prune(store, 90)

	// Then: all entries pruned
	require.NoError(t, err)
	assert.Equal(t, 3, pruned)

	entries, err := store.Read()
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestPrune_BoundaryExactly90Days(t *testing.T) {
	t.Parallel()

	// Given: entry exactly at 90-day boundary
	dir := t.TempDir()
	store, err := NewStore(dir)
	require.NoError(t, err)

	entry := LearningEntry{
		ID:        "L-001",
		Type:      EntryTypeGateFail,
		Timestamp: time.Now().Add(-90 * 24 * time.Hour),
	}
	require.NoError(t, store.Append(entry))

	// When: pruning at 90-day threshold
	pruned, err := Prune(store, 90)

	// Then: boundary entry should be pruned (>= 90 days)
	require.NoError(t, err)
	assert.Equal(t, 1, pruned)
}
