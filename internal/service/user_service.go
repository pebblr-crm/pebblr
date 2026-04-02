package service

import (
	"context"
	"fmt"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

// UserService handles user read operations with role-based access control.
type UserService struct {
	users store.UserRepository
}

// NewUserService constructs a UserService with the given repository.
func NewUserService(users store.UserRepository) *UserService {
	return &UserService{users: users}
}

// List returns users visible to the actor. Admins and managers see all users;
// reps see only themselves.
func (s *UserService) List(ctx context.Context, actor *domain.User) ([]*domain.User, error) {
	switch actor.Role {
	case domain.RoleAdmin, domain.RoleManager:
		result, err := s.users.ListPaginated(ctx, 1, 10000)
		if err != nil {
			return nil, fmt.Errorf("listing users: %w", err)
		}
		return result.Users, nil
	case domain.RoleRep:
		user, err := s.users.GetByID(ctx, actor.ID)
		if err != nil {
			return nil, fmt.Errorf("getting own user: %w", err)
		}
		return []*domain.User{user}, nil
	default:
		return nil, ErrForbidden
	}
}

// ListPaginated returns a paginated list of users.
func (s *UserService) ListPaginated(ctx context.Context, page, limit int) (*store.UserPage, error) {
	result, err := s.users.ListPaginated(ctx, page, limit)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	return result, nil
}

// Get retrieves a user by their internal ID. Admins and managers may view any
// user; reps may only view themselves.
func (s *UserService) Get(ctx context.Context, actor *domain.User, id string) (*domain.User, error) {
	switch actor.Role {
	case domain.RoleAdmin, domain.RoleManager:
		// permitted to view any user
	case domain.RoleRep:
		if actor.ID != id {
			return nil, ErrForbidden
		}
	default:
		return nil, ErrForbidden
	}

	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}
	return user, nil
}
