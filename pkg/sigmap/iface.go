package sigmap

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Extractor defines the interface for language-specific signature extraction.
type Extractor interface {
	// Extract extracts signatures from the given directory.
	Extract(dir string) (*SignatureMap, error)
	// Language returns the language this extractor supports.
	Language() string
}

// ExtractorRegistry manages language-specific extractors keyed by file extension.
type ExtractorRegistry struct {
	extractors map[string]Extractor
}

// NewExtractorRegistry returns a registry pre-populated with built-in extractors.
func NewExtractorRegistry() *ExtractorRegistry {
	r := &ExtractorRegistry{
		extractors: make(map[string]Extractor),
	}
	r.Register(&GoExtractor{})
	r.Register(&tsExtractorDir{ts: NewTSExtractor()})
	return r
}

// Register adds an extractor to the registry under its supported extensions.
func (r *ExtractorRegistry) Register(ext Extractor) {
	switch ext.Language() {
	case "go":
		r.extractors[".go"] = ext
	case "typescript":
		r.extractors[".ts"] = ext
		r.extractors[".tsx"] = ext
	default:
		r.extractors["."+strings.ToLower(ext.Language())] = ext
	}
}

// Get returns the extractor for the given filename based on its extension.
// Returns an error if no extractor is registered for the file's extension.
func (r *ExtractorRegistry) Get(filename string) (Extractor, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return nil, fmt.Errorf("unsupported file extension: no extension in %q", filename)
	}
	ex, ok := r.extractors[ext]
	if !ok {
		return nil, fmt.Errorf("unsupported file extension: %q", ext)
	}
	return ex, nil
}

// tsExtractorDir adapts TSExtractor to the Extractor interface for directory scanning.
type tsExtractorDir struct {
	ts *TSExtractor
}

func (a *tsExtractorDir) Language() string {
	return a.ts.Language()
}

func (a *tsExtractorDir) Extract(dir string) (*SignatureMap, error) {
	return a.ts.ExtractDir(dir)
}
