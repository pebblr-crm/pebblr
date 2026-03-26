package store

import (
	"context"

	"github.com/pebblr/pebblr/internal/domain"
)

// TerritoryFilter specifies optional filter criteria for territory list queries.
type TerritoryFilter struct {
	TeamID *string
	Region *string
}

// TerritoryRepository provides CRUD access for territories.
type TerritoryRepository interface {
	// Get retrieves a territory by ID.
	Get(ctx context.Context, id string) (*domain.Territory, error)

	// List returns territories matching the optional filter.
	List(ctx context.Context, filter TerritoryFilter) ([]*domain.Territory, error)

	// Create persists a new territory and returns it with generated fields.
	Create(ctx context.Context, t *domain.Territory) (*domain.Territory, error)

	// Update persists changes to an existing territory.
	Update(ctx context.Context, t *domain.Territory) (*domain.Territory, error)

	// Delete removes a territory by ID.
	Delete(ctx context.Context, id string) error
}
