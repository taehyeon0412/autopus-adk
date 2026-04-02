package parallel

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTaskSemaphore_PanicsOnZero(t *testing.T) {
	t.Parallel()
	assert.Panics(t, func() { NewTaskSemaphore(0) })
}

func TestNewTaskSemaphore_PanicsOnNegative(t *testing.T) {
	t.Parallel()
	assert.Panics(t, func() { NewTaskSemaphore(-1) })
}

func TestTaskSemaphore_AcquireRelease(t *testing.T) {
	t.Parallel()

	sem := NewTaskSemaphore(2)
	ctx := context.Background()

	require.NoError(t, sem.Acquire(ctx))
	assert.Equal(t, 1, sem.Available())

	require.NoError(t, sem.Acquire(ctx))
	assert.Equal(t, 0, sem.Available())

	sem.Release()
	assert.Equal(t, 1, sem.Available())

	sem.Release()
	assert.Equal(t, 2, sem.Available())
}

func TestTaskSemaphore_ConcurrencyLimitEnforced(t *testing.T) {
	t.Parallel()

	const limit = 3
	sem := NewTaskSemaphore(limit)
	ctx := context.Background()

	// Fill all slots.
	for i := 0; i < limit; i++ {
		require.NoError(t, sem.Acquire(ctx))
	}
	assert.Equal(t, 0, sem.Available())

	// Next acquire should block; use a short-lived context to prove it.
	shortCtx, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
	defer cancel()

	err := sem.Acquire(shortCtx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	// Release all.
	for i := 0; i < limit; i++ {
		sem.Release()
	}
}

func TestTaskSemaphore_ContextCancellation(t *testing.T) {
	t.Parallel()

	sem := NewTaskSemaphore(1)
	ctx := context.Background()
	require.NoError(t, sem.Acquire(ctx))

	cancelCtx, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately.

	err := sem.Acquire(cancelCtx)
	assert.ErrorIs(t, err, context.Canceled)

	sem.Release()
}

func TestTaskSemaphore_Available(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		limit    int
		acquired int
		want     int
	}{
		{"none acquired", 5, 0, 5},
		{"some acquired", 5, 3, 2},
		{"all acquired", 5, 5, 0},
		{"single slot", 1, 0, 1},
		{"single slot full", 1, 1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sem := NewTaskSemaphore(tt.limit)
			ctx := context.Background()
			for i := 0; i < tt.acquired; i++ {
				require.NoError(t, sem.Acquire(ctx))
			}
			assert.Equal(t, tt.want, sem.Available())
			for i := 0; i < tt.acquired; i++ {
				sem.Release()
			}
		})
	}
}

func TestTaskSemaphore_Limit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		limit int
	}{
		{"one", 1},
		{"five", 5},
		{"hundred", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sem := NewTaskSemaphore(tt.limit)
			assert.Equal(t, tt.limit, sem.Limit())
		})
	}
}

func TestTaskSemaphore_ReleasePanicsWhenEmpty(t *testing.T) {
	t.Parallel()

	sem := NewTaskSemaphore(1)
	assert.Panics(t, func() { sem.Release() })
}

func TestTaskSemaphore_ConcurrentStress(t *testing.T) {
	t.Parallel()

	const (
		limit      = 5
		goroutines = 50
	)

	sem := NewTaskSemaphore(limit)
	ctx := context.Background()

	var maxConcurrent atomic.Int32
	var current atomic.Int32
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			require.NoError(t, sem.Acquire(ctx))

			n := current.Add(1)
			// Track the max observed concurrency.
			for {
				old := maxConcurrent.Load()
				if n <= old || maxConcurrent.CompareAndSwap(old, n) {
					break
				}
			}

			// Simulate work.
			time.Sleep(time.Millisecond)
			current.Add(-1)
			sem.Release()
		}()
	}

	wg.Wait()
	assert.LessOrEqual(t, int(maxConcurrent.Load()), limit)
	assert.Equal(t, limit, sem.Available())
}
