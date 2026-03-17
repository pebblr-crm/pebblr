ALTER TABLE leads ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS leads_deleted_at_idx ON leads(deleted_at) WHERE deleted_at IS NULL;
