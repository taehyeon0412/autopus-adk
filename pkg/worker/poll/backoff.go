package poll

import (
	"sync"
	"time"
)

const (
	defaultMin    = 2 * time.Second
	defaultMax    = 60 * time.Second
	defaultFactor = 2.0
)

// AdaptiveBackoff implements exponential backoff with min/max bounds.
type AdaptiveBackoff struct {
	min     time.Duration
	max     time.Duration
	factor  float64
	current time.Duration
	mu      sync.Mutex
}

// NewAdaptiveBackoff returns a backoff starting at 2s, doubling up to 60s.
func NewAdaptiveBackoff() *AdaptiveBackoff {
	return &AdaptiveBackoff{
		min:     defaultMin,
		max:     defaultMax,
		factor:  defaultFactor,
		current: defaultMin,
	}
}

// Next returns the current backoff duration and increases it for the next call.
func (b *AdaptiveBackoff) Next() time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	d := b.current
	next := time.Duration(float64(b.current) * b.factor)
	if next > b.max {
		next = b.max
	}
	b.current = next
	return d
}

// Reset sets the backoff back to the minimum value.
// Typically called when a WebSocket push arrives.
func (b *AdaptiveBackoff) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.current = b.min
}

// Current returns the current backoff duration without advancing.
func (b *AdaptiveBackoff) Current() time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.current
}
