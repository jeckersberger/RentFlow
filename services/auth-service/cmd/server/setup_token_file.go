package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const defaultSetupTokenFile = "/tmp/cratedesk-setup-token"

func persistSetupToken(token string) (string, error) {
	path := strings.TrimSpace(os.Getenv("SETUP_TOKEN_FILE"))
	if path == "" {
		path = defaultSetupTokenFile
	}

	if token == "" {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return path, fmt.Errorf("remove stale setup token file: %w", err)
		}
		return path, nil
	}

	if err := os.WriteFile(path, []byte(token+"\n"), 0o600); err != nil {
		return path, fmt.Errorf("write setup token file: %w", err)
	}
	if err := os.Chmod(path, 0o600); err != nil {
		return path, fmt.Errorf("secure setup token file permissions: %w", err)
	}
	return path, nil
}
