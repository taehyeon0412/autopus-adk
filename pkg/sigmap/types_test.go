package sigmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignature_Fields(t *testing.T) {
	t.Parallel()

	sig := Signature{
		Name:       "NewServer",
		Kind:       "func",
		Receiver:   "",
		Params:     "(addr string, port int)",
		Returns:    "(error)",
		TypeParams: "",
		Doc:        "NewServer creates a new server.",
	}

	assert.Equal(t, "NewServer", sig.Name)
	assert.Equal(t, "func", sig.Kind)
	assert.Empty(t, sig.Receiver)
	assert.Equal(t, "(addr string, port int)", sig.Params)
	assert.Equal(t, "(error)", sig.Returns)
	assert.Empty(t, sig.TypeParams)
	assert.Equal(t, "NewServer creates a new server.", sig.Doc)
}

func TestSignature_MethodWithReceiver(t *testing.T) {
	t.Parallel()

	sig := Signature{
		Name:     "Run",
		Kind:     "method",
		Receiver: "(s *Server)",
		Params:   "(ctx context.Context)",
		Returns:  "(error)",
	}

	assert.Equal(t, "method", sig.Kind)
	assert.Equal(t, "(s *Server)", sig.Receiver)
}

func TestSignature_GenericTypeParams(t *testing.T) {
	t.Parallel()

	sig := Signature{
		Name:       "Map",
		Kind:       "func",
		TypeParams: "[T any, U comparable]",
		Params:     "(items []T, fn func(T) U)",
		Returns:    "([]U)",
	}

	assert.Equal(t, "[T any, U comparable]", sig.TypeParams)
}

func TestPackage_Fields(t *testing.T) {
	t.Parallel()

	pkg := Package{
		Path:  "pkg/adapter",
		Name:  "adapter",
		FanIn: 3,
		Depth: 2,
		Signatures: []Signature{
			{Name: "New", Kind: "func"},
		},
	}

	assert.Equal(t, "pkg/adapter", pkg.Path)
	assert.Equal(t, "adapter", pkg.Name)
	assert.Equal(t, 3, pkg.FanIn)
	assert.Equal(t, 2, pkg.Depth)
	assert.Len(t, pkg.Signatures, 1)
}

func TestPackage_EmptySignatures(t *testing.T) {
	t.Parallel()

	pkg := Package{
		Path: "pkg/empty",
		Name: "empty",
	}

	assert.Empty(t, pkg.Signatures)
	assert.Equal(t, 0, pkg.FanIn)
	assert.Equal(t, 0, pkg.Depth)
}

func TestSignatureMap_Fields(t *testing.T) {
	t.Parallel()

	sm := SignatureMap{
		ModulePath: "github.com/example/app",
		Packages: []Package{
			{Path: "pkg/a", Name: "a"},
			{Path: "pkg/b", Name: "b"},
		},
		Warnings: []string{"parse error in pkg/c"},
	}

	assert.Equal(t, "github.com/example/app", sm.ModulePath)
	assert.Len(t, sm.Packages, 2)
	assert.Len(t, sm.Warnings, 1)
	assert.Equal(t, "parse error in pkg/c", sm.Warnings[0])
}

func TestSignatureMap_Empty(t *testing.T) {
	t.Parallel()

	sm := SignatureMap{}

	assert.Empty(t, sm.ModulePath)
	assert.Empty(t, sm.Packages)
	assert.Empty(t, sm.Warnings)
}
