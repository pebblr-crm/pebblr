package store

import (
	"context"

	"github.com/pebblr/pebblr/internal/domain"
)

// UserPage holds a paginated result set of users.
type UserPage struct {
	Users []*domain.User
	Total int
	Page  int
	Limit int
}

// UserRepository provides access to user records.
type UserRepository interface {
	// GetByID retrieves a user by their internal ID.
	GetByID(ctx context.Context, id string) (*domain.User, error)

	// GetByExternalID retrieves a user by their Azure AD OID.
	GetByExternalID(ctx context.Context, externalID string) (*domain.User, error)

	// List returns all users. Intended for admin/manager views.
	//
	// Deprecated: Use ListPaginated for new code.
	List(ctx context.Context) ([]*domain.User, error)

	// ListPaginated returns a paginated list of users.
	ListPaginated(ctx context.Context, page, limit int) (*UserPage, error)

	// Upsert creates or updates a user record based on ExternalID.
	// Used to sync user details from OIDC claims on login.
	Upsert(ctx context.Context, user *domain.User) (*domain.User, error)
}
