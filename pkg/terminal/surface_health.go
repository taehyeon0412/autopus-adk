// Package terminal provides optional signal-based communication interfaces.
package terminal

import (
	"context"
	"errors"
	"time"
)

// SurfaceStatus represents the health status of a terminal surface.
type SurfaceStatus struct {
	Valid      bool   // Surface is responsive
	SurfaceRef string // Surface reference (e.g., "surface:7")
	InWindow   bool   // Surface is visible in a window
}

// SignalCapable is an optional interface for terminal adapters that support
// signal-based communication. Use type assertion to check:
//
//	if sc, ok := term.(SignalCapable); ok { ... }
//
// @AX:ANCHOR [AUTO] optional interface — implemented by CmuxAdapter; checked via type assertion in NewCompletionDetector, NewSurfaceManager
type SignalCapable interface {
	// SurfaceHealth returns the health status of a surface without reading its content.
	SurfaceHealth(ctx context.Context, paneID PaneID) (SurfaceStatus, error)
	// WaitForSignal blocks until the named signal is received or timeout expires.
	WaitForSignal(ctx context.Context, name string, timeout time.Duration) error
	// SendSignal sends a named signal that unblocks any WaitForSignal waiters.
	SendSignal(ctx context.Context, name string) error
}

// ErrSignalNotSupported is returned by adapters that do not support signal operations.
var ErrSignalNotSupported = errors.New("terminal adapter does not support signal operations")
