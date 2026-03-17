package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/events"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

// --- stub implementations ---

type stubLeadRepo struct {
	lead    *domain.Lead
	created *domain.Lead
	updated *domain.Lead
	getErr  error
	saveErr error
	deleted bool
}

func (r *stubLeadRepo) Get(_ context.Context, _ string) (*domain.Lead, error) {
	if r.getErr != nil {
		return nil, r.getErr
	}
	return r.lead, nil
}

func (r *stubLeadRepo) List(_ context.Context, _ rbac.LeadScope, _ store.LeadFilter, _, _ int) (*store.LeadPage, error) {
	if r.lead != nil {
		return &store.LeadPage{Leads: []*domain.Lead{r.lead}, Total: 1, Page: 1, Limit: 20}, nil
	}
	return &store.LeadPage{}, nil
}

func (r *stubLeadRepo) Create(_ context.Context, lead *domain.Lead) (*domain.Lead, error) {
	if r.saveErr != nil {
		return nil, r.saveErr
	}
	lead.ID = "lead-1"
	lead.CreatedAt = time.Now().UTC()
	r.created = lead
	return lead, nil
}

func (r *stubLeadRepo) Update(_ context.Context, lead *domain.Lead) (*domain.Lead, error) {
	if r.saveErr != nil {
		return nil, r.saveErr
	}
	r.updated = lead
	return lead, nil
}

func (r *stubLeadRepo) Delete(_ context.Context, _ string) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.deleted = true
	return nil
}

type stubEventRepo struct {
	recorded []events.LeadEvent
	recordErr error
}

func (r *stubEventRepo) Record(_ context.Context, evt *events.LeadEvent) error {
	if r.recordErr != nil {
		return r.recordErr
	}
	r.recorded = append(r.recorded, *evt)
	return nil
}

func (r *stubEventRepo) ListByLead(_ context.Context, _ string) ([]events.LeadEvent, error) {
	return nil, nil
}

func (r *stubEventRepo) ListByActor(_ context.Context, _ string, _, _ time.Time) ([]events.LeadEvent, error) {
	return nil, nil
}

func (r *stubEventRepo) CountByType(_ context.Context, _, _ time.Time) (map[events.EventType]int, error) {
	return nil, nil
}

// --- helpers ---

func adminUser() *domain.User {
	return &domain.User{ID: "admin-1", Role: domain.RoleAdmin, TeamIDs: []string{"team-1"}}
}

func managerUser() *domain.User {
	return &domain.User{ID: "mgr-1", Role: domain.RoleManager, TeamIDs: []string{"team-1"}}
}

func repUser() *domain.User {
	return &domain.User{ID: "rep-1", Role: domain.RoleRep, TeamIDs: []string{"team-1"}}
}

func sampleLead() *domain.Lead {
	return &domain.Lead{
		ID:         "lead-1",
		Title:      "Acme Corp",
		Status:     domain.LeadStatusNew,
		AssigneeID: "rep-1",
		TeamID:     "team-1",
		CustomerID: "cust-1",
	}
}

func newSvc(repo *stubLeadRepo, evtRepo *stubEventRepo) *service.LeadService {
	return service.NewLeadService(repo, evtRepo, rbac.NewEnforcer())
}

// --- Create tests ---

func TestCreate_AdminSucceeds(t *testing.T) {
	t.Parallel()
	repo := &stubLeadRepo{}
	evts := &stubEventRepo{}
	svc := newSvc(repo, evts)

	lead := &domain.Lead{Title: "New Lead", TeamID: "team-1", CustomerID: "cust-1"}
	created, err := svc.Create(context.Background(), adminUser(), lead)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID == "" {
		t.Error("expected ID to be set")
	}
	if created.Status != domain.LeadStatusNew {
		t.Errorf("expected status new, got %s", created.Status)
	}
	if len(evts.recorded) != 1 || evts.recorded[0].EventType != events.EventTypeCreated {
		t.Errorf("expected created event, got %v", evts.recorded)
	}
}

