package learn

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_AppendAtomic_ConcurrentNoDuplicateIDs(t *testing.T) {
	t.Parallel()

	// Given: a fresh store
	dir := t.TempDir()
	store, err := NewStore(dir)
	require.NoError(t, err)

	const goroutines = 20
	var wg sync.WaitGroup
	errs := make([]error, goroutines)

	// When: 20 goroutines call AppendAtomic concurrently
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			errs[idx] = store.AppendAtomic(EntryTypeGateFail, RecordOpts{
				Phase:    "validate",
				SpecID:   fmt.Sprintf("SPEC-%03d", idx),
				Pattern:  fmt.Sprintf("concurrent pattern %d", idx),
				Severity: SeverityLow,
			})
		}(i)
	}
	wg.Wait()

	// Then: no errors occurred
	for i, e := range errs {
		require.NoError(t, e, "goroutine %d returned error", i)
	}

	// Then: exactly goroutines entries written, no ID collisions, no data loss
	entries, err := store.Read()
	require.NoError(t, err)
	assert.Len(t, entries, goroutines)

	ids := make(map[string]struct{}, goroutines)
	for _, entry := range entries {
		assert.True(t, IsValidEntryID(entry.ID), "invalid ID format: %s", entry.ID)
		_, duplicate := ids[entry.ID]
		assert.False(t, duplicate, "duplicate ID found: %s", entry.ID)
		ids[entry.ID] = struct{}{}
	}
}

func TestStore_AppendAtomic_ConcurrentNoDataLoss(t *testing.T) {
	t.Parallel()

	// Given: a store pre-populated with one entry
	dir := t.TempDir()
	store, err := NewStore(dir)
	require.NoError(t, err)

	seed := LearningEntry{
		ID:       "L-001",
		Type:     EntryTypeFixPattern,
		Pattern:  "seed pattern",
		Severity: SeverityHigh,
	}
	require.NoError(t, store.Append(seed))

	const goroutines = 10
	var wg sync.WaitGroup

	// When: 10 goroutines append concurrently on top of existing entry
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_ = store.AppendAtomic(EntryTypeCoverageGap, RecordOpts{
				Phase:   "implement",
				Pattern: fmt.Sprintf("coverage gap %d", idx),
			})
		}(i)
	}
	wg.Wait()

	// Then: seed entry is preserved and total count is 1 + goroutines
	entries, err := store.Read()
	require.NoError(t, err)
	assert.Len(t, entries, goroutines+1)

	// Seed entry must still be intact
	assert.Equal(t, "L-001", entries[0].ID)
	assert.Equal(t, "seed pattern", entries[0].Pattern)
}
