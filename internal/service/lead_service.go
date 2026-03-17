package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/events"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/store"
)

// ErrForbidden is returned when an actor lacks permission for an operation.
var ErrForbidden = errors.New("access denied")

// ErrInvalidInput is returned when request data fails validation.
var ErrInvalidInput = errors.New("invalid input")

// LeadService handles lead business logic with RBAC enforcement and event emission.
type LeadService struct {
	leads    store.LeadRepository
	events   store.EventRepository
	enforcer rbac.Enforcer
}

// NewLeadService constructs a LeadService with the given dependencies.
func NewLeadService(leads store.LeadRepository, evts store.EventRepository, enforcer rbac.Enforcer) *LeadService {
	return &LeadService{leads: leads, events: evts, enforcer: enforcer}
}

// Create persists a new lead. Only managers and admins may create leads.
func (s *LeadService) Create(ctx context.Context, actor *domain.User, lead *domain.Lead) (*domain.Lead, error) {
	if actor.Role == domain.RoleRep {
		return nil, ErrForbidden
	}

	lead.Status = domain.LeadStatusNew

	created, err := s.leads.Create(ctx, lead)
	if err != nil {
		return nil, fmt.Errorf("creating lead: %w", err)
	}

	if err := s.events.Record(ctx, &events.LeadEvent{
		LeadID:    created.ID,
		EventType: events.EventTypeCreated,
		ActorID:   actor.ID,
		Timestamp: time.Now().UTC(),
	}); err != nil {
		return nil, fmt.Errorf("recording created event: %w", err)
	}

	return created, nil
}

// Get retrieves a lead by ID, enforcing view permission for the actor.
func (s *LeadService) Get(ctx context.Context, actor *domain.User, id string) (*domain.Lead, error) {
	lead, err := s.leads.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting lead: %w", err)
	}

	if !s.enforcer.CanViewLead(ctx, actor, lead) {
		return nil, ErrForbidden
	}

	return lead, nil
}

// List returns a paginated, RBAC-scoped list of leads matching the filter.
func (s *LeadService) List(ctx context.Context, actor *domain.User, filter store.LeadFilter, page, limit int) (*store.LeadPage, error) {
	scope := s.enforcer.ScopeLeadQuery(ctx, actor)
	result, err := s.leads.List(ctx, scope, filter, page, limit)
	if err != nil {
		return nil, fmt.Errorf("listing leads: %w", err)
	}
	return result, nil
}

// Update replaces a lead's mutable fields, enforcing update permission.
// Emits status_changed or closed events if status changed; assigned if assignee changed.
func (s *LeadService) Update(ctx context.Context, actor *domain.User, lead *domain.Lead) (*domain.Lead, error) {
	existing, err := s.leads.Get(ctx, lead.ID)
	if err != nil {
		return nil, fmt.Errorf("getting lead: %w", err)
	}

	if !s.enforcer.CanUpdateLead(ctx, actor, existing) {
		return nil, ErrForbidden
	}

	updated, err := s.leads.Update(ctx, lead)
	if err != nil {
		return nil, fmt.Errorf("updating lead: %w", err)
	}

	if existing.Status != lead.Status {
		evtType := events.EventTypeStatusChanged
		if lead.Status.Terminal() {
			evtType = events.EventTypeClosed
		}
		if err := s.events.Record(ctx, &events.LeadEvent{
			LeadID:    updated.ID,
			EventType: evtType,
			ActorID:   actor.ID,
			Timestamp: time.Now().UTC(),
		}); err != nil {
			return nil, fmt.Errorf("recording status event: %w", err)
		}
	}

	if existing.AssigneeID != lead.AssigneeID {
		if err := s.events.Record(ctx, &events.LeadEvent{
			LeadID:    updated.ID,
			EventType: events.EventTypeAssigned,
			ActorID:   actor.ID,
			Timestamp: time.Now().UTC(),
		}); err != nil {
			return nil, fmt.Errorf("recording assigned event: %w", err)
		}
	}

	return updated, nil
}

// PatchStatus transitions a lead to a new status, enforcing update permission.
func (s *LeadService) PatchStatus(ctx context.Context, actor *domain.User, id string, status domain.LeadStatus) (*domain.Lead, error) {
	if !status.Valid() {
		return nil, ErrInvalidInput
	}

	existing, err := s.leads.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting lead: %w", err)
	}

	if !s.enforcer.CanUpdateLead(ctx, actor, existing) {
		return nil, ErrForbidden
	}

	existing.Status = status
	updated, err := s.leads.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("patching lead status: %w", err)
	}

	evtType := events.EventTypeStatusChanged
	if status.Terminal() {
		evtType = events.EventTypeClosed
	}
	if err := s.events.Record(ctx, &events.LeadEvent{
		LeadID:    updated.ID,
		EventType: evtType,
		ActorID:   actor.ID,
		Timestamp: time.Now().UTC(),
	}); err != nil {
		return nil, fmt.Errorf("recording status event: %w", err)
	}

	return updated, nil
}

// Delete removes a lead by ID, enforcing delete permission.
func (s *LeadService) Delete(ctx context.Context, actor *domain.User, id string) error {
	existing, err := s.leads.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("getting lead: %w", err)
	}

	if !s.enforcer.CanDeleteLead(ctx, actor, existing) {
		return ErrForbidden
	}

	if err := s.leads.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting lead: %w", err)
	}

	return nil
}
