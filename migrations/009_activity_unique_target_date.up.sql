-- Soft-delete duplicate activities before adding the unique constraint.
-- Keeps the oldest activity per (target_id, due_date, creator_id) group.
UPDATE activities SET deleted_at = NOW()
WHERE id IN (
    SELECT id FROM (
        SELECT id,
               ROW_NUMBER() OVER (
                   PARTITION BY target_id, due_date, creator_id
                   ORDER BY created_at ASC
               ) AS rn
        FROM activities
        WHERE deleted_at IS NULL AND target_id IS NOT NULL
    ) ranked
    WHERE rn > 1
);

-- Prevent duplicate activities for the same target on the same date by the same creator.
-- Only applies to non-deleted activities with a target_id.
CREATE UNIQUE INDEX idx_activities_unique_target_date
    ON activities (target_id, due_date, creator_id)
    WHERE deleted_at IS NULL AND target_id IS NOT NULL;
