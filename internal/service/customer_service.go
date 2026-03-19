package service

import (
	"context"
	"fmt"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

// CustomerDetail holds a customer and its associated leads for the GET by ID response.
type CustomerDetail struct {
	Customer *domain.Customer
	Leads    []*domain.Lead
}

// CustomerService handles customer business logic with RBAC enforcement.
type CustomerService struct {
	customers store.CustomerRepository
}

// NewCustomerService constructs a CustomerService with the given dependencies.
func NewCustomerService(customers store.CustomerRepository) *CustomerService {
	return &CustomerService{customers: customers}
}

// Create persists a new customer. Only managers and admins may create customers.
func (s *CustomerService) Create(ctx context.Context, actor *domain.User, customer *domain.Customer) (*domain.Customer, error) {
	if actor.Role == domain.RoleRep {
		return nil, ErrForbidden
	}
	if customer.Name == "" {
		return nil, ErrInvalidInput
	}
	if !customer.Type.Valid() {
		return nil, ErrInvalidInput
	}

	created, err := s.customers.Create(ctx, customer)
	if err != nil {
		return nil, fmt.Errorf("creating customer: %w", err)
	}
	return created, nil
}

// Get retrieves a customer by ID along with its associated leads.
// All authenticated users may view customers.
func (s *CustomerService) Get(ctx context.Context, id string) (*CustomerDetail, error) {
	customer, err := s.customers.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting customer: %w", err)
	}

	leads, err := s.customers.ListLeads(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("listing customer leads: %w", err)
	}
	if leads == nil {
		leads = []*domain.Lead{}
	}

	return &CustomerDetail{Customer: customer, Leads: leads}, nil
}

// List returns a paginated list of customers matching the filter.
// All authenticated users may list customers.
func (s *CustomerService) List(ctx context.Context, filter store.CustomerFilter, page, limit int) (*store.CustomerPage, error) {
	result, err := s.customers.List(ctx, filter, page, limit)
	if err != nil {
		return nil, fmt.Errorf("listing customers: %w", err)
	}
	return result, nil
}

// Update replaces a customer's mutable fields. Only managers and admins may update.
func (s *CustomerService) Update(ctx context.Context, actor *domain.User, customer *domain.Customer) (*domain.Customer, error) {
	if actor.Role == domain.RoleRep {
		return nil, ErrForbidden
	}
	if customer.Name == "" {
		return nil, ErrInvalidInput
	}
	if !customer.Type.Valid() {
		return nil, ErrInvalidInput
	}

	// Verify the customer exists before updating.
	if _, err := s.customers.Get(ctx, customer.ID); err != nil {
		return nil, fmt.Errorf("getting customer: %w", err)
	}

	updated, err := s.customers.Update(ctx, customer)
	if err != nil {
		return nil, fmt.Errorf("updating customer: %w", err)
	}
	return updated, nil
}
