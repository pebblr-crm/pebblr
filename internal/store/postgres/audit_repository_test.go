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
	"github.com/pebblr/pebblr/internal/store"
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
		"audit-1", "activity", "act-1", "status_changed", "rep-1",
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

	mock.ExpectExec("INSERT INTO audit_log").
		WithArgs(anyArgs(6)...).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err := repo.Record(ctx, entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

	mock.ExpectExec("INSERT INTO audit_log").
		WithArgs(anyArgs(6)...).
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	err := repo.Record(ctx, entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

	mock.ExpectExec("INSERT INTO audit_log").
		WithArgs(anyArgs(6)...).
		WillReturnError(fmt.Errorf("db error"))

	err := repo.Record(ctx, entry)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- ListByEntity tests ---

func TestAuditListByEntity_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM audit_log").
		WithArgs("activity", "act-1").
		WillReturnRows(auditRow(mock))

	entries, err := repo.ListByEntity(ctx, "activity", "act-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].ID != "audit-1" {
		t.Errorf("expected ID audit-1, got %s", entries[0].ID)
	}
}

func TestAuditListByEntity_Empty(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT .+ FROM audit_log").
		WithArgs("activity", "act-1").
		WillReturnRows(emptyAuditRows(mock))

	entries, err := repo.ListByEntity(ctx, "activity", "act-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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

	mock.ExpectQuery("SELECT .+ FROM audit_log").
		WithArgs("activity", "act-1").
		WillReturnError(fmt.Errorf("db error"))

	_, err := repo.ListByEntity(ctx, "activity", "act-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- List tests ---

func TestAuditList_NoFilters(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery("SELECT .+ FROM audit_log").
		WithArgs(anyArgs(1)...). // limit
		WillReturnRows(auditRow(mock))

	entries, total, err := repo.List(ctx, store.AuditFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
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

	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
		WithArgs("activity", "rep-1", "pending").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(15))

	mock.ExpectQuery("SELECT .+ FROM audit_log").
		WithArgs("activity", "rep-1", "pending", 10, 10). // filters + limit + offset
		WillReturnRows(auditRow(mock))

	entries, total, err := repo.List(ctx, filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 15 {
		t.Errorf("expected total 15, got %d", total)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 entry, got %d", len(entries))
	}
}

func TestAuditList_CountError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
		WillReturnError(fmt.Errorf("db error"))

	_, _, err := repo.List(ctx, store.AuditFilter{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAuditList_QueryError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	mock.ExpectQuery("SELECT COUNT\\(\\*\\)").
		WillReturnRows(mock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery("SELECT .+ FROM audit_log").
		WithArgs(anyArgs(1)...).
		WillReturnError(fmt.Errorf("db error"))

	_, _, err := repo.List(ctx, store.AuditFilter{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- UpdateStatus tests ---

func TestAuditUpdateStatus_Success(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	mock.ExpectExec("UPDATE audit_log SET status").
		WithArgs("accepted", "admin-1", "audit-1").
		WillReturnResult(pgxmock.NewResult("UPDATE", 1))

	err := repo.UpdateStatus(ctx, "audit-1", "accepted", "admin-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuditUpdateStatus_NotFound(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	mock.ExpectExec("UPDATE audit_log SET status").
		WithArgs("accepted", "admin-1", "missing").
		WillReturnResult(pgxmock.NewResult("UPDATE", 0))

	err := repo.UpdateStatus(ctx, "missing", "accepted", "admin-1")
	if !errors.Is(err, store.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}

func TestAuditUpdateStatus_DBError(t *testing.T) {
	t.Parallel()
	mock := newMockPool(t)
	repo := newAuditRepo(mock)
	ctx := context.Background()

	mock.ExpectExec("UPDATE audit_log SET status").
		WithArgs("accepted", "admin-1", "audit-1").
		WillReturnError(fmt.Errorf("db error"))

	err := repo.UpdateStatus(ctx, "audit-1", "accepted", "admin-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- unmarshalJSONValue tests ---

func TestUnmarshalJSONValue_Empty(t *testing.T) {
	t.Parallel()
	result, err := unmarshalJSONValue(nil, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
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
		t.Fatalf("unexpected error: %v", err)
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
