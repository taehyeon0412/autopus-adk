package a2a

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskLifecycle_ValidTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		from TaskStatus
		to   TaskStatus
	}{
		{"working to completed", StatusWorking, StatusCompleted},
		{"working to failed", StatusWorking, StatusFailed},
		{"working to canceled", StatusWorking, StatusCanceled},
		{"working to input-required", StatusWorking, StatusInputRequired},
		{"input-required to working", StatusInputRequired, StatusWorking},
		{"input-required to canceled", StatusInputRequired, StatusCanceled},
		{"input-required to failed", StatusInputRequired, StatusFailed},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			task := &Task{ID: "t-1", Status: tc.from}
			lc := NewTaskLifecycle(task)
			require.NoError(t, lc.Transition(tc.to))
			assert.Equal(t, tc.to, lc.Status())
		})
	}
}

func TestTaskLifecycle_InvalidTransitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		from TaskStatus
		to   TaskStatus
	}{
		{"completed to working", StatusCompleted, StatusWorking},
		{"failed to working", StatusFailed, StatusWorking},
		{"canceled to working", StatusCanceled, StatusWorking},
		{"completed to failed", StatusCompleted, StatusFailed},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			task := &Task{ID: "t-1", Status: tc.from}
			lc := NewTaskLifecycle(task)
			err := lc.Transition(tc.to)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid transition")
		})
	}
}

func TestTaskLifecycle_IsTerminal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status   TaskStatus
		terminal bool
	}{
		{StatusWorking, false},
		{StatusInputRequired, false},
		{StatusCompleted, true},
		{StatusFailed, true},
		{StatusCanceled, true},
	}

	for _, tc := range tests {
		t.Run(string(tc.status), func(t *testing.T) {
			t.Parallel()
			task := &Task{ID: "t-1", Status: tc.status}
			lc := NewTaskLifecycle(task)
			assert.Equal(t, tc.terminal, lc.IsTerminal())
		})
	}
}

func TestTaskLifecycle_TransitionListener(t *testing.T) {
	t.Parallel()

	task := &Task{ID: "t-listen", Status: StatusWorking}
	lc := NewTaskLifecycle(task)

	var captured struct {
		taskID string
		from   TaskStatus
		to     TaskStatus
	}
	lc.AddListener(func(taskID string, from, to TaskStatus) {
		captured.taskID = taskID
		captured.from = from
		captured.to = to
	})

	require.NoError(t, lc.Transition(StatusCompleted))
	assert.Equal(t, "t-listen", captured.taskID)
	assert.Equal(t, StatusWorking, captured.from)
	assert.Equal(t, StatusCompleted, captured.to)
}

func TestLifecycleManager_TrackAndGet(t *testing.T) {
	t.Parallel()

	mgr := NewLifecycleManager()
	task := &Task{ID: "mgr-1", Status: StatusWorking}

	lc := mgr.Track(task)
	require.NotNil(t, lc)

	got, ok := mgr.Get("mgr-1")
	require.True(t, ok)
	assert.Equal(t, lc, got)

	_, ok = mgr.Get("nonexistent")
	assert.False(t, ok)
}

func TestLifecycleManager_ActiveTasks(t *testing.T) {
	t.Parallel()

	mgr := NewLifecycleManager()
	mgr.Track(&Task{ID: "active-1", Status: StatusWorking})
	mgr.Track(&Task{ID: "active-2", Status: StatusInputRequired})
	mgr.Track(&Task{ID: "done-1", Status: StatusCompleted})
	mgr.Track(&Task{ID: "done-2", Status: StatusFailed})

	active := mgr.ActiveTasks()
	ids := make(map[string]bool)
	for _, t := range active {
		ids[t.ID] = true
	}
	assert.Len(t, active, 2)
	assert.True(t, ids["active-1"])
	assert.True(t, ids["active-2"])
}

func TestLifecycleManager_Remove(t *testing.T) {
	t.Parallel()

	mgr := NewLifecycleManager()
	mgr.Track(&Task{ID: "rm-1", Status: StatusWorking})

	_, ok := mgr.Get("rm-1")
	require.True(t, ok)

	mgr.Remove("rm-1")
	_, ok = mgr.Get("rm-1")
	assert.False(t, ok)

	// Removing a non-existent task should not panic.
	mgr.Remove("rm-1")
}

func TestLifecycleManager_AddListener_Propagation(t *testing.T) {
	t.Parallel()

	mgr := NewLifecycleManager()
	task := &Task{ID: "prop-1", Status: StatusWorking}
	mgr.Track(task)

	var called atomic.Bool
	mgr.AddListener(func(_ string, _, _ TaskStatus) {
		called.Store(true)
	})

	lc, _ := mgr.Get("prop-1")
	require.NoError(t, lc.Transition(StatusCompleted))
	assert.True(t, called.Load())
}

func TestTaskLifecycle_ConcurrentTransitions(t *testing.T) {
	t.Parallel()

	// All goroutines try to transition to terminal states from working.
	// Only terminal targets ensure exactly one succeeds (no chain transitions).
	task := &Task{ID: "race-1", Status: StatusWorking}
	lc := NewTaskLifecycle(task)

	var wg sync.WaitGroup
	targets := []TaskStatus{StatusCompleted, StatusFailed, StatusCanceled}

	var successCount atomic.Int32
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			to := targets[idx%len(targets)]
			if err := lc.Transition(to); err == nil {
				successCount.Add(1)
			}
		}(i)
	}
	wg.Wait()

	// Exactly one transition should succeed; the rest fail because the
	// first success moves to a terminal state with no valid outbound transitions.
	assert.Equal(t, int32(1), successCount.Load())
	assert.True(t, lc.IsTerminal())
}
