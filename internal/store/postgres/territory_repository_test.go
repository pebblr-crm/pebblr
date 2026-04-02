package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

const (
	querySelectTerritory  = "SELECT .+ FROM territories"
	fmtExpectedTer1       = "expected ID ter-1, got %s"
	fmtExpected1Territory = "expected 1 territory, got %d"
	queryInsertTerritory  = "INSERT INTO territories"
	queryUpdateTerritory  = "UPDATE territories SET"
	queryDeleteTerritory  = "DELETE FROM territories"
)

func territoryColumns() []string {
	return []string{"id", "name", "team_id", "region", "boundary", "created_at", "updated_at"}
}

func territoryRow(mock pgxmock.PgxPoolIface) *pgxmock.Rows {
	now := testTime()
	boundary, _ := json.Marshal(map[string]any{"type": "Polygon", "coordinates": []any{}})
	return mock.NewRows(territoryColumns()).AddRow(
		"ter-1", "Bucharest North", testTeamID, "Bucharest", boundary, now, now,
	)
}

func emptyTerritoryRows(mock pgxmock.PgxPoolIface) *pgxmock.Rows {
	return mock.NewRows(territoryColumns())
}

func newTerritoryRepo(pool pgxmock.PgxPoolIface) *territoryRepository {
	return &territoryRepository{pool: pool}
}

// --- Get tests ---

func TestTerritoryGet_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery(querySelectTerritory).
		WithArgs("ter-1").
		WillReturnRows(territoryRow(mock))

	ter, err := repo.Get(ctx, "ter-1")
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if ter.ID != "ter-1" {
		t.Errorf(fmtExpectedTer1, ter.ID)
	}
	if ter.Name != "Bucharest North" {
		t.Errorf("expected name Bucharest North, got %s", ter.Name)
	}
}

func TestTerritoryGet_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery(querySelectTerritory).
		WithArgs("missing").
		WillReturnRows(emptyTerritoryRows(mock))

	_, err := repo.Get(ctx, "missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf(fmtExpectedNotFound, err)
	}
}

func TestTerritoryGet_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery(querySelectTerritory).
		WithArgs("ter-1").
		WillReturnError(fmt.Errorf("connection refused"))

	_, err := repo.Get(ctx, "ter-1")
	if err == nil {
		t.Fatal(msgExpectedErr)
	}
}

// --- List tests ---

func TestTerritoryList_NoFilter(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery(querySelectTerritory).
		WillReturnRows(territoryRow(mock))

	territories, err := repo.List(ctx, store.TerritoryFilter{})
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if len(territories) != 1 {
		t.Errorf(fmtExpected1Territory, len(territories))
	}
}

func TestTerritoryList_WithTeamFilter(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()
	teamID := testTeamID

	mock.ExpectQuery(querySelectTerritory).
		WithArgs(testTeamID).
		WillReturnRows(territoryRow(mock))

	territories, err := repo.List(ctx, store.TerritoryFilter{TeamID: &teamID})
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if len(territories) != 1 {
		t.Errorf(fmtExpected1Territory, len(territories))
	}
}

func TestTerritoryList_WithRegionFilter(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()
	region := "Bucharest"

	mock.ExpectQuery(querySelectTerritory).
		WithArgs("Bucharest").
		WillReturnRows(territoryRow(mock))

	territories, err := repo.List(ctx, store.TerritoryFilter{Region: &region})
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if len(territories) != 1 {
		t.Errorf(fmtExpected1Territory, len(territories))
	}
}

func TestTerritoryList_Empty(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery(querySelectTerritory).
		WillReturnRows(emptyTerritoryRows(mock))

	territories, err := repo.List(ctx, store.TerritoryFilter{})
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if len(territories) != 0 {
		t.Errorf("expected 0 territories, got %d", len(territories))
	}
}

func TestTerritoryList_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery(querySelectTerritory).
		WillReturnError(errors.New(errDBMsg))

	_, err := repo.List(ctx, store.TerritoryFilter{})
	if err == nil {
		t.Fatal(msgExpectedErr)
	}
}

// --- Create tests ---

