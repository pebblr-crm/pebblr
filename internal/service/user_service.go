

package service

import (
	"context"
	"fmt"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

// UserService handles user read operations.
type UserService struct {
	users store.UserRepository
}

// NewUserService constructs a UserService with the given repository.
func NewUserService(users store.UserRepository) *UserService {
	return &UserService{users: users}
}

// List returns all users.
func (s *UserService) List(ctx context.Context) ([]*domain.User, error) {
	users, err := s.users.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}
	return users, nil
}

// Get retrieves a user by their internal ID.
func (s *UserService) Get(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting user: %w", err)
	}
	return user, nil
}
