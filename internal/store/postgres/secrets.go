package postgres

import (
	"fmt"
	"os"
	"strings"
)

// readSecret reads a secret value from a mounted file.
// Trailing whitespace (newlines) is trimmed.
// Secrets are never read from environment variables.
func readSecret(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading secret file %s: %w", path, err)
	}
	return strings.TrimSpace(string(data)), nil
}
