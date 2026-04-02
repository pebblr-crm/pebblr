package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pebblr/pebblr/internal/auth"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
)

const (
	testTeamID       = "team-1"
	fmtUserFromCtxErr = "UserFromContext error: %v"
	fmtExpected200   = "expected 200, got %d"
)

func TestClaimsBridge_SetsUser(t *testing.T) {
	t.Parallel()
	claims := &auth.UserClaims{
		Sub:     "user-123",
		Email:   "test@example.com",
		Name:    "Test User",
		Roles:   []domain.Role{domain.RoleManager},
		TeamIDs: []string{testTeamID},
	}

	var got *domain.User
	handler := auth.ClaimsBridge(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		got, err = rbac.UserFromContext(r.Context())
		if err != nil {
			t.Fatalf(fmtUserFromCtxErr, err)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", http.NoBody)
	req = req.WithContext(auth.WithClaims(req.Context(), claims))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, rec.Code)
	}
	if got.ID != "user-123" {
		t.Errorf("expected user ID user-123, got %q", got.ID)
	}
	if got.Role != domain.RoleManager {
		t.Errorf("expected role manager, got %q", got.Role)
	}
	if len(got.TeamIDs) != 1 || got.TeamIDs[0] != testTeamID {
		t.Errorf("expected teamIDs [team-1], got %v", got.TeamIDs)
	}
}

func TestClaimsBridge_MultiRole_PicksHighest(t *testing.T) {
	t.Parallel()
	claims := &auth.UserClaims{
		Sub:     "user-multi",
		Email:   "multi@example.com",
		Name:    "Multi-Role User",
		Roles:   []domain.Role{domain.RoleRep, domain.RoleAdmin, domain.RoleManager},
		TeamIDs: []string{testTeamID},
	}

	var got *domain.User
	handler := auth.ClaimsBridge(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		got, err = rbac.UserFromContext(r.Context())
		if err != nil {
			t.Fatalf(fmtUserFromCtxErr, err)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", http.NoBody)
	req = req.WithContext(auth.WithClaims(req.Context(), claims))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, rec.Code)
	}
	if got.Role != domain.RoleAdmin {
		t.Errorf("expected highest role admin, got %q", got.Role)
	}
}

func TestClaimsBridge_EmptyRoles_DefaultsToRep(t *testing.T) {
	t.Parallel()
	claims := &auth.UserClaims{
		Sub:     "user-norole",
		Email:   "norole@example.com",
		Name:    "No Role User",
		Roles:   []domain.Role{},
		TeamIDs: []string{},
	}

	var got *domain.User
	handler := auth.ClaimsBridge(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		got, err = rbac.UserFromContext(r.Context())
		if err != nil {
			t.Fatalf(fmtUserFromCtxErr, err)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", http.NoBody)
	req = req.WithContext(auth.WithClaims(req.Context(), claims))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(fmtExpected200, rec.Code)
	}
	if got.Role != domain.RoleRep {
		t.Errorf("expected default role rep, got %q", got.Role)
	}
}

func TestClaimsBridge_NoClaims_Returns401(t *testing.T) {
	t.Parallel()
	handler := auth.ClaimsBridge(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called without claims")
	}))

	req := httptest.NewRequest("GET", "/", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
