package service

import (
	"context"
	"fmt"

	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

// CalendarEventService handles calendar event business logic.
type CalendarEventService struct {
	events store.CalendarEventRepository
}

// NewCalendarEventService constructs a CalendarEventService with the given repository.
func NewCalendarEventService(events store.CalendarEventRepository) *CalendarEventService {
	return &CalendarEventService{events: events}
}

// Get retrieves a calendar event by ID.
func (s *CalendarEventService) Get(ctx context.Context, id string) (*domain.CalendarEvent, error) {
	evt, err := s.events.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting calendar event: %w", err)
	}
	return evt, nil
}

// List returns a paginated list of calendar events matching the filter.
func (s *CalendarEventService) List(ctx context.Context, filter store.CalendarEventFilter, page, limit int) (*store.CalendarEventPage, error) {
	result, err := s.events.List(ctx, filter, page, limit)
	if err != nil {
		return nil, fmt.Errorf("listing calendar events: %w", err)
	}
	return result, nil
}

// Create persists a new calendar event. The actor's ID is set as the creator.
func (s *CalendarEventService) Create(ctx context.Context, actor *domain.User, event *domain.CalendarEvent) (*domain.CalendarEvent, error) {
	event.CreatorID = actor.ID

	created, err := s.events.Create(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("creating calendar event: %w", err)
	}
	return created, nil
}

// Update replaces an existing calendar event's fields.
// Only the creator or an admin may update an event.
func (s *CalendarEventService) Update(ctx context.Context, actor *domain.User, event *domain.CalendarEvent) (*domain.CalendarEvent, error) {
	existing, err := s.events.Get(ctx, event.ID)
	if err != nil {
		return nil, fmt.Errorf("getting calendar event: %w", err)
	}

	if actor.Role != domain.RoleAdmin && existing.CreatorID != actor.ID {
		return nil, ErrForbidden
	}

	updated, err := s.events.Update(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("updating calendar event: %w", err)
	}
	return updated, nil
}

// Delete removes a calendar event by ID.
// Only the creator or an admin may delete an event.
func (s *CalendarEventService) Delete(ctx context.Context, actor *domain.User, id string) error {
	existing, err := s.events.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("getting calendar event: %w", err)
	}

	if actor.Role != domain.RoleAdmin && existing.CreatorID != actor.ID {
		return ErrForbidden
	}

	if err := s.events.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting calendar event: %w", err)
	}
	return nil
}
