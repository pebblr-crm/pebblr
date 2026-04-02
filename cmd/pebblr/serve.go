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
	"github.com/pebblr/pebblr/internal/auth/azuread"
	"github.com/pebblr/pebblr/internal/auth/demo"
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
	authProvider := fs.String("auth-provider", "static", "authentication provider: static, azuread, demo")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if err := serve(*configPath, *authProvider); err != nil {
		slog.Error("server error", "err", err)
		return 1
	}
	return 0
}

func serve(configPath, authProvider string) error {
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
	targetOpts = append(targetOpts, service.WithUsers(db.Users()), service.WithAudit(db.Audit()))
	targetSvc := service.NewTargetService(db.Targets(), enforcer, tenantCfg, targetOpts...)
	activitySvc := service.NewActivityService(db.Activities(), db.Targets(), db.Users(), db.Audit(), enforcer, tenantCfg, service.WithDashboard(db.Dashboard()))
	dashboardSvc := service.NewDashboardService(db.Dashboard(), enforcer, tenantCfg)
	collectionSvc := service.NewCollectionService(db.Collections())
	territorySvc := service.NewTerritoryService(db.Territories())
	auditSvc := service.NewAuditService(db.Audit())

	targetHandler := api.NewTargetHandler(targetSvc)
	activityHandler := api.NewActivityHandler(activitySvc)
	dashboardHandler := api.NewDashboardHandler(dashboardSvc)
	teamHandler := api.NewTeamHandler(teamSvc)
	userHandler := api.NewUserHandler(userSvc)
	collectionHandler := api.NewCollectionHandler(collectionSvc)
	territoryHandler := api.NewTerritoryHandler(territorySvc)
	auditHandler := api.NewAuditHandler(auditSvc)
	configHandler := api.NewConfigHandler(tenantCfg)

	webDistPath := os.Getenv("WEB_DIST_PATH")

	secretPath := os.Getenv("SECRET_MOUNT_PATH")
	if secretPath == "" {
		secretPath = "/run/secrets"
	}

	authenticator, demoHandler, err := buildAuthenticator(ctx, logger, authProvider, secretPath, db.Users())
	if err != nil {
		return fmt.Errorf("setting up auth provider: %w", err)
	}

	router := api.NewRouter(api.RouterConfig{
		Logger:           logger,
		Authenticator:    authenticator,
		TargetHandler:    targetHandler,
		ActivityHandler:  activityHandler,
		DashboardHandler: dashboardHandler,
		TeamHandler:      teamHandler,
		UserHandler:      userHandler,
		ConfigHandler:      configHandler,
		CollectionHandler:  collectionHandler,
		TerritoryHandler:   territoryHandler,
		AuditHandler:       auditHandler,
		DemoHandler:        demoHandler,
		WebDistPath:        webDistPath,
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

	listenErr := make(chan error, 1)
	go func() {
		logger.Info("starting server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			listenErr <- err
			cancel()
		}
	}()

	<-ctx.Done()

	// If ListenAndServe failed (e.g., port already in use), return that error
	// instead of attempting a graceful shutdown.
	select {
	case err := <-listenErr:
		return fmt.Errorf("server listen: %w", err)
	default:
	}

	logger.Info("shutting down server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	logger.Info("server stopped")
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

// buildAuthenticator creates the appropriate Authenticator based on the provider name.
// Returns the authenticator and an optional demo handler (non-nil only for "demo" provider).
func buildAuthenticator(ctx context.Context, logger *slog.Logger, provider, secretPath string, users demo.UserLister) (auth.Authenticator, *demo.Handler, error) {
	switch provider {
	case "static":
		if env := os.Getenv("PEBBLR_ENV"); env == "production" {
			return nil, nil, fmt.Errorf("static auth provider is not allowed in production (PEBBLR_ENV=%s)", env)
		}
		jwtSecret, err := readSecretFile(secretPath + "/jwt-secret")
		if err != nil {
			return nil, nil, fmt.Errorf("reading jwt secret: %w", err)
		}
		logger.Info("using static token authenticator (dev/test only)")
		return auth.NewStaticAuthenticator(jwtSecret), nil, nil

	case "azuread":
		tenantID, err := readSecretFile(secretPath + "/azuread-tenant-id")
		if err != nil {
			return nil, nil, fmt.Errorf("reading Azure AD tenant ID: %w", err)
		}
		clientID, err := readSecretFile(secretPath + "/azuread-client-id")
		if err != nil {
			return nil, nil, fmt.Errorf("reading Azure AD client ID: %w", err)
		}
		issuer, _ := readOptionalSecret(secretPath + "/azuread-issuer")

		a, err := azuread.New(ctx, azuread.Config{
			TenantID:  tenantID,
			ClientID:  clientID,
			IssuerURL: issuer,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("creating Azure AD authenticator: %w", err)
		}
		logger.Info("using Azure AD OIDC authenticator", "tenant_id", tenantID)
		return a, nil, nil

	case "demo":
		if env := os.Getenv("PEBBLR_ENV"); env == "production" {
			return nil, nil, fmt.Errorf("demo auth provider is not allowed in production (PEBBLR_ENV=%s)", env)
		}
		signingKey, _ := readOptionalSecret(secretPath + "/demo-signing-key")
		a, err := demo.New([]byte(signingKey))
		if err != nil {
			return nil, nil, fmt.Errorf("creating demo authenticator: %w", err)
		}
		h := demo.NewHandler(a, users)
		logger.Info("using demo authenticator — NOT FOR PRODUCTION")
		return a, h, nil

	default:
		return nil, nil, fmt.Errorf("unknown auth provider %q (expected static, azuread, or demo)", provider)
	}
}
