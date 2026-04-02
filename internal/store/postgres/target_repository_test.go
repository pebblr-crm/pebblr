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

const (
	tgtTestExtID          = "EXT-001"
	tgtTestPharmacyA      = "Pharmacy A"
	tgtTestTeamID         = "team-1"
	tgtTestUserID         = "user-1"
	sqlSelectTarget       = "SELECT .+ FROM targets WHERE id"
	tgtErrUnexpected      = "unexpected error: %v"
	tgtTestMissing        = "tgt-missing"
	tgtErrExpectedError   = "expected error, got nil"
	sqlSelectTargets      = "SELECT .+ FROM targets"
	tgtErrExpectedTotal1  = "expected total 1, got %d"
	tgtTestDBError        = "db error"
	tgtTestDrSmith        = "Dr. Smith"
	sqlInsertTargets      = "INSERT INTO targets"
	tgtErrMarshalError    = "expected marshal error, got nil"
	tgtUpdatedPharmacy    = "Updated Pharmacy"
	sqlSelectVisitStatus  = "SELECT t.id::TEXT, MAX\\(a.due_date\\)"
	tgtErrExpected1Result = "expected 1 result, got %d"
	tgtErrExpectedEmpty   = "expected empty=false"
	queryCountTargets     = "SELECT COUNT\\(\\*\\) FROM targets"
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
		"tgt-1", tgtTestExtID, "pharmacy", tgtTestPharmacyA, fieldsJSON,
		tgtTestUserID, tgtTestTeamID,
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

	mock.ExpectQuery(sqlSelectTarget).
		WithArgs("tgt-1").
		WillReturnRows(targetRow(mock))

	tgt, err := repo.Get(ctx, "tgt-1")
	if err != nil {
		t.Fatalf(tgtErrUnexpected, err)
	}
	if tgt.ID != "tgt-1" {
		t.Errorf("expected ID tgt-1, got %s", tgt.ID)
	}
	if tgt.Name != tgtTestPharmacyA {
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

	mock.ExpectQuery(sqlSelectTarget).
		WithArgs(tgtTestMissing).
		WillReturnRows(emptyTargetRows(mock))

	_, err := repo.Get(ctx, tgtTestMissing)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestTargetGet_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery(sqlSelectTarget).
		WithArgs("tgt-1").
		WillReturnError(fmt.Errorf("connection refused"))

	_, err := repo.Get(ctx, "tgt-1")
	if err == nil {
		t.Fatal(tgtErrExpectedError)
	}
}

func TestTargetList_AllTargets(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()
	scope := rbac.TargetScope{AllTargets: true}

	mock.ExpectQuery(queryCountTargets).
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(1))

	// AllTargets, no filters: 2 pagination args
	mock.ExpectQuery(sqlSelectTargets).
		WithArgs(anyArgs(2)...).
		WillReturnRows(targetRow(mock))

	page, err := repo.List(ctx, scope, store.TargetFilter{}, 1, 20)
	if err != nil {
		t.Fatalf(tgtErrUnexpected, err)
	}
	if page.Total != 1 {
		t.Errorf(tgtErrExpectedTotal1, page.Total)
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
		t.Fatalf(tgtErrUnexpected, err)
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
	scope := rbac.TargetScope{AssigneeIDs: []string{tgtTestUserID}}

	mock.ExpectQuery(queryCountTargets).
		WithArgs(tgtTestUserID).
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(sqlSelectTargets).
		WithArgs(tgtTestUserID, 20, 0).
		WillReturnRows(targetRow(mock))

	page, err := repo.List(ctx, scope, store.TargetFilter{}, 1, 20)
	if err != nil {
		t.Fatalf(tgtErrUnexpected, err)
	}
	if page.Total != 1 {
		t.Errorf(tgtErrExpectedTotal1, page.Total)
	}
}

func TestTargetList_ScopedByTeamIDs(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()
	scope := rbac.TargetScope{TeamIDs: []string{tgtTestTeamID}}

	mock.ExpectQuery(queryCountTargets).
		WithArgs(tgtTestTeamID).
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(sqlSelectTargets).
		WithArgs(tgtTestTeamID, 20, 0).
		WillReturnRows(emptyTargetRows(mock))

	page, err := repo.List(ctx, scope, store.TargetFilter{}, 1, 20)
	if err != nil {
		t.Fatalf(tgtErrUnexpected, err)
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
		AssigneeID: strPtr(tgtTestUserID),
		TeamID:     strPtr(tgtTestTeamID),
		Query:      strPtr("Pharm"),
	}

	mock.ExpectQuery(queryCountTargets).
		WithArgs("pharmacy", tgtTestUserID, tgtTestTeamID, "%Pharm%").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(sqlSelectTargets).
		WithArgs("pharmacy", tgtTestUserID, tgtTestTeamID, "%Pharm%", 20, 0).
		WillReturnRows(targetRow(mock))

	page, err := repo.List(ctx, scope, filter, 1, 20)
	if err != nil {
		t.Fatalf(tgtErrUnexpected, err)
	}
	if page.Total != 1 {
		t.Errorf(tgtErrExpectedTotal1, page.Total)
	}
}

func TestTargetList_PaginationDefaults(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()
	scope := rbac.TargetScope{AllTargets: true}

	mock.ExpectQuery(queryCountTargets).
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(sqlSelectTargets).
		WithArgs(anyArgs(2)...).
		WillReturnRows(emptyTargetRows(mock))

	page, err := repo.List(ctx, scope, store.TargetFilter{}, -1, 500)
	if err != nil {
		t.Fatalf(tgtErrUnexpected, err)
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

	mock.ExpectQuery(queryCountTargets).
		WillReturnError(errors.New(tgtTestDBError))

	_, err := repo.List(ctx, scope, store.TargetFilter{}, 1, 20)
	if err == nil {
		t.Fatal(tgtErrExpectedError)
	}
}

func TestTargetList_QueryError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()
	scope := rbac.TargetScope{AllTargets: true}

	mock.ExpectQuery(queryCountTargets).
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(sqlSelectTargets).
		WithArgs(anyArgs(2)...).
		WillReturnError(errors.New(tgtTestDBError))

	_, err := repo.List(ctx, scope, store.TargetFilter{}, 1, 20)
	if err == nil {
		t.Fatal(tgtErrExpectedError)
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
		Name:       tgtTestDrSmith,
		Fields:     map[string]any{"specialty": "cardiology"},
		AssigneeID: tgtTestUserID,
		TeamID:     tgtTestTeamID,
	}

	now := testTime()
	fieldsJSON, _ := json.Marshal(map[string]any{"specialty": "cardiology"})
	// Create has 7 args
	mock.ExpectQuery(sqlInsertTargets).
		WithArgs(anyArgs(7)...).
		WillReturnRows(mock.NewRows([]string{
			"id", "external_id", "target_type", "name", "fields",
			"assignee_id", "team_id",
			"imported_at", "created_at", "updated_at",
		}).AddRow(
			"tgt-2", "EXT-002", "doctor", tgtTestDrSmith, fieldsJSON,
			tgtTestUserID, tgtTestTeamID,
			(*time.Time)(nil), now, now,
		))

	created, err := repo.Create(ctx, tgt)
	if err != nil {
		t.Fatalf(tgtErrUnexpected, err)
	}
	if created.ID != "tgt-2" {
		t.Errorf("expected ID tgt-2, got %s", created.ID)
	}
	if created.Name != tgtTestDrSmith {
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
		t.Fatal(tgtErrMarshalError)
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

	mock.ExpectQuery(sqlInsertTargets).
		WithArgs(anyArgs(7)...).
		WillReturnError(errors.New(tgtTestDBError))

	_, err := repo.Create(ctx, tgt)
	if err == nil {
		t.Fatal(tgtErrExpectedError)
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
		Name:       tgtUpdatedPharmacy,
		Fields:     map[string]any{"city": "Cluj"},
		AssigneeID: "user-2",
		TeamID:     tgtTestTeamID,
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
			"tgt-1", "", "pharmacy", tgtUpdatedPharmacy, fieldsJSON,
			"user-2", tgtTestTeamID,
			(*time.Time)(nil), now, now,
		))

	updated, err := repo.Update(ctx, tgt)
	if err != nil {
		t.Fatalf(tgtErrUnexpected, err)
	}
	if updated.Name != tgtUpdatedPharmacy {
		t.Errorf("expected name Updated Pharmacy, got %s", updated.Name)
	}
}

func TestTargetUpdate_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	tgt := &domain.Target{
		ID:     tgtTestMissing,
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
		t.Fatal(tgtErrMarshalError)
	}
}

