package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

type auditRepository struct {
	pool dbPool
}

func (r *auditRepository) Record(ctx context.Context, entry *domain.AuditEntry) error {
	oldJSON, err := json.Marshal(entry.OldValue)
	if err != nil {
		return fmt.Errorf("marshalling audit old_value: %w", err)
	}
	newJSON, err := json.Marshal(entry.NewValue)
	if err != nil {
		return fmt.Errorf("marshalling audit new_value: %w", err)
	}

	_, err = r.pool.Exec(ctx,
		`INSERT INTO audit_log (entity_type, entity_id, event_type, actor_id, old_value, new_value)
		 VALUES ($1, $2::UUID, $3, $4::UUID, $5, $6)`,
		entry.EntityType, entry.EntityID, entry.EventType, entry.ActorID,
		nullJSONIfNil(entry.OldValue, oldJSON), nullJSONIfNil(entry.NewValue, newJSON),
	)
	if err != nil {
		return fmt.Errorf("recording audit entry: %w", err)
	}
	return nil
}

func (r *auditRepository) ListByEntity(ctx context.Context, entityType, entityID string) ([]*domain.AuditEntry, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id::TEXT, entity_type, entity_id::TEXT, event_type, actor_id::TEXT,
		        old_value, new_value, status, reviewed_by::TEXT, reviewed_at, created_at
		 FROM audit_log
		 WHERE entity_type = $1 AND entity_id = $2::UUID
		 ORDER BY created_at DESC`,
		entityType, entityID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying audit log: %w", err)
	}
	defer rows.Close()

	return scanAuditEntries(rows)
}

func (r *auditRepository) List(ctx context.Context, filter store.AuditFilter) ([]*domain.AuditEntry, int, error) {
	query := `SELECT id::TEXT, entity_type, entity_id::TEXT, event_type, actor_id::TEXT,
	                 old_value, new_value, status, reviewed_by::TEXT, reviewed_at, created_at
	          FROM audit_log WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM audit_log WHERE 1=1`
	args := []any{}
	argIdx := 1

	if filter.EntityType != nil {
		clause := fmt.Sprintf(" AND entity_type = $%d", argIdx)
		query += clause
		countQuery += clause
		args = append(args, *filter.EntityType)
		argIdx++
	}
	if filter.ActorID != nil {
		clause := fmt.Sprintf(" AND actor_id = $%d::UUID", argIdx)
		query += clause
		countQuery += clause
		args = append(args, *filter.ActorID)
		argIdx++
	}
	if filter.Status != nil {
		clause := fmt.Sprintf(" AND status = $%d", argIdx)
		query += clause
		countQuery += clause
		args = append(args, *filter.Status)
		argIdx++
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting audit entries: %w", err)
	}

	query += " ORDER BY created_at DESC"
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	query += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, limit)
	argIdx++

	if filter.Page > 1 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, (filter.Page-1)*limit)
		argIdx++
	}
	_ = argIdx

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing audit entries: %w", err)
	}
	defer rows.Close()

	entries, err := scanAuditEntries(rows)
	if err != nil {
		return nil, 0, err
	}
	return entries, total, nil
}

func (r *auditRepository) UpdateStatus(ctx context.Context, id, status, reviewerID string) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE audit_log SET status = $1, reviewed_by = $2::UUID, reviewed_at = NOW()
		 WHERE id = $3::UUID`,
		status, reviewerID, id,
	)
	if err != nil {
		return fmt.Errorf("updating audit status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}

func scanAuditEntries(rows pgx.Rows) ([]*domain.AuditEntry, error) {
	var entries []*domain.AuditEntry
	for rows.Next() {
		e, err := scanOneAuditEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("iterating audit log: %w", err)
	}
	return entries, nil
}

// scanOneAuditEntry scans a single row into a domain.AuditEntry.
func scanOneAuditEntry(rows pgx.Rows) (*domain.AuditEntry, error) {
	var e domain.AuditEntry
	var oldJSON, newJSON []byte
	var reviewedBy *string
	if err := rows.Scan(
		&e.ID, &e.EntityType, &e.EntityID, &e.EventType, &e.ActorID,
		&oldJSON, &newJSON, &e.Status, &reviewedBy, &e.ReviewedAt, &e.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("scanning audit entry: %w", err)
	}
	if reviewedBy != nil {
		e.ReviewedBy = *reviewedBy
	}
	var unmarshalErr error
	if e.OldValue, unmarshalErr = unmarshalJSONValue(oldJSON, "old_value"); unmarshalErr != nil {
		return nil, unmarshalErr
	}
	if e.NewValue, unmarshalErr = unmarshalJSONValue(newJSON, "new_value"); unmarshalErr != nil {
		return nil, unmarshalErr
	}
	return &e, nil
}

// unmarshalJSONValue decodes a JSON byte slice into a map if non-empty.
func unmarshalJSONValue(data []byte, label string) (map[string]any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	m := make(map[string]any)
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("unmarshalling audit %s: %w", label, err)
	}
	return m, nil
}

