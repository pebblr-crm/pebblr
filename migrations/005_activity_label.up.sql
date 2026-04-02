-- Add a label column to activities for display grouping.

ALTER TABLE activities ADD COLUMN label TEXT NOT NULL DEFAULT '';