func TestCreate_ManagerSucceeds(t *testing.T) {
	t.Parallel()
	repo := &stubLeadRepo{}
	evts := &stubEventRepo{}
	svc := newSvc(repo, evts)

	lead := &domain.Lead{Title: "New Lead", TeamID: "team-1", CustomerID: "cust-1"}
	_, err := svc.Create(context.Background(), managerUser(), lead)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreate_RepForbidden(t *testing.T) {
	t.Parallel()
	repo := &stubLeadRepo{}
	evts := &stubEventRepo{}
	svc := newSvc(repo, evts)

	lead := &domain.Lead{Title: "New Lead", TeamID: "team-1", CustomerID: "cust-1"}
	_, err := svc.Create(context.Background(), repUser(), lead)
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// --- Get tests ---

func TestGet_RepOwnsLead(t *testing.T) {
	t.Parallel()
	repo := &stubLeadRepo{lead: sampleLead()}
	svc := newSvc(repo, &stubEventRepo{})

	lead, err := svc.Get(context.Background(), repUser(), "lead-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lead.ID != "lead-1" {
		t.Errorf("expected lead-1, got %s", lead.ID)
	}
}

func TestGet_RepForbiddenOtherLead(t *testing.T) {
	t.Parallel()
	otherLead := &domain.Lead{ID: "lead-2", AssigneeID: "other-rep", TeamID: "team-1"}
	repo := &stubLeadRepo{lead: otherLead}
	svc := newSvc(repo, &stubEventRepo{})

	_, err := svc.Get(context.Background(), repUser(), "lead-2")
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestGet_NotFound(t *testing.T) {
	t.Parallel()
	repo := &stubLeadRepo{getErr: store.ErrNotFound}
	svc := newSvc(repo, &stubEventRepo{})

	_, err := svc.Get(context.Background(), adminUser(), "missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- List tests ---

func TestList_RepScopedToOwnLeads(t *testing.T) {
	t.Parallel()
	repo := &stubLeadRepo{lead: sampleLead()}
	svc := newSvc(repo, &stubEventRepo{})

	page, err := svc.List(context.Background(), repUser(), store.LeadFilter{}, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 1 {
		t.Errorf("expected 1 lead, got %d", page.Total)
	}
}

// --- Update tests ---

func TestUpdate_RepCanUpdateOwnLead(t *testing.T) {
	t.Parallel()
	existing := sampleLead()
	updated := *existing
	updated.Title = "Updated Title"
	repo := &stubLeadRepo{lead: existing}
	evts := &stubEventRepo{}
	svc := newSvc(repo, evts)

	result, err := svc.Update(context.Background(), repUser(), &updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Title != "Updated Title" {
		t.Errorf("expected updated title, got %s", result.Title)
	}
	// No status change, no events expected
	if len(evts.recorded) != 0 {
		t.Errorf("expected no events, got %d", len(evts.recorded))
	}
}

func TestUpdate_StatusChangeEmitsEvent(t *testing.T) {
	t.Parallel()
	existing := sampleLead()
	updated := *existing
	updated.Status = domain.LeadStatusInProgress
	repo := &stubLeadRepo{lead: existing}
	evts := &stubEventRepo{}
	svc := newSvc(repo, evts)

	_, err := svc.Update(context.Background(), adminUser(), &updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(evts.recorded) != 1 || evts.recorded[0].EventType != events.EventTypeStatusChanged {
		t.Errorf("expected status_changed event, got %v", evts.recorded)
	}
}

func TestUpdate_TerminalStatusEmitsClosedEvent(t *testing.T) {
	t.Parallel()
	existing := sampleLead()
	updated := *existing
	updated.Status = domain.LeadStatusClosedWon
	repo := &stubLeadRepo{lead: existing}
	evts := &stubEventRepo{}
	svc := newSvc(repo, evts)

	_, err := svc.Update(context.Background(), adminUser(), &updated)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(evts.recorded) != 1 || evts.recorded[0].EventType != events.EventTypeClosed {
		t.Errorf("expected closed event, got %v", evts.recorded)
	}
}

func TestUpdate_RepForbiddenOnOtherLead(t *testing.T) {
	t.Parallel()
	otherLead := &domain.Lead{ID: "lead-2", AssigneeID: "other-rep", TeamID: "team-1"}
	repo := &stubLeadRepo{lead: otherLead}
	svc := newSvc(repo, &stubEventRepo{})

	updated := *otherLead
	_, err := svc.Update(context.Background(), repUser(), &updated)
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// --- PatchStatus tests ---

func TestPatchStatus_InvalidStatusReturnsErrInvalidInput(t *testing.T) {
	t.Parallel()
	repo := &stubLeadRepo{lead: sampleLead()}
	svc := newSvc(repo, &stubEventRepo{})

	_, err := svc.PatchStatus(context.Background(), adminUser(), "lead-1", domain.LeadStatus("bogus"))
	if !errors.Is(err, service.ErrInvalidInput) {
		t.Errorf("expected ErrInvalidInput, got %v", err)
	}
}

func TestPatchStatus_RepCanUpdateOwnLead(t *testing.T) {
	t.Parallel()
	repo := &stubLeadRepo{lead: sampleLead()}
	evts := &stubEventRepo{}
	svc := newSvc(repo, evts)

	result, err := svc.PatchStatus(context.Background(), repUser(), "lead-1", domain.LeadStatusInProgress)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != domain.LeadStatusInProgress {
		t.Errorf("expected in_progress, got %s", result.Status)
	}
	if len(evts.recorded) != 1 {
		t.Errorf("expected 1 event, got %d", len(evts.recorded))
	}
}

func TestPatchStatus_RepForbiddenOnOtherLead(t *testing.T) {
	t.Parallel()
	otherLead := &domain.Lead{ID: "lead-2", AssigneeID: "other-rep", TeamID: "team-1"}
	repo := &stubLeadRepo{lead: otherLead}
	svc := newSvc(repo, &stubEventRepo{})

	_, err := svc.PatchStatus(context.Background(), repUser(), "lead-2", domain.LeadStatusInProgress)
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

// --- Delete tests ---

func TestDelete_AdminCanDelete(t *testing.T) {
	t.Parallel()
	repo := &stubLeadRepo{lead: sampleLead()}
	svc := newSvc(repo, &stubEventRepo{})

	if err := svc.Delete(context.Background(), adminUser(), "lead-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.deleted {
		t.Error("expected lead to be deleted")
	}
}

func TestDelete_ManagerCanDeleteTeamLead(t *testing.T) {
	t.Parallel()
	repo := &stubLeadRepo{lead: sampleLead()}
	svc := newSvc(repo, &stubEventRepo{})

	if err := svc.Delete(context.Background(), managerUser(), "lead-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDelete_RepForbidden(t *testing.T) {
	t.Parallel()
	repo := &stubLeadRepo{lead: sampleLead()}
	svc := newSvc(repo, &stubEventRepo{})

	err := svc.Delete(context.Background(), repUser(), "lead-1")
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}
