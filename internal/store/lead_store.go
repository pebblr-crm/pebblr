package store

import (
	"context"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
)

// LeadFilter specifies optional filter criteria for lead list queries.
type LeadFilter struct {
	Status   *domain.LeadStatus
	Assignee *string
	Team     *string
}

// LeadPage holds a paginated result set of leads.
type LeadPage struct {
	Leads  []*domain.Lead
	Total  int
	Page   int
	Limit  int
}

// LeadRepository provides CRUD and scoped query access for leads.
// All list methods accept a LeadScope from RBAC to enforce row-level access.
type LeadRepository interface {
	// Get retrieves a single lead by ID.
	// Returns ErrNotFound if no lead exists with that ID.
	Get(ctx context.Context, id string) (*domain.Lead, error)

	// List returns a paginated, scoped list of leads matching the filter.
	List(ctx context.Context, scope rbac.LeadScope, filter LeadFilter, page, limit int) (*LeadPage, error)

	// Create persists a new lead and returns it with its generated ID.
	Create(ctx context.Context, lead *domain.Lead) (*domain.Lead, error)

	// Update persists changes to an existing lead.
	Update(ctx context.Context, lead *domain.Lead) (*domain.Lead, error)

	// Delete removes a lead by ID.
	Delete(ctx context.Context, id string) error
}
