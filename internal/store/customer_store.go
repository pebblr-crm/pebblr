package store

import (
	"context"

	"github.com/pebblr/pebblr/internal/domain"
)

// CustomerFilter specifies optional filter criteria for customer list queries.
type CustomerFilter struct {
	Type *domain.CustomerType
}

// CustomerPage holds a paginated result set of customers.
type CustomerPage struct {
	Customers []*domain.Customer
	Total     int
	Page      int
	Limit     int
}

// CustomerRepository provides CRUD and query access for customers.
type CustomerRepository interface {
	// Get retrieves a single customer by ID.
	// Returns ErrNotFound if no customer exists with that ID.
	Get(ctx context.Context, id string) (*domain.Customer, error)

	// List returns a paginated list of customers matching the filter.
	List(ctx context.Context, filter CustomerFilter, page, limit int) (*CustomerPage, error)

	// Create persists a new customer and returns it with its generated ID.
	Create(ctx context.Context, customer *domain.Customer) (*domain.Customer, error)

	// Update persists changes to an existing customer.
	Update(ctx context.Context, customer *domain.Customer) (*domain.Customer, error)

	// ListLeads returns all non-deleted leads associated with the given customer ID.
	ListLeads(ctx context.Context, customerID string) ([]*domain.Lead, error)
}
