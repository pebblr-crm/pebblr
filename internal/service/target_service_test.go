package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pebblr/pebblr/internal/config"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// --- stub target repo ---

type stubTargetRepo struct {
	target  *domain.Target
	created *domain.Target
	updated *domain.Target
	getErr  error
	saveErr error
}

func (r *stubTargetRepo) Get(_ context.Context, _ string) (*domain.Target, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.target, nil
}

func (r *stubTargetRepo) List(_ context.Context, _ rbac.TargetScope, _ store.TargetFilter, _, _ int) (*store.TargetPage, error) {
	if r.target != nil {
		return &store.TargetPage{Targets: []*domain.Target{r.target}, Total: 1, Page: 1, Limit: 20}, nil
	}
	return &store.TargetPage{}, nil
}

func (r *stubTargetRepo) Create(_ context.Context, target *domain.Target) (*domain.Target, error) {
	if r.saveErr != nil {
		return nil, r.saveErr
	}
	target.ID = "target-1"
	target.CreatedAt = time.Now().UTC()
	r.created = target
	return target, nil
}

func (r *stubTargetRepo) Update(_ context.Context, target *domain.Target) (*domain.Target, error) {
	if r.saveErr != nil {
		return nil, r.saveErr
	}
	r.updated = target
	return target, nil
}

// --- test config ---

func testConfig() *config.TenantConfig {
	return &config.TenantConfig{
		Accounts: config.AccountsConfig{
			Types: []config.AccountTypeConfig{
				{Key: "doctor", Label: "Doctor"},
				{Key: "pharmacy", Label: "Pharmacy"},
			},
		},
	}
}

func sampleTarget() *domain.Target {
	return &domain.Target{
		ID:         "target-1",
		TargetType: "doctor",
		Name:       "Dr. Smith",
		Fields:     map[string]any{},
		AssigneeID: "rep-1",
		TeamID:     "team-1",
	}
}

func newTargetSvc(repo *stubTargetRepo) *service.TargetService {
	return service.NewTargetService(repo, rbac.NewEnforcer(), testConfig())
}

// --- Create tests ---

func TestTargetCreate_AdminSucceeds(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{}
	svc := newTargetSvc(repo)

	target := &domain.Target{TargetType: "doctor", Name: "Dr. Jones", Fields: map[string]any{}}
	created, err := svc.Create(context.Background(), adminUser(), target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == "" {
		t.Error("expected ID to be set")
	}
}

func TestTargetCreate_ManagerSucceeds(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{}
	svc := newTargetSvc(repo)

	target := &domain.Target{TargetType: "pharmacy", Name: "Central Pharmacy", Fields: map[string]any{}}
	_, err := svc.Create(context.Background(), managerUser(), target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTargetCreate_RepForbidden(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{}
	svc := newTargetSvc(repo)

	target := &domain.Target{TargetType: "doctor", Name: "Dr. Jones", Fields: map[string]any{}}
	_, err := svc.Create(context.Background(), repUser(), target)
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestTargetCreate_InvalidType(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{}
	svc := newTargetSvc(repo)

	target := &domain.Target{TargetType: "unknown", Name: "Test", Fields: map[string]any{}}
	_, err := svc.Create(context.Background(), adminUser(), target)
	if !errors.Is(err, service.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestTargetCreate_EmptyName(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{}
	svc := newTargetSvc(repo)

	target := &domain.Target{TargetType: "doctor", Name: "", Fields: map[string]any{}}
	_, err := svc.Create(context.Background(), adminUser(), target)
	if !errors.Is(err, service.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

// --- Get tests ---

func TestTargetGet_RepOwnsTarget(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{target: sampleTarget()}
	svc := newTargetSvc(repo)

	target, err := svc.Get(context.Background(), repUser(), "target-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.ID != "target-1" {
		t.Errorf("expected target-1, got %s", target.ID)
	}
}

func TestTargetGet_RepForbiddenOtherTarget(t *testing.T) {
	t.Parallel()
	otherTarget := &domain.Target{ID: "target-2", AssigneeID: "other-rep", TeamID: "team-1"}
	repo := &stubTargetRepo{target: otherTarget}
	svc := newTargetSvc(repo)

	_, err := svc.Get(context.Background(), repUser(), "target-2")
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestTargetGet_NotFound(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{getErr: store.ErrNotFound}
	svc := newTargetSvc(repo)

	_, err := svc.Get(context.Background(), adminUser(), "missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- List tests ---

func TestTargetList_RepScopedToOwnTargets(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{target: sampleTarget()}
	svc := newTargetSvc(repo)

	page, err := svc.List(context.Background(), repUser(), store.TargetFilter{}, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 1 {
		t.Errorf("expected 1 target, got %d", page.Total)
	}
}

// --- Update tests ---

func TestTargetUpdate_RepCanUpdateOwnTarget(t *testing.T) {
	t.Parallel()
	existing := sampleTarget()
	updated := *existing
	updated.Name = "Dr. Smith Jr."
	repo := &stubTargetRepo{target: existing}
	svc := newTargetSvc(repo)

	result, err := svc.Update(context.Background(), repUser(), &updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "Dr. Smith Jr." {
		t.Errorf("expected updated name, got %s", result.Name)
	}
}

func TestTargetUpdate_RepForbiddenOnOtherTarget(t *testing.T) {
	t.Parallel()
	otherTarget := &domain.Target{ID: "target-2", TargetType: "doctor", Name: "Other", AssigneeID: "other-rep", TeamID: "team-1", Fields: map[string]any{}}
	repo := &stubTargetRepo{target: otherTarget}
	svc := newTargetSvc(repo)

	updated := *otherTarget
	_, err := svc.Update(context.Background(), repUser(), &updated)
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}
