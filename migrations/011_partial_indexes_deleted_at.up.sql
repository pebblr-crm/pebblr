-- Partial indexes on activities for queries that filter on deleted_at IS NULL.
-- These significantly improve performance for the common case where soft-deleted
-- rows are excluded, since the planner can skip deleted rows entirely.

CREATE INDEX IF NOT EXISTS idx_activities_creator_not_deleted
    ON activities(creator_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_activities_target_not_deleted
    ON activities(target_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_activities_due_date_not_deleted
    ON activities(due_date) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_activities_status_not_deleted
    ON activities(status) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_activities_team_not_deleted
    ON activities(team_id) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_activities_type_not_deleted
    ON activities(activity_type) WHERE deleted_at IS NULL;
