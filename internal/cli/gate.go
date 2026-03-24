package cli

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

// GateMode defines whether a gate failure blocks execution or is advisory only.
type GateMode string

const (
	// GateModeMandatory causes GateCheck to return Passed=false on gate failure.
	GateModeMandatory GateMode = "mandatory"
	// GateModeAdvisory causes GateCheck to return Passed=true with a Warning on gate failure.
	GateModeAdvisory GateMode = "advisory"
)

// GateConfig holds configuration for a single gate check invocation.
type GateConfig struct {
	// GateName identifies which gate logic to apply (e.g. "phase2").
	GateName string
	// Mode controls whether a failure blocks (mandatory) or warns (advisory).
	Mode GateMode
	// Dir is the project root directory to inspect.
	Dir string
}

// GateResult is the outcome of a GateCheck call.
type GateResult struct {
	// Passed is true when the gate condition is met or mode is advisory.
	Passed bool
	// Message contains the failure reason when Passed is false.
	Message string
	// Warning contains an advisory message when the gate condition failed but
	// mode is advisory (Passed is still true in this case).
	Warning string
	// Err is set when the gate name is unknown or an internal error occurs.
	Err error
}

// @AX:NOTE [AUTO] [downgraded from ANCHOR — fan_in < 3] internal registry for gate check functions; only GateCheck reads this map
// knownGates lists gate names that GateCheck recognises.
var knownGates = map[string]gateChecker{
	"phase2": checkPhase2Gate,
}

type gateChecker func(dir string) (bool, string)

// GateCheck evaluates the named gate against the given directory.
// Returns a GateResult indicating pass/fail and any messages.
func GateCheck(cfg GateConfig) GateResult {
	checker, ok := knownGates[cfg.GateName]
	if !ok {
		return GateResult{
			Passed: false,
			Err:    fmt.Errorf("unknown gate: %q", cfg.GateName),
		}
	}

	passed, reason := checker(cfg.Dir)
	if passed {
		return GateResult{Passed: true}
	}

	if cfg.Mode == GateModeAdvisory {
		return GateResult{
			Passed:  true,
			Warning: fmt.Sprintf("gate %q advisory: %s", cfg.GateName, reason),
		}
	}
	return GateResult{
		Passed:  false,
		Message: fmt.Sprintf("gate %q failed: %s", cfg.GateName, reason),
	}
}

// checkPhase2Gate verifies that at least one *_test.go file exists in dir or
// any of its subdirectories. Returns (true, "") when found, (false, reason) otherwise.
func checkPhase2Gate(dir string) (bool, string) {
	found := false
	err := filepath.WalkDir(dir, func(_ string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil // skip unreadable entries
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), "_test.go") {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil {
		return false, fmt.Sprintf("cannot walk directory: %v", err)
	}
	if !found {
		return false, "no test files found (expected *_test.go)"
	}
	return true, ""
}
