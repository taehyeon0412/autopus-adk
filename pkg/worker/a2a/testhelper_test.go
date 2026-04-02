package a2a

import (
	"encoding/json"
	"fmt"
)

// mustMarshal is a test-only helper that panics on marshal failure.
func mustMarshal(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("mustMarshal: %v", err))
	}
	return data
}
