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

func activityRow(mock pgxmock.PgxPoolIface) *pgxmock.Rows {
	now := testTime()
	fields := map[string]any{"notes": "test"}
	fieldsJSON, _ := json.Marshal(fields)
	targetFields := map[string]any{"city": "Bucharest"}
	targetFieldsJSON, _ := json.Marshal(targetFields)

	return mock.NewRows([]string{
		"id", "activity_type", "label", "status", "due_date", "duration",
		"routing", "fields",
		"target_id", "target_name", "creator_id",
		"joint_visit_user_id", "team_id",
		"submitted_at", "created_at", "updated_at", "deleted_at",
		"target_type", "target_fields",
	}).AddRow(
		"act-1", "visit", "Morning visit", "planned", now, "full_day",
		"week-1", fieldsJSON,
		"tgt-1", "Pharmacy A", "user-1",
		"", "team-1",
		(*time.Time)(nil), now, now, (*time.Time)(nil),
		"pharmacy", targetFieldsJSON,
	)
}

func emptyActivityRows(mock pgxmock.PgxPoolIface) *pgxmock.Rows {
	return mock.NewRows([]string{
		"id", "activity_type", "label", "status", "due_date", "duration",
		"routing", "fields",
		"target_id", "target_name", "creator_id",
		"joint_visit_user_id", "team_id",
		"submitted_at", "created_at", "updated_at", "deleted_at",
		"target_type", "target_fields",
	})
}

func newActivityRepo(pool pgxmock.PgxPoolIface) *activityRepository {
	return &activityRepository{pool: pool}
}

func TestActivityGet_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM activities a LEFT JOIN targets t").
		WithArgs("act-1").
		WillReturnRows(activityRow(mock))

	a, err := repo.Get(ctx, "act-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.ID != "act-1" {
		t.Errorf("expected ID act-1, got %s", a.ID)
	}
	if a.ActivityType != "visit" {
		t.Errorf("expected activity_type visit, got %s", a.ActivityType)
	}
	if a.TargetSummary == nil {
		t.Fatal("expected TargetSummary to be populated")
	}
	if a.TargetSummary.Name != "Pharmacy A" {
		t.Errorf("expected target name Pharmacy A, got %s", a.TargetSummary.Name)
	}
	if a.Fields["notes"] != "test" {
		t.Errorf("expected fields.notes=test, got %v", a.Fields["notes"])
	}
}

