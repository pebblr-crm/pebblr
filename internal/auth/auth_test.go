package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pebblr/pebblr/internal/auth"
	"github.com/pebblr/pebblr/internal/domain"
)

// stubAuthenticator is a test double that always succeeds.
type stubAuthenticator struct {
	claims *auth.UserClaims
}

func (s *stubAuthenticator) ValidateToken(_ context.Context, _ string) (*auth.UserClaims, error) {
	return s.claims, nil
}

func TestMiddlewareAttachesClaims(t *testing.T) {
	t.Parallel()
	want := &auth.UserClaims{
		Sub:   "user-oid-123",
		Email: "rep@example.com",
		Roles: []domain.Role{domain.RoleRep},
	}

	mw := auth.Middleware(&stubAuthenticator{claims: want})

	var gotClaims *auth.UserClaims
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotClaims = auth.ClaimsFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("Authorization", "Bearer some-valid-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotClaims == nil {
		t.Fatal("expected claims in context, got nil")
	}
	if gotClaims.Sub != want.Sub {
		t.Errorf("expected sub %q, got %q", want.Sub, gotClaims.Sub)
	}
}

func TestMiddlewareRejectsNoBearerToken(t *testing.T) {
	t.Parallel()
	mw := auth.Middleware(&stubAuthenticator{})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestMiddlewareReturnsJSONContentType(t *testing.T) {
	t.Parallel()
	mw := auth.Middleware(&stubAuthenticator{})

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// No Authorization header -- should get JSON error response.
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
