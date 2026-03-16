// Command migrate runs database migrations using golang-migrate.
// Usage: migrate -dsn-file /run/secrets/db-dsn -migrations-path ./migrations [up|down]
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if err := run(); err != nil {
		slog.Error("migration failed", "err", err)
		os.Exit(1)
	}
}

func run() error {
	dsnFile := flag.String("dsn-file", "/run/secrets/db-dsn", "path to file containing the database DSN")
	migrationsPath := flag.String("migrations-path", "./migrations", "path to migrations directory")
	flag.Parse()

	direction := flag.Arg(0)
	if direction == "" {
		direction = "up"
	}

	dsn, err := readSecret(*dsnFile)
	if err != nil {
		return fmt.Errorf("reading dsn: %w", err)
	}

	m, err := migrate.New("file://"+*migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	switch direction {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("running up migrations: %w", err)
		}
		slog.Info("migrations applied")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("running down migrations: %w", err)
		}
		slog.Info("migrations rolled back")
	default:
		return fmt.Errorf("unknown direction %q (use 'up' or 'down')", direction)
	}

	return nil
}

func readSecret(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading secret file %s: %w", path, err)
	}
	return strings.TrimSpace(string(data)), nil
}
