package a2a

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeartbeat_SendsAtInterval(t *testing.T) {
	t.Parallel()

	var count atomic.Int32
	hb := NewHeartbeat(func() error {
		count.Add(1)
		return nil
	}, func() {})
	hb.interval = 10 * time.Millisecond
	hb.timeout = 200 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	hb.Start(ctx)

	time.Sleep(55 * time.Millisecond)
	cancel()
	time.Sleep(5 * time.Millisecond)

	got := count.Load()
	assert.GreaterOrEqual(t, got, int32(3), "expected at least 3 sends in 55ms with 10ms interval")
}

func TestHeartbeat_AckResetsTimeout(t *testing.T) {
	t.Parallel()

	var timedOut atomic.Bool
	hb := NewHeartbeat(func() error { return nil }, func() {
		timedOut.Store(true)
	})
	hb.interval = 10 * time.Millisecond
	hb.timeout = 30 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hb.Start(ctx)

	// Keep acking to prevent timeout.
	for i := 0; i < 5; i++ {
		time.Sleep(15 * time.Millisecond)
		hb.Ack()
	}

	assert.False(t, timedOut.Load(), "should not timeout when Ack is called regularly")
}

func TestHeartbeat_TimeoutFires(t *testing.T) {
	t.Parallel()

	var timedOut atomic.Bool
	hb := NewHeartbeat(func() error { return nil }, func() {
		timedOut.Store(true)
	})
	hb.interval = 5 * time.Millisecond
	hb.timeout = 20 * time.Millisecond
	// Set lastAck far in the past so timeout triggers immediately.
	hb.lastAck = time.Now().Add(-1 * time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	hb.Start(ctx)

	require.Eventually(t, func() bool {
		return timedOut.Load()
	}, 100*time.Millisecond, 2*time.Millisecond, "timeout callback should fire")
}

func TestHeartbeat_ContextCancellation(t *testing.T) {
	t.Parallel()

	var count atomic.Int32
	hb := NewHeartbeat(func() error {
		count.Add(1)
		return nil
	}, func() {})
	hb.interval = 10 * time.Millisecond
	hb.timeout = 1 * time.Second

	ctx, cancel := context.WithCancel(context.Background())
	hb.Start(ctx)

	time.Sleep(50 * time.Millisecond)
	cancel()
	// Give the goroutine time to observe the cancellation.
	time.Sleep(20 * time.Millisecond)
	snapshot := count.Load()
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, snapshot, count.Load(), "no more sends after context cancelled")
}
