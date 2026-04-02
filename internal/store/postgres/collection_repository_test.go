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

const (
	testCollectionName   = "Monday Route"
	querySelectCollection = "SELECT .+ FROM target_collections c"
	fmtExpectedCol1      = "expected ID col-1, got %s"
	queryInsertCollection = "INSERT INTO target_collections"
	queryDeleteItems     = "DELETE FROM target_collection_items"
	queryDeleteCollection = "DELETE FROM target_collections"
	queryUpdateCollection = "UPDATE target_collections SET"
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
		"col-1", testCollectionName, "rep-1", testTeamID, now, now,
	)
}

func collectionWithItemsRow(mock pgxmock.PgxPoolIface) *pgxmock.Rows {
	now := testTime()
	return mock.NewRows(collectionWithItemsColumns()).AddRow(
		"col-1", testCollectionName, "rep-1", testTeamID, now, now, []string{"t1", "t2"},
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

	mock.ExpectQuery(querySelectCollection).
		WithArgs("col-1").
		WillReturnRows(collectionWithItemsRow(mock))

	c, err := repo.Get(ctx, "col-1")
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if c.ID != "col-1" {
		t.Errorf(fmtExpectedCol1, c.ID)
	}
	if c.Name != testCollectionName {
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

	mock.ExpectQuery(querySelectCollection).
		WithArgs("missing").
		WillReturnRows(emptyCollectionWithItemsRows(mock))

	_, err := repo.Get(ctx, "missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf(fmtExpectedNotFound, err)
	}
}

func TestCollectionGet_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery(querySelectCollection).
		WithArgs("col-1").
		WillReturnError(errors.New(errDBMsg))

	_, err := repo.Get(ctx, "col-1")
	if err == nil {
		t.Fatal(msgExpectedErr)
	}
}

// --- List tests ---

func TestCollectionList_NoFilter(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery(querySelectCollection).
		WillReturnRows(collectionWithItemsRow(mock))

	collections, err := repo.List(ctx, store.CollectionFilter{})
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
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

	mock.ExpectQuery(querySelectCollection).
		WithArgs("rep-1").
		WillReturnRows(collectionWithItemsRow(mock))

	collections, err := repo.List(ctx, store.CollectionFilter{CreatorID: &creatorID})
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
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

	mock.ExpectQuery(querySelectCollection).
		WillReturnRows(emptyCollectionWithItemsRows(mock))

	collections, err := repo.List(ctx, store.CollectionFilter{})
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
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

	mock.ExpectQuery(querySelectCollection).
		WillReturnError(errors.New(errDBMsg))

	_, err := repo.List(ctx, store.CollectionFilter{})
	if err == nil {
		t.Fatal(msgExpectedErr)
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
		TeamID:    testTeamID,
		TargetIDs: []string{"t1", "t2"},
	}

	mock.ExpectBegin()
	mock.ExpectQuery(queryInsertCollection).
		WithArgs("New Collection", "rep-1", testTeamID).
		WillReturnRows(collectionRow(mock))
	mock.ExpectExec(queryDeleteItems).
		WithArgs("col-1").
		WillReturnResult(pgxmock.NewResult("DELETE", 0))
	mock.ExpectExec("INSERT INTO target_collection_items").
		WithArgs("col-1", "t1", "t2").
		WillReturnResult(pgxmock.NewResult("INSERT", 2))
	mock.ExpectCommit()

	created, err := repo.Create(ctx, c)
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if created.ID != "col-1" {
		t.Errorf(fmtExpectedCol1, created.ID)
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
		TeamID:    testTeamID,
		TargetIDs: []string{},
	}

	mock.ExpectBegin()
	mock.ExpectQuery(queryInsertCollection).
		WithArgs("Empty Collection", "rep-1", testTeamID).
		WillReturnRows(collectionRow(mock))
	mock.ExpectExec(queryDeleteItems).
		WithArgs("col-1").
		WillReturnResult(pgxmock.NewResult("DELETE", 0))
	// No INSERT INTO target_collection_items since targetIDs is empty.
	mock.ExpectCommit()

	created, err := repo.Create(ctx, c)
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if created.ID != "col-1" {
		t.Errorf(fmtExpectedCol1, created.ID)
	}
}

func TestCollectionCreate_BeginError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	c := &domain.Collection{Name: "Fail", CreatorID: "rep-1", TeamID: testTeamID}

	mock.ExpectBegin().WillReturnError(fmt.Errorf("begin error"))

	_, err := repo.Create(ctx, c)
	if err == nil {
		t.Fatal(msgExpectedErr)
	}
}

func TestCollectionCreate_InsertError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	c := &domain.Collection{Name: "Fail", CreatorID: "rep-1", TeamID: testTeamID}

	mock.ExpectBegin()
	mock.ExpectQuery(queryInsertCollection).
		WithArgs("Fail", "rep-1", testTeamID).
		WillReturnError(fmt.Errorf("insert error"))
	mock.ExpectRollback()

	_, err := repo.Create(ctx, c)
	if err == nil {
		t.Fatal(msgExpectedErr)
	}
}

// --- Delete tests ---

func TestCollectionDelete_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	mock.ExpectExec(queryDeleteCollection).
		WithArgs("col-1").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := repo.Delete(ctx, "col-1")
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
}

func TestCollectionDelete_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	mock.ExpectExec(queryDeleteCollection).
		WithArgs("missing").
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err := repo.Delete(ctx, "missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf(fmtExpectedNotFound, err)
	}
}

func TestCollectionDelete_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	mock.ExpectExec(queryDeleteCollection).
		WithArgs("col-1").
		WillReturnError(errors.New(errDBMsg))

	err := repo.Delete(ctx, "col-1")
	if err == nil {
		t.Fatal(msgExpectedErr)
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
	mock.ExpectQuery(queryUpdateCollection).
		WithArgs("Updated Name", "col-1").
		WillReturnRows(collectionRow(mock))
	mock.ExpectExec(queryDeleteItems).
		WithArgs("col-1").
		WillReturnResult(pgxmock.NewResult("DELETE", 2))
	mock.ExpectExec("INSERT INTO target_collection_items").
		WithArgs("col-1", "t3").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))
	mock.ExpectCommit()

	updated, err := repo.Update(ctx, c)
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if updated.ID != "col-1" {
		t.Errorf(fmtExpectedCol1, updated.ID)
	}
}

func TestCollectionUpdate_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newCollectionRepo(mock)
	ctx := context.Background()

	c := &domain.Collection{ID: "missing", Name: "Ghost"}

	mock.ExpectBegin()
	mock.ExpectQuery(queryUpdateCollection).
		WithArgs("Ghost", "missing").
		WillReturnRows(mock.NewRows(collectionColumns()))
	mock.ExpectRollback()

	_, err := repo.Update(ctx, c)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf(fmtExpectedNotFound, err)
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
	mock.ExpectQuery(queryUpdateCollection).
		WithArgs("Name Only", "col-1").
		WillReturnRows(collectionRow(mock))
	// No item replacement when TargetIDs is nil.
	mock.ExpectCommit()

	updated, err := repo.Update(ctx, c)
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if updated.ID != "col-1" {
		t.Errorf(fmtExpectedCol1, updated.ID)
	}
}
