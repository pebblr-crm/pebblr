-- Target collections: user-created saved groups of targets for reuse across planning cycles.

CREATE TABLE target_collections (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL,
    creator_id UUID        NOT NULL REFERENCES users(id),
    team_id    UUID        NOT NULL REFERENCES teams(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_target_collections_creator ON target_collections(creator_id);
CREATE INDEX idx_target_collections_team    ON target_collections(team_id);

CREATE TABLE target_collection_items (
    collection_id UUID NOT NULL REFERENCES target_collections(id) ON DELETE CASCADE,
    target_id     UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    PRIMARY KEY (collection_id, target_id)
);

-- Row-level security: rep sees own, manager sees team, admin sees all.
ALTER TABLE target_collections ENABLE ROW LEVEL SECURITY;

CREATE POLICY target_collections_access ON target_collections FOR ALL
    USING (
        current_setting('app.user_role', true) IN ('manager', 'admin')
        OR creator_id = current_setting('app.user_id', true)::uuid
    );

ALTER TABLE target_collection_items ENABLE ROW LEVEL SECURITY;

CREATE POLICY target_collection_items_access ON target_collection_items FOR ALL
    USING (
        collection_id IN (
            SELECT id FROM target_collections
        )
    );
