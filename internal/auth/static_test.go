package auth_test

import (
	"context"
	"testing"

	"github.com/pebblr/pebblr/internal/auth"
	"github.com/pebblr/pebblr/internal/domain"
)

func TestStaticAuthenticator_ValidToken(t *testing.T) {
	t.Parallel()
	a := auth.NewStaticAuthenticator("test-secret")
	claims, err := a.ValidateToken(context.Background(), "test-secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.Sub != "a0000000-0000-0000-0000-000000000001" {
		t.Errorf("expected sub a0000000-..., got %q", claims.Sub)
	}
	if claims.Email != "admin@pebblr.dev" {
		t.Errorf("expected email admin@pebblr.dev, got %q", claims.Email)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != domain.RoleAdmin {
		t.Errorf("expected [admin] role, got %v", claims.Roles)
	}
}

func TestStaticAuthenticator_InvalidToken(t *testing.T) {
	t.Parallel()
	a := auth.NewStaticAuthenticator("test-secret")
	_, err := a.ValidateToken(context.Background(), "wrong-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestStaticAuthenticator_EmptyToken(t *testing.T) {
	t.Parallel()
	a := auth.NewStaticAuthenticator("test-secret")
	_, err := a.ValidateToken(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}
