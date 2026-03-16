package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps a pgx connection pool and implements store.Store.
type DB struct {
	pool *pgxpool.Pool
}

// New creates a new DB using the given connection pool.
func New(pool *pgxpool.Pool) *DB {
	return &DB{pool: pool}
}

// Connect opens a connection pool using the DSN read from the given file path.
// Secrets are read from mounted files, never from environment variables.
func Connect(ctx context.Context, dsnFile string) (*pgxpool.Pool, error) {
	dsn, err := readSecret(dsnFile)
	if err != nil {
		return nil, fmt.Errorf("reading db dsn from %s: %w", dsnFile, err)
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parsing db config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("connecting to db: %w", err)
	}

	return pool, nil
}
