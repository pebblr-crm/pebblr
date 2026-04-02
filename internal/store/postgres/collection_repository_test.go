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

func collectionColumns() []string {
	return []string{"id", "name", "creator_id", "team_id", "created_at", "updated_at"}
}

func collectionWithItemsColumns() []string {
	return []string{"id", "name", "creator_id", "team_id", "created_at", "updated_at", "target_ids"}
}

func collectionRow(mock pgxmock.PgxPoolIface) *pgxmock.Rows {
	now := testTime()
	return mock.NewRows(collectionColumns()).AddRow(
		"col-1", "Monday Route", "rep-1", "team-1", now, now,
	)
}

func collectionWithItemsRow(mock pgxmock.PgxPoolIface) *pgxmock.Rows {
	now := testTime()
	return mock.NewRows(collectionWithItemsColumns()).AddRow(
		"col-1", "Monday Route", "rep-1", "team-1", now, now, []string{"t1", "t2"},
	)
}

func emptyCollectionWithItemsRows(mock pgxmock.PgxPoolIface) *pgxmock.Rows {
	return mock.NewRows(collectionWithItemsColumns())
}

func newCollectionRepo(pool pgxmock.PgxPoolIface) *collectionRepository {
	return &collectionRepository{pool: pool}
}

// --- Get tests ---

func TestCollectionGet_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM target_collections c").
		WithArgs("col-1").
		WillReturnRows(collectionWithItemsRow(mock))

	c, err := repo.Get(ctx, "col-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.ID != "col-1" {
		t.Errorf("expected ID col-1, got %s", c.ID)
	}
	if c.Name != "Monday Route" {
		t.Errorf("expected name Monday Route, got %s", c.Name)
	}
	if len(c.TargetIDs) != 2 {
		t.Errorf("expected 2 targetIDs, got %d", len(c.TargetIDs))
	}
}

func TestCollectionGet_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM target_collections c").
		WithArgs("missing").
		WillReturnRows(emptyCollectionWithItemsRows(mock))

	_, err := repo.Get(ctx, "missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestCollectionGet_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM target_collections c").
		WithArgs("col-1").
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.Get(ctx, "col-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- List tests ---

func TestCollectionList_NoFilter(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM target_collections c").
		WillReturnRows(collectionWithItemsRow(mock))

	collections, err := repo.List(ctx, store.CollectionFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(collections) != 1 {
		t.Errorf("expected 1 collection, got %d", len(collections))
	}
}

func TestCollectionList_WithCreatorFilter(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()
	creatorID := "rep-1"

	mock.ExpectQuery("SELECT .+ FROM target_collections c").
		WithArgs("rep-1").
		WillReturnRows(collectionWithItemsRow(mock))

	collections, err := repo.List(ctx, store.CollectionFilter{CreatorID: &creatorID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(collections) != 1 {
		t.Errorf("expected 1 collection, got %d", len(collections))
	}
}

func TestCollectionList_Empty(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM target_collections c").
		WillReturnRows(emptyCollectionWithItemsRows(mock))

	collections, err := repo.List(ctx, store.CollectionFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(collections) != 0 {
		t.Errorf("expected 0 collections, got %d", len(collections))
	}
}

func TestCollectionList_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM target_collections c").
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.List(ctx, store.CollectionFilter{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Create tests ---

func TestCollectionCreate_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	c := &domain.Collection{
		Name:      "New Collection",
		CreatorID: "rep-1",
		TeamID:    "team-1",
		TargetIDs: []string{"t1", "t2"},
	}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO target_collections").
		WithArgs("New Collection", "rep-1", "team-1").
		WillReturnRows(collectionRow(mock))
	mock.ExpectExec("DELETE FROM target_collection_items").
		WithArgs("col-1").
		WillReturnResult(pgxmock.NewResult("DELETE", 0))
	mock.ExpectExec("INSERT INTO target_collection_items").
		WithArgs("col-1", "t1", "t2").
		WillReturnResult(pgxmock.NewResult("INSERT", 2))
	mock.ExpectCommit()

	created, err := repo.Create(ctx, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != "col-1" {
		t.Errorf("expected ID col-1, got %s", created.ID)
	}
}

func TestCollectionCreate_EmptyTargets(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	c := &domain.Collection{
		Name:      "Empty Collection",
		CreatorID: "rep-1",
		TeamID:    "team-1",
		TargetIDs: []string{},
	}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO target_collections").
		WithArgs("Empty Collection", "rep-1", "team-1").
		WillReturnRows(collectionRow(mock))
	mock.ExpectExec("DELETE FROM target_collection_items").
		WithArgs("col-1").
		WillReturnResult(pgxmock.NewResult("DELETE", 0))
	// No INSERT INTO target_collection_items since targetIDs is empty.
	mock.ExpectCommit()

	created, err := repo.Create(ctx, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != "col-1" {
		t.Errorf("expected ID col-1, got %s", created.ID)
	}
}

func TestCollectionCreate_BeginError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	c := &domain.Collection{Name: "Fail", CreatorID: "rep-1", TeamID: "team-1"}

	mock.ExpectBegin().WillReturnError(fmt.Errorf("begin error"))

	_, err := repo.Create(ctx, c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCollectionCreate_InsertError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	c := &domain.Collection{Name: "Fail", CreatorID: "rep-1", TeamID: "team-1"}

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO target_collections").
		WithArgs("Fail", "rep-1", "team-1").
		WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	_, err := repo.Create(ctx, c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Delete tests ---

func TestCollectionDelete_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM target_collections").
		WithArgs("col-1").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := repo.Delete(ctx, "col-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCollectionDelete_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM target_collections").
		WithArgs("missing").
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err := repo.Delete(ctx, "missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestCollectionDelete_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM target_collections").
		WithArgs("col-1").
		WillReturnError(fmt.Errorf("db error"))

	err := repo.Delete(ctx, "col-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Update tests ---

func TestCollectionUpdate_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	c := &domain.Collection{
		ID:        "col-1",
		Name:      "Updated Name",
		TargetIDs: []string{"t3"},
	}

	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE target_collections SET").
		WithArgs("Updated Name", "col-1").
		WillReturnRows(collectionRow(mock))
	mock.ExpectExec("DELETE FROM target_collection_items").
		WithArgs("col-1").
		WillReturnResult(pgxmock.NewResult("DELETE", 2))
	mock.ExpectExec("INSERT INTO target_collection_items").
		WithArgs("col-1", "t3").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	updated, err := repo.Update(ctx, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.ID != "col-1" {
		t.Errorf("expected ID col-1, got %s", updated.ID)
	}
}

func TestCollectionUpdate_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	c := &domain.Collection{ID: "missing", Name: "Ghost"}

	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE target_collections SET").
		WithArgs("Ghost", "missing").
		WillReturnRows(mock.NewRows(collectionColumns()))
	mock.ExpectRollback()

	_, err := repo.Update(ctx, c)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestCollectionUpdate_NilTargetIDs(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	c := &domain.Collection{
		ID:        "col-1",
		Name:      "Name Only",
		TargetIDs: nil, // nil means don't update items
	}

	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE target_collections SET").
		WithArgs("Name Only", "col-1").
		WillReturnRows(collectionRow(mock))
	// No item replacement when TargetIDs is nil.
	mock.ExpectCommit()

	updated, err := repo.Update(ctx, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.ID != "col-1" {
		t.Errorf("expected ID col-1, got %s", updated.ID)
	}
}
