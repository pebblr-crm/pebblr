-- migrate:up

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS avatar        TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS online_status TEXT NOT NULL DEFAULT 'offline'
        CHECK (online_status IN ('online', 'away', 'offline'));
