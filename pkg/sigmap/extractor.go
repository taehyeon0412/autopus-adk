// Package sigmap provides an AST-based extractor for exported Go symbols.
package sigmap

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

// Extract scans projectDir for all exported Go symbols and returns a SignatureMap.
func Extract(projectDir string) (*SignatureMap, error) {
	modPath, err := extractModulePath(projectDir)
	if err != nil {
		return nil, fmt.Errorf("sigmap: read module path: %w", err)
	}

	// Collect unique package directories via filepath.Walk.
	dirs := map[string]struct{}{}
	err = filepath.Walk(projectDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil // skip unreadable entries
		}
		if info.IsDir() {
			base := info.Name()
			if base == "vendor" || base == "node_modules" || strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			dirs[filepath.Dir(path)] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("sigmap: walk project: %w", err)
	}

	sm := &SignatureMap{ModulePath: modPath}

	for dir := range dirs {
		rel, relErr := filepath.Rel(projectDir, dir)
		if relErr != nil {
			rel = dir
		}
		// Normalize path separator to forward slash.
		rel = filepath.ToSlash(rel)

		pkg, warnings := extractPackageSignatures(dir, rel)
		sm.Warnings = append(sm.Warnings, warnings...)
		if pkg != nil {
			sm.Packages = append(sm.Packages, *pkg)
		}
	}

	// Sort packages for deterministic output.
	sort.Slice(sm.Packages, func(i, j int) bool {
		return sm.Packages[i].Path < sm.Packages[j].Path
	})

	return sm, nil
}

// extractModulePath parses go.mod in projectDir to find the module path.
func extractModulePath(projectDir string) (string, error) {
	f, err := os.Open(filepath.Join(projectDir, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("open go.mod: %w", err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module directive not found in go.mod")
}

// extractPackageSignatures parses all non-test .go files in dir and returns
// the Package and any parse-error warnings. Returns nil Package when no
// exported symbols are found.
func extractPackageSignatures(dir string, relPath string) (*Package, []string) {
	fset := token.NewFileSet()
	var warnings []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, []string{fmt.Sprintf("read dir %s: %v", dir, err)}
	}

	pkg := &Package{
		Path:  relPath,
		Depth: strings.Count(relPath, "/"),
	}
	if relPath == "." {
		pkg.Depth = 0
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		fullPath := filepath.Join(dir, name)
		af, parseErr := parser.ParseFile(fset, fullPath, nil, parser.ParseComments)
		if parseErr != nil {
			warnings = append(warnings, fmt.Sprintf("parse %s: %v", fullPath, parseErr))
			continue
		}

		// Set package name from first successfully parsed file.
		if pkg.Name == "" {
			pkg.Name = af.Name.Name
		}

		for _, decl := range af.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if isExported(d.Name.Name) {
					pkg.Signatures = append(pkg.Signatures, extractFuncSignature(d))
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok || !isExported(ts.Name.Name) {
						continue
					}
					// Prefer TypeSpec doc, fall back to GenDecl doc.
					docText := ""
					if ts.Comment != nil {
						docText = ts.Comment.Text()
					} else if ts.Doc != nil {
						docText = ts.Doc.Text()
					} else if d.Doc != nil {
						docText = d.Doc.Text()
					}
					pkg.Signatures = append(pkg.Signatures, extractTypeSignature(ts, docText))
				}
			}
		}
	}

	if len(pkg.Signatures) == 0 {
		return nil, warnings
	}
	return pkg, warnings
}

// extractFuncSignature builds a Signature from an exported *ast.FuncDecl.
func extractFuncSignature(fn *ast.FuncDecl) Signature {
	sig := Signature{
		Name:   fn.Name.Name,
		Kind:   "func",
		Params: formatFieldList(fn.Type.Params),
	}

	if fn.Type.Results != nil {
		sig.Returns = formatFieldList(fn.Type.Results)
	}

	if fn.Type.TypeParams != nil {
		sig.TypeParams = formatTypeParams(fn.Type.TypeParams)
	}

	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		sig.Kind = "method"
		sig.Receiver = formatFieldList(fn.Recv)
	}

	if fn.Doc != nil {
		sig.Doc = firstSentence(fn.Doc.Text())
	}
	return sig
}

// extractTypeSignature builds a Signature from an exported *ast.TypeSpec.
func extractTypeSignature(spec *ast.TypeSpec, doc string) Signature {
	kind := "type"
	if _, ok := spec.Type.(*ast.InterfaceType); ok {
		kind = "interface"
	}

	sig := Signature{
		Name: spec.Name.Name,
		Kind: kind,
		Doc:  firstSentence(doc),
	}

	if spec.TypeParams != nil {
		sig.TypeParams = formatTypeParams(spec.TypeParams)
	}
	return sig
}

// formatFieldList renders an *ast.FieldList to a parenthesised string such as
// "(addr string, port int)". Returns "()" for nil or empty lists.
func formatFieldList(fl *ast.FieldList) string {
	if fl == nil || len(fl.List) == 0 {
		return "()"
	}

	fset := token.NewFileSet()
	var parts []string
	for _, field := range fl.List {
		typeStr := nodeToString(fset, field.Type)
		if len(field.Names) == 0 {
			parts = append(parts, typeStr)
		} else {
			names := make([]string, len(field.Names))
			for i, n := range field.Names {
				names[i] = n.Name
			}
			parts = append(parts, strings.Join(names, ", ")+" "+typeStr)
		}
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

// formatTypeParams renders generic type parameters to "[T any, U comparable]".
func formatTypeParams(fl *ast.FieldList) string {
	if fl == nil || len(fl.List) == 0 {
		return ""
	}

	fset := token.NewFileSet()
	var parts []string
	for _, field := range fl.List {
		constraint := nodeToString(fset, field.Type)
		for _, name := range field.Names {
			parts = append(parts, name.Name+" "+constraint)
		}
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// nodeToString renders an AST node to its Go source representation.
func nodeToString(fset *token.FileSet, node ast.Node) string {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return "<unknown>"
	}
	return buf.String()
}

// firstSentence returns the first sentence from a GoDoc string (up to the
// first "." followed by a space or end of string).
func firstSentence(doc string) string {
	doc = strings.TrimSpace(doc)
	if doc == "" {
		return ""
	}
	// Search for ". " (period followed by space).
	if idx := strings.Index(doc, ". "); idx >= 0 {
		return doc[:idx+1]
	}
	// If the doc ends with ".", keep it.
	return doc
}

// isExported reports whether name is an exported Go identifier.
func isExported(name string) bool {
	if name == "" {
		return false
	}
	r := []rune(name)
	return unicode.IsUpper(r[0])
}
