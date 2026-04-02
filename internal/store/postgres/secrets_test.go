package postgres

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadSecret_Success(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "secret.txt")
	if err := os.WriteFile(path, []byte("my-secret-value\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	val, err := readSecret(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "my-secret-value" {
		t.Errorf("expected my-secret-value, got %q", val)
	}
}

func TestReadSecret_TrimsWhitespace(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "secret.txt")
	if err := os.WriteFile(path, []byte("  spaces  \n\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	val, err := readSecret(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "spaces" {
		t.Errorf("expected trimmed value 'spaces', got %q", val)
	}
}

func TestReadSecret_FileNotFound(t *testing.T) {
	t.Parallel()
	_, err := readSecret("/nonexistent/secret.txt")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
