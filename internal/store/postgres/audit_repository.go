package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pebblr/pebblr/internal/domain"
)

type auditRepository struct {
	pool *pgxpool.Pool
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
		        old_value, new_value, created_at
		 FROM audit_log
		 WHERE entity_type = $1 AND entity_id = $2::UUID
		 ORDER BY created_at DESC`,
		entityType, entityID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying audit log: %w", err)
	}
	defer rows.Close()

	var entries []*domain.AuditEntry
	for rows.Next() {
		var e domain.AuditEntry
		var oldJSON, newJSON []byte
		if err := rows.Scan(
			&e.ID, &e.EntityType, &e.EntityID, &e.EventType, &e.ActorID,
			&oldJSON, &newJSON, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning audit entry: %w", err)
		}
		if len(oldJSON) > 0 {
			e.OldValue = make(map[string]any)
			if err := json.Unmarshal(oldJSON, &e.OldValue); err != nil {
				return nil, fmt.Errorf("unmarshalling audit old_value: %w", err)
			}
		}
		if len(newJSON) > 0 {
			e.NewValue = make(map[string]any)
			if err := json.Unmarshal(newJSON, &e.NewValue); err != nil {
				return nil, fmt.Errorf("unmarshalling audit new_value: %w", err)
			}
		}
		entries = append(entries, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating audit log: %w", err)
	}
	return entries, nil
}

// nullJSONIfNil returns nil if the map is nil (stores NULL in db), otherwise the marshalled JSON.
func nullJSONIfNil(m map[string]any, marshalled []byte) []byte {
	if m == nil {
		return nil
	}
	return marshalled
}
