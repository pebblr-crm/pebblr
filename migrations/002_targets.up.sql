-- Targets: entities that reps visit (doctors, pharmacies, etc.).
-- Target types and dynamic fields are driven by the tenant configuration.

CREATE TABLE IF NOT EXISTS targets (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    target_type TEXT        NOT NULL,
    name        TEXT        NOT NULL,
    fields      JSONB       NOT NULL DEFAULT '{}',
    external_id TEXT,                              -- external system ID for import upsert
    assignee_id UUID        REFERENCES users(id),
    team_id     UUID        REFERENCES teams(id),
    imported_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_targets_type     ON targets(target_type);
CREATE INDEX IF NOT EXISTS idx_targets_assignee ON targets(assignee_id);
CREATE INDEX IF NOT EXISTS idx_targets_team     ON targets(team_id);
CREATE INDEX IF NOT EXISTS idx_targets_fields   ON targets USING GIN(fields);

-- Unique on (target_type, external_id) for upsert; NULLs excluded.
CREATE UNIQUE INDEX IF NOT EXISTS idx_targets_type_external_id
    ON targets(target_type, external_id)
    WHERE external_id IS NOT NULL;

-- Row-Level Security
ALTER TABLE targets ENABLE ROW LEVEL SECURITY;

CREATE POLICY targets_rbac ON targets
    FOR ALL
    USING (
        current_setting('app.user_role', true) IN ('manager', 'admin')
        OR assignee_id::TEXT = current_setting('app.user_id', true)
    );
