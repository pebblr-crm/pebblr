package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pebblr/pebblr/internal/domain"
	"github.com/pebblr/pebblr/internal/store"
)

type calendarEventRepository struct {
	pool *pgxpool.Pool
}

const calendarEventColumns = `
	id, title, event_type, start_time, end_time, client,
	COALESCE(creator_id::TEXT, ''), COALESCE(team_id::TEXT, ''),
	created_at, updated_at`

func scanCalendarEvent(row pgx.Row) (*domain.CalendarEvent, error) {
	var e domain.CalendarEvent
	var endTime *time.Time
	err := row.Scan(
		&e.ID, &e.Title, &e.EventType, &e.StartTime, &endTime, &e.Client,
		&e.CreatorID, &e.TeamID,
		&e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, fmt.Errorf("scanning calendar event: %w", err)
	}
	e.EndTime = endTime
	return &e, nil
}

func (r *calendarEventRepository) Get(ctx context.Context, id string) (*domain.CalendarEvent, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT `+calendarEventColumns+` FROM calendar_events WHERE id = $1::UUID`,
		id,
	)
	return scanCalendarEvent(row)
}

func (r *calendarEventRepository) List(ctx context.Context, filter store.CalendarEventFilter, page, limit int) (*store.CalendarEventPage, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 20
	}
	offset := (page - 1) * limit

	args := []any{}
	argIdx := 1
	var conditions []string

	if filter.CreatorID != nil {
		conditions = append(conditions, fmt.Sprintf("creator_id::TEXT = $%d", argIdx))
		args = append(args, *filter.CreatorID)
		argIdx++
	}
	if filter.TeamID != nil {
		conditions = append(conditions, fmt.Sprintf("team_id::TEXT = $%d", argIdx))
		args = append(args, *filter.TeamID)
		argIdx++
	}
	if filter.From != nil {
		conditions = append(conditions, fmt.Sprintf("start_time >= $%d", argIdx))
		args = append(args, *filter.From)
		argIdx++
	}
	if filter.To != nil {
		conditions = append(conditions, fmt.Sprintf("start_time <= $%d", argIdx))
		args = append(args, *filter.To)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := `SELECT COUNT(*) FROM calendar_events ` + where
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("counting calendar events: %w", err)
	}

	listQuery := `SELECT ` + calendarEventColumns + ` FROM calendar_events ` + where +
		fmt.Sprintf(` ORDER BY start_time ASC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("querying calendar events: %w", err)
	}
	defer rows.Close()

	var evts []*domain.CalendarEvent
	for rows.Next() {
		evt, err := scanCalendarEvent(rows)
		if err != nil {
			return nil, err
		}
		evts = append(evts, evt)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating calendar events: %w", err)
	}

	return &store.CalendarEventPage{Events: evts, Total: total, Page: page, Limit: limit}, nil
}

func (r *calendarEventRepository) Create(ctx context.Context, event *domain.CalendarEvent) (*domain.CalendarEvent, error) {
	var teamID *string
	if event.TeamID != "" {
		teamID = &event.TeamID
	}

	row := r.pool.QueryRow(ctx,
		`INSERT INTO calendar_events (title, event_type, start_time, end_time, client, creator_id, team_id)
		 VALUES ($1, $2, $3, $4, $5, $6::UUID, $7::UUID)
		 RETURNING `+calendarEventColumns,
		event.Title, string(event.EventType), event.StartTime, event.EndTime, event.Client,
		event.CreatorID, teamID,
	)
	return scanCalendarEvent(row)
}

func (r *calendarEventRepository) Update(ctx context.Context, event *domain.CalendarEvent) (*domain.CalendarEvent, error) {
	var teamID *string
	if event.TeamID != "" {
		teamID = &event.TeamID
	}

	row := r.pool.QueryRow(ctx,
		`UPDATE calendar_events
		 SET title = $1, event_type = $2, start_time = $3, end_time = $4,
		     client = $5, creator_id = $6::UUID, team_id = $7::UUID,
		     updated_at = NOW()
		 WHERE id = $8::UUID
		 RETURNING `+calendarEventColumns,
		event.Title, string(event.EventType), event.StartTime, event.EndTime, event.Client,
		event.CreatorID, teamID, event.ID,
	)
	return scanCalendarEvent(row)
}

func (r *calendarEventRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM calendar_events WHERE id = $1::UUID`,
		id,
	)
	if err != nil {
		return fmt.Errorf("deleting calendar event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return store.ErrNotFound
	}
	return nil
}
