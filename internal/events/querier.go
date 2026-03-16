package events

import (
	"context"
	"time"
)

// EventQuerier reads lead lifecycle events for telemetry and audit purposes.
type EventQuerier interface {
	// ListByLead returns all events for the given lead, ordered by timestamp ascending.
	ListByLead(ctx context.Context, leadID string) ([]LeadEvent, error)

	// ListByActor returns all events triggered by the given actor within the time range.
	ListByActor(ctx context.Context, actorID string, from, to time.Time) ([]LeadEvent, error)

	// CountByType returns the event count grouped by EventType within the time range.
	CountByType(ctx context.Context, from, to time.Time) (map[EventType]int, error)
}
