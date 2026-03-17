DROP INDEX IF EXISTS leads_deleted_at_idx;

ALTER TABLE leads DROP COLUMN IF EXISTS deleted_at;
