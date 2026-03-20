package lsp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/lsp"
)

// TestMockClient는 모의 클라이언트를 이용한 커맨드 테스트이다.
func TestMockClient_Diagnostics(t *testing.T) {
	t.Parallel()

	client := lsp.NewMockClient([]lsp.Diagnostic{
		{File: "main.go", Line: 10, Col: 5, Message: "undefined: foo", Severity: "error"},
		{File: "main.go", Line: 20, Col: 3, Message: "unused variable", Severity: "warning"},
	})

	diags, err := client.Diagnostics("main.go")
	require.NoError(t, err)
	require.Len(t, diags, 2)
	assert.Equal(t, "undefined: foo", diags[0].Message)
	assert.Equal(t, "error", diags[0].Severity)
}

func TestMockClient_References(t *testing.T) {
	t.Parallel()

	client := lsp.NewMockClient(nil)
	client.SetRefs("HandleRequest", []lsp.Location{
		{File: "main.go", Line: 5},
		{File: "handler_test.go", Line: 12},
	})

	refs, err := client.References("HandleRequest")
	require.NoError(t, err)
	assert.Len(t, refs, 2)
}

func TestMockClient_Symbols(t *testing.T) {
	t.Parallel()

	client := lsp.NewMockClient(nil)
	client.SetSymbols("api.go", []lsp.Symbol{
		{Name: "Handler", Kind: "struct", Location: lsp.Location{File: "api.go", Line: 5}},
		{Name: "Handle", Kind: "function", Location: lsp.Location{File: "api.go", Line: 15}},
	})

	syms, err := client.Symbols("api.go")
	require.NoError(t, err)
	assert.Len(t, syms, 2)
}

func TestMockClient_Definition(t *testing.T) {
	t.Parallel()

	client := lsp.NewMockClient(nil)
	client.SetDefinition("MyFunc", &lsp.Location{File: "impl.go", Line: 42})

	loc, err := client.Definition("MyFunc")
	require.NoError(t, err)
	require.NotNil(t, loc)
	assert.Equal(t, "impl.go", loc.File)
	assert.Equal(t, 42, loc.Line)
}

func TestMockClient_Rename(t *testing.T) {
	t.Parallel()

	client := lsp.NewMockClient(nil)
	err := client.Rename("oldName", "newName")
	assert.NoError(t, err)
}

func TestMockClient_DefinitionNotFound(t *testing.T) {
	t.Parallel()

	client := lsp.NewMockClient(nil)
	loc, err := client.Definition("unknownSymbol")
	require.NoError(t, err)
	assert.Nil(t, loc)
}
