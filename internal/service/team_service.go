package service

import (
	"context"
	"fmt"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

// TeamService handles team read operations.
type TeamService struct {
	teams store.TeamRepository
}

// NewTeamService constructs a TeamService with the given repository.
func NewTeamService(teams store.TeamRepository) *TeamService {
	return &TeamService{teams: teams}
}

// List returns all teams.
func (s *TeamService) List(ctx context.Context) ([]*domain.Team, error) {
	teams, err := s.teams.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing teams: %w", err)
	}
	return teams, nil
}

// Get retrieves a team by ID.
func (s *TeamService) Get(ctx context.Context, id string) (*domain.Team, error) {
	team, err := s.teams.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting team: %w", err)
	}
	return team, nil
}

// ListMembers returns the users that belong to the given team.
func (s *TeamService) ListMembers(ctx context.Context, teamID string) ([]*domain.User, error) {
	members, err := s.teams.ListMembers(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("listing team members: %w", err)
	}
	return members, nil
}
