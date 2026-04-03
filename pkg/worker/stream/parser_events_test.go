package stream

import (
	"testing"
)

// TestParseLine_AllEventTypes verifies parsing of every defined event type.
func TestParseLine_AllEventTypes(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantTyp string
		wantSub string
	}{
		{
			name:    "system.init",
			input:   `{"type":"system.init","mcp_servers":["fs"]}`,
			wantTyp: "system",
			wantSub: "init",
		},
		{
			name:    "system.task_started",
			input:   `{"type":"system.task_started"}`,
			wantTyp: "system",
			wantSub: "task_started",
		},
		{
			name:    "system.task_progress",
			input:   `{"type":"system.task_progress","progress":50}`,
			wantTyp: "system",
			wantSub: "task_progress",
		},
		{
			name:    "system.task_notification",
			input:   `{"type":"system.task_notification","message":"hello"}`,
			wantTyp: "system",
			wantSub: "task_notification",
		},
		{
			name:    "result",
			input:   `{"type":"result","cost_usd":0.05,"duration_ms":1200}`,
			wantTyp: "result",
			wantSub: "",
		},
		{
			name:    "error",
			input:   `{"type":"error","message":"something failed"}`,
			wantTyp: "error",
			wantSub: "",
		},
		{
			name:    "tool_call",
			input:   `{"type":"tool_call","name":"read_file"}`,
			wantTyp: "tool_call",
			wantSub: "",
		},
		{
			name:    "tool_use",
			input:   `{"type":"tool_use","name":"write_file"}`,
			wantTyp: "tool_use",
			wantSub: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt, err := ParseLine([]byte(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if evt.Type != tt.wantTyp {
				t.Errorf("Type = %q, want %q", evt.Type, tt.wantTyp)
			}
			if evt.Subtype != tt.wantSub {
				t.Errorf("Subtype = %q, want %q", evt.Subtype, tt.wantSub)
			}
		})
	}
}

// TestParseLine_MalformedInputs verifies error handling for various bad inputs.
func TestParseLine_MalformedInputs(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "empty string", input: ""},
		{name: "whitespace only", input: "   "},
		{name: "plain text", input: "hello world"},
		{name: "array not object", input: `[1,2,3]`},
		{name: "missing type field", input: `{"foo":"bar"}`},
		{name: "empty type value", input: `{"type":""}`},
		{name: "truncated json", input: `{"type":"resu`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseLine([]byte(tt.input))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

// TestSplitType verifies the type.subtype splitting logic.
func TestSplitType(t *testing.T) {
	tests := []struct {
		input   string
		wantTyp string
		wantSub string
	}{
		{"system.init", "system", "init"},
		{"system.task_started", "system", "task_started"},
		{"result", "result", ""},
		{"a.b.c", "a", "b.c"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			typ, sub := splitType(tt.input)
			if typ != tt.wantTyp || sub != tt.wantSub {
				t.Errorf("splitType(%q) = (%q, %q), want (%q, %q)",
					tt.input, typ, sub, tt.wantTyp, tt.wantSub)
			}
		})
	}
}
