package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pebblr/pebblr/internal/events"
)

type eventRepository struct {
	pool *pgxpool.Pool
}

func (r *eventRepository) Record(_ context.Context, _ *events.LeadEvent) error {
	// TODO: implement
	return nil
}

func (r *eventRepository) ListByLead(_ context.Context, _ string) ([]events.LeadEvent, error) {
	// TODO: implement
	return nil, nil
}

func (r *eventRepository) ListByActor(_ context.Context, _ string, _, _ time.Time) ([]events.LeadEvent, error) {
	// TODO: implement
	return nil, nil
}

func (r *eventRepository) CountByType(_ context.Context, _, _ time.Time) (map[events.EventType]int, error) {
	// TODO: implement
	return nil, nil
}
