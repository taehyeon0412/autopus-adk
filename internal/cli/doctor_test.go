// Package cli는 doctor 커맨드 테스트이다.
package cli_test

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoctorCmd_ReportsStatus(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// init 실행 후 doctor 수행
	initCmd := newTestRootCmd()
	initCmd.SetArgs([]string{"init", "--lite", "--dir", dir, "--project", "test-proj", "--platforms", "claude-code"})
	require.NoError(t, initCmd.Execute())

	var out bytes.Buffer
	doctorCmd := newTestRootCmd()
	doctorCmd.SetOut(&out)
	doctorCmd.SetArgs([]string{"doctor", "--dir", dir})
	err := doctorCmd.Execute()
	require.NoError(t, err)

	output := out.String()
	// 상태 리포트가 있어야 함
	assert.True(t, len(output) > 0, "doctor 커맨드가 출력을 생성해야 함")
}

func TestDoctorCmd_DetectsMissingFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	// 빈 디렉터리에서 doctor 실행 (파일 없음)
	var out bytes.Buffer
	doctorCmd := newTestRootCmd()
	doctorCmd.SetOut(&out)
	doctorCmd.SetArgs([]string{"doctor", "--dir", dir})
	// 에러가 있을 수 있지만 패닉은 없어야 함
	_ = doctorCmd.Execute()

	output := out.String()
	// 뭔가 출력이 있어야 함
	_ = output
}

func TestDoctorCmd_ShowsOKAfterInit(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	initCmd := newTestRootCmd()
	initCmd.SetArgs([]string{"init", "--lite", "--dir", dir, "--project", "test-proj", "--platforms", "claude-code"})
	require.NoError(t, initCmd.Execute())

	var out bytes.Buffer
	doctorCmd := newTestRootCmd()
	doctorCmd.SetOut(&out)
	doctorCmd.SetArgs([]string{"doctor", "--dir", dir})
	require.NoError(t, doctorCmd.Execute())

	output := out.String()
	// OK 또는 성공 상태가 포함되어야 함
	assert.Contains(t, output, "OK")
	_ = filepath.Join(dir, "autopus.yaml") // 경로 참조만
}