func TestTargetUpsert_Empty(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	result, err := repo.Upsert(ctx, []*domain.Target{})
	if err != nil {
		t.Fatalf(tgtErrUnexpected, err)
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
			ExternalID: tgtTestExtID,
			TargetType: "pharmacy",
			Name:       tgtTestPharmacyA,
			Fields:     map[string]any{"city": "Bucharest"},
			AssigneeID: tgtTestUserID,
			TeamID:     tgtTestTeamID,
		},
	}

	fieldsJSON, _ := json.Marshal(map[string]any{"city": "Bucharest"})

	mock.ExpectBegin()
	// Upsert has 6 args
	mock.ExpectQuery(sqlInsertTargets).
		WithArgs(anyArgs(6)...).
		WillReturnRows(mock.NewRows([]string{
			"id", "external_id", "target_type", "name", "fields",
			"assignee_id", "team_id",
			"imported_at", "created_at", "updated_at",
			"is_new",
		}).AddRow(
			"tgt-1", tgtTestExtID, "pharmacy", tgtTestPharmacyA, fieldsJSON,
			tgtTestUserID, tgtTestTeamID,
			&now, now, now,
			true,
		))
	mock.ExpectCommit()

	result, err := repo.Upsert(ctx, targets)
	if err != nil {
		t.Fatalf(tgtErrUnexpected, err)
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
		t.Fatal(tgtErrExpectedError)
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
			ExternalID: tgtTestExtID,
			TargetType: "pharmacy",
			Name:       tgtTestPharmacyA,
			Fields:     map[string]any{},
		},
	}

	fieldsJSON, _ := json.Marshal(map[string]any{})

	mock.ExpectBegin()
	mock.ExpectQuery(sqlInsertTargets).
		WithArgs(anyArgs(6)...).
		WillReturnRows(mock.NewRows([]string{
			"id", "external_id", "target_type", "name", "fields",
			"assignee_id", "team_id",
			"imported_at", "created_at", "updated_at",
			"is_new",
		}).AddRow(
			"tgt-1", tgtTestExtID, "pharmacy", tgtTestPharmacyA, fieldsJSON,
			"", "",
			&now, now, now,
			true,
		))
	mock.ExpectCommit().WillReturnError(fmt.Errorf("commit failed"))
	mock.ExpectRollback()

	_, err := repo.Upsert(ctx, targets)
	if err == nil {
		t.Fatal(tgtErrExpectedError)
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
		t.Fatal(tgtErrMarshalError)
	}
}

