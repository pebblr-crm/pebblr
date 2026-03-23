package main

import (
	"context"
	"flag"
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
	"github.com/pebblr/pebblr/internal/auth"
	"github.com/pebblr/pebblr/internal/config"
	"github.com/pebblr/pebblr/internal/geo"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store/postgres"
)

const (
	defaultDSNFile           = "/run/secrets/db-dsn"
	defaultGeoAPIKeyFile     = "/run/secrets/google-geocoding-api-key"
	defaultMigrationsPath    = "./migrations"
	defaultConfigPath        = "./config/tenant.json"
)

func runServe(args []string) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	configPath := fs.String("config", defaultConfigPath, "path to tenant config JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if err := serve(*configPath); err != nil {
		slog.Error("server error", "err", err)
		return 1
	}
	return 0
}

func serve(configPath string) error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Validate config using the same pipeline as `pebblr config validate`.
	tenantCfg, validationErrors, err := config.LoadAndValidate(configPath)
	if err != nil {
		return fmt.Errorf("config validation: %w", err)
	}
	if len(validationErrors) > 0 {
		for _, e := range validationErrors {
			logger.Error("config validation error", "error", e)
		}
		return fmt.Errorf("tenant config has %d validation error(s)", len(validationErrors))
	}
	logger.Info("tenant config validated", "path", configPath)

	if err := runMigrations(logger); err != nil {
		return fmt.Errorf("migrations: %w", err)
	}

	ctx := context.Background()
	pool, err := postgres.Connect(ctx, defaultDSNFile)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()

	db := postgres.New(pool)
	enforcer := rbac.NewEnforcer()

	teamSvc := service.NewTeamService(db.Teams())
	userSvc := service.NewUserService(db.Users())
	var targetOpts []service.TargetServiceOption
	if apiKey, err := readOptionalSecret(defaultGeoAPIKeyFile); err == nil && apiKey != "" {
		logger.Info("geocoding enabled (Google Maps API key loaded)")
		targetOpts = append(targetOpts, service.WithGeocoder(geo.NewGoogleGeocoder(apiKey)))
	} else {
		logger.Info("geocoding disabled (no API key file)")
	}
	targetSvc := service.NewTargetService(db.Targets(), enforcer, tenantCfg, targetOpts...)
	activitySvc := service.NewActivityService(db.Activities(), db.Users(), db.Audit(), enforcer, tenantCfg)
	dashboardSvc := service.NewDashboardService(db.Dashboard(), enforcer, tenantCfg)

	targetHandler := api.NewTargetHandler(targetSvc)
	activityHandler := api.NewActivityHandler(activitySvc)
	dashboardHandler := api.NewDashboardHandler(dashboardSvc)
	teamHandler := api.NewTeamHandler(teamSvc)
	userHandler := api.NewUserHandler(userSvc)
	configHandler := api.NewConfigHandler(tenantCfg)

	webDistPath := os.Getenv("WEB_DIST_PATH")

	secretPath := os.Getenv("SECRET_MOUNT_PATH")
	if secretPath == "" {
		secretPath = "/run/secrets"
	}
	jwtSecret, err := readSecretFile(secretPath + "/jwt-secret")
	if err != nil {
		return fmt.Errorf("reading jwt secret: %w", err)
	}
	authenticator := auth.NewStaticAuthenticator(jwtSecret)
	logger.Info("using static token authenticator")

	router := api.NewRouter(api.RouterConfig{
		Logger:           logger,
		Authenticator:    authenticator,
		TargetHandler:    targetHandler,
		ActivityHandler:  activityHandler,
		DashboardHandler: dashboardHandler,
		TeamHandler:      teamHandler,
		UserHandler:      userHandler,
		ConfigHandler:    configHandler,
		WebDistPath:      webDistPath,
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

// readOptionalSecret reads a secret file, returning "" if the file doesn't exist.
func readOptionalSecret(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
