package demo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pebblr/pebblr/internal/domain"
)

const (
	testKey           = "test-key"
	testTokenPath     = "/demo/token"
	testContentType   = "Content-Type"
	testContentJSON   = "application/json"
	errExpected400Fmt = "expected 400, got %d"
)

// stubUserLister is a test double for UserLister.
type stubUserLister struct {
	users []*domain.User
}

func (s *stubUserLister) List(_ context.Context) ([]*domain.User, error) {
	return s.users, nil
}

func testUsers() *stubUserLister {
	return &stubUserLister{
		users: []*domain.User{
			{ID: "u-rep", Email: "rep@demo.pebblr.com", Name: "Riley Rep", Role: domain.RoleRep},
			{ID: "u-mgr", Email: "mgr@demo.pebblr.com", Name: "Morgan Manager", Role: domain.RoleManager},
			{ID: "u-adm", Email: "adm@demo.pebblr.com", Name: "Alex Admin", Role: domain.RoleAdmin},
		},
	}
}

func TestListAccounts(t *testing.T) {
	t.Parallel()

	a, _ := New([]byte(testKey))
	h := NewHandler(a, testUsers())

	req := httptest.NewRequest(http.MethodGet, "/demo/accounts", http.NoBody)
	rec := httptest.NewRecorder()

	h.ListAccounts(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var accounts []Account
	if err := json.NewDecoder(rec.Body).Decode(&accounts); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if len(accounts) != 3 {
		t.Errorf("expected 3 accounts, got %d", len(accounts))
	}
	if accounts[0].ID != "u-rep" {
		t.Errorf("expected first account ID u-rep, got %q", accounts[0].ID)
	}
	if accounts[0].Role != domain.RoleRep {
		t.Errorf("expected first account role rep, got %q", accounts[0].Role)
	}
}

func TestIssueToken_ValidUser(t *testing.T) {
	t.Parallel()

	a, _ := New([]byte(testKey))
	h := NewHandler(a, testUsers())

	body := `{"user_id":"u-rep"}`
	req := httptest.NewRequest(http.MethodPost, testTokenPath, strings.NewReader(body))
	req.Header.Set(testContentType, testContentJSON)
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
	if resp.Account.ID != "u-rep" {
		t.Errorf("expected account ID u-rep, got %q", resp.Account.ID)
	}
	if resp.Account.Role != domain.RoleRep {
		t.Errorf("expected role rep, got %q", resp.Account.Role)
	}
}

func TestIssueToken_UnknownUser(t *testing.T) {
	t.Parallel()

	a, _ := New([]byte(testKey))
	h := NewHandler(a, testUsers())

	body := `{"user_id":"nonexistent"}`
	req := httptest.NewRequest(http.MethodPost, testTokenPath, strings.NewReader(body))
	req.Header.Set(testContentType, testContentJSON)
	rec := httptest.NewRecorder()

	h.IssueToken(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(errExpected400Fmt, rec.Code)
	}
}

func TestIssueToken_MissingUserID(t *testing.T) {
	t.Parallel()

	a, _ := New([]byte(testKey))
	h := NewHandler(a, testUsers())

	body := `{}`
	req := httptest.NewRequest(http.MethodPost, testTokenPath, strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.IssueToken(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(errExpected400Fmt, rec.Code)
	}
}

func TestIssueToken_InvalidBody(t *testing.T) {
	t.Parallel()

	a, _ := New([]byte(testKey))
	h := NewHandler(a, testUsers())

	req := httptest.NewRequest(http.MethodPost, testTokenPath, strings.NewReader("not json"))
	rec := httptest.NewRecorder()

	h.IssueToken(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(errExpected400Fmt, rec.Code)
	}
}

func TestIssueToken_RoundTrip(t *testing.T) {
	t.Parallel()

	a, _ := New([]byte(testKey))
	h := NewHandler(a, testUsers())

	body := `{"user_id":"u-mgr"}`
	req := httptest.NewRequest(http.MethodPost, testTokenPath, strings.NewReader(body))
	req.Header.Set(testContentType, testContentJSON)
	rec := httptest.NewRecorder()
	h.IssueToken(rec, req)

	var resp tokenResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	claims, err := a.ValidateToken(req.Context(), resp.Token)
	if err != nil {
		t.Fatalf("validating issued token: %v", err)
	}
	if claims.Sub != "u-mgr" {
		t.Errorf("expected sub u-mgr, got %q", claims.Sub)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != domain.RoleManager {
		t.Errorf("expected [manager] role, got %v", claims.Roles)
	}
}
