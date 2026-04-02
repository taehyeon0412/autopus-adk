package stream

import (
	"io"
	"strings"
	"testing"
)

func TestParseLine_ValidEvent(t *testing.T) {
	line := []byte(`{"type":"system.init","mcp_servers":["fs"]}`)
	evt, err := ParseLine(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if evt.Type != "system" {
		t.Errorf("Type = %q, want %q", evt.Type, "system")
	}
	if evt.Subtype != "init" {
		t.Errorf("Subtype = %q, want %q", evt.Subtype, "init")
	}
}

func TestParseLine_NoSubtype(t *testing.T) {
	line := []byte(`{"type":"result","cost_usd":0.05}`)
	evt, err := ParseLine(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if evt.Type != "result" {
		t.Errorf("Type = %q, want %q", evt.Type, "result")
	}
	if evt.Subtype != "" {
		t.Errorf("Subtype = %q, want empty", evt.Subtype)
	}
}

func TestParseLine_MissingType(t *testing.T) {
	line := []byte(`{"foo":"bar"}`)
	_, err := ParseLine(line)
	if err == nil {
		t.Fatal("expected error for missing type field")
	}
}

func TestParseLine_NotJSON(t *testing.T) {
	_, err := ParseLine([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for non-JSON input")
	}
}

func TestParseLine_EmptyLine(t *testing.T) {
	_, err := ParseLine([]byte(""))
	if err == nil {
		t.Fatal("expected error for empty line")
	}
}

func TestParser_Next(t *testing.T) {
	input := strings.Join([]string{
		`{"type":"system.init","tools":["read"]}`,
		``,
		`not json`,
		`{"type":"system.task_started"}`,
		`{"type":"result","cost_usd":0.01}`,
	}, "\n")

	p := NewParser(strings.NewReader(input))

	// First event: system.init
	evt, err := p.Next()
	if err != nil {
		t.Fatalf("event 1: %v", err)
	}
	if evt.Type != "system" || evt.Subtype != "init" {
		t.Errorf("event 1: got %s.%s, want system.init", evt.Type, evt.Subtype)
	}

	// Second event: system.task_started (skips empty and non-JSON lines)
	evt, err = p.Next()
	if err != nil {
		t.Fatalf("event 2: %v", err)
	}
	if evt.Type != "system" || evt.Subtype != "task_started" {
		t.Errorf("event 2: got %s.%s, want system.task_started", evt.Type, evt.Subtype)
	}

	// Third event: result
	evt, err = p.Next()
	if err != nil {
		t.Fatalf("event 3: %v", err)
	}
	if evt.Type != "result" {
		t.Errorf("event 3: got %s, want result", evt.Type)
	}

	// EOF
	_, err = p.Next()
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}
}

func TestParser_RawPreserved(t *testing.T) {
	line := `{"type":"error","message":"boom"}`
	p := NewParser(strings.NewReader(line))
	evt, err := p.Next()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(evt.Raw) != line {
		t.Errorf("Raw = %s, want %s", evt.Raw, line)
	}
}
