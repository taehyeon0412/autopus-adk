//go:build windows

package cli

import "fmt"

// isWritable on Windows always returns true — elevated permission handling is out of scope.
func isWritable(_ string) bool {
	return true
}

// reExecWithSudo is not supported on Windows.
func reExecWithSudo() error {
	return fmt.Errorf("Windows에서는 관리자 권한으로 직접 실행해주세요")
}
