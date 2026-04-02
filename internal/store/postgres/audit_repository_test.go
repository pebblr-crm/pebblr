package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

const (
	testAuditID       = "audit-1"
	queryInsertAudit  = "INSERT INTO audit_log"
	errDBMsg          = "db error"
	msgExpectedErr    = "expected error, got nil"
	querySelectAudit  = "SELECT .+ FROM audit_log"
	fmtExpected1Entry = "expected 1 entry, got %d"
	auditSelectCount  = "SELECT COUNT\\(\\*\\)"
	queryUpdateStatus = "UPDATE audit_log SET status"
	testAdminID       = "admin-1"
)

func auditColumns() []string {
	return []string{
		"id", "entity_type", "entity_id", "event_type", "actor_id",
		"old_value", "new_value", "status", "reviewed_by", "reviewed_at", "created_at",
	}
}

func auditRow(mock pgxmock.PgxPoolIface) *pgxmock.Rows {
	now := testTime()
	oldJSON, _ := json.Marshal(map[string]any{"status": "planned"})
	newJSON, _ := json.Marshal(map[string]any{"status": "completed"})
	return mock.NewRows(auditColumns()).AddRow(
		testAuditID, "activity", "act-1", "status_changed", "rep-1",
		oldJSON, newJSON, "pending", (*string)(nil), (*time.Time)(nil), now,
	)
}

func emptyAuditRows(mock pgxmock.PgxPoolIface) *pgxmock.Rows {
	return mock.NewRows(auditColumns())
}

func newAuditRepo(pool pgxmock.PgxPoolIface) *auditRepository {
	return &auditRepository{pool: pool}
}

// --- Record tests ---

func TestAuditRecord_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	entry := &domain.AuditEntry{
		EntityType: "activity",
		EntityID:   "act-1",
		EventType:  "status_changed",
		ActorID:    "rep-1",
		OldValue:   map[string]any{"status": "planned"},
		NewValue:   map[string]any{"status": "completed"},
	}

	mock.ExpectExec(queryInsertAudit).
		WithArgs(anyArgs(6)...).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err := repo.Record(ctx, entry)
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
}

func TestAuditRecord_NilValues(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	entry := &domain.AuditEntry{
		EntityType: "activity",
		EntityID:   "act-1",
		EventType:  "created",
		ActorID:    "rep-1",
		OldValue:   nil,
		NewValue:   nil,
	}

	mock.ExpectExec(queryInsertAudit).
		WithArgs(anyArgs(6)...).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err := repo.Record(ctx, entry)
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
}

func TestAuditRecord_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	entry := &domain.AuditEntry{
		EntityType: "activity",
		EntityID:   "act-1",
		EventType:  "created",
		ActorID:    "rep-1",
	}

	mock.ExpectExec(queryInsertAudit).
		WithArgs(anyArgs(6)...).
		WillReturnError(errors.New(errDBMsg))

	err := repo.Record(ctx, entry)
	if err == nil {
		t.Fatal(msgExpectedErr)
	}
}

// --- ListByEntity tests ---

func TestAuditListByEntity_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery(querySelectAudit).
		WithArgs("activity", "act-1").
		WillReturnRows(auditRow(mock))

	entries, err := repo.ListByEntity(ctx, "activity", "act-1")
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if len(entries) != 1 {
		t.Errorf(fmtExpected1Entry, len(entries))
	}
	if entries[0].ID != testAuditID {
		t.Errorf("expected ID %s, got %s", testAuditID, entries[0].ID)
	}
}

func TestAuditListByEntity_Empty(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery(querySelectAudit).
		WithArgs("activity", "act-1").
		WillReturnRows(emptyAuditRows(mock))

	entries, err := repo.ListByEntity(ctx, "activity", "act-1")
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestAuditListByEntity_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery(querySelectAudit).
		WithArgs("activity", "act-1").
		WillReturnError(errors.New(errDBMsg))

	_, err := repo.ListByEntity(ctx, "activity", "act-1")
	if err == nil {
		t.Fatal(msgExpectedErr)
	}
}

// --- List tests ---

func TestAuditList_NoFilters(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery(auditSelectCount).
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(querySelectAudit).
		WithArgs(anyArgs(1)...). // limit
		WillReturnRows(auditRow(mock))

	entries, total, err := repo.List(ctx, store.AuditFilter{})
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(entries) != 1 {
		t.Errorf(fmtExpected1Entry, len(entries))
	}
}

func TestAuditList_WithFilters(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	entityType := "activity"
	actorID := "rep-1"
	status := "pending"
	filter := store.AuditFilter{
		EntityType: &entityType,
		ActorID:    &actorID,
		Status:     &status,
		Page:       2,
		Limit:      10,
	}

	mock.ExpectQuery(auditSelectCount).
		WithArgs("activity", "rep-1", "pending").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(15))

	mock.ExpectQuery(querySelectAudit).
		WithArgs("activity", "rep-1", "pending", 10, 10). // filters + limit + offset
		WillReturnRows(auditRow(mock))

	entries, total, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if total != 15 {
		t.Errorf("expected total 15, got %d", total)
	}
	if len(entries) != 1 {
		t.Errorf(fmtExpected1Entry, len(entries))
	}
}

func TestAuditList_CountError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery(auditSelectCount).
		WillReturnError(errors.New(errDBMsg))

	_, _, err := repo.List(ctx, store.AuditFilter{})
	if err == nil {
		t.Fatal(msgExpectedErr)
	}
}

func TestAuditList_QueryError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery(auditSelectCount).
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(querySelectAudit).
		WithArgs(anyArgs(1)...).
		WillReturnError(errors.New(errDBMsg))

	_, _, err := repo.List(ctx, store.AuditFilter{})
	if err == nil {
		t.Fatal(msgExpectedErr)
	}
}

// --- UpdateStatus tests ---

func TestAuditUpdateStatus_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	mock.ExpectExec(queryUpdateStatus).
		WithArgs("accepted", testAdminID, testAuditID).
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := repo.UpdateStatus(ctx, testAuditID, "accepted", testAdminID)
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
}

func TestAuditUpdateStatus_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	mock.ExpectExec(queryUpdateStatus).
		WithArgs("accepted", testAdminID, "missing").
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err := repo.UpdateStatus(ctx, "missing", "accepted", testAdminID)
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestAuditUpdateStatus_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	mock.ExpectExec(queryUpdateStatus).
		WithArgs("accepted", testAdminID, testAuditID).
		WillReturnError(errors.New(errDBMsg))

	err := repo.UpdateStatus(ctx, testAuditID, "accepted", testAdminID)
	if err == nil {
		t.Fatal(msgExpectedErr)
	}
}

// --- unmarshalJSONValue tests ---

func TestUnmarshalJSONValue_Empty(t *testing.T) {
	t.Parallel()
	result, err := unmarshalJSONValue(nil, "test")
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if result != nil {
		t.Error("expected nil for empty data")
	}
}

func TestUnmarshalJSONValue_Valid(t *testing.T) {
	t.Parallel()
	data, _ := json.Marshal(map[string]any{"key": "value"})
	result, err := unmarshalJSONValue(data, "test")
	if err != nil {
		t.Fatalf(fmtUnexpectedErr, err)
	}
	if result["key"] != "value" {
		t.Errorf("expected key=value, got %v", result["key"])
	}
}

func TestUnmarshalJSONValue_InvalidJSON(t *testing.T) {
	t.Parallel()
	_, err := unmarshalJSONValue([]byte("{bad"), "test")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
