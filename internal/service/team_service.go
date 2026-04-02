package service

import (
	"context"
	"fmt"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

// TeamService handles team read operations with role-based access control.
type TeamService struct {
	teams store.TeamRepository
}

// NewTeamService constructs a TeamService with the given repository.
func NewTeamService(teams store.TeamRepository) *TeamService {
	return &TeamService{teams: teams}
}

// List returns teams visible to the actor. Admins see all teams; managers
// see their own teams; reps see only teams they belong to.
func (s *TeamService) List(ctx context.Context, actor *domain.User) ([]*domain.Team, error) {
	switch actor.Role {
	case domain.RoleAdmin:
		teams, err := s.teams.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing teams: %w", err)
		}
		return teams, nil
	case domain.RoleManager, domain.RoleRep:
		teams, err := s.teams.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("listing teams: %w", err)
		}
		// Filter to only teams the actor belongs to.
		var visible []*domain.Team
		for _, t := range teams {
			if containsString(actor.TeamIDs, t.ID) {
				visible = append(visible, t)
			}
		}
		return visible, nil
	default:
		return nil, ErrForbidden
	}
}

// Get retrieves a team by ID. Admins may view any team; managers and reps may
// only view teams they belong to.
func (s *TeamService) Get(ctx context.Context, actor *domain.User, id string) (*domain.Team, error) {
	if actor.Role != domain.RoleAdmin && !containsString(actor.TeamIDs, id) {
		return nil, ErrForbidden
	}

	team, err := s.teams.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting team: %w", err)
	}
	return team, nil
}

// ListMembers returns the users that belong to the given team. Admins may view
// any team's members; managers and reps may only view members of their own teams.
func (s *TeamService) ListMembers(ctx context.Context, actor *domain.User, teamID string) ([]*domain.User, error) {
	if actor.Role != domain.RoleAdmin && !containsString(actor.TeamIDs, teamID) {
		return nil, ErrForbidden
	}

	members, err := s.teams.ListMembers(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("listing team members: %w", err)
	}
	return members, nil
}

