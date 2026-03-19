package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// --- stub implementations ---

type stubCustomerRepo struct {
	customer  *domain.Customer
	created   *domain.Customer
	updated   *domain.Customer
	leads     []*domain.Lead
	getErr    error
	saveErr   error
	leadsErr  error
}

func (r *stubCustomerRepo) Get(_ context.Context, _ string) (*domain.Customer, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.customer, nil
}

func (r *stubCustomerRepo) List(_ context.Context, _ store.CustomerFilter, _, _ int) (*store.CustomerPage, error) {
	if r.customer != nil {
		return &store.CustomerPage{Customers: []*domain.Customer{r.customer}, Total: 1, Page: 1, Limit: 20}, nil
	}
	return &store.CustomerPage{Customers: []*domain.Customer{}, Total: 0, Page: 1, Limit: 20}, nil
}

func (r *stubCustomerRepo) Create(_ context.Context, c *domain.Customer) (*domain.Customer, error) {
	if r.saveErr != nil {
		return nil, r.saveErr
	}
	c.ID = "cust-1"
	r.created = c
	return c, nil
}

func (r *stubCustomerRepo) Update(_ context.Context, c *domain.Customer) (*domain.Customer, error) {
	if r.saveErr != nil {
		return nil, r.saveErr
	}
	r.updated = c
	return c, nil
}

func (r *stubCustomerRepo) ListLeads(_ context.Context, _ string) ([]*domain.Lead, error) {
	if r.leadsErr != nil {
		return nil, r.leadsErr
	}
	return r.leads, nil
}

// --- helpers ---

func newCustomerSvc(repo *stubCustomerRepo) *service.CustomerService {
	return service.NewCustomerService(repo)
}

func sampleCustomer() *domain.Customer {
	return &domain.Customer{
		ID:   "cust-1",
		Name: "Acme Corp",
		Type: domain.CustomerTypeRetail,
	}
}

// --- Create tests ---

func TestCustomerCreate_AdminSucceeds(t *testing.T) {
	t.Parallel()
	repo := &stubCustomerRepo{}
	svc := newCustomerSvc(repo)

	c := &domain.Customer{Name: "Acme Corp", Type: domain.CustomerTypeRetail}
	created, err := svc.Create(context.Background(), adminUser(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == "" {
		t.Error("expected ID to be set")
	}
}

func TestCustomerCreate_ManagerSucceeds(t *testing.T) {
	t.Parallel()
	repo := &stubCustomerRepo{}
	svc := newCustomerSvc(repo)

	c := &domain.Customer{Name: "Beta Inc", Type: domain.CustomerTypeWholesale}
	_, err := svc.Create(context.Background(), managerUser(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCustomerCreate_RepForbidden(t *testing.T) {
	t.Parallel()
	repo := &stubCustomerRepo{}
	svc := newCustomerSvc(repo)

	c := &domain.Customer{Name: "X Corp", Type: domain.CustomerTypeRetail}
	_, err := svc.Create(context.Background(), repUser(), c)
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestCustomerCreate_MissingName(t *testing.T) {
	t.Parallel()
	repo := &stubCustomerRepo{}
	svc := newCustomerSvc(repo)

	c := &domain.Customer{Name: "", Type: domain.CustomerTypeRetail}
	_, err := svc.Create(context.Background(), adminUser(), c)
	if !errors.Is(err, service.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCustomerCreate_InvalidType(t *testing.T) {
	t.Parallel()
	repo := &stubCustomerRepo{}
	svc := newCustomerSvc(repo)

	c := &domain.Customer{Name: "X Corp", Type: domain.CustomerType("bogus")}
	_, err := svc.Create(context.Background(), adminUser(), c)
	if !errors.Is(err, service.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

// --- Get tests ---

func TestCustomerGet_Succeeds(t *testing.T) {
	t.Parallel()
	repo := &stubCustomerRepo{customer: sampleCustomer(), leads: []*domain.Lead{}}
	svc := newCustomerSvc(repo)

	detail, err := svc.Get(context.Background(), "cust-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.Customer.ID != "cust-1" {
		t.Errorf("expected cust-1, got %s", detail.Customer.ID)
	}
	if detail.Leads == nil {
		t.Error("expected non-nil leads slice")
	}
}

func TestCustomerGet_NotFound(t *testing.T) {
	t.Parallel()
	repo := &stubCustomerRepo{getErr: store.ErrNotFound}
	svc := newCustomerSvc(repo)

	_, err := svc.Get(context.Background(), "missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- List tests ---

func TestCustomerList_Succeeds(t *testing.T) {
	t.Parallel()
	repo := &stubCustomerRepo{customer: sampleCustomer()}
	svc := newCustomerSvc(repo)

	page, err := svc.List(context.Background(), store.CustomerFilter{}, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 1 {
		t.Errorf("expected 1 customer, got %d", page.Total)
	}
}

// --- Update tests ---

func TestCustomerUpdate_AdminSucceeds(t *testing.T) {
	t.Parallel()
	existing := sampleCustomer()
	repo := &stubCustomerRepo{customer: existing}
	svc := newCustomerSvc(repo)

	updated := *existing
	updated.Name = "Acme Corp Updated"
	result, err := svc.Update(context.Background(), adminUser(), &updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "Acme Corp Updated" {
		t.Errorf("expected updated name, got %s", result.Name)
	}
}

func TestCustomerUpdate_RepForbidden(t *testing.T) {
	t.Parallel()
	existing := sampleCustomer()
	repo := &stubCustomerRepo{customer: existing}
	svc := newCustomerSvc(repo)

	updated := *existing
	_, err := svc.Update(context.Background(), repUser(), &updated)
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestCustomerUpdate_NotFound(t *testing.T) {
	t.Parallel()
	repo := &stubCustomerRepo{getErr: store.ErrNotFound}
	svc := newCustomerSvc(repo)

	c := &domain.Customer{ID: "missing", Name: "X", Type: domain.CustomerTypeRetail}
	_, err := svc.Update(context.Background(), adminUser(), c)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
