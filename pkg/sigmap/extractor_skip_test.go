package sigmap

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- Extract: vendor directory skipped ----

func TestExtract_VendorSkipped(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "pkg/real/real.go", `package real

func Real() {}
`)
	writeFile(t, dir, "vendor/github.com/foo/bar/bar.go", `package bar

func Vendor() {}
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 1)
	assert.Equal(t, "pkg/real", sm.Packages[0].Path)
}

// ---- Extract: parse error → warning, not failure ----

func TestExtract_ParseErrorWarning(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "pkg/good/good.go", `package good

func Good() {}
`)
	writeFile(t, dir, "pkg/bad/bad.go", `package bad

this is not valid go code !!!
`)

	sm, err := Extract(dir)
	require.NoError(t, err) // must not fail
	assert.NotEmpty(t, sm.Warnings)

	// good package is still present
	found := false
	for _, p := range sm.Packages {
		if p.Path == "pkg/good" {
			found = true
		}
	}
	assert.True(t, found, "pkg/good should still be present despite parse error in pkg/bad")
}

// ---- Extract: empty package skipped ----

func TestExtract_EmptyPackageSkipped(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "pkg/empty/empty.go", `package empty

func unexported() {}
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	assert.Empty(t, sm.Packages) // no exported symbols → package skipped
}

// ---- Extract: module path ----

func TestExtract_ModulePath(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "cmd/main.go", `package main

func Main() {}
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	assert.Equal(t, "example.com/testmod", sm.ModulePath)
}

// ---- Extract: packages sorted by path ----

func TestExtract_SortedByPath(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "pkg/z/z.go", `package z
func Z() {}
`)
	writeFile(t, dir, "pkg/a/a.go", `package a
func A() {}
`)
	writeFile(t, dir, "pkg/m/m.go", `package m
func M() {}
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 3)

	paths := make([]string, len(sm.Packages))
	for i, p := range sm.Packages {
		paths[i] = p.Path
	}
	assert.Equal(t, []string{"pkg/a", "pkg/m", "pkg/z"}, paths)
}

// ---- Extract: package depth ----

func TestExtract_PackageDepth(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "pkg/a/b/c/deep.go", `package c
func Deep() {}
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 1)
	// "pkg/a/b/c" has 3 slashes → depth 3
	assert.Equal(t, 3, sm.Packages[0].Depth)
}

// ---- Extract: multiple files in same package ----

func TestExtract_MultipleFilesInPackage(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "pkg/multi/a.go", `package multi
func FuncA() {}
`)
	writeFile(t, dir, "pkg/multi/b.go", `package multi
func FuncB() {}
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 1)
	assert.Len(t, sm.Packages[0].Signatures, 2)

	names := []string{sm.Packages[0].Signatures[0].Name, sm.Packages[0].Signatures[1].Name}
	assert.ElementsMatch(t, []string{"FuncA", "FuncB"}, names)
}

// ---- Extract: root-level package (no subdirectory) ----

func TestExtract_RootPackage(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "root.go", `package main

func MainFunc() {}
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	// Root package path should be "."
	require.Len(t, sm.Packages, 1)
	assert.Equal(t, ".", sm.Packages[0].Path)
	assert.Equal(t, 0, sm.Packages[0].Depth)

	found := false
	for _, sig := range sm.Packages[0].Signatures {
		if sig.Name == "MainFunc" {
			found = true
		}
	}
	assert.True(t, found)
}

// ---- Extract: node_modules directory skipped ----

func TestExtract_NodeModulesSkipped(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "pkg/real/real.go", `package real
func Real() {}
`)
	writeFile(t, dir, "node_modules/pkg/thing.go", `package thing
func Thing() {}
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	for _, p := range sm.Packages {
		assert.False(t, strings.HasPrefix(p.Path, "node_modules"),
			"node_modules package %q must be skipped", p.Path)
	}
}

// ---- Extract: exported const/var via TypeSpec is not extracted (only types) ----

func TestExtract_ConstVarNotExtracted(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "pkg/consts/consts.go", `package consts

const ExportedConst = 42
var ExportedVar = "hello"

func ExportedFunc() {}
`)
	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 1)
	// Only ExportedFunc is extracted (consts/vars are not TypeSpecs)
	assert.Len(t, sm.Packages[0].Signatures, 1)
	assert.Equal(t, "ExportedFunc", sm.Packages[0].Signatures[0].Name)
}

// ---- Extract: path separator normalization ----

func TestExtract_RelPathNormalization(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "pkg/alpha/beta/gamma/deep.go", `package gamma
func Deep() {}
`)
	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 1)
	// Path must use forward slashes even on Windows
	assert.NotContains(t, sm.Packages[0].Path, "\\")
}
