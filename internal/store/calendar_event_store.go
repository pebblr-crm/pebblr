package store

import (
	"context"
	"time"

	"github.com/pebblr/pebblr/internal/domain"
)

// CalendarEventFilter specifies optional filter criteria for calendar event list queries.
type CalendarEventFilter struct {
	// CreatorID filters events by the creating user's ID.
	CreatorID *string
	// TeamID filters events by associated team ID.
	TeamID *string
	// From filters events with start_time >= From.
	From *time.Time
	// To filters events with start_time <= To.
	To *time.Time
}

// CalendarEventPage holds a paginated result set of calendar events.
type CalendarEventPage struct {
	Events []*domain.CalendarEvent
	Total  int
	Page   int
	Limit  int
}

// CalendarEventRepository provides CRUD and query access for calendar events.
type CalendarEventRepository interface {
	// Get retrieves a single calendar event by ID.
	// Returns ErrNotFound if no event exists with that ID.
	Get(ctx context.Context, id string) (*domain.CalendarEvent, error)

	// List returns a paginated list of calendar events matching the filter.
	List(ctx context.Context, filter CalendarEventFilter, page, limit int) (*CalendarEventPage, error)

	// Create persists a new calendar event and returns it with its generated ID.
	Create(ctx context.Context, event *domain.CalendarEvent) (*domain.CalendarEvent, error)

	// Update persists changes to an existing calendar event.
	Update(ctx context.Context, event *domain.CalendarEvent) (*domain.CalendarEvent, error)

	// Delete removes a calendar event by ID.
	Delete(ctx context.Context, id string) error
}
