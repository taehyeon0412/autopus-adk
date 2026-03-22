package sigmap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- Extract: no go.mod → error ----

func TestExtract_NoGoMod(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	// No go.mod file written intentionally.
	_, err := Extract(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "read module path")
}

// ---- extractPackageSignatures: ReadDir failure ----

func TestExtractPackageSignatures_ReadDirFail(t *testing.T) {
	t.Parallel()
	// Use a path that does not exist so ReadDir returns an error.
	pkg, warnings := extractPackageSignatures("/nonexistent/path/that/cannot/exist", ".")
	assert.Nil(t, pkg)
	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0], "read dir")
}

// ---- extractPackageSignatures: directory entries are skipped ----

func TestExtractPackageSignatures_SkipsSubDirs(t *testing.T) {
	t.Parallel()
	// Create a directory with a subdirectory and a .go file.
	root := t.TempDir()
	subDir := filepath.Join(root, "sub")
	require.NoError(t, os.MkdirAll(subDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(root, "main.go"),
		[]byte("package main\n\nfunc Exported() {}\n"),
		0o644,
	))

	pkg, warnings := extractPackageSignatures(root, ".")
	assert.Empty(t, warnings)
	require.NotNil(t, pkg)
	// The subdirectory must not be processed as a file.
	assert.Len(t, pkg.Signatures, 1)
	assert.Equal(t, "Exported", pkg.Signatures[0].Name)
}

// ---- extractPackageSignatures: TypeSpec.Comment used for doc ----

func TestExtract_TypeSpecCommentAsDoc(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	// Go source where a grouped type declaration uses an inline comment
	// on the TypeSpec (ts.Comment). This is the "Prefer TypeSpec doc" path
	// that differs from ts.Doc and d.Doc.
	writeFile(t, dir, "pkg/grouped/grouped.go", `package grouped

type (
	// Ignored group-level comment
	Foo struct{} // Foo is the grouped type.
)
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 1)
	require.Len(t, sm.Packages[0].Signatures, 1)

	sig := sm.Packages[0].Signatures[0]
	assert.Equal(t, "Foo", sig.Name)
	// ts.Comment takes priority over d.Doc when present.
	assert.Contains(t, sig.Doc, "Foo")
}

// ---- extractPackageSignatures: GenDecl doc used when no TypeSpec doc ----

func TestExtract_GenDeclDocFallback(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	// A grouped declaration where the doc comment is on the GenDecl (var/type block)
	// but not on the individual TypeSpec.
	writeFile(t, dir, "pkg/gendecl/gendecl.go", `package gendecl

// Bar represents a bar value.
type Bar struct{}
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 1)
	require.Len(t, sm.Packages[0].Signatures, 1)

	sig := sm.Packages[0].Signatures[0]
	assert.Equal(t, "Bar", sig.Name)
	assert.Contains(t, sig.Doc, "Bar")
}

// ---- extractPackageSignatures: non-TypeSpec specs in GenDecl are skipped ----

func TestExtract_ValueSpecSkipped(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	// var/const declarations are ValueSpec, not TypeSpec — must be skipped.
	writeFile(t, dir, "pkg/values/values.go", `package values

const Limit = 100
var Debug = false

func Utility() {}
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 1)
	// Only Utility; const and var must not appear.
	assert.Len(t, sm.Packages[0].Signatures, 1)
	assert.Equal(t, "Utility", sm.Packages[0].Signatures[0].Name)
}

// ---- Extract: hidden directory skipped ----

func TestExtract_HiddenDirSkipped(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "pkg/visible/visible.go", `package visible
func Visible() {}
`)
	writeFile(t, dir, ".hidden/pkg/hidden.go", `package hidden
func Hidden() {}
`)

	sm, err := Extract(dir)
	require.NoError(t, err)

	for _, p := range sm.Packages {
		assert.False(t, len(p.Path) > 0 && p.Path[0] == '.',
			"hidden directory package %q must be skipped", p.Path)
	}
}

// ---- Extract: multiple packages with warnings (parse error + good package) ----

func TestExtract_MultipleWarnings(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "pkg/alpha/alpha.go", `package alpha
func Alpha() {}
`)
	writeFile(t, dir, "pkg/alpha/broken.go", `package alpha
this is broken !!!
`)
	writeFile(t, dir, "pkg/beta/beta.go", `package beta
func Beta() {}
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	// Warnings contain the parse error; packages still extracted.
	assert.NotEmpty(t, sm.Warnings)

	paths := make(map[string]bool)
	for _, p := range sm.Packages {
		paths[p.Path] = true
	}
	assert.True(t, paths["pkg/beta"], "pkg/beta should be extracted")
}

// ---- Extract: type with TypeSpec.Doc (not just GenDecl doc) ----

func TestExtract_TypeSpecDocField(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	// TypeSpec inside a group block where each spec has its own doc comment
	// (this populates ts.Doc, not d.Doc or ts.Comment).
	writeFile(t, dir, "pkg/tsdoc/tsdoc.go", `package tsdoc

type (
	// Widget is a UI element.
	Widget struct{}

	// Gadget is a smart device.
	Gadget struct{}
)
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 1)
	assert.Len(t, sm.Packages[0].Signatures, 2)

	docs := make(map[string]string)
	for _, sig := range sm.Packages[0].Signatures {
		docs[sig.Name] = sig.Doc
	}
	assert.Contains(t, docs["Widget"], "Widget")
	assert.Contains(t, docs["Gadget"], "Gadget")
}
