package selfupdate

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// Replacer atomically replaces a binary file.
type Replacer struct{}

// NewReplacer creates a new Replacer.
func NewReplacer() *Replacer {
	return &Replacer{}
}

// Replace atomically replaces the target binary with the new one.
func (r *Replacer) Replace(newBinaryPath, targetPath string) error {
	targetInfo, err := os.Stat(targetPath)
	if err != nil {
		return err
	}

	// @AX:NOTE: [AUTO] os.Rename is atomic only within the same filesystem — will fail cross-device (e.g., tmpfs → /usr/local/bin)
	if err := os.Rename(newBinaryPath, targetPath); err != nil {
		// Windows: running exe cannot be overwritten, but CAN be renamed.
		// Move the running binary aside, then place the new one.
		if runtime.GOOS == "windows" && os.IsPermission(err) {
			if winErr := r.replaceWindows(newBinaryPath, targetPath); winErr != nil {
				return fmt.Errorf("permission denied replacing %s: %w. Check directory permissions", targetPath, winErr)
			}
		} else if os.IsPermission(err) {
			return fmt.Errorf("permission denied replacing %s: %w. Check directory permissions", targetPath, err)
		} else {
			return err
		}
	}

	if err := os.Chmod(targetPath, targetInfo.Mode().Perm()); err != nil {
		return err
	}

	// macOS: clear extended attributes and ad-hoc codesign to prevent Gatekeeper from killing the binary.
	// com.apple.quarantine and com.apple.provenance both trigger SIGKILL on macOS Sequoia+.
	if runtime.GOOS == "darwin" {
		_ = exec.Command("xattr", "-cr", targetPath).Run()
		_ = exec.Command("codesign", "--force", "--sign", "-", targetPath).Run()
	}

	return nil
}

// replaceWindows handles binary replacement on Windows where a running exe
// cannot be overwritten but can be renamed. Moves the old binary to .old,
// then renames the new binary into place.
func (r *Replacer) replaceWindows(newBinaryPath, targetPath string) error {
	oldPath := targetPath + ".old"

	// Clean up any previous .old file.
	_ = os.Remove(oldPath)

	// Rename the running binary out of the way.
	if err := os.Rename(targetPath, oldPath); err != nil {
		return fmt.Errorf("rename old binary: %w", err)
	}

	// Place the new binary at the target path.
	if err := os.Rename(newBinaryPath, targetPath); err != nil {
		// Attempt to restore the old binary on failure.
		_ = os.Rename(oldPath, targetPath)
		return fmt.Errorf("rename new binary: %w", err)
	}

	// Best-effort cleanup — the .old file may still be locked by the OS.
	_ = os.Remove(oldPath)
	return nil
}
