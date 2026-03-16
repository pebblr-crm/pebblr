package store

import (
	"context"

	"github.com/pebblr/pebblr/internal/domain"
)

// UserRepository provides access to user records.
type UserRepository interface {
	// GetByID retrieves a user by their internal ID.
	GetByID(ctx context.Context, id string) (*domain.User, error)

	// GetByExternalID retrieves a user by their Azure AD OID.
	GetByExternalID(ctx context.Context, externalID string) (*domain.User, error)

	// List returns all users. Intended for admin/manager views.
	List(ctx context.Context) ([]*domain.User, error)

	// Upsert creates or updates a user record based on ExternalID.
	// Used to sync user details from OIDC claims on login.
	Upsert(ctx context.Context, user *domain.User) (*domain.User, error)
}
