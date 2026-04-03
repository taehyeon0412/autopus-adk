package learn

import (
	"encoding/json"
	"fmt"
	"os"
)

// rewriteStore rewrites the store file with the given entries.
func rewriteStore(store *Store, entries []LearningEntry) error {
	f, err := os.Create(store.path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	for _, e := range entries {
		data, err := json.Marshal(e)
		if err != nil {
			return fmt.Errorf("marshal entry: %w", err)
		}
		if _, err := f.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("write entry: %w", err)
		}
	}
	return nil
}
