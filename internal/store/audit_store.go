package store

import (
	"context"

	"github.com/pebblr/pebblr/internal/domain"
)

// AuditFilter specifies optional filter criteria for audit log queries.
type AuditFilter struct {
	EntityType *string
	ActorID    *string
	Status     *string
	Page       int
	Limit      int
}

// AuditRepository provides write and query access for the audit log.
type AuditRepository interface {
	// Record persists a new audit entry.
	Record(ctx context.Context, entry *domain.AuditEntry) error

	// ListByEntity returns audit entries for a given entity, ordered by created_at desc.
	ListByEntity(ctx context.Context, entityType string, entityID string) ([]*domain.AuditEntry, error)

	// List returns audit entries matching the filter with pagination.
	List(ctx context.Context, filter AuditFilter) ([]*domain.AuditEntry, int, error)

	// UpdateStatus updates the review status of an audit entry.
	UpdateStatus(ctx context.Context, id, status, reviewerID string) error
}
