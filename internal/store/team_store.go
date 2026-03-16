package store

import (
	"context"

	"github.com/pebblr/pebblr/internal/domain"
)

// TeamRepository provides access to team records.
type TeamRepository interface {
	// Get retrieves a team by ID.
	Get(ctx context.Context, id string) (*domain.Team, error)

	// List returns all teams.
	List(ctx context.Context) ([]*domain.Team, error)

	// Create persists a new team.
	Create(ctx context.Context, team *domain.Team) (*domain.Team, error)

	// Update persists changes to an existing team.
	Update(ctx context.Context, team *domain.Team) (*domain.Team, error)

	// Delete removes a team by ID.
	Delete(ctx context.Context, id string) error

	// ListMembers returns the users that belong to the given team.
	ListMembers(ctx context.Context, teamID string) ([]*domain.User, error)
}
