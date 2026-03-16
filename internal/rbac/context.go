package rbac

import (
	"context"
	"fmt"

	"github.com/pebblr/pebblr/internal/domain"
)

type contextKey string

const userKey contextKey = "rbac_user"

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

// MustUserFromContext retrieves the User from context or panics.
// Only use in handlers that are guaranteed to run after auth middleware.
func MustUserFromContext(ctx context.Context) *domain.User {
	user, err := UserFromContext(ctx)
	if err != nil {
		panic(fmt.Sprintf("rbac: %v", err))
	}
	return user
}
