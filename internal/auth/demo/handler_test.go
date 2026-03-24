package demo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pebblr/pebblr/internal/domain"
)

func TestListPersonas(t *testing.T) {
	t.Parallel()

	a, _ := New([]byte("test-key"))
	h := NewHandler(a, DefaultPersonas())

	req := httptest.NewRequest(http.MethodGet, "/demo/personas", http.NoBody)
	rec := httptest.NewRecorder()

	h.ListPersonas(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var personas []Persona
	if err := json.NewDecoder(rec.Body).Decode(&personas); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(personas) != 3 {
		t.Errorf("expected 3 personas, got %d", len(personas))
	}
}

func TestIssueToken_ValidPersona(t *testing.T) {
	t.Parallel()

	a, _ := New([]byte("test-key"))
	h := NewHandler(a, DefaultPersonas())

	body := `{"persona_id":"demo-rep"}`
	req := httptest.NewRequest(http.MethodPost, "/demo/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.IssueToken(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp tokenResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
	if resp.Persona.ID != "demo-rep" {
		t.Errorf("expected persona demo-rep, got %q", resp.Persona.ID)
	}
	if resp.Persona.Role != domain.RoleRep {
		t.Errorf("expected role rep, got %q", resp.Persona.Role)
	}
}

func TestIssueToken_UnknownPersona(t *testing.T) {
	t.Parallel()

	a, _ := New([]byte("test-key"))
	h := NewHandler(a, DefaultPersonas())

	body := `{"persona_id":"nonexistent"}`
	req := httptest.NewRequest(http.MethodPost, "/demo/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.IssueToken(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestIssueToken_InvalidBody(t *testing.T) {
	t.Parallel()

	a, _ := New([]byte("test-key"))
	h := NewHandler(a, DefaultPersonas())

	req := httptest.NewRequest(http.MethodPost, "/demo/token", strings.NewReader("not json"))
	rec := httptest.NewRecorder()

	h.IssueToken(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestIssueToken_RoundTrip(t *testing.T) {
	t.Parallel()

	a, _ := New([]byte("test-key"))
	h := NewHandler(a, DefaultPersonas())

	// Issue a token via the handler.
	body := `{"persona_id":"demo-manager"}`
	req := httptest.NewRequest(http.MethodPost, "/demo/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.IssueToken(rec, req)

	var resp tokenResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	// Validate the token using the authenticator directly.
	claims, err := a.ValidateToken(req.Context(), resp.Token)
	if err != nil {
		t.Fatalf("validating issued token: %v", err)
	}
	if claims.Sub != "demo-manager" {
		t.Errorf("expected sub demo-manager, got %q", claims.Sub)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != domain.RoleManager {
		t.Errorf("expected [manager] role, got %v", claims.Roles)
	}
}
