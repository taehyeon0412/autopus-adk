package docs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFormat_EmptyResults verifies that an error is returned for empty results.
// Given: no DocResult items
// When: FormatPromptInjection is called with nil or empty slice
// Then: an error is returned
func TestFormat_EmptyResults(t *testing.T) {
	t.Parallel()

	_, err := FormatPromptInjection(nil)
	require.Error(t, err)

	_, err = FormatPromptInjection([]*DocResult{})
	require.Error(t, err)
}

// TestFormat_SourceLabel_CacheSource verifies that "cache" source maps to "Cache" label.
// Given: a DocResult with source "cache"
// When: FormatPromptInjection is called
// Then: output contains "(via Cache)"
func TestFormat_SourceLabel_CacheSource(t *testing.T) {
	t.Parallel()

	result := &DocResult{
		LibraryName: "cobra",
		Source:      "cache",
		Content:     "cached cobra docs",
		Tokens:      10,
	}

	formatted, err := FormatPromptInjection([]*DocResult{result})
	require.NoError(t, err)
	assert.Contains(t, formatted, "(via Cache)")
}

// TestFormat_SourceLabel_ScraperSource verifies that "scraper" source maps to "WebSearch" label.
// Given: a DocResult with source "scraper"
// When: FormatPromptInjection is called
// Then: output contains "(via WebSearch)"
func TestFormat_SourceLabel_ScraperSource(t *testing.T) {
	t.Parallel()

	result := &DocResult{
		LibraryName: "express",
		Source:      "scraper",
		Content:     "express docs",
		Tokens:      8,
	}

	formatted, err := FormatPromptInjection([]*DocResult{result})
	require.NoError(t, err)
	assert.Contains(t, formatted, "(via WebSearch)")
}

// TestFormat_SourceLabel_UnknownSource verifies that unknown sources are title-cased.
// Given: a DocResult with an unknown source string
// When: FormatPromptInjection is called
// Then: output uses the title-cased source name
func TestFormat_SourceLabel_UnknownSource(t *testing.T) {
	t.Parallel()

	result := &DocResult{
		LibraryName: "somelib",
		Source:      "custom",
		Content:     "custom docs",
		Tokens:      5,
	}

	formatted, err := FormatPromptInjection([]*DocResult{result})
	require.NoError(t, err)
	assert.Contains(t, formatted, "(via Custom)")
}
