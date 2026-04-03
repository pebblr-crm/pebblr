package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

const (
	querySelectCount = "SELECT COUNT"
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

	mock.ExpectQuery(querySelectCount).
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

	mock.ExpectQuery(querySelectCount).
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

	mock.ExpectQuery(querySelectCount).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.ListPaginated(context.Background(), 1, 20)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUserGetByID_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := &userRepository{pool: mock}

	mock.ExpectQuery("SELECT .+ FROM users WHERE id").
		WithArgs("user-1").
		WillReturnRows(mock.NewRows([]string{
			"id", "external_id", "email", "name", "role", "avatar", "online_status",
		}).AddRow("user-1", "ext-1", "user@test.com", "Test User", "rep", "", "offline"))

	user, err := repo.GetByID(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "user-1" {
		t.Errorf("expected user ID user-1, got %s", user.ID)
	}
	if user.Email != "user@test.com" {
		t.Errorf("expected email user@test.com, got %s", user.Email)
	}
}

func TestUserGetByID_ScanError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := &userRepository{pool: mock}

	mock.ExpectQuery("SELECT .+ FROM users WHERE id").
		WithArgs("user-1").
		WillReturnError(fmt.Errorf("connection error"))

	_, err := repo.GetByID(context.Background(), "user-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUserGetByExternalID_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := &userRepository{pool: mock}

	mock.ExpectQuery("SELECT .+ FROM users WHERE external_id").
		WithArgs("ext-1").
		WillReturnRows(mock.NewRows([]string{
			"id", "external_id", "email", "name", "role", "avatar", "online_status",
		}).AddRow("user-1", "ext-1", "user@test.com", "Test User", "rep", "", "offline"))

	user, err := repo.GetByExternalID(context.Background(), "ext-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ExternalID != "ext-1" {
		t.Errorf("expected external ID ext-1, got %s", user.ExternalID)
	}
}

func TestUserGetByExternalID_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := &userRepository{pool: mock}

	mock.ExpectQuery("SELECT .+ FROM users WHERE external_id").
		WithArgs("missing").
		WillReturnRows(mock.NewRows([]string{
			"id", "external_id", "email", "name", "role", "avatar", "online_status",
		}))

	_, err := repo.GetByExternalID(context.Background(), "missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestUserList_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := &userRepository{pool: mock}

	mock.ExpectQuery("SELECT u.id.+FROM users u").
		WillReturnRows(mock.NewRows([]string{
			"id", "external_id", "email", "name", "role", "avatar", "online_status", "team_ids",
		}).
			AddRow("user-1", "ext-1", "user1@test.com", "User One", "rep", "", "offline", []string{"team-1"}).
			AddRow("user-2", "ext-2", "user2@test.com", "User Two", "admin", "", "online", []string{"team-1", "team-2"}))

	users, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestUserList_QueryError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := &userRepository{pool: mock}

	mock.ExpectQuery("SELECT u.id.+FROM users u").
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.List(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUserList_ScanError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := &userRepository{pool: mock}

	// Return a row with wrong column count to trigger scan error
	mock.ExpectQuery("SELECT u.id.+FROM users u").
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow("user-1"))

	_, err := repo.List(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUserUpsert_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := &userRepository{pool: mock}

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("ext-1", "user@test.com", "Test User", "rep", "", "offline").
		WillReturnRows(mock.NewRows([]string{
			"id", "external_id", "email", "name", "role", "avatar", "online_status",
		}).AddRow("user-1", "ext-1", "user@test.com", "Test User", "rep", "", "offline"))

	user := &domain.User{
		ExternalID:   "ext-1",
		Email:        "user@test.com",
		Name:         "Test User",
		Role:         "rep",
		Avatar:       "",
		OnlineStatus: "offline",
	}
	result, err := repo.Upsert(context.Background(), user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "user-1" {
		t.Errorf("expected user ID user-1, got %s", result.ID)
	}
}

func TestUserUpsert_Error(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := &userRepository{pool: mock}

	mock.ExpectQuery("INSERT INTO users").
		WithArgs("ext-1", "user@test.com", "Test User", "rep", "", "offline").
		WillReturnError(fmt.Errorf("unique constraint violation"))

	user := &domain.User{
		ExternalID:   "ext-1",
		Email:        "user@test.com",
		Name:         "Test User",
		Role:         "rep",
		Avatar:       "",
		OnlineStatus: "offline",
	}
	_, err := repo.Upsert(context.Background(), user)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUserListPaginated_QueryError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := &userRepository{pool: mock}

	mock.ExpectQuery(querySelectCount).
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(5))

	mock.ExpectQuery("SELECT u.id.+FROM users u").
		WithArgs(20, 0).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.ListPaginated(context.Background(), 1, 20)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUserListPaginated_ScanError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := &userRepository{pool: mock}

	mock.ExpectQuery(querySelectCount).
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(5))

	mock.ExpectQuery("SELECT u.id.+FROM users u").
		WithArgs(20, 0).
		WillReturnRows(mock.NewRows([]string{"id"}).AddRow("user-1"))

	_, err := repo.ListPaginated(context.Background(), 1, 20)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
