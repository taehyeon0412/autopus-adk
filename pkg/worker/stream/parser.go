package stream

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
)

// Parser reads stream-json lines from an io.Reader and emits typed events.
type Parser struct {
	scanner *bufio.Scanner
}

// NewParser creates a parser that reads from r.
func NewParser(r io.Reader) *Parser {
	return &Parser{
		scanner: bufio.NewScanner(r),
	}
}

// Next returns the next event or io.EOF when the stream ends.
// Malformed lines are logged and skipped.
func (p *Parser) Next() (Event, error) {
	for p.scanner.Scan() {
		line := p.scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		evt, err := ParseLine(line)
		if err != nil {
			log.Printf("[stream] skipping malformed line: %v", err)
			continue
		}
		return evt, nil
	}

	if err := p.scanner.Err(); err != nil {
		return Event{}, fmt.Errorf("stream scan: %w", err)
	}
	return Event{}, io.EOF
}

// ParseLine parses a single JSON line into an Event.
// The line must be a JSON object with a "type" field.
func ParseLine(line []byte) (Event, error) {
	// Quick check: must start with '{' to be a JSON object.
	trimmed := strings.TrimSpace(string(line))
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return Event{}, fmt.Errorf("not a JSON object")
	}

	var raw struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(line, &raw); err != nil {
		return Event{}, fmt.Errorf("json unmarshal: %w", err)
	}
	if raw.Type == "" {
		return Event{}, fmt.Errorf("missing type field")
	}

	typ, subtype := splitType(raw.Type)

	return Event{
		Type:    typ,
		Subtype: subtype,
		Raw:     json.RawMessage(append([]byte(nil), line...)),
	}, nil
}

// splitType splits "system.init" into ("system", "init").
// If there is no dot, subtype is empty.
func splitType(full string) (string, string) {
	if before, after, ok := strings.Cut(full, "."); ok {
		return before, after
	}
	return full, ""
}
