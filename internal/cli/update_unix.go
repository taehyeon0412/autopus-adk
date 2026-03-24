//go:build !windows

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// isWritable checks if the directory is writable by the current user.
func isWritable(dir string) bool {
	return syscall.Access(dir, syscall.O_RDWR) == nil
}

// reExecWithSudo re-executes the current command with sudo, inheriting stdin/stdout/stderr.
func reExecWithSudo() error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("현재 바이너리 경로를 가져올 수 없음: %w", err)
	}
	args := append([]string{exe}, os.Args[1:]...)
	cmd := exec.Command("sudo", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
