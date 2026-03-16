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
	dsnFile := flag.String("dsn-file", "/run/secrets/db-dsn", "path to file containing the database DSN")
	migrationsPath := flag.String("migrations-path", "./migrations", "path to migrations directory")
	flag.Parse()

	direction := flag.Arg(0)
	if direction == "" {
		direction = "up"
	}

	dsn, err := readSecret(*dsnFile)
	if err != nil {
		slog.Error("reading dsn", "err", err)
		os.Exit(1)
	}

	m, err := migrate.New("file://"+*migrationsPath, dsn)
	if err != nil {
		slog.Error("creating migrator", "err", err)
		os.Exit(1)
	}
	defer m.Close()

	switch direction {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			slog.Error("running up migrations", "err", err)
			os.Exit(1)
		}
		slog.Info("migrations applied")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			slog.Error("running down migrations", "err", err)
			os.Exit(1)
		}
		slog.Info("migrations rolled back")
	default:
		fmt.Fprintf(os.Stderr, "unknown direction %q (use 'up' or 'down')\n", direction)
		os.Exit(1)
	}
}

func readSecret(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading secret file %s: %w", path, err)
	}
	return strings.TrimSpace(string(data)), nil
}
