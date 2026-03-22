-- Recreate leads and lead_events tables (rollback of 008_drop_leads).
-- This restores the original schema from migrations 001-003.

CREATE TABLE IF NOT EXISTS leads (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status      TEXT NOT NULL DEFAULT 'new',
    assignee_id TEXT NOT NULL DEFAULT '',
    team_id     TEXT NOT NULL DEFAULT '',
    customer_id TEXT NOT NULL DEFAULT '',
    customer_type TEXT NOT NULL DEFAULT '',
    company     TEXT NOT NULL DEFAULT '',
    industry    TEXT NOT NULL DEFAULT '',
    location    TEXT NOT NULL DEFAULT '',
    value_cents BIGINT NOT NULL DEFAULT 0,
    initials    TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at  TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS lead_events (
    id         TEXT PRIMARY KEY,
    lead_id    TEXT NOT NULL REFERENCES leads(id),
    event_type TEXT NOT NULL,
    actor_id   TEXT NOT NULL,
    payload    JSONB,
    timestamp  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_lead_events_lead_id ON lead_events(lead_id);
CREATE INDEX IF NOT EXISTS idx_lead_events_actor_id ON lead_events(actor_id);
