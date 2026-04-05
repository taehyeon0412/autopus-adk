package learn

import (
	"fmt"
)

func recordEntry(store *Store, entryType EntryType, opts RecordOpts) error {
	if store == nil {
		return fmt.Errorf("store is nil")
	}
	if opts.Pattern == "" {
		return fmt.Errorf("pattern is required")
	}
	return store.AppendAtomic(entryType, opts)
}

// RecordGateFail records a gate failure learning entry.
func RecordGateFail(store *Store, opts RecordOpts) error {
	return recordEntry(store, EntryTypeGateFail, opts)
}

// RecordCoverageGap records a coverage gap learning entry.
func RecordCoverageGap(store *Store, opts RecordOpts) error {
	return recordEntry(store, EntryTypeCoverageGap, opts)
}

// RecordReviewIssue records a review issue learning entry.
func RecordReviewIssue(store *Store, opts RecordOpts) error {
	return recordEntry(store, EntryTypeReviewIssue, opts)
}

// RecordExecutorError records an executor error learning entry.
func RecordExecutorError(store *Store, opts RecordOpts) error {
	return recordEntry(store, EntryTypeExecutorError, opts)
}

// RecordFixPattern records a fix pattern learning entry.
func RecordFixPattern(store *Store, opts RecordOpts) error {
	return recordEntry(store, EntryTypeFixPattern, opts)
}
