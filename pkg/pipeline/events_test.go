package pipeline

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEventType_Constants verifies all 7 event types are defined with correct values.
func TestEventType_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		got      EventType
		expected string
	}{
		{"phase_start", EventPhaseStart, "phase_start"},
		{"phase_end", EventPhaseEnd, "phase_end"},
		{"agent_spawn", EventAgentSpawn, "agent_spawn"},
		{"agent_done", EventAgentDone, "agent_done"},
		{"checkpoint", EventCheckpoint, "checkpoint"},
		{"error", EventError, "error"},
		{"blocker", EventBlocker, "blocker"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, EventType(tt.expected), tt.got)
		})
	}

	// Verify exactly 7 types are defined.
	allTypes := []EventType{
		EventPhaseStart, EventPhaseEnd,
		EventAgentSpawn, EventAgentDone,
		EventCheckpoint, EventError, EventBlocker,
	}
	assert.Len(t, allTypes, 7, "expected exactly 7 event types")
}

// TestNewEvent verifies the constructor sets timestamp and fields correctly.
func TestNewEvent(t *testing.T) {
	t.Parallel()

	before := time.Now()
	event := NewEvent(EventPhaseStart, "starting phase 1")
	after := time.Now()

	// NewEvent must set Type to the given event type.
	assert.Equal(t, EventPhaseStart, event.Type,
		"NewEvent should set Type field")

	// NewEvent must set Message to the given message.
	assert.Equal(t, "starting phase 1", event.Message,
		"NewEvent should set Message field")

	// NewEvent must set Timestamp to approximately now.
	assert.False(t, event.Timestamp.IsZero(),
		"NewEvent should set a non-zero Timestamp")
	assert.True(t, !event.Timestamp.Before(before) && !event.Timestamp.After(after),
		"Timestamp should be between before and after call")
}

// TestEvent_MarshalJSONL verifies Event serializes to single-line JSON.
func TestEvent_MarshalJSONL(t *testing.T) {
	t.Parallel()

	event := Event{
		Type:      EventAgentSpawn,
		Timestamp: time.Date(2026, 3, 26, 12, 0, 0, 0, time.UTC),
		Phase:     "phase2",
		Agent:     "executor-1",
		Message:   "spawning executor",
	}

	data, err := event.MarshalJSONL()
	require.NoError(t, err, "MarshalJSONL should not return error")
	require.NotNil(t, data, "MarshalJSONL should return non-nil data")

	// Must be a single line (no embedded newlines).
	line := string(data)
	assert.False(t, strings.Contains(line, "\n"),
		"JSONL output must be a single line, got: %s", line)

	// Must be valid JSON.
	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err, "JSONL output must be valid JSON")

	// Must contain required fields.
	assert.Equal(t, "agent_spawn", parsed["type"])
	assert.Equal(t, "phase2", parsed["phase"])
	assert.Equal(t, "executor-1", parsed["agent"])
	assert.Equal(t, "spawning executor", parsed["message"])
}

// TestEvent_MarshalJSONL_OmitsEmptyFields verifies omitempty behavior.
func TestEvent_MarshalJSONL_OmitsEmptyFields(t *testing.T) {
	t.Parallel()

	event := Event{
		Type:      EventError,
		Timestamp: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Message:   "something failed",
		// Phase and Agent intentionally empty.
	}

	data, err := event.MarshalJSONL()
	require.NoError(t, err)

	var parsed map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &parsed))

	assert.Equal(t, "error", parsed["type"])
	assert.Equal(t, "something failed", parsed["message"])
	// Phase and Agent should be omitted (omitempty).
	_, hasPhase := parsed["phase"]
	_, hasAgent := parsed["agent"]
	assert.False(t, hasPhase, "empty Phase should be omitted")
	assert.False(t, hasAgent, "empty Agent should be omitted")
}

// TestNewEvent_DifferentTypes verifies NewEvent works for all event types.
func TestNewEvent_DifferentTypes(t *testing.T) {
	t.Parallel()

	types := []EventType{
		EventPhaseStart, EventPhaseEnd, EventAgentSpawn,
		EventAgentDone, EventCheckpoint, EventError, EventBlocker,
	}

	for _, et := range types {
		t.Run(string(et), func(t *testing.T) {
			t.Parallel()
			event := NewEvent(et, "test msg")
			assert.Equal(t, et, event.Type)
			assert.Equal(t, "test msg", event.Message)
			assert.False(t, event.Timestamp.IsZero())
		})
	}
}
