package store

import (
	"context"

	"github.com/pebblr/pebblr/internal/domain"
)

// AuditRepository provides write and query access for the audit log.
type AuditRepository interface {
	// Record persists a new audit entry.
	Record(ctx context.Context, entry *domain.AuditEntry) error

	// ListByEntity returns audit entries for a given entity, ordered by created_at desc.
	ListByEntity(ctx context.Context, entityType string, entityID string) ([]*domain.AuditEntry, error)
}
