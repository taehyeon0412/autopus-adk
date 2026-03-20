// Package cli_test는 LSP/search CLI 커맨드의 에러 경로 테스트를 제공한다.
package cli_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLSPDiagnosticsCmd_NoServer는 LSP 서버 없이 diagnostics를 실행할 때 에러를 테스트한다.
func TestLSPDiagnosticsCmd_NoServer(t *testing.T) {
	t.Parallel()

	cmd := newTestRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"lsp", "diagnostics", "nonexistent.go"})
	err := cmd.Execute()
	// LSP 서버가 없으므로 에러가 발생해야 함
	assert.Error(t, err)
}

// TestLSPRefsCmd_NoServer는 LSP 서버 없이 refs를 실행할 때 에러를 테스트한다.
func TestLSPRefsCmd_NoServer(t *testing.T) {
	t.Parallel()

	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"lsp", "refs", "SomeSymbol"})
	err := cmd.Execute()
	assert.Error(t, err)
}

// TestLSPRenameCmd_NoServer는 LSP 서버 없이 rename을 실행할 때 에러를 테스트한다.
func TestLSPRenameCmd_NoServer(t *testing.T) {
	t.Parallel()

	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"lsp", "rename", "OldName", "NewName"})
	err := cmd.Execute()
	assert.Error(t, err)
}

// TestLSPSymbolsCmd_NoServer는 LSP 서버 없이 symbols를 실행할 때 에러를 테스트한다.
func TestLSPSymbolsCmd_NoServer(t *testing.T) {
	t.Parallel()

	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"lsp", "symbols", "some_file.go"})
	err := cmd.Execute()
	assert.Error(t, err)
}

// TestLSPDefinitionCmd_NoServer는 LSP 서버 없이 definition을 실행할 때 에러를 테스트한다.
func TestLSPDefinitionCmd_NoServer(t *testing.T) {
	t.Parallel()

	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"lsp", "definition", "SomeFunc"})
	err := cmd.Execute()
	assert.Error(t, err)
}

// TestLSPDiagnosticsCmd_NoArgs는 인자 없이 diagnostics를 실행할 때 에러를 테스트한다.
func TestLSPDiagnosticsCmd_NoArgs(t *testing.T) {
	t.Parallel()

	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"lsp", "diagnostics"})
	err := cmd.Execute()
	assert.Error(t, err, "인자 없이 실행 시 에러가 발생해야 함")
}

// TestLSPRenameCmd_OneArg는 인자 1개로 rename을 실행할 때 에러를 테스트한다.
func TestLSPRenameCmd_OneArg(t *testing.T) {
	t.Parallel()

	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"lsp", "rename", "OnlyOldName"})
	err := cmd.Execute()
	assert.Error(t, err, "인자 부족 시 에러가 발생해야 함")
}

// TestDocsCmd_InvalidLibrary2는 잘못된 라이브러리 이름으로 docs를 테스트한다.
func TestDocsCmd_InvalidLibrary2(t *testing.T) {
	t.Parallel()

	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"docs", "nonexistent-library-xyz-12345"})
	err := cmd.Execute()
	// 존재하지 않는 라이브러리이므로 에러 발생
	assert.Error(t, err)
}

// TestSearchCmd_WithNumFlag2는 --num 플래그를 테스트한다.
func TestSearchCmd_WithNumFlag2(t *testing.T) {
	// t.Setenv과 t.Parallel은 동시 사용 불가
	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"search", "--num", "3", "test query"})
	err := cmd.Execute()
	// API 키 없으므로 에러 (플래그 파싱은 성공)
	assert.Error(t, err)
}

// TestDocsCmd_WithTopicFlag2는 --topic 플래그를 테스트한다.
func TestDocsCmd_WithTopicFlag2(t *testing.T) {
	t.Parallel()

	cmd := newTestRootCmd()
	cmd.SetArgs([]string{"docs", "--topic", "installation", "nonexistent-lib"})
	err := cmd.Execute()
	assert.Error(t, err)
}
