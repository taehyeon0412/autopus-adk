package sigmap

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TSExtractor extracts exported symbols from TypeScript source files.
type TSExtractor struct{}

// NewTSExtractor returns a new TSExtractor.
func NewTSExtractor() *TSExtractor {
	return &TSExtractor{}
}

// Language returns "typescript".
func (t *TSExtractor) Language() string {
	return "typescript"
}

// ExtractDir implements the Extractor interface by scanning dir for .ts/.tsx
// files and aggregating their exported signatures into a SignatureMap.
func (t *TSExtractor) ExtractDir(dir string) (*SignatureMap, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("ts_extractor: read dir %s: %w", dir, err)
	}

	sm := &SignatureMap{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".ts") && !strings.HasSuffix(name, ".tsx") {
			continue
		}
		src, readErr := os.ReadFile(filepath.Join(dir, name))
		if readErr != nil {
			sm.Warnings = append(sm.Warnings, fmt.Sprintf("read %s: %v", name, readErr))
			continue
		}
		sigs, extractErr := t.Extract(src, name)
		if extractErr != nil {
			sm.Warnings = append(sm.Warnings, fmt.Sprintf("extract %s: %v", name, extractErr))
			continue
		}
		if len(sigs) > 0 {
			sm.Packages = append(sm.Packages, Package{
				Path:       dir,
				Name:       strings.TrimSuffix(strings.TrimSuffix(name, ".tsx"), ".ts"),
				Signatures: sigs,
			})
		}
	}
	return sm, nil
}

// @AX:NOTE [AUTO] @AX:REASON: regex patterns match TypeScript export syntax; changes break signature detection
// tsPattern matches TypeScript export declarations.
var (
	reExportDefault   = regexp.MustCompile(`(?m)^export\s+default\s+(?:async\s+)?function`)
	reExportFunction  = regexp.MustCompile(`(?m)^export\s+(?:async\s+)?function\s+(\w+)`)
	reExportClass     = regexp.MustCompile(`(?m)^export\s+class\s+(\w+)`)
	reExportInterface = regexp.MustCompile(`(?m)^export\s+interface\s+(\w+)`)
	reExportConst     = regexp.MustCompile(`(?m)^export\s+const\s+(\w+)`)
	reReExport        = regexp.MustCompile(`(?m)^export\s+\{([^}]+)\}\s+from\s+`)
)

// Extract extracts exported signatures from a single TypeScript source file.
// src is the raw file content; filename is used for context only.
func (t *TSExtractor) Extract(src []byte, _ string) ([]Signature, error) {
	content := string(src)
	var sigs []Signature

	// default export function
	if reExportDefault.MatchString(content) {
		sigs = append(sigs, Signature{Name: "default", Kind: "func"})
	}

	// named export functions (skip "export default function" matches)
	for _, m := range reExportFunction.FindAllStringSubmatch(content, -1) {
		sigs = append(sigs, Signature{Name: m[1], Kind: "func"})
	}

	// export class
	for _, m := range reExportClass.FindAllStringSubmatch(content, -1) {
		sigs = append(sigs, Signature{Name: m[1], Kind: "type"})
	}

	// export interface
	for _, m := range reExportInterface.FindAllStringSubmatch(content, -1) {
		sigs = append(sigs, Signature{Name: m[1], Kind: "interface"})
	}

	// export const
	for _, m := range reExportConst.FindAllStringSubmatch(content, -1) {
		sigs = append(sigs, Signature{Name: m[1], Kind: "const"})
	}

	// re-exports: export { foo, bar } from './baz'
	for _, m := range reReExport.FindAllStringSubmatch(content, -1) {
		names := strings.Split(m[1], ",")
		for _, name := range names {
			name = strings.TrimSpace(name)
			if idx := strings.Index(name, " as "); idx >= 0 {
				name = strings.TrimSpace(name[:idx])
			}
			if name != "" {
				sigs = append(sigs, Signature{Name: name, Kind: "reexport"})
			}
		}
	}

	return sigs, nil
}