func TestTerritoryCreate_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	ter := &domain.Territory{
		Name:     "Cluj Region",
		TeamID:   testTeamID,
		Region:   "Cluj",
		Boundary: map[string]any{"type": "Polygon", "coordinates": []any{}},
	}

	mock.ExpectQuery(queryInsertTerritory).
		WithArgs(anyArgs(4)...).
		WillReturnRows(territoryRow(mock))

	created, err := repo.Create(ctx, ter)
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if created.ID != "ter-1" {
		t.Errorf(fmtExpectedTer1, created.ID)
	}
}

func TestTerritoryCreate_NilBoundary(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	now := testTime()
	ter := &domain.Territory{
		Name:   "No Boundary",
		TeamID: testTeamID,
	}

	mock.ExpectQuery(queryInsertTerritory).
		WithArgs(anyArgs(4)...).
		WillReturnRows(mock.NewRows(territoryColumns()).AddRow(
			"ter-2", "No Boundary", testTeamID, "", []byte(nil), now, now,
		))

	created, err := repo.Create(ctx, ter)
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if created.Boundary != nil {
		t.Error("expected nil boundary")
	}
}

func TestTerritoryCreate_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	ter := &domain.Territory{Name: "Fail", TeamID: testTeamID}

	mock.ExpectQuery(queryInsertTerritory).
		WithArgs(anyArgs(4)...).
		WillReturnError(errors.New(errDBMsg))

	_, err := repo.Create(ctx, ter)
	if err == nil {
		t.Fatal(msgExpectedErr)
	}
}

// --- Update tests ---

func TestTerritoryUpdate_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	ter := &domain.Territory{
		ID:     "ter-1",
		Name:   "Updated Name",
		TeamID: testTeamID,
		Region: "Updated Region",
	}

	mock.ExpectQuery(queryUpdateTerritory).
		WithArgs(anyArgs(5)...).
		WillReturnRows(territoryRow(mock))

	updated, err := repo.Update(ctx, ter)
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if updated.ID != "ter-1" {
		t.Errorf(fmtExpectedTer1, updated.ID)
	}
}

func TestTerritoryUpdate_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	ter := &domain.Territory{
		ID:     "missing",
		Name:   "Ghost",
		TeamID: testTeamID,
	}

	mock.ExpectQuery(queryUpdateTerritory).
		WithArgs(anyArgs(5)...).
		WillReturnRows(emptyTerritoryRows(mock))

	_, err := repo.Update(ctx, ter)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf(fmtExpectedNotFound, err)
	}
}

func TestTerritoryUpdate_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	ter := &domain.Territory{ID: "ter-1", Name: "Fail", TeamID: testTeamID}

	mock.ExpectQuery(queryUpdateTerritory).
		WithArgs(anyArgs(5)...).
		WillReturnError(errors.New(errDBMsg))

	_, err := repo.Update(ctx, ter)
	if err == nil {
		t.Fatal(msgExpectedErr)
	}
}

// --- Delete tests ---

func TestTerritoryDelete_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	mock.ExpectExec(queryDeleteTerritory).
		WithArgs("ter-1").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := repo.Delete(ctx, "ter-1")
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
}

func TestTerritoryDelete_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	mock.ExpectExec(queryDeleteTerritory).
		WithArgs("missing").
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err := repo.Delete(ctx, "missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf(fmtExpectedNotFound, err)
	}
}

func TestTerritoryDelete_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	mock.ExpectExec(queryDeleteTerritory).
		WithArgs("ter-1").
		WillReturnError(errors.New(errDBMsg))

	err := repo.Delete(ctx, "ter-1")
	if err == nil {
		t.Fatal(msgExpectedErr)
	}
}

// --- unmarshalBoundary tests ---

func TestUnmarshalBoundary_Empty(t *testing.T) {
	t.Parallel()
	result, err := unmarshalBoundary(nil)
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if result != nil {
		t.Error("expected nil for empty data")
	}
}

func TestUnmarshalBoundary_Valid(t *testing.T) {
	t.Parallel()
	data, _ := json.Marshal(map[string]any{"type": "Polygon"})
	result, err := unmarshalBoundary(data)
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if result["type"] != "Polygon" {
		t.Errorf("expected type=Polygon, got %v", result["type"])
	}
}

func TestUnmarshalBoundary_InvalidJSON(t *testing.T) {
	t.Parallel()
	_, err := unmarshalBoundary([]byte("{bad"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
