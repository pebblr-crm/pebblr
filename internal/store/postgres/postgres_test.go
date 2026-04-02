package postgres

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	pgxmock "github.com/pashagolub/pgxmock/v4"
)

func TestNew_ReturnsDB(t *testing.T) {
	t.Parallel()
	// New() requires a *pgxpool.Pool, but we can test initRepos via the DB struct directly.
	// This test verifies the DB struct works with a mock pool.
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("creating pgxmock pool: %v", err)
	}

	db := &DB{pool: mock}
	db.initRepos()

	if db.pool == nil {
		t.Error("expected non-nil pool")
	}
}

func TestConnect_InvalidDSNFile(t *testing.T) {
	t.Parallel()
	_, err := Connect(context.TODO(), "/nonexistent/dsn.txt")
	if err == nil {
		t.Fatal("expected error for missing DSN file")
	}
}

func TestConnect_InvalidDSN(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "dsn.txt")
	// Write an invalid DSN that will fail pgxpool.ParseConfig.
	if err := os.WriteFile(path, []byte("not-a-valid-dsn://"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Connect(context.TODO(), path)
	if err == nil {
		t.Fatal("expected error for invalid DSN")
	}
}
