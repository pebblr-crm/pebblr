package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/pebblr/pebblr/internal/store"
)

func TestUserGetByID_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := &userRepository{pool: mock}

	mock.ExpectQuery("SELECT .+ FROM users WHERE id").
		WithArgs("missing").
		WillReturnRows(mock.NewRows([]string{
			"id", "external_id", "email", "name", "role", "avatar", "online_status",
		}))

	_, err := repo.GetByID(context.Background(), "missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestUserListPaginated_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := &userRepository{pool: mock}

	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(25))

	mock.ExpectQuery("SELECT u.id.+FROM users u").
		WithArgs(20, 0).
		WillReturnRows(mock.NewRows([]string{
			"id", "external_id", "email", "name", "role", "avatar", "online_status", "team_ids",
		}).AddRow("user-1", "ext-1", "user@test.com", "Test User", "rep", "", "offline", []string{"team-1"}))

	page, err := repo.ListPaginated(context.Background(), 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 25 {
		t.Errorf("expected total 25, got %d", page.Total)
	}
	if len(page.Users) != 1 {
		t.Errorf("expected 1 user, got %d", len(page.Users))
	}
	if page.Page != 1 {
		t.Errorf("expected page 1, got %d", page.Page)
	}
}

func TestUserListPaginated_Defaults(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := &userRepository{pool: mock}

	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("SELECT u.id.+FROM users u").
		WithArgs(20, 0).
		WillReturnRows(mock.NewRows([]string{
			"id", "external_id", "email", "name", "role", "avatar", "online_status", "team_ids",
		}))

	page, err := repo.ListPaginated(context.Background(), -1, 500)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Page != 1 {
		t.Errorf("expected page 1, got %d", page.Page)
	}
	if page.Limit != 20 {
		t.Errorf("expected limit 20, got %d", page.Limit)
	}
}

func TestUserListPaginated_CountError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := &userRepository{pool: mock}

	mock.ExpectQuery("SELECT COUNT").
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.ListPaginated(context.Background(), 1, 20)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
