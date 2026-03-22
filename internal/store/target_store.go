package store

import (
	"context"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
)

// TargetFilter specifies optional filter criteria for target list queries.
type TargetFilter struct {
	TargetType *string
	AssigneeID *string
	TeamID     *string
	Query      *string // name search
}

// TargetPage holds a paginated result set of targets.
type TargetPage struct {
	Targets []*domain.Target
	Total   int
	Page    int
	Limit   int
}

// ImportResult holds the outcome of a bulk import operation.
type ImportResult struct {
	Created  int
	Updated  int
	Imported []*domain.Target
}

// TargetRepository provides CRUD and scoped query access for targets.
type TargetRepository interface {
	// Get retrieves a single target by ID.
	// Returns ErrNotFound if no target exists with that ID.
	Get(ctx context.Context, id string) (*domain.Target, error)

	// List returns a paginated, scoped list of targets matching the filter.
	List(ctx context.Context, scope rbac.TargetScope, filter TargetFilter, page, limit int) (*TargetPage, error)

	// Create persists a new target and returns it with its generated ID.
	Create(ctx context.Context, target *domain.Target) (*domain.Target, error)

	// Update persists changes to an existing target.
	Update(ctx context.Context, target *domain.Target) (*domain.Target, error)

	// Upsert inserts or updates targets keyed by (target_type, external_id).
	// Returns the number of created and updated records.
	Upsert(ctx context.Context, targets []*domain.Target) (*ImportResult, error)
}
