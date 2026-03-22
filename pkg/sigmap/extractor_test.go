package sigmap

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeFile creates a file at path with content inside a temp directory.
func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, rel)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
}

// makeProject builds a minimal Go module in a temp dir and returns the path.
func makeProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/testmod\n\ngo 1.21\n")
	return dir
}

// ---- Extract: exported function ----

func TestExtract_ExportedFunc(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "pkg/server/server.go", `// Package server handles HTTP serving.
package server

// NewServer creates a new HTTP server. Additional detail here.
func NewServer(addr string, port int) error {
	return nil
}
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 1)

	pkg := sm.Packages[0]
	assert.Equal(t, "pkg/server", pkg.Path)
	assert.Equal(t, "server", pkg.Name)
	assert.Equal(t, 1, pkg.Depth) // "pkg/server" has 1 slash → depth 1
	require.Len(t, pkg.Signatures, 1)

	sig := pkg.Signatures[0]
	assert.Equal(t, "NewServer", sig.Name)
	assert.Equal(t, "func", sig.Kind)
	assert.Empty(t, sig.Receiver)
	assert.Equal(t, "NewServer creates a new HTTP server.", sig.Doc)
	assert.Contains(t, sig.Params, "addr")
	assert.Contains(t, sig.Params, "port")
}

// ---- Extract: unexported function skipped ----

func TestExtract_UnexportedFuncSkipped(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "internal/util/util.go", `package util

func helper() {}
func AnotherHelper() string { return "" }
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 1)
	assert.Len(t, sm.Packages[0].Signatures, 1)
	assert.Equal(t, "AnotherHelper", sm.Packages[0].Signatures[0].Name)
}

// ---- Extract: method with receiver ----

func TestExtract_Method(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "pkg/srv/srv.go", `package srv

type Server struct{}

// Run starts the server.
func (s *Server) Run() error { return nil }
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 1)

	var methodSig *Signature
	for i := range sm.Packages[0].Signatures {
		if sm.Packages[0].Signatures[i].Kind == "method" {
			methodSig = &sm.Packages[0].Signatures[i]
		}
	}
	require.NotNil(t, methodSig)
	assert.Equal(t, "Run", methodSig.Name)
	assert.Equal(t, "method", methodSig.Kind)
	assert.Contains(t, methodSig.Receiver, "Server")
	assert.Equal(t, "Run starts the server.", methodSig.Doc)
}

// ---- Extract: struct type ----

func TestExtract_StructType(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "pkg/model/model.go", `package model

// User represents an authenticated user.
type User struct {
	ID   int
	Name string
}
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 1)
	require.Len(t, sm.Packages[0].Signatures, 1)

	sig := sm.Packages[0].Signatures[0]
	assert.Equal(t, "User", sig.Name)
	assert.Equal(t, "type", sig.Kind)
	assert.Equal(t, "User represents an authenticated user.", sig.Doc)
}

// ---- Extract: interface type ----

func TestExtract_InterfaceType(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "pkg/iface/iface.go", `package iface

// Runner can run tasks.
type Runner interface {
	Run() error
}
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 1)
	require.Len(t, sm.Packages[0].Signatures, 1)

	sig := sm.Packages[0].Signatures[0]
	assert.Equal(t, "Runner", sig.Name)
	assert.Equal(t, "interface", sig.Kind)
}

// ---- Extract: generics ----

func TestExtract_GenericFunc(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "pkg/gen/gen.go", `package gen

// Map transforms a slice using a function.
func Map[T any, U comparable](items []T, fn func(T) U) []U {
	return nil
}
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 1)
	require.Len(t, sm.Packages[0].Signatures, 1)

	sig := sm.Packages[0].Signatures[0]
	assert.Equal(t, "Map", sig.Name)
	assert.Contains(t, sig.TypeParams, "T")
	assert.Contains(t, sig.TypeParams, "U")
}

func TestExtract_GenericType(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "pkg/container/container.go", `package container

// Stack is a generic LIFO data structure.
type Stack[T any] struct {
	items []T
}
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 1)
	require.Len(t, sm.Packages[0].Signatures, 1)

	sig := sm.Packages[0].Signatures[0]
	assert.Equal(t, "Stack", sig.Name)
	assert.Equal(t, "type", sig.Kind)
	assert.Contains(t, sig.TypeParams, "T")
}

// ---- Extract: test files skipped ----

func TestExtract_TestFilesSkipped(t *testing.T) {
	t.Parallel()
	dir := makeProject(t)
	writeFile(t, dir, "pkg/foo/foo.go", `package foo

func Exported() {}
`)
	writeFile(t, dir, "pkg/foo/foo_test.go", `package foo

func TestSomething(t *testing.T) {}
func ExportedTest() {}
`)

	sm, err := Extract(dir)
	require.NoError(t, err)
	require.Len(t, sm.Packages, 1)
	// Only Exported from foo.go; ExportedTest from _test.go must be skipped
	assert.Len(t, sm.Packages[0].Signatures, 1)
	assert.Equal(t, "Exported", sm.Packages[0].Signatures[0].Name)
}
