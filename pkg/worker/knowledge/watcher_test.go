package knowledge

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileWatcher_DetectsNewFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	var mu sync.Mutex
	var changed []string

	w := NewFileWatcher(dir, 50*time.Millisecond, func(path string) {
		mu.Lock()
		changed = append(changed, path)
		mu.Unlock()
	}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- w.Start(ctx) }()

	// Wait for initial scan to complete.
	time.Sleep(100 * time.Millisecond)

	// Create a new file after the initial scan.
	err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("hello"), 0644)
	require.NoError(t, err)

	// Wait for detection.
	time.Sleep(200 * time.Millisecond)

	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()
	assert.NotEmpty(t, changed, "should detect new file")
}

func TestFileWatcher_DetectsModifiedFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "existing.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("v1"), 0644))

	var mu sync.Mutex
	var changed []string

	w := NewFileWatcher(dir, 50*time.Millisecond, func(path string) {
		mu.Lock()
		changed = append(changed, path)
		mu.Unlock()
	}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- w.Start(ctx) }()

	// Wait for initial scan.
	time.Sleep(100 * time.Millisecond)

	// Modify the file with a future mtime to ensure detection.
	require.NoError(t, os.WriteFile(filePath, []byte("v2"), 0644))
	future := time.Now().Add(1 * time.Second)
	require.NoError(t, os.Chtimes(filePath, future, future))

	time.Sleep(200 * time.Millisecond)

	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()
	assert.NotEmpty(t, changed, "should detect modified file")
}

func TestFileWatcher_InitialScanNoCallback(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "pre.txt"), []byte("data"), 0644))

	callCount := 0
	var mu sync.Mutex

	w := NewFileWatcher(dir, 50*time.Millisecond, func(_ string) {
		mu.Lock()
		callCount++
		mu.Unlock()
	}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- w.Start(ctx) }()

	// Wait enough for initial scan + one poll cycle with no changes.
	time.Sleep(200 * time.Millisecond)

	cancel()
	<-done

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 0, callCount, "initial scan should not trigger callback")
}

func TestFileWatcher_StopCancelsPolling(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	w := NewFileWatcher(dir, 50*time.Millisecond, func(_ string) {}, nil)

	ctx := context.Background()
	done := make(chan error, 1)
	go func() { done <- w.Start(ctx) }()

	time.Sleep(100 * time.Millisecond)
	w.Stop()

	err := <-done
	assert.ErrorIs(t, err, context.Canceled)
}

func TestFileWatcher_DefaultInterval(t *testing.T) {
	t.Parallel()

	w := NewFileWatcher(t.TempDir(), 0, func(_ string) {}, nil)
	assert.Equal(t, 500*time.Millisecond, w.interval)
}
