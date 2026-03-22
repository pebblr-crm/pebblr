-- migrate:up

-- Add enrichment columns to leads
ALTER TABLE leads
    ADD COLUMN IF NOT EXISTS company    TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS industry   TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS location   TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS value_cents BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS initials   TEXT NOT NULL DEFAULT '';

-- Expand the status CHECK constraint to include new frontend-aligned statuses.
-- PostgreSQL does not support ALTER CONSTRAINT, so drop and recreate.
ALTER TABLE leads DROP CONSTRAINT IF EXISTS leads_status_check;
ALTER TABLE leads ADD CONSTRAINT leads_status_check
    CHECK (status IN (
        'new', 'assigned', 'in_progress', 'visited', 'closed_won', 'closed_lost',
        'scheduled', 'done', 'hot_lead', 'follow_up', 'inquiry'
    ));
