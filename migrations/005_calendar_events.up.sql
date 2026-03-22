-- migrate:up

CREATE TABLE IF NOT EXISTS calendar_events (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    title      TEXT        NOT NULL,
    event_type TEXT        NOT NULL CHECK (event_type IN ('call', 'meeting', 'sync', 'visit', 'review', 'callback', 'lunch', 'demo', 'other')),
    start_time TIMESTAMPTZ NOT NULL,
    end_time   TIMESTAMPTZ,
    client     TEXT        NOT NULL DEFAULT '',
    creator_id UUID        NOT NULL REFERENCES users(id),
    team_id    UUID        REFERENCES teams(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS calendar_events_creator_id_idx  ON calendar_events(creator_id);
CREATE INDEX IF NOT EXISTS calendar_events_team_id_idx     ON calendar_events(team_id);
CREATE INDEX IF NOT EXISTS calendar_events_start_time_idx  ON calendar_events(start_time);

ALTER TABLE calendar_events ENABLE ROW LEVEL SECURITY;

-- Reps see only events they created; managers and admins see all events.
CREATE POLICY calendar_events_policy ON calendar_events
    FOR ALL
    USING (
        current_setting('app.user_role', true) = 'admin'
        OR current_setting('app.user_role', true) = 'manager'
        OR (
            current_setting('app.user_role', true) = 'rep'
            AND creator_id::TEXT = current_setting('app.user_id', true)
        )
    );
