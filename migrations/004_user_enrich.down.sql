-- migrate:down

ALTER TABLE users
    DROP COLUMN IF EXISTS avatar,
    DROP COLUMN IF EXISTS online_status;
