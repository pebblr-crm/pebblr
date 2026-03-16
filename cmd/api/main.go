package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/pebblr/pebblr/internal/api"
)

const (
	defaultDSNFile        = "/run/secrets/db-dsn"
	defaultMigrationsPath = "./migrations"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := runMigrations(logger); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	router := api.NewRouter(api.RouterConfig{
		Logger: logger,
	})

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		logger.Info("starting server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	fmt.Println("server stopped")
	return nil
}

func runMigrations(logger *slog.Logger) error {
	dsn, err := readSecretFile(defaultDSNFile)
	if err != nil {
		return fmt.Errorf("reading dsn: %w", err)
	}

	m, err := migrate.New("file://"+defaultMigrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	before, _, _ := m.Version()
	logger.Info("running migrations", "current_version", before)

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("applying migrations: %w", err)
	}

	after, _, _ := m.Version()
	logger.Info("migrations complete", "version", after)

	return nil
}

func readSecretFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading secret file %s: %w", path, err)
	}
	return strings.TrimSpace(string(data)), nil
}
