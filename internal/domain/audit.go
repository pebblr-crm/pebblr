package domain

import "time"

// AuditEntry records a single change to a tracked entity for audit purposes.
type AuditEntry struct {
	ID         string         `json:"id"`
	EntityType string         `json:"entityType"` // e.g. "activity", "target"
	EntityID   string         `json:"entityId"`
	EventType  string         `json:"eventType"`  // e.g. "created", "status_changed", "submitted", "field_updated"
	ActorID    string         `json:"actorId"`
	OldValue   map[string]any `json:"oldValue,omitempty"`
	NewValue   map[string]any `json:"newValue,omitempty"`
	Status     string         `json:"status"`               // "pending", "accepted", "false_positive"
	ReviewedBy string         `json:"reviewedBy,omitempty"`
	ReviewedAt *time.Time     `json:"reviewedAt,omitempty"`
	CreatedAt  time.Time      `json:"createdAt"`
}
