package poll

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAdaptiveBackoff_InitialValue(t *testing.T) {
	t.Parallel()
	b := NewAdaptiveBackoff()
	assert.Equal(t, 2*time.Second, b.Current())
}

func TestAdaptiveBackoff_ExponentialIncrease(t *testing.T) {
	t.Parallel()
	b := NewAdaptiveBackoff()

	expected := []time.Duration{
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		16 * time.Second,
		32 * time.Second,
		60 * time.Second, // capped at max
		60 * time.Second, // stays at max
	}

	for i, want := range expected {
		got := b.Next()
		assert.Equal(t, want, got, "step %d", i)
	}
}

func TestAdaptiveBackoff_Reset(t *testing.T) {
	t.Parallel()
	b := NewAdaptiveBackoff()

	// Advance a few times.
	b.Next()
	b.Next()
	b.Next()
	assert.NotEqual(t, 2*time.Second, b.Current())

	b.Reset()
	assert.Equal(t, 2*time.Second, b.Current())
	assert.Equal(t, 2*time.Second, b.Next(), "first Next after Reset should return min")
}

func TestAdaptiveBackoff_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	b := NewAdaptiveBackoff()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Next()
			b.Current()
			b.Reset()
		}()
	}
	wg.Wait()
	// No race detector panic = pass.
}
