-- Generic audit log for tracking entity changes.
CREATE TABLE audit_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type TEXT NOT NULL,
    entity_id   UUID NOT NULL,
    event_type  TEXT NOT NULL,
    actor_id    UUID NOT NULL REFERENCES users(id),
    old_value   JSONB,
    new_value   JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_entity  ON audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_actor   ON audit_log(actor_id);
CREATE INDEX idx_audit_created ON audit_log(created_at);
