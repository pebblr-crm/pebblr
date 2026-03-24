package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/rbac"
	"github.com/pebblr/pebblr/internal/store"
)

func targetRow(mock pgxmock.PgxPoolIface) *pgxmock.Rows {
	now := testTime()
	fields := map[string]any{"city": "Bucharest"}
	fieldsJSON, _ := json.Marshal(fields)

	return mock.NewRows([]string{
		"id", "external_id", "target_type", "name", "fields",
		"assignee_id", "team_id",
		"imported_at", "created_at", "updated_at",
	}).AddRow(
		"tgt-1", "EXT-001", "pharmacy", "Pharmacy A", fieldsJSON,
		"user-1", "team-1",
		(*time.Time)(nil), now, now,
	)
}

func emptyTargetRows(mock pgxmock.PgxPoolIface) *pgxmock.Rows {
	return mock.NewRows([]string{
		"id", "external_id", "target_type", "name", "fields",
		"assignee_id", "team_id",
		"imported_at", "created_at", "updated_at",
	})
}

func newTargetRepo(pool pgxmock.PgxPoolIface) *targetRepository {
	return &targetRepository{pool: pool}
}

func TestTargetGet_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM targets WHERE id").
		WithArgs("tgt-1").
		WillReturnRows(targetRow(mock))

	tgt, err := repo.Get(ctx, "tgt-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tgt.ID != "tgt-1" {
		t.Errorf("expected ID tgt-1, got %s", tgt.ID)
	}
	if tgt.Name != "Pharmacy A" {
		t.Errorf("expected name Pharmacy A, got %s", tgt.Name)
	}
	if tgt.Fields["city"] != "Bucharest" {
		t.Errorf("expected fields.city=Bucharest, got %v", tgt.Fields["city"])
	}
}

