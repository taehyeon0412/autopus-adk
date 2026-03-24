package sigmap_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/sigmap"
)

// TestExtractorRegistry_Get_Go verifies that the registry returns the Go
// extractor for a file with a .go extension.
func TestExtractorRegistry_Get_Go(t *testing.T) {
	t.Parallel()

	// Given: the default extractor registry
	reg := sigmap.NewExtractorRegistry()

	// When: requesting the extractor for a .go file
	ex, err := reg.Get("main.go")
	require.NoError(t, err)

	// Then: a non-nil Go extractor is returned
	require.NotNil(t, ex)
	assert.Equal(t, "go", ex.Language())
}

// TestExtractorRegistry_Get_TypeScript verifies that the registry returns the
// TypeScript extractor for files with .ts extensions.
func TestExtractorRegistry_Get_TypeScript(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
	}{
		{"ts file", "component.ts"},
		{"tsx file", "Component.tsx"},
	}

	// Given: the default extractor registry
	reg := sigmap.NewExtractorRegistry()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// When: requesting the extractor for a TypeScript file
			ex, err := reg.Get(tt.filename)
			require.NoError(t, err)

			// Then: a non-nil TypeScript extractor is returned
			require.NotNil(t, ex)
			assert.Equal(t, "typescript", ex.Language())
		})
	}
}

// TestExtractorRegistry_Get_Unknown_ReturnsError verifies that requesting an
// extractor for an unsupported file extension returns an error.
func TestExtractorRegistry_Get_Unknown_ReturnsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filename string
	}{
		{"python file", "script.py"},
		{"ruby file", "app.rb"},
		{"no extension", "Makefile"},
	}

	// Given: the default extractor registry
	reg := sigmap.NewExtractorRegistry()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// When: requesting the extractor for an unsupported file
			ex, err := reg.Get(tt.filename)

			// Then: an error is returned and extractor is nil
			require.Error(t, err)
			assert.Nil(t, ex)
			assert.Contains(t, err.Error(), "unsupported")
		})
	}
}
