package store

import (
	"context"
	"time"

	"github.com/pebblr/pebblr/internal/events"
)

// EventRepository provides persistence for lead lifecycle events.
type EventRepository interface {
	// Record persists a new lead event.
	Record(ctx context.Context, event *events.LeadEvent) error

	// ListByLead returns all events for the given lead ordered by timestamp ascending.
	ListByLead(ctx context.Context, leadID string) ([]events.LeadEvent, error)

	// ListByActor returns events triggered by the given actor within the time range.
	ListByActor(ctx context.Context, actorID string, from, to time.Time) ([]events.LeadEvent, error)

	// CountByType returns the count of each event type within the time range.
	CountByType(ctx context.Context, from, to time.Time) (map[events.EventType]int, error)
}
