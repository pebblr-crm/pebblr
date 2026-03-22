-- Migration 006: Drop customers table, create targets table.
-- Targets replace customers as the entities reps visit (doctors, pharmacies, etc.).

-- ============================================================
-- Remove customer foreign key from leads
-- ============================================================
ALTER TABLE leads DROP CONSTRAINT IF EXISTS leads_customer_id_fkey;

-- ============================================================
-- Drop customers table
-- ============================================================
DROP TABLE IF EXISTS customers;

-- ============================================================
-- Targets
-- ============================================================
CREATE TABLE IF NOT EXISTS targets (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    target_type   TEXT        NOT NULL,
    name          TEXT        NOT NULL,
    fields        JSONB       NOT NULL DEFAULT '{}',
    assignee_id   UUID        REFERENCES users(id),
    team_id       UUID        REFERENCES teams(id),
    imported_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_targets_type ON targets(target_type);
CREATE INDEX IF NOT EXISTS idx_targets_assignee ON targets(assignee_id);
CREATE INDEX IF NOT EXISTS idx_targets_team ON targets(team_id);
CREATE INDEX IF NOT EXISTS idx_targets_fields ON targets USING GIN(fields);

-- ============================================================
-- Row-Level Security
-- ============================================================
ALTER TABLE targets ENABLE ROW LEVEL SECURITY;

CREATE POLICY targets_rbac ON targets
    FOR ALL
    USING (
        current_setting('app.user_role', true) IN ('manager', 'admin')
        OR assignee_id::TEXT = current_setting('app.user_id', true)
    );
