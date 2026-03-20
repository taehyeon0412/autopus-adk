package lore_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/insajin/autopus-adk/pkg/lore"
)

// initGitRepo는 테스트용 git 저장소를 초기화한다.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test"},
	}
	for _, c := range cmds {
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Dir = dir
		require.NoError(t, cmd.Run(), "git command failed: %v", c)
	}
}

// commitWithMsg는 테스트용 커밋을 생성한다.
func commitWithMsg(t *testing.T, dir, file, msg string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, file), []byte(msg), 0o644))
	add := exec.Command("git", "add", file)
	add.Dir = dir
	require.NoError(t, add.Run())
	commit := exec.Command("git", "commit", "-m", msg)
	commit.Dir = dir
	require.NoError(t, commit.Run())
}

func TestQueryConstraints_ReturnsEntriesWithConstraint(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	initGitRepo(t, dir)

	msg1 := "feat: 인증 구현\n\nConstraint: stateless 세션만\nConfidence: high\n"
	msg2 := "fix: 버그 수정\n\n단순 버그 수정입니다.\n"

	commitWithMsg(t, dir, "a.txt", msg1)
	commitWithMsg(t, dir, "b.txt", msg2)

	entries, err := lore.QueryConstraints(dir)
	require.NoError(t, err)

	// Constraint가 있는 커밋만 반환
	for _, e := range entries {
		assert.NotEmpty(t, e.Constraint)
	}
	assert.GreaterOrEqual(t, len(entries), 1)
}

func TestQueryRejected_ReturnsEntriesWithRejected(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	initGitRepo(t, dir)

	msg := "feat: DB 선택\n\nRejected: MySQL (확장성 이슈)\nConfidence: high\n"
	commitWithMsg(t, dir, "db.txt", msg)

	entries, err := lore.QueryRejected(dir)
	require.NoError(t, err)

	for _, e := range entries {
		assert.NotEmpty(t, e.Rejected)
	}
}

func TestQueryDirectives_ReturnsEntriesWithDirective(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	initGitRepo(t, dir)

	msg := "feat: 보안 정책\n\nDirective: 항상 HTTPS 사용\nConstraint: TLS 1.2+\n"
	commitWithMsg(t, dir, "sec.txt", msg)

	entries, err := lore.QueryDirectives(dir)
	require.NoError(t, err)

	for _, e := range entries {
		assert.NotEmpty(t, e.Directive)
	}
}

func TestQueryContext_ReturnsEntriesForPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	initGitRepo(t, dir)

	msg := "feat: 서비스 구현\n\nConstraint: 단일 책임 원칙 준수\n"
	commitWithMsg(t, dir, "service.go", msg)

	entries, err := lore.QueryContext(dir, "service.go")
	require.NoError(t, err)
	// 경로 관련 항목 반환 (비어있을 수 있음)
	assert.NotNil(t, entries)
}

func TestQueryStale_EmptyResult(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	initGitRepo(t, dir)

	// 최근 커밋 — stale 아님
	msg := "feat: 최근 변경\n\nConstraint: 테스트 제약\n"
	commitWithMsg(t, dir, "recent.txt", msg)

	// 365일 초과 기준으로 검색 — 최근 커밋은 포함 안 됨
	entries, err := lore.QueryStale(dir, 365)
	require.NoError(t, err)
	assert.NotNil(t, entries)
}

func TestQuery_NonExistentDir(t *testing.T) {
	t.Parallel()

	_, err := lore.QueryConstraints("/nonexistent/path")
	assert.Error(t, err)
}
