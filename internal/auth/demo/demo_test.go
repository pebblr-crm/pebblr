package demo

import (
	"context"
	"testing"
	"time"

	"github.com/pebblr/pebblr/internal/domain"
)

const (
	testUserID    = "user-1"
	testUserEmail = "user@demo.pebblr.com"
)

func TestIssueAndValidate(t *testing.T) {
	t.Parallel()

	a, err := New([]byte("test-secret-key-for-demo-signing"))
	if err != nil {
		t.Fatalf("creating authenticator: %v", err)
	}

	token, err := a.IssueToken(Persona{
		ID:    "demo-rep-1",
		Email: "rep@demo.pebblr.com",
		Name:  "Demo Rep",
		Role:  domain.RoleRep,
	})
	if err != nil {
		t.Fatalf("issuing token: %v", err)
	}

	claims, err := a.ValidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("validating token: %v", err)
	}

	if claims.Sub != "demo-rep-1" {
		t.Errorf("expected sub demo-rep-1, got %q", claims.Sub)
	}
	if claims.Email != "rep@demo.pebblr.com" {
		t.Errorf("expected email rep@demo.pebblr.com, got %q", claims.Email)
	}
	if claims.Name != "Demo Rep" {
		t.Errorf("expected name Demo Rep, got %q", claims.Name)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != domain.RoleRep {
		t.Errorf("expected [rep] role, got %v", claims.Roles)
	}
}

func TestIssueAndValidate_AllRoles(t *testing.T) {
	t.Parallel()

	a, err := New([]byte(testKey))
	if err != nil {
		t.Fatalf("creating authenticator: %v", err)
	}

	for _, role := range []domain.Role{domain.RoleRep, domain.RoleManager, domain.RoleAdmin} {
		token, err := a.IssueToken(Persona{
			ID:    "user-" + string(role),
			Email: string(role) + "@demo.pebblr.com",
			Name:  "Demo " + string(role),
			Role:  role,
		})
		if err != nil {
			t.Fatalf("issuing token for role %s: %v", role, err)
		}

		claims, err := a.ValidateToken(context.Background(), token)
		if err != nil {
			t.Fatalf("validating token for role %s: %v", role, err)
		}
		if len(claims.Roles) != 1 || claims.Roles[0] != role {
			t.Errorf("role %s: expected [%s], got %v", role, role, claims.Roles)
		}
	}
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	t.Parallel()

	a1, _ := New([]byte("key-one"))
	a2, _ := New([]byte("key-two"))

	token, _ := a1.IssueToken(Persona{
		ID:   testUserID,
		Email: testUserEmail,
		Name:  "User",
		Role:  domain.RoleRep,
	})

	_, err := a2.ValidateToken(context.Background(), token)
	if err == nil {
		t.Fatal("expected error for wrong signing key")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	t.Parallel()

	a, _ := New([]byte(testKey))
	// Set a very short TTL to create an already-expired token.
	a.tokenTTL = -1 * time.Hour

	token, _ := a.IssueToken(Persona{
		ID:    testUserID,
		Email: testUserEmail,
		Name:  "User",
		Role:  domain.RoleRep,
	})

	_, err := a.ValidateToken(context.Background(), token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestValidateToken_Malformed(t *testing.T) {
	t.Parallel()

	a, _ := New([]byte(testKey))

	for _, token := range []string{
		"not-a-jwt",
		"a.b",
		"a.b.c.d",
		"",
	} {
		_, err := a.ValidateToken(context.Background(), token)
		if err == nil {
			t.Errorf("expected error for malformed token %q", token)
		}
	}
}

func TestIssueToken_InvalidRole(t *testing.T) {
	t.Parallel()

	a, _ := New([]byte(testKey))
	_, err := a.IssueToken(Persona{
		ID:   testUserID,
		Email: testUserEmail,
		Name:  "User",
		Role:  domain.Role("invalid"),
	})
	if err == nil {
		t.Fatal("expected error for invalid role")
	}
}

func TestNew_GeneratesRandomKey(t *testing.T) {
	t.Parallel()

	a, err := New(nil)
	if err != nil {
		t.Fatalf("creating authenticator with nil key: %v", err)
	}

	// Should be able to issue and validate with the generated key.
	token, err := a.IssueToken(Persona{
		ID:    testUserID,
		Email: testUserEmail,
		Name:  "User",
		Role:  domain.RoleRep,
	})
	if err != nil {
		t.Fatalf("issuing token: %v", err)
	}
	_, err = a.ValidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("validating token: %v", err)
	}
}
