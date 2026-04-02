package rbac

import (
	"context"
	"fmt"

	"github.com/pebblr/pebblr/internal/domain"
)

// contextKey is an unexported type for context keys in this package,
// preventing collisions with keys defined in other packages.
type contextKey struct{}

var userKey contextKey

// WithUser stores the authenticated User in the context for RBAC checks.
func WithUser(ctx context.Context, user *domain.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext retrieves the authenticated User from the context.
// Returns an error if no user is present — callers must handle missing users
// as an authentication failure rather than silently granting access.
func UserFromContext(ctx context.Context) (*domain.User, error) {
	user, ok := ctx.Value(userKey).(*domain.User)
	if !ok || user == nil {
		return nil, fmt.Errorf("no authenticated user in context")
	}
	return user, nil
}

