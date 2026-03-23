-- Create the activities table for field sales activity tracking.
CREATE TABLE activities (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activity_type       TEXT NOT NULL,
    status              TEXT NOT NULL,
    due_date            DATE NOT NULL,
    duration            TEXT NOT NULL,
    routing             TEXT,
    fields              JSONB NOT NULL DEFAULT '{}',
    target_id           UUID REFERENCES targets(id),
    creator_id          UUID NOT NULL REFERENCES users(id),
    joint_visit_user_id UUID REFERENCES users(id),
    team_id             UUID REFERENCES teams(id),
    submitted_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE INDEX idx_activities_type     ON activities(activity_type);
CREATE INDEX idx_activities_status   ON activities(status);
CREATE INDEX idx_activities_due_date ON activities(due_date);
CREATE INDEX idx_activities_creator  ON activities(creator_id);
CREATE INDEX idx_activities_target   ON activities(target_id);
CREATE INDEX idx_activities_team     ON activities(team_id);
CREATE INDEX idx_activities_fields   ON activities USING GIN(fields);

-- Row-level security: reps see own + joint-visit activities; managers/admins see all.
ALTER TABLE activities ENABLE ROW LEVEL SECURITY;

CREATE POLICY activities_access ON activities FOR ALL
    USING (
        current_setting('app.user_role', true) IN ('manager', 'admin')
        OR creator_id = current_setting('app.user_id', true)::uuid
        OR joint_visit_user_id = current_setting('app.user_id', true)::uuid
    );
