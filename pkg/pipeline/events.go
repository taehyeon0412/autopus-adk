// Package pipeline provides pipeline state management types and persistence.
package pipeline

import (
	"encoding/json"
	"time"
)

// EventType identifies the kind of pipeline event.
type EventType string

// @AX:NOTE [AUTO] @AX:REASON: magic constants — event type strings define the JSONL wire format; changing values breaks log consumers
const (
	// EventPhaseStart is emitted when a pipeline phase begins.
	EventPhaseStart EventType = "phase_start"
	// EventPhaseEnd is emitted when a pipeline phase completes.
	EventPhaseEnd EventType = "phase_end"
	// EventAgentSpawn is emitted when an agent is spawned.
	EventAgentSpawn EventType = "agent_spawn"
	// EventAgentDone is emitted when an agent finishes.
	EventAgentDone EventType = "agent_done"
	// EventCheckpoint is emitted when a checkpoint is saved.
	EventCheckpoint EventType = "checkpoint"
	// EventError is emitted when an error occurs.
	EventError EventType = "error"
	// EventBlocker is emitted when a blocker is detected.
	EventBlocker EventType = "blocker"
)

// Event represents a single pipeline lifecycle event.
type Event struct {
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Phase     string    `json:"phase,omitempty"`
	Agent     string    `json:"agent,omitempty"`
	Message   string    `json:"message,omitempty"`
}

// NewEvent creates an Event with the given type and message, setting Timestamp to now.
func NewEvent(eventType EventType, message string) Event {
	return Event{
		Type:      eventType,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// MarshalJSONL serializes the Event as a single-line JSON (JSONL format).
func (e Event) MarshalJSONL() ([]byte, error) {
	return json.Marshal(e)
}