func TestTargetVisitStatus_EmptyFieldTypes(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	result, err := repo.VisitStatus(ctx, rbac.TargetScope{AllTargets: true}, []string{})
	if err != nil {
		t.Fatalf(tgtErrUnexpected, err)
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
		t.Fatalf(tgtErrUnexpected, err)
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

	mock.ExpectQuery(sqlSelectVisitStatus).
		WithArgs([]string{"visit"}).
		WillReturnRows(mock.NewRows([]string{"target_id", "last_visit_date"}).
			AddRow("tgt-1", now))

	result, err := repo.VisitStatus(ctx, scope, []string{"visit"})
	if err != nil {
		t.Fatalf(tgtErrUnexpected, err)
	}
	if len(result) != 1 {
		t.Fatalf(tgtErrExpected1Result, len(result))
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
	scope := rbac.TargetScope{AssigneeIDs: []string{tgtTestUserID}}

	// 1 scope arg + 1 fieldTypes arg = 2
	mock.ExpectQuery(sqlSelectVisitStatus).
		WithArgs(anyArgs(2)...).
		WillReturnRows(mock.NewRows([]string{"target_id", "last_visit_date"}).
			AddRow("tgt-1", now))

	result, err := repo.VisitStatus(ctx, scope, []string{"visit"})
	if err != nil {
		t.Fatalf(tgtErrUnexpected, err)
	}
	if len(result) != 1 {
		t.Fatalf(tgtErrExpected1Result, len(result))
	}
}

func TestTargetVisitStatus_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()
	scope := rbac.TargetScope{AllTargets: true}

	mock.ExpectQuery(sqlSelectVisitStatus).
		WithArgs([]string{"visit"}).
		WillReturnError(errors.New(tgtTestDBError))

	_, err := repo.VisitStatus(ctx, scope, []string{"visit"})
	if err == nil {
		t.Fatal(tgtErrExpectedError)
	}
}

func TestTargetFrequencyStatus_EmptyFieldTypes(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newTargetRepo(mock)
	ctx := context.Background()

	result, err := repo.FrequencyStatus(ctx, rbac.TargetScope{AllTargets: true}, []string{}, testTime(), testTime())
	if err != nil {
		t.Fatalf(tgtErrUnexpected, err)
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
		t.Fatalf(tgtErrUnexpected, err)
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
		t.Fatalf(tgtErrUnexpected, err)
	}
	if len(result) != 1 {
		t.Fatalf(tgtErrExpected1Result, len(result))
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
		WillReturnError(errors.New(tgtTestDBError))

	_, err := repo.FrequencyStatus(ctx, scope, []string{"visit"}, testTime(), testTime())
	if err == nil {
		t.Fatal(tgtErrExpectedError)
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
		t.Error(tgtErrExpectedEmpty)
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
		t.Error(tgtErrExpectedEmpty)
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
		t.Error(tgtErrExpectedEmpty)
	}
	// The shared scope builder produces a single combined "(assignee OR team)" condition.
	if len(result.conditions) != 1 {
		t.Errorf("expected 1 combined condition, got %d", len(result.conditions))
	}
	if len(result.args) != 2 {
		t.Errorf("expected 2 args, got %d", len(result.args))
	}
	if result.argIdx != 3 {
		t.Errorf("expected argIdx 3, got %d", result.argIdx)
	}
}