func TestActivityGet_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM activities a LEFT JOIN targets t").
		WithArgs("act-missing").
		WillReturnRows(emptyActivityRows(mock))

	_, err := repo.Get(ctx, "act-missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestActivityGet_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM activities a LEFT JOIN targets t").
		WithArgs("act-1").
		WillReturnError(fmt.Errorf("connection refused"))

	_, err := repo.Get(ctx, "act-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestActivityList_AllActivities(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}

	// All scope, no filters: count query has 0 args, list query has 2 (limit, offset)
	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery("SELECT .+ FROM activities a LEFT JOIN targets t").
		WithArgs(anyArgs(2)...).
		WillReturnRows(activityRow(mock))

	page, err := repo.List(ctx, scope, store.ActivityFilter{}, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 1 {
		t.Errorf("expected total 1, got %d", page.Total)
	}
	if len(page.Activities) != 1 {
		t.Errorf("expected 1 activity, got %d", len(page.Activities))
	}
}

func TestActivityList_EmptyScope(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()

	// Empty scope (not AllActivities, no CreatorIDs, no TeamIDs) should return empty immediately.
	scope := rbac.ActivityScope{}

	page, err := repo.List(ctx, scope, store.ActivityFilter{}, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 0 {
		t.Errorf("expected total 0, got %d", page.Total)
	}
	if len(page.Activities) != 0 {
		t.Errorf("expected 0 activities, got %d", len(page.Activities))
	}
}

func TestActivityList_ScopedByCreatorIDs(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{CreatorIDs: []string{"user-1"}}

	// 1 scope arg for count; 1 scope arg + 2 pagination for list
	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
		WithArgs("user-1").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery("SELECT .+ FROM activities a LEFT JOIN targets t").
		WithArgs("user-1", 20, 0).
		WillReturnRows(activityRow(mock))

	page, err := repo.List(ctx, scope, store.ActivityFilter{}, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 1 {
		t.Errorf("expected total 1, got %d", page.Total)
	}
}

func TestActivityList_ScopedByTeamIDs(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{TeamIDs: []string{"team-1"}}

	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
		WithArgs("team-1").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("SELECT .+ FROM activities a LEFT JOIN targets t").
		WithArgs("team-1", 20, 0).
		WillReturnRows(emptyActivityRows(mock))

	page, err := repo.List(ctx, scope, store.ActivityFilter{}, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 0 {
		t.Errorf("expected total 0, got %d", page.Total)
	}
}

func TestActivityList_WithFilters(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}
	dateFrom := testTime()
	dateTo := testTime().Add(24 * time.Hour)

	filter := store.ActivityFilter{
		ActivityType: strPtr("visit"),
		Status:       strPtr("planned"),
		CreatorID:    strPtr("user-1"),
		TargetID:     strPtr("tgt-1"),
		TeamID:       strPtr("team-1"),
		DateFrom:     &dateFrom,
		DateTo:       &dateTo,
	}

	// 7 filter args for count
	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
		WithArgs("visit", "planned", "user-1", "tgt-1", "team-1", dateFrom, dateTo).
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(1))

	// 7 filter args + 2 pagination for list
	mock.ExpectQuery("SELECT .+ FROM activities a LEFT JOIN targets t").
		WithArgs("visit", "planned", "user-1", "tgt-1", "team-1", dateFrom, dateTo, 20, 0).
		WillReturnRows(activityRow(mock))

	page, err := repo.List(ctx, scope, filter, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 1 {
		t.Errorf("expected total 1, got %d", page.Total)
	}
}

func TestActivityList_PaginationDefaults(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}

	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(0))

	// page=0 and limit=0 should be normalized to page=1, limit=20
	mock.ExpectQuery("SELECT .+ FROM activities a LEFT JOIN targets t").
		WithArgs(anyArgs(2)...).
		WillReturnRows(emptyActivityRows(mock))

	page, err := repo.List(ctx, scope, store.ActivityFilter{}, 0, 0)
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

func TestActivityList_LimitClamp(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}

	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("SELECT .+ FROM activities a LEFT JOIN targets t").
		WithArgs(anyArgs(2)...).
		WillReturnRows(emptyActivityRows(mock))

	// limit > 200 should be clamped to 20
	page, err := repo.List(ctx, scope, store.ActivityFilter{}, 1, 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Limit != 20 {
		t.Errorf("expected limit 20, got %d", page.Limit)
	}
}

func TestActivityList_CountError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}

	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.List(ctx, scope, store.ActivityFilter{}, 1, 20)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestActivityList_QueryError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{AllActivities: true}

	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery("SELECT .+ FROM activities a LEFT JOIN targets t").
		WithArgs(anyArgs(2)...).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.List(ctx, scope, store.ActivityFilter{}, 1, 20)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestActivityCreate_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()
	now := testTime()

	a := &domain.Activity{
		ActivityType: "visit",
		Label:        "Morning",
		Status:       "planned",
		DueDate:      now,
		Duration:     "full_day",
		Routing:      "week-1",
		Fields:       map[string]any{"notes": "test"},
		TargetID:     "tgt-1",
		CreatorID:    "user-1",
		TeamID:       "team-1",
	}

	// Create has 12 args
	mock.ExpectQuery("WITH ins AS").
		WithArgs(anyArgs(12)...).
		WillReturnRows(activityRow(mock))

	created, err := repo.Create(ctx, a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.ID != "act-1" {
		t.Errorf("expected ID act-1, got %s", created.ID)
	}
}

func TestActivityCreate_MarshalError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()

	// Fields with a channel value cannot be marshalled.
	a := &domain.Activity{
		Fields: map[string]any{"bad": make(chan int)},
	}

	_, err := repo.Create(ctx, a)
	if err == nil {
		t.Fatal("expected marshal error, got nil")
	}
}

func TestActivityCreate_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()

	a := &domain.Activity{
		Fields: map[string]any{},
	}

	mock.ExpectQuery("WITH ins AS").
		WithArgs(anyArgs(12)...).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.Create(ctx, a)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestActivityUpdate_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()
	now := testTime()

	a := &domain.Activity{
		ID:           "act-1",
		ActivityType: "visit",
		Status:       "completed",
		DueDate:      now,
		Duration:     "full_day",
		Fields:       map[string]any{"notes": "updated"},
		TargetID:     "tgt-1",
		CreatorID:    "user-1",
		TeamID:       "team-1",
	}

	// Update has 12 args
	mock.ExpectQuery("WITH upd AS").
		WithArgs(anyArgs(12)...).
		WillReturnRows(activityRow(mock))

	updated, err := repo.Update(ctx, a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.ID != "act-1" {
		t.Errorf("expected ID act-1, got %s", updated.ID)
	}
}

