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

func territoryColumns() []string {
	return []string{"id", "name", "team_id", "region", "boundary", "created_at", "updated_at"}
}

func territoryRow(mock pgxmock.PgxPoolIface) *pgxmock.Rows {
	now := testTime()
	boundary, _ := json.Marshal(map[string]any{"type": "Polygon", "coordinates": []any{}})
	return mock.NewRows(territoryColumns()).AddRow(
		"ter-1", "Bucharest North", "team-1", "Bucharest", boundary, now, now,
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

	mock.ExpectQuery("SELECT .+ FROM territories").
		WithArgs("ter-1").
		WillReturnRows(territoryRow(mock))

	ter, err := repo.Get(ctx, "ter-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ter.ID != "ter-1" {
		t.Errorf("expected ID ter-1, got %s", ter.ID)
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

	mock.ExpectQuery("SELECT .+ FROM territories").
		WithArgs("missing").
		WillReturnRows(emptyTerritoryRows(mock))

	_, err := repo.Get(ctx, "missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestTerritoryGet_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM territories").
		WithArgs("ter-1").
		WillReturnError(fmt.Errorf("connection refused"))

	_, err := repo.Get(ctx, "ter-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- List tests ---

func TestTerritoryList_NoFilter(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM territories").
		WillReturnRows(territoryRow(mock))

	territories, err := repo.List(ctx, store.TerritoryFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(territories) != 1 {
		t.Errorf("expected 1 territory, got %d", len(territories))
	}
}

func TestTerritoryList_WithTeamFilter(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()
	teamID := "team-1"

	mock.ExpectQuery("SELECT .+ FROM territories").
		WithArgs("team-1").
		WillReturnRows(territoryRow(mock))

	territories, err := repo.List(ctx, store.TerritoryFilter{TeamID: &teamID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(territories) != 1 {
		t.Errorf("expected 1 territory, got %d", len(territories))
	}
}

func TestTerritoryList_WithRegionFilter(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()
	region := "Bucharest"

	mock.ExpectQuery("SELECT .+ FROM territories").
		WithArgs("Bucharest").
		WillReturnRows(territoryRow(mock))

	territories, err := repo.List(ctx, store.TerritoryFilter{Region: &region})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(territories) != 1 {
		t.Errorf("expected 1 territory, got %d", len(territories))
	}
}

func TestTerritoryList_Empty(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM territories").
		WillReturnRows(emptyTerritoryRows(mock))

	territories, err := repo.List(ctx, store.TerritoryFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

	mock.ExpectQuery("SELECT .+ FROM territories").
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.List(ctx, store.TerritoryFilter{})
	if err == nil {
		t.Fatal("expected error, got nil")
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
		TeamID:   "team-1",
		Region:   "Cluj",
		Boundary: map[string]any{"type": "Polygon", "coordinates": []any{}},
	}

	mock.ExpectQuery("INSERT INTO territories").
		WithArgs(anyArgs(4)...).
		WillReturnRows(territoryRow(mock))

	created, err := repo.Create(ctx, ter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != "ter-1" {
		t.Errorf("expected ID ter-1, got %s", created.ID)
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
		TeamID: "team-1",
	}

	mock.ExpectQuery("INSERT INTO territories").
		WithArgs(anyArgs(4)...).
		WillReturnRows(mock.NewRows(territoryColumns()).AddRow(
			"ter-2", "No Boundary", "team-1", "", []byte(nil), now, now,
		))

	created, err := repo.Create(ctx, ter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

	ter := &domain.Territory{Name: "Fail", TeamID: "team-1"}

	mock.ExpectQuery("INSERT INTO territories").
		WithArgs(anyArgs(4)...).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.Create(ctx, ter)
	if err == nil {
		t.Fatal("expected error, got nil")
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
		TeamID: "team-1",
		Region: "Updated Region",
	}

	mock.ExpectQuery("UPDATE territories SET").
		WithArgs(anyArgs(5)...).
		WillReturnRows(territoryRow(mock))

	updated, err := repo.Update(ctx, ter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.ID != "ter-1" {
		t.Errorf("expected ID ter-1, got %s", updated.ID)
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
		TeamID: "team-1",
	}

	mock.ExpectQuery("UPDATE territories SET").
		WithArgs(anyArgs(5)...).
		WillReturnRows(emptyTerritoryRows(mock))

	_, err := repo.Update(ctx, ter)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestTerritoryUpdate_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	ter := &domain.Territory{ID: "ter-1", Name: "Fail", TeamID: "team-1"}

	mock.ExpectQuery("UPDATE territories SET").
		WithArgs(anyArgs(5)...).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.Update(ctx, ter)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Delete tests ---

func TestTerritoryDelete_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM territories").
		WithArgs("ter-1").
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err := repo.Delete(ctx, "ter-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTerritoryDelete_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM territories").
		WithArgs("missing").
		WillReturnResult(pgxmock.NewResult("DELETE", 0))

	err := repo.Delete(ctx, "missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestTerritoryDelete_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTerritoryRepo(mock)
	ctx := context.Background()

	mock.ExpectExec("DELETE FROM territories").
		WithArgs("ter-1").
		WillReturnError(fmt.Errorf("db error"))

	err := repo.Delete(ctx, "ter-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- unmarshalBoundary tests ---

func TestUnmarshalBoundary_Empty(t *testing.T) {
	t.Parallel()
	result, err := unmarshalBoundary(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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
		t.Fatalf("unexpected error: %v", err)
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
