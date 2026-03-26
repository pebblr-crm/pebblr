-- Territories: geographic regions assigned to teams for coverage tracking.

CREATE TABLE territories (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL,
    team_id    UUID        NOT NULL REFERENCES teams(id),
    region     TEXT,
    boundary   JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_territories_team   ON territories(team_id);
CREATE INDEX idx_territories_region ON territories(region) WHERE region IS NOT NULL;

-- Row-level security: manager/admin see all, rep sees own team's territories.
ALTER TABLE territories ENABLE ROW LEVEL SECURITY;

CREATE POLICY territories_access ON territories FOR ALL
    USING (
        current_setting('app.user_role', true) IN ('manager', 'admin')
        OR team_id IN (
            SELECT team_id FROM team_members
            WHERE user_id = current_setting('app.user_id', true)::uuid
        )
    );
