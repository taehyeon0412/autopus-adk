package spec

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// REQ-010: Prior findings persistence to review-findings.json

func TestPersistFindings_WritesJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	findings := []ReviewFinding{
		{
			ID:           "F-001",
			Status:       FindingStatusOpen,
			Category:     FindingCategoryCorrectness,
			ScopeRef:     "REQ-001",
			Description:  "Missing error case",
			Provider:     "claude",
			FirstSeenRev: 0,
			LastSeenRev:  0,
		},
	}

	err := PersistFindings(dir, findings)
	require.NoError(t, err)

	// Verify file exists
	path := filepath.Join(dir, "review-findings.json")
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify content is valid JSON with expected fields
	var parsed []ReviewFinding
	require.NoError(t, json.Unmarshal(data, &parsed))
	require.Len(t, parsed, 1)
	assert.Equal(t, "F-001", parsed[0].ID)
	assert.Equal(t, FindingStatusOpen, parsed[0].Status)
	assert.Equal(t, FindingCategoryCorrectness, parsed[0].Category)
}

func TestLoadFindings_ReadsPersistedJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	original := []ReviewFinding{
		{ID: "F-001", Status: FindingStatusOpen, Category: FindingCategoryCorrectness, ScopeRef: "REQ-001", FirstSeenRev: 0, LastSeenRev: 1},
		{ID: "F-002", Status: FindingStatusResolved, Category: FindingCategoryFeasibility, ScopeRef: "types.go:42", FirstSeenRev: 0, LastSeenRev: 0},
	}

	require.NoError(t, PersistFindings(dir, original))

	loaded, err := LoadFindings(dir)
	require.NoError(t, err)
	require.Len(t, loaded, 2)
	assert.Equal(t, "F-001", loaded[0].ID)
	assert.Equal(t, FindingStatusResolved, loaded[1].Status)
}

func TestLoadFindings_MissingFileFallback(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	// No file exists yet

	loaded, err := LoadFindings(dir)
	// Must not error — return empty slice as fallback
	require.NoError(t, err)
	assert.Empty(t, loaded, "missing review-findings.json must return empty slice, not error")
}

func TestLoadFindings_CorruptedFileFallback(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "review-findings.json")
	require.NoError(t, os.WriteFile(path, []byte("{ invalid json }{"), 0o644))

	loaded, err := LoadFindings(dir)
	// Must error gracefully — caller decides how to handle
	assert.Error(t, err, "corrupted JSON must return an error")
	assert.Nil(t, loaded)
}

func TestPersistFindings_Roundtrip_PreservesAllFields(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	findings := []ReviewFinding{
		{
			ID:           "F-005",
			Status:       FindingStatusDeferred,
			Category:     FindingCategoryStyle,
			ScopeRef:     "pkg/spec/types.go:12",
			Description:  "Use consistent naming",
			Provider:     "gemini",
			FirstSeenRev: 1,
			LastSeenRev:  3,
			EscapeHatch:  false,
		},
	}

	require.NoError(t, PersistFindings(dir, findings))
	loaded, err := LoadFindings(dir)
	require.NoError(t, err)
	require.Len(t, loaded, 1)

	f := loaded[0]
	assert.Equal(t, "F-005", f.ID)
	assert.Equal(t, FindingStatusDeferred, f.Status)
	assert.Equal(t, FindingCategoryStyle, f.Category)
	assert.Equal(t, "pkg/spec/types.go:12", f.ScopeRef)
	assert.Equal(t, 1, f.FirstSeenRev)
	assert.Equal(t, 3, f.LastSeenRev)
}

// Finding dedup logic

func TestDeduplicateFindings_RemovesDuplicateByDescription(t *testing.T) {
	t.Parallel()

	findings := []ReviewFinding{
		{ID: "F-001", Provider: "claude", Category: FindingCategoryCorrectness, Description: "Missing nil check", ScopeRef: "REQ-002"},
		{ID: "F-002", Provider: "gemini", Category: FindingCategoryCorrectness, Description: "Missing nil check", ScopeRef: "REQ-002"},
	}

	deduped := DeduplicateFindings(findings)

	// Two providers found the same thing — must collapse to one
	assert.Len(t, deduped, 1, "duplicate findings must be collapsed to one")
}

func TestDeduplicateFindings_KeepsDifferentScopes(t *testing.T) {
	t.Parallel()

	findings := []ReviewFinding{
		{ID: "F-001", Provider: "claude", Category: FindingCategoryCorrectness, Description: "Missing nil check", ScopeRef: "REQ-002"},
		{ID: "F-002", Provider: "claude", Category: FindingCategoryCorrectness, Description: "Missing nil check", ScopeRef: "REQ-003"},
	}

	deduped := DeduplicateFindings(findings)

	// Same description but different ScopeRef — must be kept separate
	assert.Len(t, deduped, 2, "findings with different ScopeRef must not be deduplicated")
}

func TestDeduplicateFindings_EmptyInput(t *testing.T) {
	t.Parallel()

	deduped := DeduplicateFindings(nil)
	assert.Empty(t, deduped)
}

func TestDeduplicateFindings_AssignsSequentialIDs(t *testing.T) {
	t.Parallel()

	findings := []ReviewFinding{
		{Provider: "claude", Category: FindingCategoryCompleteness, Description: "No acceptance criteria", ScopeRef: ""},
		{Provider: "openai", Category: FindingCategoryFeasibility, Description: "Timeline unrealistic", ScopeRef: ""},
	}

	deduped := DeduplicateFindings(findings)

	require.Len(t, deduped, 2)
	assert.Equal(t, "F-001", deduped[0].ID, "first finding must be F-001")
	assert.Equal(t, "F-002", deduped[1].ID, "second finding must be F-002")
}
