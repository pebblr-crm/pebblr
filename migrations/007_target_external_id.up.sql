-- Migration 007: Add external_id column to targets for import upsert.

ALTER TABLE targets ADD COLUMN IF NOT EXISTS external_id TEXT;

-- Unique constraint on (target_type, external_id) to support upsert by external ID.
-- NULL external_id values are excluded — only imported targets have one.
CREATE UNIQUE INDEX IF NOT EXISTS idx_targets_type_external_id
    ON targets(target_type, external_id)
    WHERE external_id IS NOT NULL;
