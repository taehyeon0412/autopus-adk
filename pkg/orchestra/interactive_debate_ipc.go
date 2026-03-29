package orchestra

import (
	"context"
	"log"
	"time"
)

// fileIPCReadyTimeout is the default timeout for waiting for a provider's ready signal.
// @AX:NOTE [AUTO] magic constant 30s — must be less than per-round timeout; increase if providers are slow to signal ready
const fileIPCReadyTimeout = 30 * time.Second

// tryFileIPC attempts to deliver a prompt via file IPC for hook-capable providers.
// Returns true if file IPC succeeded and SendLongText should be skipped.
// Returns false if SendLongText fallback is needed.
// R5-SAFETY: On write failure after ready, sends abort signal to prevent hook deadlock.
func tryFileIPC(ctx context.Context, hookSession *HookSession, provider string, round int, prompt string) bool {
	if err := hookSession.WaitForReadyCtx(ctx, fileIPCReadyTimeout, provider, round); err != nil {
		log.Printf("[Round %d] %s WaitForReady failed: %v — falling back to SendLongText", round, provider, err)
		return false
	}

	if err := hookSession.WriteInputRound(provider, round, prompt); err != nil {
		log.Printf("[Round %d] %s WriteInputRound failed: %v — falling back to SendLongText", round, provider, err)
		_ = hookSession.WriteAbortSignal(provider, round)
		return false
	}

	return true
}
