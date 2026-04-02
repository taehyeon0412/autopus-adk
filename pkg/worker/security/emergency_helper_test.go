package security

import "syscall"

// sysProcAttr returns SysProcAttr with Setpgid for process group termination.
func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}
