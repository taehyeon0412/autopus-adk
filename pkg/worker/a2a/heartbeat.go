package a2a

import (
	"context"
	"sync"
	"time"
)

const (
	defaultHeartbeatInterval = 30 * time.Second
	defaultHeartbeatTimeout  = 60 * time.Second
)

// Heartbeat sends periodic heartbeat messages over the A2A WebSocket
// and detects connection loss via timeout.
type Heartbeat struct {
	interval  time.Duration
	timeout   time.Duration
	sendFn    func() error
	onTimeout func()
	mu        sync.Mutex
	lastAck   time.Time
}

// NewHeartbeat creates a Heartbeat that calls sendFn every 30s.
// If no Ack is received within 60s, onTimeout is called.
func NewHeartbeat(sendFn func() error, onTimeout func()) *Heartbeat {
	return &Heartbeat{
		interval:  defaultHeartbeatInterval,
		timeout:   defaultHeartbeatTimeout,
		sendFn:    sendFn,
		onTimeout: onTimeout,
		lastAck:   time.Now(),
	}
}

// Start runs the heartbeat loop in a goroutine until ctx is cancelled.
func (h *Heartbeat) Start(ctx context.Context) {
	go h.run(ctx)
}

func (h *Heartbeat) run(ctx context.Context) {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.tick()
		}
	}
}

func (h *Heartbeat) tick() {
	h.mu.Lock()
	elapsed := time.Since(h.lastAck)
	h.mu.Unlock()

	if elapsed >= h.timeout {
		h.onTimeout()
		return
	}

	// Best-effort send; timeout check handles failures.
	_ = h.sendFn()
}

// Ack records that a heartbeat response was received.
func (h *Heartbeat) Ack() {
	h.mu.Lock()
	h.lastAck = time.Now()
	h.mu.Unlock()
}
