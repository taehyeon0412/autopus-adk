package detect

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// PermissionResult holds the detected permission mode and metadata.
type PermissionResult struct {
	Mode      string `json:"mode"`       // "bypass" or "safe"
	ParentPID int    `json:"parent_pid"` // parent process PID
	FlagFound bool   `json:"flag_found"` // whether the flag was found
}

const (
	modeBypass = "bypass"
	modeSafe  = "safe"

	permissionFlag = "--dangerously-skip-permissions"
	envKey         = "AUTOPUS_PERMISSION_MODE"
	// @AX:NOTE [AUTO] magic constant — max process tree traversal depth, prevents infinite loops on circular PID references
	maxTreeDepth = 20
)

// DetectPermissionMode checks the parent process tree for the
// --dangerously-skip-permissions flag and returns the permission mode.
// Environment variable AUTOPUS_PERMISSION_MODE overrides process detection.
func DetectPermissionMode() PermissionResult {
	if env := os.Getenv(envKey); env == modeBypass || env == modeSafe {
		return PermissionResult{Mode: env}
	}
	return detectPermissionModeWith(checkParentProcessTree)
}

// detectPermissionModeWith uses an injectable checker for testability.
func detectPermissionModeWith(checker func() (bool, int, error)) PermissionResult {
	flagFound, parentPID, err := checker()
	if err != nil {
		return PermissionResult{Mode: modeSafe}
	}
	if flagFound {
		return PermissionResult{Mode: modeBypass, ParentPID: parentPID, FlagFound: true}
	}
	return PermissionResult{Mode: modeSafe, ParentPID: parentPID}
}

// checkParentProcessTree walks up the process tree looking for the
// --dangerously-skip-permissions flag in command arguments.
func checkParentProcessTree() (flagFound bool, parentPID int, err error) {
	pid := os.Getppid()
	parentPID = pid

	for i := 0; i < maxTreeDepth && pid > 1; i++ {
		args, argsErr := processArgs(pid)
		if argsErr != nil {
			return false, parentPID, argsErr
		}
		if strings.Contains(args, permissionFlag) {
			return true, parentPID, nil
		}

		nextPID, ppidErr := parentPIDOf(pid)
		if ppidErr != nil {
			return false, parentPID, ppidErr
		}
		pid = nextPID
	}
	return false, parentPID, nil
}

// processArgs returns the command-line arguments for a given PID.
func processArgs(pid int) (string, error) {
	out, err := exec.Command("ps", "-o", "args=", "-p", fmt.Sprintf("%d", pid)).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// parentPIDOf returns the parent PID of the given PID.
func parentPIDOf(pid int) (int, error) {
	out, err := exec.Command("ps", "-o", "ppid=", "-p", fmt.Sprintf("%d", pid)).Output()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(out)))
}
