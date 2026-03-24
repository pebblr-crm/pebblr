package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/pebblr/pebblr/internal/config"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/geo"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// --- stub target repo ---

type stubTargetRepo struct {
	target          *domain.Target
	created         *domain.Target
	updated         *domain.Target
	upserted        []*domain.Target
	frequencyStatus []store.TargetFrequencyStatus
	getErr          error
	saveErr         error
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

func (r *stubTargetRepo) VisitStatus(_ context.Context, _ rbac.TargetScope, _ []string) ([]store.TargetVisitStatus, error) {
	return nil, nil
}

func (r *stubTargetRepo) FrequencyStatus(_ context.Context, _ rbac.TargetScope, _ []string, _, _ time.Time) ([]store.TargetFrequencyStatus, error) {
	return r.frequencyStatus, nil
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
		Activities: config.ActivitiesConfig{
			Types: []config.ActivityTypeConfig{
				{Key: "visit", Label: "Visit", Category: "field"},
			},
		},
		Rules: config.RulesConfig{
			Frequency: map[string]int{"a": 4, "b": 2, "c": 1},
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

// --- Geocoding tests ---

type stubGeocoder struct {
	result *geo.Result
	err    error
	calls  int
}

func (g *stubGeocoder) Geocode(_ context.Context, _ string) (*geo.Result, error) {
	g.calls++
	return g.result, g.err
}

func TestTargetImport_GeocodesAddresses(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{}
	gc := &stubGeocoder{result: &geo.Result{Lat: 44.4268, Lng: 26.1025, FormattedAddress: "București, Romania"}}
	svc := service.NewTargetService(repo, rbac.NewEnforcer(), testConfig(), service.WithGeocoder(gc))

	targets := []*domain.Target{
		{ExternalID: "ext-1", TargetType: "doctor", Name: "Dr. Test", Fields: map[string]any{
			"address": "Str. Exemplu 1", "city": "București",
		}},
	}
	_, err := svc.Import(context.Background(), adminUser(), targets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gc.calls != 1 {
		t.Errorf("expected 1 geocode call, got %d", gc.calls)
	}
	if targets[0].Fields["lat"] != 44.4268 {
		t.Errorf("expected lat 44.4268, got %v", targets[0].Fields["lat"])
	}
	if targets[0].Fields["lng"] != 26.1025 {
		t.Errorf("expected lng 26.1025, got %v", targets[0].Fields["lng"])
	}
}

func TestTargetImport_SkipsAlreadyGeocoded(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{}
	gc := &stubGeocoder{result: &geo.Result{Lat: 1, Lng: 2}}
	svc := service.NewTargetService(repo, rbac.NewEnforcer(), testConfig(), service.WithGeocoder(gc))

	targets := []*domain.Target{
		{ExternalID: "ext-1", TargetType: "doctor", Name: "Dr. Test", Fields: map[string]any{
			"address": "Str. Exemplu 1", "city": "București", "lat": 44.0, "lng": 26.0,
		}},
	}
	_, err := svc.Import(context.Background(), adminUser(), targets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gc.calls != 0 {
		t.Errorf("expected 0 geocode calls for already-geocoded target, got %d", gc.calls)
	}
}

func TestTargetImport_GeocodingFailureDoesNotBlock(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{}
	gc := &stubGeocoder{err: geo.ErrNoResults}
	svc := service.NewTargetService(repo, rbac.NewEnforcer(), testConfig(), service.WithGeocoder(gc))

	targets := []*domain.Target{
		{ExternalID: "ext-1", TargetType: "doctor", Name: "Dr. Test", Fields: map[string]any{
			"address": "Nonexistent Address", "city": "Nowhere",
		}},
	}
	_, err := svc.Import(context.Background(), adminUser(), targets)
	if err != nil {
		t.Fatalf("expected import to succeed despite geocoding failure, got %v", err)
	}
	if _, hasLat := targets[0].Fields["lat"]; hasLat {
		t.Error("expected no lat field after geocoding failure")
	}
}

// --- FrequencyStatus tests ---

func TestFrequencyStatus_CalculatesCompliance(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{
		frequencyStatus: []store.TargetFrequencyStatus{
			{TargetID: "t1", Classification: "a", VisitCount: 4},
			{TargetID: "t2", Classification: "a", VisitCount: 2},
			{TargetID: "t3", Classification: "b", VisitCount: 0},
		},
	}
	svc := service.NewTargetService(repo, rbac.NewEnforcer(), testConfig())

	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	items, err := svc.FrequencyStatus(context.Background(), adminUser(), from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	// t1: classification "a", required 4, visited 4 → 100%
	if items[0].Compliance != 100 {
		t.Errorf("t1 compliance = %f, want 100", items[0].Compliance)
	}
	// t2: classification "a", required 4, visited 2 → 50%
	if items[1].Compliance != 50 {
		t.Errorf("t2 compliance = %f, want 50", items[1].Compliance)
	}
	// t3: classification "b", required 2, visited 0 → 0%
	if items[2].Compliance != 0 {
		t.Errorf("t3 compliance = %f, want 0", items[2].Compliance)
	}
}

func TestFrequencyStatus_MultiMonth(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{
		frequencyStatus: []store.TargetFrequencyStatus{
			{TargetID: "t1", Classification: "a", VisitCount: 6},
		},
	}
	svc := service.NewTargetService(repo, rbac.NewEnforcer(), testConfig())

	// Q1: 3 months, "a" requires 4/month → expected 12
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	items, err := svc.FrequencyStatus(context.Background(), adminUser(), from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 6 / 12 = 50%
	if items[0].Compliance != 50 {
		t.Errorf("compliance = %f, want 50", items[0].Compliance)
	}
}

func TestFrequencyStatus_NoConfigRule(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{
		frequencyStatus: []store.TargetFrequencyStatus{
			{TargetID: "t1", Classification: "x", VisitCount: 3},
		},
	}
	svc := service.NewTargetService(repo, rbac.NewEnforcer(), testConfig())

	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	items, err := svc.FrequencyStatus(context.Background(), adminUser(), from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if items[0].Required != 0 {
		t.Errorf("required = %d, want 0", items[0].Required)
	}
	if items[0].Compliance != 0 {
		t.Errorf("compliance = %f, want 0", items[0].Compliance)
	}
}

func TestFrequencyStatus_CapsAt100Percent(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{
		frequencyStatus: []store.TargetFrequencyStatus{
			{TargetID: "t1", Classification: "c", VisitCount: 10},
		},
	}
	svc := service.NewTargetService(repo, rbac.NewEnforcer(), testConfig())

	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	items, err := svc.FrequencyStatus(context.Background(), adminUser(), from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// "c" requires 1/month, 10 visits → capped at 100%
	if items[0].Compliance != 100 {
		t.Errorf("compliance = %f, want 100", items[0].Compliance)
	}
}

func TestFrequencyStatus_EmptyResult(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{
		frequencyStatus: []store.TargetFrequencyStatus{},
	}
	svc := service.NewTargetService(repo, rbac.NewEnforcer(), testConfig())

	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	items, err := svc.FrequencyStatus(context.Background(), adminUser(), from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

// --- Assign tests ---

// stubAssignUserRepo provides user lookup for assignment validation.
type stubAssignUserRepo struct {
	users map[string]*domain.User
}

func (r *stubAssignUserRepo) GetByID(_ context.Context, id string) (*domain.User, error) {
	if u, ok := r.users[id]; ok {
		return u, nil
	}
	return nil, store.ErrNotFound
}
func (r *stubAssignUserRepo) GetByExternalID(_ context.Context, _ string) (*domain.User, error) {
	return nil, store.ErrNotFound
}
func (r *stubAssignUserRepo) List(_ context.Context) ([]*domain.User, error) { return nil, nil }
func (r *stubAssignUserRepo) Upsert(_ context.Context, u *domain.User) (*domain.User, error) {
	return u, nil
}

func TestTargetAssign_AdminSucceeds(t *testing.T) {
	t.Parallel()
	existing := sampleTarget()
	repo := &stubTargetRepo{target: existing}
	userRepo := &stubAssignUserRepo{users: map[string]*domain.User{
		"rep-2": {ID: "rep-2", Name: "Rep Two", Role: domain.RoleRep},
	}}
	auditRepo := &stubAuditRepo{}
	svc := service.NewTargetService(repo, rbac.NewEnforcer(), testConfig(), service.WithUsers(userRepo), service.WithAudit(auditRepo))

	updated, err := svc.Assign(context.Background(), adminUser(), "target-1", "rep-2", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.AssigneeID != "rep-2" {
		t.Errorf("expected assignee rep-2, got %s", updated.AssigneeID)
	}
	if len(auditRepo.entries) != 1 {
		t.Errorf("expected 1 audit entry, got %d", len(auditRepo.entries))
	}
	if auditRepo.entries[0].EventType != "assigned" {
		t.Errorf("expected event type assigned, got %s", auditRepo.entries[0].EventType)
	}
}

func TestTargetAssign_ManagerTeamSucceeds(t *testing.T) {
	t.Parallel()
	existing := sampleTarget()
	repo := &stubTargetRepo{target: existing}
	userRepo := &stubAssignUserRepo{users: map[string]*domain.User{
		"rep-2": {ID: "rep-2", Name: "Rep Two", Role: domain.RoleRep},
	}}
	svc := service.NewTargetService(repo, rbac.NewEnforcer(), testConfig(), service.WithUsers(userRepo))

	_, err := svc.Assign(context.Background(), managerUser(), "target-1", "rep-2", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTargetAssign_ManagerOtherTeamForbidden(t *testing.T) {
	t.Parallel()
	otherTeamTarget := &domain.Target{ID: "t-3", TargetType: "doctor", Name: "Dr. Other", AssigneeID: "rep-3", TeamID: "team-9", Fields: map[string]any{}}
	repo := &stubTargetRepo{target: otherTeamTarget}
	userRepo := &stubAssignUserRepo{users: map[string]*domain.User{
		"rep-2": {ID: "rep-2"},
	}}
	svc := service.NewTargetService(repo, rbac.NewEnforcer(), testConfig(), service.WithUsers(userRepo))

	_, err := svc.Assign(context.Background(), managerUser(), "t-3", "rep-2", "")
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestTargetAssign_RepForbidden(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{target: sampleTarget()}
	svc := newTargetSvc(repo)

	_, err := svc.Assign(context.Background(), repUser(), "target-1", "rep-2", "")
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestTargetAssign_EmptyAssigneeFails(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{target: sampleTarget()}
	svc := newTargetSvc(repo)

	_, err := svc.Assign(context.Background(), adminUser(), "target-1", "", "")
	if !errors.Is(err, service.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestTargetAssign_NonExistentAssigneeFails(t *testing.T) {
	t.Parallel()
	repo := &stubTargetRepo{target: sampleTarget()}
	userRepo := &stubAssignUserRepo{users: map[string]*domain.User{}} // empty — no users
	svc := service.NewTargetService(repo, rbac.NewEnforcer(), testConfig(), service.WithUsers(userRepo))

	_, err := svc.Assign(context.Background(), adminUser(), "target-1", "nonexistent", "")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestTargetAssign_UpdatesTeamWhenProvided(t *testing.T) {
	t.Parallel()
	existing := sampleTarget()
	repo := &stubTargetRepo{target: existing}
	userRepo := &stubAssignUserRepo{users: map[string]*domain.User{
		"rep-2": {ID: "rep-2"},
	}}
	svc := service.NewTargetService(repo, rbac.NewEnforcer(), testConfig(), service.WithUsers(userRepo))

	updated, err := svc.Assign(context.Background(), adminUser(), "target-1", "rep-2", "team-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.TeamID != "team-2" {
		t.Errorf("expected team team-2, got %s", updated.TeamID)
	}
}
