-- migrate:down

-- Restore original CHECK constraint (6 statuses only)
ALTER TABLE leads DROP CONSTRAINT IF EXISTS leads_status_check;
ALTER TABLE leads ADD CONSTRAINT leads_status_check
    CHECK (status IN ('new', 'assigned', 'in_progress', 'visited', 'closed_won', 'closed_lost'));

-- Remove enrichment columns
ALTER TABLE leads
    DROP COLUMN IF EXISTS company,
    DROP COLUMN IF EXISTS industry,
    DROP COLUMN IF EXISTS location,
    DROP COLUMN IF EXISTS value_cents,
    DROP COLUMN IF EXISTS initials;
