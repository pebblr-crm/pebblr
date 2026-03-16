-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================
-- Users
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id TEXT        NOT NULL UNIQUE,  -- Azure AD OID
    email       TEXT        NOT NULL UNIQUE,
    name        TEXT        NOT NULL,
    role        TEXT        NOT NULL CHECK (role IN ('rep', 'manager', 'admin')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- Teams
-- ============================================================
CREATE TABLE IF NOT EXISTS teams (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL,
    manager_id UUID        NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Join table: user team membership
CREATE TABLE IF NOT EXISTS team_members (
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (team_id, user_id)
);

-- ============================================================
-- Customers
-- ============================================================
CREATE TABLE IF NOT EXISTS customers (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT        NOT NULL,
    customer_type TEXT        NOT NULL CHECK (customer_type IN ('retail', 'wholesale', 'hospitality', 'institutional', 'other')),
    street        TEXT,
    city          TEXT,
    state         TEXT,
    country       TEXT,
    zip           TEXT,
    phone         TEXT,
    email         TEXT,
    notes         TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================
-- Leads
-- ============================================================
CREATE TABLE IF NOT EXISTS leads (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    title         TEXT        NOT NULL,
    description   TEXT,
    status        TEXT        NOT NULL DEFAULT 'new'
                              CHECK (status IN ('new', 'assigned', 'in_progress', 'visited', 'closed_won', 'closed_lost')),
    assignee_id   UUID        REFERENCES users(id),
    team_id       UUID        NOT NULL REFERENCES teams(id),
    customer_id   UUID        NOT NULL REFERENCES customers(id),
    customer_type TEXT        NOT NULL CHECK (customer_type IN ('retail', 'wholesale', 'hospitality', 'institutional', 'other')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS leads_assignee_id_idx ON leads(assignee_id);
CREATE INDEX IF NOT EXISTS leads_team_id_idx ON leads(team_id);
CREATE INDEX IF NOT EXISTS leads_status_idx ON leads(status);
CREATE INDEX IF NOT EXISTS leads_created_at_idx ON leads(created_at);

-- ============================================================
-- Lead Events
-- ============================================================
CREATE TABLE IF NOT EXISTS lead_events (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    lead_id    UUID        NOT NULL REFERENCES leads(id) ON DELETE CASCADE,
    event_type TEXT        NOT NULL CHECK (event_type IN ('created', 'assigned', 'status_changed', 'note_added', 'visited', 'closed')),
    actor_id   UUID        NOT NULL REFERENCES users(id),
    payload    JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS lead_events_lead_id_idx    ON lead_events(lead_id);
CREATE INDEX IF NOT EXISTS lead_events_actor_id_idx   ON lead_events(actor_id);
CREATE INDEX IF NOT EXISTS lead_events_created_at_idx ON lead_events(created_at);

-- ============================================================
-- Row-Level Security (defense-in-depth)
-- ============================================================
ALTER TABLE leads       ENABLE ROW LEVEL SECURITY;
ALTER TABLE lead_events ENABLE ROW LEVEL SECURITY;

-- Application role used by the API server.
-- CREATE ROLE pebblr_app LOGIN PASSWORD '...';  -- managed via External Secrets

-- Reps see only leads assigned to them.
CREATE POLICY rep_leads_policy ON leads
    FOR ALL
    USING (
        current_setting('app.user_role', true) = 'admin'
        OR current_setting('app.user_role', true) = 'manager'
        OR (
            current_setting('app.user_role', true) = 'rep'
            AND assignee_id::TEXT = current_setting('app.user_id', true)
        )
    );

-- Lead events follow the same visibility as leads.
CREATE POLICY rep_lead_events_policy ON lead_events
    FOR ALL
    USING (
        current_setting('app.user_role', true) = 'admin'
        OR current_setting('app.user_role', true) = 'manager'
        OR lead_id IN (
            SELECT id FROM leads
            WHERE assignee_id::TEXT = current_setting('app.user_id', true)
        )
    );
