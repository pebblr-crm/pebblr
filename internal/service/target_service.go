package service

import (
	"context"
	"fmt"

	"github.com/pebblr/pebblr/internal/config"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/store"
)

// TargetService handles target business logic with RBAC enforcement.
type TargetService struct {
	targets  store.TargetRepository
	enforcer rbac.Enforcer
	cfg      *config.TenantConfig
}

// NewTargetService constructs a TargetService with the given dependencies.
func NewTargetService(targets store.TargetRepository, enforcer rbac.Enforcer, cfg *config.TenantConfig) *TargetService {
	return &TargetService{targets: targets, enforcer: enforcer, cfg: cfg}
}

// Create persists a new target. Only managers and admins may create targets.
func (s *TargetService) Create(ctx context.Context, actor *domain.User, target *domain.Target) (*domain.Target, error) {
	if actor.Role == domain.RoleRep {
		return nil, ErrForbidden
	}
	if err := s.validateTarget(target); err != nil {
		return nil, err
	}

	created, err := s.targets.Create(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("creating target: %w", err)
	}
	return created, nil
}

// Get retrieves a target by ID with RBAC enforcement.
func (s *TargetService) Get(ctx context.Context, actor *domain.User, id string) (*domain.Target, error) {
	target, err := s.targets.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting target: %w", err)
	}
	if !s.enforcer.CanViewTarget(ctx, actor, target) {
		return nil, ErrForbidden
	}
	return target, nil
}

// List returns a paginated list of targets scoped to the actor's permissions.
func (s *TargetService) List(ctx context.Context, actor *domain.User, filter store.TargetFilter, page, limit int) (*store.TargetPage, error) {
	scope := s.enforcer.ScopeTargetQuery(ctx, actor)
	result, err := s.targets.List(ctx, scope, filter, page, limit)
	if err != nil {
		return nil, fmt.Errorf("listing targets: %w", err)
	}
	return result, nil
}

// Update persists changes to an existing target. Reps can only update editable fields
// on their own targets; managers/admins can update any field.
func (s *TargetService) Update(ctx context.Context, actor *domain.User, target *domain.Target) (*domain.Target, error) {
	existing, err := s.targets.Get(ctx, target.ID)
	if err != nil {
		return nil, fmt.Errorf("getting target: %w", err)
	}
	if !s.enforcer.CanUpdateTarget(ctx, actor, existing) {
		return nil, ErrForbidden
	}
	if err := s.validateTarget(target); err != nil {
		return nil, err
	}

	updated, err := s.targets.Update(ctx, target)
	if err != nil {
		return nil, fmt.Errorf("updating target: %w", err)
	}
	return updated, nil
}

// Import bulk-upserts targets by external ID. Admin-only.
func (s *TargetService) Import(ctx context.Context, actor *domain.User, targets []*domain.Target) (*store.ImportResult, error) {
	if actor.Role != domain.RoleAdmin {
		return nil, ErrForbidden
	}
	for i, t := range targets {
		if t.ExternalID == "" {
			return nil, fmt.Errorf("target at index %d: %w: externalId is required", i, ErrInvalidInput)
		}
		if err := s.validateTarget(t); err != nil {
			return nil, fmt.Errorf("target at index %d: %w", i, err)
		}
	}
	result, err := s.targets.Upsert(ctx, targets)
	if err != nil {
		return nil, fmt.Errorf("importing targets: %w", err)
	}
	return result, nil
}

// validateTarget checks that the target has a valid type and name.
func (s *TargetService) validateTarget(target *domain.Target) error {
	if target.Name == "" {
		return ErrInvalidInput
	}
	if s.cfg != nil {
		if s.cfg.AccountType(target.TargetType) == nil {
			return ErrInvalidInput
		}
	}
	return nil
}