func TestTargetGet_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM targets WHERE id").
		WithArgs("tgt-missing").
		WillReturnRows(emptyTargetRows(mock))

	_, err := repo.Get(ctx, "tgt-missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestTargetGet_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM targets WHERE id").
		WithArgs("tgt-1").
		WillReturnError(fmt.Errorf("connection refused"))

	_, err := repo.Get(ctx, "tgt-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTargetList_AllTargets(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()
	scope := rbac.TargetScope{AllTargets: true}

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM targets").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(1))

	// AllTargets, no filters: 2 pagination args
	mock.ExpectQuery("SELECT .+ FROM targets").
		WithArgs(anyArgs(2)...).
		WillReturnRows(targetRow(mock))

	page, err := repo.List(ctx, scope, store.TargetFilter{}, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 1 {
		t.Errorf("expected total 1, got %d", page.Total)
	}
	if len(page.Targets) != 1 {
		t.Errorf("expected 1 target, got %d", len(page.Targets))
	}
}

func TestTargetList_EmptyScope(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	// Empty scope should return empty immediately without querying.
	scope := rbac.TargetScope{}

	page, err := repo.List(ctx, scope, store.TargetFilter{}, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 0 {
		t.Errorf("expected total 0, got %d", page.Total)
	}
}

func TestTargetList_ScopedByAssigneeIDs(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()
	scope := rbac.TargetScope{AssigneeIDs: []string{"user-1"}}

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM targets").
		WithArgs("user-1").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery("SELECT .+ FROM targets").
		WithArgs("user-1", 20, 0).
		WillReturnRows(targetRow(mock))

	page, err := repo.List(ctx, scope, store.TargetFilter{}, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 1 {
		t.Errorf("expected total 1, got %d", page.Total)
	}
}

func TestTargetList_ScopedByTeamIDs(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()
	scope := rbac.TargetScope{TeamIDs: []string{"team-1"}}

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM targets").
		WithArgs("team-1").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("SELECT .+ FROM targets").
		WithArgs("team-1", 20, 0).
		WillReturnRows(emptyTargetRows(mock))

	page, err := repo.List(ctx, scope, store.TargetFilter{}, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 0 {
		t.Errorf("expected total 0, got %d", page.Total)
	}
}

func TestTargetList_WithFilters(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()
	scope := rbac.TargetScope{AllTargets: true}

	filter := store.TargetFilter{
		TargetType: strPtr("pharmacy"),
		AssigneeID: strPtr("user-1"),
		TeamID:     strPtr("team-1"),
		Query:      strPtr("Pharm"),
	}

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM targets").
		WithArgs("pharmacy", "user-1", "team-1", "%Pharm%").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery("SELECT .+ FROM targets").
		WithArgs("pharmacy", "user-1", "team-1", "%Pharm%", 20, 0).
		WillReturnRows(targetRow(mock))

	page, err := repo.List(ctx, scope, filter, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 1 {
		t.Errorf("expected total 1, got %d", page.Total)
	}
}

func TestTargetList_PaginationDefaults(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()
	scope := rbac.TargetScope{AllTargets: true}

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM targets").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("SELECT .+ FROM targets").
		WithArgs(anyArgs(2)...).
		WillReturnRows(emptyTargetRows(mock))

	page, err := repo.List(ctx, scope, store.TargetFilter{}, -1, 500)
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

func TestTargetList_CountError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()
	scope := rbac.TargetScope{AllTargets: true}

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM targets").
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.List(ctx, scope, store.TargetFilter{}, 1, 20)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTargetList_QueryError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()
	scope := rbac.TargetScope{AllTargets: true}

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM targets").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery("SELECT .+ FROM targets").
		WithArgs(anyArgs(2)...).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.List(ctx, scope, store.TargetFilter{}, 1, 20)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTargetCreate_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	tgt := &domain.Target{
		ExternalID: "EXT-002",
		TargetType: "doctor",
		Name:       "Dr. Smith",
		Fields:     map[string]any{"specialty": "cardiology"},
		AssigneeID: "user-1",
		TeamID:     "team-1",
	}

	now := testTime()
	fieldsJSON, _ := json.Marshal(map[string]any{"specialty": "cardiology"})
	// Create has 7 args
	mock.ExpectQuery("INSERT INTO targets").
		WithArgs(anyArgs(7)...).
		WillReturnRows(mock.NewRows([]string{
			"id", "external_id", "target_type", "name", "fields",
			"assignee_id", "team_id",
			"imported_at", "created_at", "updated_at",
		}).AddRow(
			"tgt-2", "EXT-002", "doctor", "Dr. Smith", fieldsJSON,
			"user-1", "team-1",
			(*time.Time)(nil), now, now,
		))

	created, err := repo.Create(ctx, tgt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != "tgt-2" {
		t.Errorf("expected ID tgt-2, got %s", created.ID)
	}
	if created.Name != "Dr. Smith" {
		t.Errorf("expected name Dr. Smith, got %s", created.Name)
	}
}

func TestTargetCreate_MarshalError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	tgt := &domain.Target{
		Fields: map[string]any{"bad": make(chan int)},
	}

	_, err := repo.Create(ctx, tgt)
	if err == nil {
		t.Fatal("expected marshal error, got nil")
	}
}

func TestTargetCreate_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	tgt := &domain.Target{
		Fields: map[string]any{},
	}

	mock.ExpectQuery("INSERT INTO targets").
		WithArgs(anyArgs(7)...).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.Create(ctx, tgt)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTargetUpdate_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	tgt := &domain.Target{
		ID:         "tgt-1",
		TargetType: "pharmacy",
		Name:       "Updated Pharmacy",
		Fields:     map[string]any{"city": "Cluj"},
		AssigneeID: "user-2",
		TeamID:     "team-1",
	}

	now := testTime()
	fieldsJSON, _ := json.Marshal(map[string]any{"city": "Cluj"})
	// Update has 6 args
	mock.ExpectQuery("UPDATE targets").
		WithArgs(anyArgs(6)...).
		WillReturnRows(mock.NewRows([]string{
			"id", "external_id", "target_type", "name", "fields",
			"assignee_id", "team_id",
			"imported_at", "created_at", "updated_at",
		}).AddRow(
			"tgt-1", "", "pharmacy", "Updated Pharmacy", fieldsJSON,
			"user-2", "team-1",
			(*time.Time)(nil), now, now,
		))

	updated, err := repo.Update(ctx, tgt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "Updated Pharmacy" {
		t.Errorf("expected name Updated Pharmacy, got %s", updated.Name)
	}
}

func TestTargetUpdate_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	tgt := &domain.Target{
		ID:     "tgt-missing",
		Fields: map[string]any{},
	}

	// Update has 6 args
	mock.ExpectQuery("UPDATE targets").
		WithArgs(anyArgs(6)...).
		WillReturnRows(emptyTargetRows(mock))

	_, err := repo.Update(ctx, tgt)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestTargetUpdate_MarshalError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	tgt := &domain.Target{
		ID:     "tgt-1",
		Fields: map[string]any{"bad": make(chan int)},
	}

	_, err := repo.Update(ctx, tgt)
	if err == nil {
		t.Fatal("expected marshal error, got nil")
	}
}

func TestTargetUpsert_Empty(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	result, err := repo.Upsert(ctx, []*domain.Target{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Created != 0 || result.Updated != 0 {
		t.Errorf("expected 0/0, got %d/%d", result.Created, result.Updated)
	}
}

func TestTargetUpsert_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()
	now := testTime()

	targets := []*domain.Target{
		{
			ExternalID: "EXT-001",
			TargetType: "pharmacy",
			Name:       "Pharmacy A",
			Fields:     map[string]any{"city": "Bucharest"},
			AssigneeID: "user-1",
			TeamID:     "team-1",
		},
	}

	fieldsJSON, _ := json.Marshal(map[string]any{"city": "Bucharest"})

	mock.ExpectBegin()
	// Upsert has 6 args
	mock.ExpectQuery("INSERT INTO targets").
		WithArgs(anyArgs(6)...).
		WillReturnRows(mock.NewRows([]string{
			"id", "external_id", "target_type", "name", "fields",
			"assignee_id", "team_id",
			"imported_at", "created_at", "updated_at",
			"is_new",
		}).AddRow(
			"tgt-1", "EXT-001", "pharmacy", "Pharmacy A", fieldsJSON,
			"user-1", "team-1",
			&now, now, now,
			true,
		))
	mock.ExpectCommit()

	result, err := repo.Upsert(ctx, targets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Created != 1 {
		t.Errorf("expected 1 created, got %d", result.Created)
	}
	if result.Updated != 0 {
		t.Errorf("expected 0 updated, got %d", result.Updated)
	}
	if len(result.Imported) != 1 {
		t.Errorf("expected 1 imported, got %d", len(result.Imported))
	}
}

func TestTargetUpsert_BeginError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	mock.ExpectBegin().WillReturnError(fmt.Errorf("begin failed"))

	_, err := repo.Upsert(ctx, []*domain.Target{{Fields: map[string]any{}}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTargetUpsert_CommitError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()
	now := testTime()

	targets := []*domain.Target{
		{
			ExternalID: "EXT-001",
			TargetType: "pharmacy",
			Name:       "Pharmacy A",
			Fields:     map[string]any{},
		},
	}

	fieldsJSON, _ := json.Marshal(map[string]any{})

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO targets").
		WithArgs(anyArgs(6)...).
		WillReturnRows(mock.NewRows([]string{
			"id", "external_id", "target_type", "name", "fields",
			"assignee_id", "team_id",
			"imported_at", "created_at", "updated_at",
			"is_new",
		}).AddRow(
			"tgt-1", "EXT-001", "pharmacy", "Pharmacy A", fieldsJSON,
			"", "",
			&now, now, now,
			true,
		))
	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit failed"))
	mock.ExpectRollback()

	_, err := repo.Upsert(ctx, targets)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTargetUpsert_MarshalError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	targets := []*domain.Target{
		{Fields: map[string]any{"bad": make(chan int)}},
	}

	mock.ExpectBegin()
	mock.ExpectRollback()

	_, err := repo.Upsert(ctx, targets)
	if err == nil {
		t.Fatal("expected marshal error, got nil")
	}
}

func TestTargetVisitStatus_EmptyFieldTypes(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	result, err := repo.VisitStatus(ctx, rbac.TargetScope{AllTargets: true}, []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty field types, got %v", result)
	}
}

func TestTargetVisitStatus_EmptyScope(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	result, err := repo.VisitStatus(ctx, rbac.TargetScope{}, []string{"visit"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for empty scope, got %v", result)
	}
}

func TestTargetVisitStatus_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()
	now := testTime()
	scope := rbac.TargetScope{AllTargets: true}

	mock.ExpectQuery("SELECT t.id::TEXT, MAX\\(a.due_date\\)").
		WithArgs([]string{"visit"}).
		WillReturnRows(mock.NewRows([]string{"target_id", "last_visit_date"}).
			AddRow("tgt-1", now))

	result, err := repo.VisitStatus(ctx, scope, []string{"visit"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].TargetID != "tgt-1" {
		t.Errorf("expected tgt-1, got %s", result[0].TargetID)
	}
}

func TestTargetVisitStatus_WithScopedAssignees(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()
	now := testTime()
	scope := rbac.TargetScope{AssigneeIDs: []string{"user-1"}}

	// 1 scope arg + 1 fieldTypes arg = 2
	mock.ExpectQuery("SELECT t.id::TEXT, MAX\\(a.due_date\\)").
		WithArgs(anyArgs(2)...).
		WillReturnRows(mock.NewRows([]string{"target_id", "last_visit_date"}).
			AddRow("tgt-1", now))

	result, err := repo.VisitStatus(ctx, scope, []string{"visit"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
}

func TestTargetVisitStatus_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()
	scope := rbac.TargetScope{AllTargets: true}

	mock.ExpectQuery("SELECT t.id::TEXT, MAX\\(a.due_date\\)").
		WithArgs([]string{"visit"}).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.VisitStatus(ctx, scope, []string{"visit"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestTargetFrequencyStatus_EmptyFieldTypes(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	result, err := repo.FrequencyStatus(ctx, rbac.TargetScope{AllTargets: true}, []string{}, testTime(), testTime())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestTargetFrequencyStatus_EmptyScope(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	result, err := repo.FrequencyStatus(ctx, rbac.TargetScope{}, []string{"visit"}, testTime(), testTime())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestTargetFrequencyStatus_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()
	scope := rbac.TargetScope{AllTargets: true}
	dateFrom := testTime()
	dateTo := testTime().Add(30 * 24 * time.Hour)

	mock.ExpectQuery("SELECT t.id::TEXT, t.fields").
		WithArgs([]string{"visit"}, dateFrom, dateTo).
		WillReturnRows(mock.NewRows([]string{"target_id", "classification", "visit_count"}).
			AddRow("tgt-1", "A", 5))

	result, err := repo.FrequencyStatus(ctx, scope, []string{"visit"}, dateFrom, dateTo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].Classification != "A" {
		t.Errorf("expected classification A, got %s", result[0].Classification)
	}
	if result[0].VisitCount != 5 {
		t.Errorf("expected visit count 5, got %d", result[0].VisitCount)
	}
}

func TestTargetFrequencyStatus_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()
	scope := rbac.TargetScope{AllTargets: true}

	mock.ExpectQuery("SELECT t.id::TEXT, t.fields").
		WithArgs([]string{"visit"}, testTime(), testTime()).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.FrequencyStatus(ctx, scope, []string{"visit"}, testTime(), testTime())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestBuildTargetScopeConditions_AllTargets(t *testing.T) {
	t.Parallel()
	result := buildTargetScopeConditions(rbac.TargetScope{AllTargets: true}, "t.", 1)
	if len(result.conditions) != 0 {
		t.Errorf("expected no conditions, got %v", result.conditions)
	}
	if result.empty {
		t.Error("expected empty=false for AllTargets")
	}
}

func TestBuildTargetScopeConditions_EmptyScope(t *testing.T) {
	t.Parallel()
	result := buildTargetScopeConditions(rbac.TargetScope{}, "t.", 1)
	if !result.empty {
		t.Error("expected empty=true for no assignees and no teams")
	}
}

func TestBuildTargetScopeConditions_WithAssignees(t *testing.T) {
	t.Parallel()
	scope := rbac.TargetScope{AssigneeIDs: []string{"u1", "u2"}}
	result := buildTargetScopeConditions(scope, "t.", 1)
	if result.empty {
		t.Error("expected empty=false")
	}
	if len(result.conditions) != 1 {
		t.Errorf("expected 1 condition, got %d", len(result.conditions))
	}
	if len(result.args) != 2 {
		t.Errorf("expected 2 args, got %d", len(result.args))
	}
	if result.argIdx != 3 {
		t.Errorf("expected argIdx 3, got %d", result.argIdx)
	}
}

func TestBuildTargetScopeConditions_WithTeams(t *testing.T) {
	t.Parallel()
	scope := rbac.TargetScope{TeamIDs: []string{"t1"}}
	result := buildTargetScopeConditions(scope, "", 5)
	if result.empty {
		t.Error("expected empty=false")
	}
	if result.argIdx != 6 {
		t.Errorf("expected argIdx 6, got %d", result.argIdx)
	}
}

func TestBuildTargetScopeConditions_AssigneesAndTeams(t *testing.T) {
	t.Parallel()
	scope := rbac.TargetScope{AssigneeIDs: []string{"u1"}, TeamIDs: []string{"t1"}}
	result := buildTargetScopeConditions(scope, "t.", 1)
	if result.empty {
		t.Error("expected empty=false")
	}
	if len(result.conditions) != 2 {
		t.Errorf("expected 2 conditions, got %d", len(result.conditions))
	}
	if len(result.args) != 2 {
		t.Errorf("expected 2 args, got %d", len(result.args))
	}
	if result.argIdx != 3 {
		t.Errorf("expected argIdx 3, got %d", result.argIdx)
	}
}
