package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const workerCredentialService = "autopus-worker"

var newCredentialStoreFunc = func() (CredentialStore, string) {
	return NewCredentialStore()
}

func loadCredentialBytes() ([]byte, error) {
	store, _ := newCredentialStoreFunc()
	if store != nil {
		raw, err := store.Load(workerCredentialService)
		if err == nil && raw != "" {
			return []byte(raw), nil
		}
	}

	// Fallback to the encrypted file store so previously migrated credentials
	// remain readable even when the preferred backend changes (for example,
	// keychain becomes available after an earlier file-backed save).
	if raw, err := newEncryptedFileStore(defaultCredentialDir()).Load(workerCredentialService); err == nil && raw != "" {
		return []byte(raw), nil
	}

	return os.ReadFile(DefaultCredentialsPath())
}

func saveCredentialBytes(data []byte) error {
	store, _ := newCredentialStoreFunc()
	if store != nil {
		if err := store.Save(workerCredentialService, string(data)); err == nil {
			// Best effort cleanup for legacy plaintext credentials after secure save.
			_ = os.Remove(DefaultCredentialsPath())
			return nil
		}
	}

	path := DefaultCredentialsPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write credentials: %w", err)
	}
	return nil
}

func loadCredentialsMap() (map[string]any, error) {
	data, err := loadCredentialBytes()
	if err != nil {
		return nil, err
	}

	var creds map[string]any
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	return creds, nil
}
