DROP INDEX IF EXISTS idx_audit_status;
ALTER TABLE audit_log DROP COLUMN IF EXISTS reviewed_at;
ALTER TABLE audit_log DROP COLUMN IF EXISTS reviewed_by;
ALTER TABLE audit_log DROP COLUMN IF EXISTS status;
