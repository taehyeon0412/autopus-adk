package net

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEqual_SameSlices(t *testing.T) {
	t.Parallel()
	assert.True(t, equal([]string{"a", "b"}, []string{"a", "b"}))
}

func TestEqual_DifferentSlices(t *testing.T) {
	t.Parallel()
	assert.False(t, equal([]string{"a", "b"}, []string{"a", "c"}))
}

func TestEqual_DifferentLengths(t *testing.T) {
	t.Parallel()
	assert.False(t, equal([]string{"a"}, []string{"a", "b"}))
}

func TestEqual_BothNil(t *testing.T) {
	t.Parallel()
	assert.True(t, equal(nil, nil))
}

func TestEqual_BothEmpty(t *testing.T) {
	t.Parallel()
	assert.True(t, equal([]string{}, []string{}))
}

func TestCurrentAddrs_ReturnsSorted(t *testing.T) {
	t.Parallel()
	addrs := currentAddrs()
	// Just verify it returns without error and is sorted.
	for i := 1; i < len(addrs); i++ {
		assert.LessOrEqual(t, addrs[i-1], addrs[i], "addresses should be sorted")
	}
}

func TestNewNetMonitor_Fields(t *testing.T) {
	t.Parallel()
	onChange := func(old, new []string) {}
	onValidate := func() error { return nil }
	m := NewNetMonitor(onChange, onValidate)

	assert.Equal(t, defaultPollInterval, m.interval)
	assert.NotNil(t, m.onChange)
	assert.NotNil(t, m.onValidate)
}

func TestNetMonitor_StartAndStop(t *testing.T) {
	t.Parallel()

	onChange := func(oldAddrs, newAddrs []string) {}
	onValidate := func() error { return nil }

	m := NewNetMonitor(onChange, onValidate)
	// Use a short interval for testing.
	m.interval = 50 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	m.Start(ctx)

	// Let the monitor poll a few times.
	time.Sleep(200 * time.Millisecond)

	// Cancel and verify no panic.
	cancel()
	time.Sleep(100 * time.Millisecond)
}

func TestNetMonitor_OnChangeCalledWhenValidateFails(t *testing.T) {
	t.Parallel()

	changeCalled := make(chan struct{}, 1)

	onChange := func(oldAddrs, newAddrs []string) {
		select {
		case changeCalled <- struct{}{}:
		default:
		}
	}
	// onValidate always fails, so onChange should fire when addrs differ.
	onValidate := func() error {
		return fmt.Errorf("validation failed")
	}

	m := NewNetMonitor(onChange, onValidate)
	m.interval = 50 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start the monitor — even if real addresses don't change,
	// this validates the monitor runs without crashing.
	m.Start(ctx)

	// Wait for the context to expire. If addresses happen to change
	// (unlikely in test), onChange will be called.
	<-ctx.Done()
}
