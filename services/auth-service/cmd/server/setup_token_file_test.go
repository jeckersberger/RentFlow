package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPersistSetupTokenWritesOwnerOnlyFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "setup-token")
	t.Setenv("SETUP_TOKEN_FILE", path)

	gotPath, err := persistSetupToken("secret-token")
	if err != nil {
		t.Fatalf("persistSetupToken: %v", err)
	}
	if gotPath != path {
		t.Fatalf("path = %q, want %q", gotPath, path)
	}

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(contents) != "secret-token\n" {
		t.Fatalf("contents = %q", contents)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode = %o, want 600", info.Mode().Perm())
	}
}

func TestPersistSetupTokenRemovesStaleFileWhenSetupComplete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "setup-token")
	t.Setenv("SETUP_TOKEN_FILE", path)
	if err := os.WriteFile(path, []byte("stale"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := persistSetupToken(""); err != nil {
		t.Fatalf("persistSetupToken: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("stale setup token file still exists: %v", err)
	}
}
