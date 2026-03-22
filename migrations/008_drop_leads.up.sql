-- Drop lead-related tables that are no longer used.
-- The Target domain replaces Leads as of Phase 1.

DROP TABLE IF EXISTS lead_events;
DROP TABLE IF EXISTS leads;
