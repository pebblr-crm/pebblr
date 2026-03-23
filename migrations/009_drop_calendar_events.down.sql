-- Recreate calendar_events table (rollback of 009_drop_calendar_events).
CREATE TABLE IF NOT EXISTS calendar_events (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    start_time  TIMESTAMPTZ NOT NULL,
    end_time    TIMESTAMPTZ NOT NULL,
    all_day     BOOLEAN NOT NULL DEFAULT false,
    location    TEXT NOT NULL DEFAULT '',
    event_type  TEXT NOT NULL DEFAULT 'meeting',
    creator_id  TEXT NOT NULL,
    assignee_id TEXT NOT NULL DEFAULT '',
    team_id     TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
