package store

import (
	"context"

	"github.com/pebblr/pebblr/internal/domain"
)

// CollectionFilter specifies optional filter criteria for collection list queries.
type CollectionFilter struct {
	CreatorID *string
	TeamID    *string
}

// CollectionRepository provides CRUD access for target collections.
type CollectionRepository interface {
	// Create persists a new collection with its target IDs and returns it.
	Create(ctx context.Context, c *domain.Collection) (*domain.Collection, error)

	// List returns all collections visible to the given scope.
	List(ctx context.Context, filter CollectionFilter) ([]*domain.Collection, error)

	// Get retrieves a single collection by ID, including its target IDs.
	Get(ctx context.Context, id string) (*domain.Collection, error)

	// Update persists changes to a collection's name and/or target list.
	Update(ctx context.Context, c *domain.Collection) (*domain.Collection, error)

	// Delete permanently removes a collection and its items.
	Delete(ctx context.Context, id string) error
}