func TestActivityUpdate_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()

	a := &domain.Activity{
		ID:     "act-missing",
		Fields: map[string]any{},
	}

	mock.ExpectQuery("WITH upd AS").
		WithArgs(anyArgs(12)...).
		WillReturnRows(emptyActivityRows(mock))

	_, err := repo.Update(ctx, a)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestActivityUpdate_MarshalError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()

	a := &domain.Activity{
		ID:     "act-1",
		Fields: map[string]any{"bad": make(chan int)},
	}

	_, err := repo.Update(ctx, a)
	if err == nil {
		t.Fatal("expected marshal error, got nil")
	}
}

func TestActivitySoftDelete_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()

	mock.ExpectExec("UPDATE activities SET deleted_at").
		WithArgs("act-1").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := repo.SoftDelete(ctx, "act-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestActivitySoftDelete_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()

	mock.ExpectExec("UPDATE activities SET deleted_at").
		WithArgs("act-missing").
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err := repo.SoftDelete(ctx, "act-missing")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestActivitySoftDelete_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()

	mock.ExpectExec("UPDATE activities SET deleted_at").
		WithArgs("act-1").
		WillReturnError(fmt.Errorf("db error"))

	err := repo.SoftDelete(ctx, "act-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestActivityCountByDate_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()
	date := testTime()

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM activities").
		WithArgs("user-1", date).
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(3))

	count, err := repo.CountByDate(ctx, "user-1", date)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3, got %d", count)
	}
}

func TestActivityCountByDate_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()
	date := testTime()

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM activities").
		WithArgs("user-1", date).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.CountByDate(ctx, "user-1", date)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestActivityHasActivityWithTypes_EmptyTypes(t *testing.T) {
	t.Parallel()
	// Should return false immediately without querying.
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()

	has, err := repo.HasActivityWithTypes(ctx, "user-1", testTime(), []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if has {
		t.Error("expected false for empty types")
	}
}

func TestActivityHasActivityWithTypes_True(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()
	date := testTime()

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("user-1", date, []string{"visit", "call"}).
		WillReturnRows(mock.NewRows([]string{"exists"}).AddRow(true))

	has, err := repo.HasActivityWithTypes(ctx, "user-1", date, []string{"visit", "call"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !has {
		t.Error("expected true")
	}
}

func TestActivityHasActivityWithTypes_False(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()
	date := testTime()

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("user-1", date, []string{"visit"}).
		WillReturnRows(mock.NewRows([]string{"exists"}).AddRow(false))

	has, err := repo.HasActivityWithTypes(ctx, "user-1", date, []string{"visit"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if has {
		t.Error("expected false")
	}
}

func TestActivityHasActivityWithTypes_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs("user-1", testTime(), []string{"visit"}).
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.HasActivityWithTypes(ctx, "user-1", testTime(), []string{"visit"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestScanActivity_NoTarget(t *testing.T) {
	t.Parallel()
	// When target_id is empty, TargetSummary should be nil.
	mock := newMockPool(t)
	now := testTime()
	fieldsJSON, _ := json.Marshal(map[string]any{})

	rows := mock.NewRows([]string{
		"id", "activity_type", "label", "status", "due_date", "duration",
		"routing", "fields",
		"target_id", "target_name", "creator_id",
		"joint_visit_user_id", "team_id",
		"submitted_at", "created_at", "updated_at", "deleted_at",
		"target_type", "target_fields",
	}).AddRow(
		"act-2", "admin", "", "planned", now, "half_day",
		"", fieldsJSON,
		"", "", "user-1",
		"", "team-1",
		(*time.Time)(nil), now, now, (*time.Time)(nil),
		"", []byte("{}"),
	)

	mock.ExpectQuery("SELECT").
		WithArgs(anyArgs(1)...).
		WillReturnRows(rows)
	row := mock.QueryRow(context.Background(), "SELECT", "dummy")
	a, err := scanActivity(row)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.TargetSummary != nil {
		t.Error("expected TargetSummary to be nil when target_id is empty")
	}
}

func TestActivityList_CombinedScopeCreatorsAndTeams(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newActivityRepo(mock)
	ctx := context.Background()
	scope := rbac.ActivityScope{
		CreatorIDs: []string{"user-1"},
		TeamIDs:    []string{"team-1"},
	}

	// 2 scope args for count; 2 scope args + 2 pagination for list
	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
		WithArgs("user-1", "team-1").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery("SELECT .+ FROM activities a LEFT JOIN targets t").
		WithArgs("user-1", "team-1", 20, 0).
		WillReturnRows(emptyActivityRows(mock))

	page, err := repo.List(ctx, scope, store.ActivityFilter{}, 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.Total != 0 {
		t.Errorf("expected total 0, got %d", page.Total)
	}
}
