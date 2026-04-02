package knowledge

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileWatcher polls a directory tree for file changes at a regular interval.
type FileWatcher struct {
	dir      string
	interval time.Duration
	onChange func(path string)
	excluder *Excluder

	mu     sync.Mutex
	mtimes map[string]time.Time
	cancel context.CancelFunc
}

// NewFileWatcher creates a new polling-based file watcher.
// The default poll interval is 500ms. An optional excluder filters out
// paths that should be ignored during scanning.
func NewFileWatcher(dir string, interval time.Duration, onChange func(path string), excluder *Excluder) *FileWatcher {
	if interval <= 0 {
		interval = 500 * time.Millisecond
	}
	return &FileWatcher{
		dir:      dir,
		interval: interval,
		onChange: onChange,
		excluder: excluder,
		mtimes:   make(map[string]time.Time),
	}
}

// Start begins polling the directory tree. It blocks until the context is
// cancelled or Stop is called.
func (w *FileWatcher) Start(ctx context.Context) error {
	ctx, w.cancel = context.WithCancel(ctx)

	// Initial scan to populate mtimes without triggering callbacks.
	w.scan(true)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			w.scan(false)
		}
	}
}

// Stop cancels the polling loop started by Start.
func (w *FileWatcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.cancel != nil {
		w.cancel()
	}
}

// scan walks the directory tree and detects changed or new files.
// After walking, it prunes entries for files that no longer exist on disk.
func (w *FileWatcher) scan(initial bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Track paths seen in this scan for delete pruning.
	seen := make(map[string]struct{})

	_ = filepath.Walk(w.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Skip excluded paths when an excluder is configured.
		if w.excluder != nil && w.excluder.IsExcluded(path) {
			return nil
		}

		seen[path] = struct{}{}
		mtime := info.ModTime()
		prev, known := w.mtimes[path]
		w.mtimes[path] = mtime

		if !initial && (!known || !mtime.Equal(prev)) {
			w.onChange(path)
		}
		return nil
	})

	// Prune deleted files from the mtimes map.
	for path := range w.mtimes {
		if _, ok := seen[path]; !ok {
			delete(w.mtimes, path)
		}
	}
}
