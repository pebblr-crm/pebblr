package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"

	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

func newTeamRepo(pool pgxmock.PgxPoolIface) *teamRepository {
	return &teamRepository{pool: pool}
}

func teamRows(mock pgxmock.PgxPoolIface) *pgxmock.Rows {
	return mock.NewRows([]string{"id", "name", "manager_id"}).
		AddRow("team-1", "Alpha", "mgr-1")
}

func TestTeamGet_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTeamRepo(mock)

	mock.ExpectQuery("SELECT .+ FROM teams WHERE id").
		WithArgs("team-1").
		WillReturnRows(teamRows(mock))

	team, err := repo.Get(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if team.ID != "team-1" {
		t.Errorf("expected team-1, got %s", team.ID)
	}
	if team.Name != "Alpha" {
		t.Errorf("expected Alpha, got %s", team.Name)
	}
	if team.ManagerID != "mgr-1" {
		t.Errorf("expected mgr-1, got %s", team.ManagerID)
	}
}

func TestTeamGet_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTeamRepo(mock)

	mock.ExpectQuery("SELECT .+ FROM teams WHERE id").
		WithArgs("missing").
		WillReturnRows(mock.NewRows([]string{"id", "name", "manager_id"}))

	_, err := repo.Get(context.Background(), "missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestTeamList_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTeamRepo(mock)

	mock.ExpectQuery("SELECT .+ FROM teams ORDER BY name").
		WillReturnRows(teamRows(mock))

	teams, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(teams) != 1 {
		t.Errorf("expected 1 team, got %d", len(teams))
	}
}

func TestTeamList_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTeamRepo(mock)

	mock.ExpectQuery("SELECT .+ FROM teams ORDER BY name").
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.List(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTeamCreate_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTeamRepo(mock)

	mock.ExpectQuery("INSERT INTO teams").
		WithArgs("Beta", "mgr-2").
		WillReturnRows(mock.NewRows([]string{"id", "name", "manager_id"}).
			AddRow("team-2", "Beta", "mgr-2"))

	team, err := repo.Create(context.Background(), &domain.Team{Name: "Beta", ManagerID: "mgr-2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if team.ID != "team-2" {
		t.Errorf("expected team-2, got %s", team.ID)
	}
}

func TestTeamUpdate_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTeamRepo(mock)

	mock.ExpectQuery("UPDATE teams").
		WithArgs("Updated", "mgr-1", "missing").
		WillReturnRows(mock.NewRows([]string{"id", "name", "manager_id"}))

	_, err := repo.Update(context.Background(), &domain.Team{ID: "missing", Name: "Updated", ManagerID: "mgr-1"})
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestTeamDelete_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTeamRepo(mock)

	mock.ExpectExec("DELETE FROM teams WHERE id").
		WithArgs("team-1").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := repo.Delete(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTeamDelete_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTeamRepo(mock)

	mock.ExpectExec("DELETE FROM teams WHERE id").
		WithArgs("missing").
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err := repo.Delete(context.Background(), "missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestTeamListMembers_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTeamRepo(mock)

	mock.ExpectQuery("SELECT u.id.+FROM users u.+JOIN team_members").
		WithArgs("team-1").
		WillReturnRows(mock.NewRows([]string{
			"id", "external_id", "email", "name", "role", "avatar", "online_status",
		}).AddRow("user-1", "ext-1", "user@test.com", "Test User", "rep", "", "offline"))

	members, err := repo.ListMembers(context.Background(), "team-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(members) != 1 {
		t.Errorf("expected 1 member, got %d", len(members))
	}
}

func TestTeamListMembers_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTeamRepo(mock)

	mock.ExpectQuery("SELECT u.id.+FROM users u.+JOIN team_members").
		WithArgs("team-1").
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.ListMembers(context.Background(), "team-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
