DROP POLICY IF EXISTS targets_rbac ON targets;
ALTER TABLE targets DISABLE ROW LEVEL SECURITY;
DROP TABLE IF EXISTS targets;

-- Restore customers table
CREATE TABLE IF NOT EXISTS customers (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT        NOT NULL,
    customer_type TEXT        NOT NULL CHECK (customer_type IN ('retail', 'wholesale', 'hospitality', 'institutional', 'other')),
    street        TEXT,
    city          TEXT,
    state         TEXT,
    country       TEXT,
    zip           TEXT,
    phone         TEXT,
    email         TEXT,
    notes         TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Restore FK on leads
ALTER TABLE leads ADD CONSTRAINT leads_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES customers(id);
