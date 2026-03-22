package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pebblr/pebblr/internal/api"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/service"
)

// --- stub DashboardService ---

type stubDashboardSvc struct{}

func (s *stubDashboardSvc) Stats(_ context.Context) (*service.DashboardStats, error) {
	return &service.DashboardStats{
		TotalLeads:      42,
		ConversionRate:  0.67,
		UnassignedCount: 5,
	}, nil
}

// --- helpers ---

func newTestDashboardHandler(user *domain.User) http.Handler {
	h := api.NewDashboardHandler(&stubDashboardSvc{})
	mux := http.NewServeMux()
	mux.HandleFunc("/stats", h.Stats)
	if user != nil {
		return injectUser(user, mux)
	}
	return mux
}

// --- tests ---

func TestDashboardStats_ReturnsOK(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/stats", http.NoBody)
	w := httptest.NewRecorder()
	newTestDashboardHandler(testAdminUser()).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var stats service.DashboardStats
	if err := json.NewDecoder(w.Body).Decode(&stats); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if stats.TotalLeads != 42 {
		t.Errorf("expected totalLeads=42, got %d", stats.TotalLeads)
	}
}

func TestDashboardStats_NoUser_Returns401(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/stats", http.NoBody)
	w := httptest.NewRecorder()
	newTestDashboardHandler(nil).ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}
