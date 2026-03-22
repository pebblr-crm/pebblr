DROP INDEX IF EXISTS idx_targets_type_external_id;
ALTER TABLE targets DROP COLUMN IF EXISTS external_id;
