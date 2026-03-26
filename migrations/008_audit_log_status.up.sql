-- Add review workflow columns to the audit log.

ALTER TABLE audit_log ADD COLUMN status      TEXT        NOT NULL DEFAULT 'pending';
ALTER TABLE audit_log ADD COLUMN reviewed_by UUID        REFERENCES users(id);
ALTER TABLE audit_log ADD COLUMN reviewed_at TIMESTAMPTZ;

CREATE INDEX idx_audit_status ON audit_log(status);
