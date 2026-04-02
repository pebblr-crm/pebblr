package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/service"
	"github.com/pebblr/pebblr/internal/store"
)

const (
	testTeamID      = "team-1"
	testTeam2ID     = "team-2"
	fmtUnexpectedErr = "unexpected error: %v"
	fmtExpectedTeam1 = "expected team-1, got %s"
)

// --- stub team repo ---

type stubTeamRepo struct {
	teams   []*domain.Team
	members []*domain.User
}

func (r *stubTeamRepo) Get(_ context.Context, id string) (*domain.Team, error) {
	for _, t := range r.teams {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, store.ErrNotFound
}

func (r *stubTeamRepo) List(_ context.Context) ([]*domain.Team, error) {
	return r.teams, nil
}

func (r *stubTeamRepo) Create(_ context.Context, _ *domain.Team) (*domain.Team, error) {
	return nil, nil
}

func (r *stubTeamRepo) Update(_ context.Context, _ *domain.Team) (*domain.Team, error) {
	return nil, nil
}

func (r *stubTeamRepo) Delete(_ context.Context, _ string) error {
	return nil
}

func (r *stubTeamRepo) ListMembers(_ context.Context, _ string) ([]*domain.User, error) {
	return r.members, nil
}

func defaultTeamRepo() *stubTeamRepo {
	return &stubTeamRepo{
		teams: []*domain.Team{
			{ID: testTeamID, Name: "Alpha", ManagerID: "mgr-1"},
			{ID: testTeam2ID, Name: "Beta", ManagerID: "mgr-2"},
		},
		members: []*domain.User{
			{ID: "rep-1", Name: "Rep", Role: domain.RoleRep},
		},
	}
}

// --- tests ---

func TestTeamService_List_AdminSeesAll(t *testing.T) {
	t.Parallel()
	svc := service.NewTeamService(defaultTeamRepo())
	teams, err := svc.List(context.Background(), adminUser())
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if len(teams) != 2 {
		t.Errorf("admin should see all 2 teams, got %d", len(teams))
	}
}

func TestTeamService_List_ManagerSeesOwnTeams(t *testing.T) {
	t.Parallel()
	svc := service.NewTeamService(defaultTeamRepo())
	// managerUser() has TeamIDs: ["team-1"]
	teams, err := svc.List(context.Background(), managerUser())
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if len(teams) != 1 {
		t.Fatalf("manager should see 1 team, got %d", len(teams))
	}
	if teams[0].ID != testTeamID {
		t.Errorf(fmtExpectedTeam1, teams[0].ID)
	}
}

func TestTeamService_List_RepSeesOwnTeams(t *testing.T) {
	t.Parallel()
	svc := service.NewTeamService(defaultTeamRepo())
	// repUser() has TeamIDs: ["team-1"]
	teams, err := svc.List(context.Background(), repUser())
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if len(teams) != 1 {
		t.Fatalf("rep should see 1 team, got %d", len(teams))
	}
	if teams[0].ID != testTeamID {
		t.Errorf(fmtExpectedTeam1, teams[0].ID)
	}
}

func TestTeamService_Get_RepCannotAccessOtherTeam(t *testing.T) {
	t.Parallel()
	svc := service.NewTeamService(defaultTeamRepo())
	_, err := svc.Get(context.Background(), repUser(), testTeam2ID)
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestTeamService_Get_RepCanAccessOwnTeam(t *testing.T) {
	t.Parallel()
	svc := service.NewTeamService(defaultTeamRepo())
	team, err := svc.Get(context.Background(), repUser(), testTeamID)
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if team.ID != testTeamID {
		t.Errorf(fmtExpectedTeam1, team.ID)
	}
}

func TestTeamService_ListMembers_Forbidden(t *testing.T) {
	t.Parallel()
	svc := service.NewTeamService(defaultTeamRepo())
	_, err := svc.ListMembers(context.Background(), repUser(), testTeam2ID)
	if !errors.Is(err, service.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestTeamService_ListMembers_Allowed(t *testing.T) {
	t.Parallel()
	svc := service.NewTeamService(defaultTeamRepo())
	members, err := svc.ListMembers(context.Background(), adminUser(), testTeamID)
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if len(members) != 1 {
		t.Errorf("expected 1 member, got %d", len(members))
	}
}
