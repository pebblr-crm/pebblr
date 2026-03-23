package service

import (
	"context"
	"fmt"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

// CollectionService handles target collection business logic with RBAC enforcement.
type CollectionService struct {
	collections store.CollectionRepository
}

// NewCollectionService constructs a CollectionService.
func NewCollectionService(collections store.CollectionRepository) *CollectionService {
	return &CollectionService{collections: collections}
}

// Create persists a new collection owned by the actor.
func (s *CollectionService) Create(ctx context.Context, actor *domain.User, name string, targetIDs []string) (*domain.Collection, error) {
	if name == "" {
		return nil, ErrInvalidInput
	}
	if len(actor.TeamIDs) == 0 {
		return nil, fmt.Errorf("creating collection: %w", ErrInvalidInput)
	}

	c := &domain.Collection{
		Name:      name,
		CreatorID: actor.ID,
		TeamID:    actor.TeamIDs[0],
		TargetIDs: targetIDs,
	}
	if c.TargetIDs == nil {
		c.TargetIDs = []string{}
	}

	created, err := s.collections.Create(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("creating collection: %w", err)
	}
	return created, nil
}

// List returns collections visible to the actor.
func (s *CollectionService) List(ctx context.Context, actor *domain.User) ([]*domain.Collection, error) {
	filter := s.scopeFilter(actor)
	result, err := s.collections.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("listing collections: %w", err)
	}
	if result == nil {
		result = []*domain.Collection{}
	}
	return result, nil
}

// Get retrieves a collection by ID with RBAC enforcement.
func (s *CollectionService) Get(ctx context.Context, actor *domain.User, id string) (*domain.Collection, error) {
	c, err := s.collections.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting collection: %w", err)
	}
	if !s.canView(actor, c) {
		return nil, ErrForbidden
	}
	return c, nil
}

// Update modifies a collection. Only the creator or an admin can update.
func (s *CollectionService) Update(ctx context.Context, actor *domain.User, id string, name string, targetIDs []string) (*domain.Collection, error) {
	if name == "" {
		return nil, ErrInvalidInput
	}

	existing, err := s.collections.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting collection for update: %w", err)
	}
	if !s.canModify(actor, existing) {
		return nil, ErrForbidden
	}

	existing.Name = name
	existing.TargetIDs = targetIDs

	updated, err := s.collections.Update(ctx, existing)
	if err != nil {
		return nil, fmt.Errorf("updating collection: %w", err)
	}
	return updated, nil
}

// Delete removes a collection. Only the creator or an admin can delete.
func (s *CollectionService) Delete(ctx context.Context, actor *domain.User, id string) error {
	existing, err := s.collections.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("getting collection for delete: %w", err)
	}
	if !s.canModify(actor, existing) {
		return ErrForbidden
	}
	if err := s.collections.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting collection: %w", err)
	}
	return nil
}

// scopeFilter returns a CollectionFilter scoped to the actor's visibility.
func (s *CollectionService) scopeFilter(actor *domain.User) store.CollectionFilter {
	switch actor.Role {
	case domain.RoleAdmin:
		return store.CollectionFilter{} // all collections
	case domain.RoleManager:
		if len(actor.TeamIDs) > 0 {
			return store.CollectionFilter{TeamID: &actor.TeamIDs[0]}
		}
		return store.CollectionFilter{CreatorID: &actor.ID}
	default: // rep
		return store.CollectionFilter{CreatorID: &actor.ID}
	}
}

// canView checks if the actor can see the collection.
func (s *CollectionService) canView(actor *domain.User, c *domain.Collection) bool {
	switch actor.Role {
	case domain.RoleAdmin:
		return true
	case domain.RoleManager:
		return containsString(actor.TeamIDs, c.TeamID)
	default:
		return actor.ID == c.CreatorID
	}
}

// canModify checks if the actor can update or delete the collection.
func (s *CollectionService) canModify(actor *domain.User, c *domain.Collection) bool {
	if actor.Role == domain.RoleAdmin {
		return true
	}
	return actor.ID == c.CreatorID
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
