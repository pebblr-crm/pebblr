package service

import (
	"context"
	"fmt"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

// TerritoryService handles territory business logic with RBAC enforcement.
type TerritoryService struct {
	territories store.TerritoryRepository
}

// NewTerritoryService constructs a TerritoryService.
func NewTerritoryService(territories store.TerritoryRepository) *TerritoryService {
	return &TerritoryService{territories: territories}
}

// List returns territories scoped to the actor's visibility.
func (s *TerritoryService) List(ctx context.Context, actor *domain.User) ([]*domain.Territory, error) {
	filter := s.scopeFilter(actor)
	result, err := s.territories.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing territories: %w", err)
	}
	if result == nil {
		result = []*domain.Territory{}
	}
	return result, nil
}

// Get retrieves a territory by ID with RBAC enforcement.
func (s *TerritoryService) Get(ctx context.Context, actor *domain.User, id string) (*domain.Territory, error) {
	t, err := s.territories.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting territory: %w", err)
	}
	if !s.canView(actor, t) {
		return nil, ErrForbidden
	}
	return t, nil
}

// Create persists a new territory. Only managers and admins can create.
func (s *TerritoryService) Create(ctx context.Context, actor *domain.User, t *domain.Territory) (*domain.Territory, error) {
	if actor.Role == domain.RoleRep {
		return nil, ErrForbidden
	}
	if t.Name == "" {
		return nil, ErrInvalidInput
	}

	created, err := s.territories.Create(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("creating territory: %w", err)
	}
	return created, nil
}

// Update modifies an existing territory. Only managers and admins can update.
func (s *TerritoryService) Update(ctx context.Context, actor *domain.User, t *domain.Territory) (*domain.Territory, error) {
	if actor.Role == domain.RoleRep {
		return nil, ErrForbidden
	}
	if t.Name == "" {
		return nil, ErrInvalidInput
	}

	existing, err := s.territories.Get(ctx, t.ID)
	if err != nil {
		return nil, fmt.Errorf("getting territory for update: %w", err)
	}
	if !s.canModify(actor, existing) {
		return nil, ErrForbidden
	}

	updated, err := s.territories.Update(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("updating territory: %w", err)
	}
	return updated, nil
}

// Delete removes a territory. Only managers and admins can delete.
func (s *TerritoryService) Delete(ctx context.Context, actor *domain.User, id string) error {
	if actor.Role == domain.RoleRep {
		return ErrForbidden
	}

	existing, err := s.territories.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("getting territory for delete: %w", err)
	}
	if !s.canModify(actor, existing) {
		return ErrForbidden
	}

	if err := s.territories.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting territory: %w", err)
	}
	return nil
}

func (s *TerritoryService) scopeFilter(actor *domain.User) store.TerritoryFilter {
	switch actor.Role {
	case domain.RoleAdmin:
		return store.TerritoryFilter{}
	case domain.RoleManager:
		if len(actor.TeamIDs) > 0 {
			return store.TerritoryFilter{TeamID: &actor.TeamIDs[0]}
		}
		return store.TerritoryFilter{}
	default: // rep
		if len(actor.TeamIDs) > 0 {
			return store.TerritoryFilter{TeamID: &actor.TeamIDs[0]}
		}
		return store.TerritoryFilter{}
	}
}

func (s *TerritoryService) canView(actor *domain.User, t *domain.Territory) bool {
	switch actor.Role {
	case domain.RoleAdmin:
		return true
	default:
		return containsString(actor.TeamIDs, t.TeamID)
	}
}

func (s *TerritoryService) canModify(actor *domain.User, t *domain.Territory) bool {
	switch actor.Role {
	case domain.RoleAdmin:
		return true
	case domain.RoleManager:
		return containsString(actor.TeamIDs, t.TeamID)
	default:
		return false
	}
}
