package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pebblr/pebblr/internal/api"
	"github.com/pebblr/pebblr/internal/config"
	"github.com/pebblr/pebblr/internal/domain"
)

func testTenantConfig() *config.TenantConfig {
	return &config.TenantConfig{
		Tenant: config.TenantInfo{Name: "Test Tenant", Locale: "en"},
		Accounts: config.AccountsConfig{
			Types: []config.AccountTypeConfig{
				{Key: "doctor", Label: "Doctor", Fields: []config.FieldConfig{
					{Key: "name", Type: "text", Required: true},
				}},
			},
		},
		Activities: config.ActivitiesConfig{
			Statuses: []config.StatusDef{
				{Key: "planned", Label: "Planned", Initial: true},
				{Key: "done", Label: "Done"},
			},
			StatusTransitions: map[string][]string{"planned": {"done"}},
			Durations:         []config.OptionDef{{Key: "full_day", Label: "Full Day"}},
			Types: []config.ActivityTypeConfig{
				{Key: "visit", Label: "Visit", Category: "field"},
			},
		},
		Options: map[string][]config.OptionDef{
			"specialties": {{Key: "cardio", Label: "Cardiology"}},
		},
		Rules: config.RulesConfig{
			MaxActivitiesPerDay: 10,
		},
	}
}

func newTestConfigHandler(user *domain.User) http.Handler {
	h := api.NewConfigHandler(testTenantConfig())
	mux := http.NewServeMux()
	mux.HandleFunc("/config", h.Get)
	if user != nil {
		return injectUser(user, mux)
	}
	return mux
}

func TestConfigGet_ReturnsOK(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/config", http.NoBody)
	w := httptest.NewRecorder()
	newTestConfigHandler(testRepUser()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var cfg config.TenantConfig
	if err := json.NewDecoder(w.Body).Decode(&cfg); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if cfg.Tenant.Name != "Test Tenant" {
		t.Errorf("expected tenant name 'Test Tenant', got %q", cfg.Tenant.Name)
	}
	if len(cfg.Accounts.Types) != 1 {
		t.Errorf("expected 1 account type, got %d", len(cfg.Accounts.Types))
	}
	if len(cfg.Activities.Types) != 1 {
		t.Errorf("expected 1 activity type, got %d", len(cfg.Activities.Types))
	}
}

func TestConfigGet_NoUser_Returns401(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/config", http.NoBody)
	w := httptest.NewRecorder()
	newTestConfigHandler(nil).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}
