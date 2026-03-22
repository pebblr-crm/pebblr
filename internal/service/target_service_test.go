package service_test

import (
	"context"
	"errors"
	"fmt"
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
	target   *domain.Target
	created  *domain.Target
	updated  *domain.Target
	upserted []*domain.Target
	getErr   error
	saveErr  error
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

func (r *stubTargetRepo) Upsert(_ context.Context, targets []*domain.Target) (*store.ImportResult, error) {
	if r.saveErr != nil {
		return nil, r.saveErr
	}
	r.upserted = targets
	for i, t := range targets {
		t.ID = fmt.Sprintf("target-%d", i+1)
		t.CreatedAt = time.Now().UTC()
		now := time.Now().UTC()
		t.ImportedAt = &now
	}
	return &store.ImportResult{Created: len(targets), Imported: targets}, nil
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

// --- Import tests ---

func TestTargetImport_AdminSucceeds(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{}
	svc := newTargetSvc(repo)

	targets := []*domain.Target{
		{ExternalID: "ext-1", TargetType: "doctor", Name: "Dr. Import", Fields: map[string]any{}},
		{ExternalID: "ext-2", TargetType: "pharmacy", Name: "Central Pharmacy", Fields: map[string]any{}},
	}
	result, err := svc.Import(context.Background(), adminUser(), targets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Created != 2 {
		t.Errorf("expected 2 created, got %d", result.Created)
	}
	if len(repo.upserted) != 2 {
		t.Errorf("expected 2 upserted, got %d", len(repo.upserted))
	}
}

func TestTargetImport_RepForbidden(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{}
	svc := newTargetSvc(repo)

	targets := []*domain.Target{
		{ExternalID: "ext-1", TargetType: "doctor", Name: "Dr. Import", Fields: map[string]any{}},
	}
	_, err := svc.Import(context.Background(), repUser(), targets)
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestTargetImport_ManagerForbidden(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{}
	svc := newTargetSvc(repo)

	targets := []*domain.Target{
		{ExternalID: "ext-1", TargetType: "doctor", Name: "Dr. Import", Fields: map[string]any{}},
	}
	_, err := svc.Import(context.Background(), managerUser(), targets)
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestTargetImport_MissingExternalID(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{}
	svc := newTargetSvc(repo)

	targets := []*domain.Target{
		{ExternalID: "", TargetType: "doctor", Name: "Dr. Import", Fields: map[string]any{}},
	}
	_, err := svc.Import(context.Background(), adminUser(), targets)
	if !errors.Is(err, service.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestTargetImport_InvalidType(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{}
	svc := newTargetSvc(repo)

	targets := []*domain.Target{
		{ExternalID: "ext-1", TargetType: "invalid", Name: "Test", Fields: map[string]any{}},
	}
	_, err := svc.Import(context.Background(), adminUser(), targets)
	if !errors.Is(err, service.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestTargetImport_EmptyName(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{}
	svc := newTargetSvc(repo)

	targets := []*domain.Target{
		{ExternalID: "ext-1", TargetType: "doctor", Name: "", Fields: map[string]any{}},
	}
	_, err := svc.Import(context.Background(), adminUser(), targets)
	if !errors.Is(err, service.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}
