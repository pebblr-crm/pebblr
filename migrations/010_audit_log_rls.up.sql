-- Enable Row-Level Security on audit_log.
-- Reps can only see audit entries they created; managers and admins see all.
-- This closes the gap where audit_log was the only data table without RLS.

ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;

CREATE POLICY audit_log_access ON audit_log FOR ALL
    USING (
        current_setting('app.user_role', true) IN ('manager', 'admin')
        OR actor_id::TEXT = current_setting('app.user_id', true)
    );
