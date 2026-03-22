-- Seed data for local development
-- Realistic field-sales CRM data: users, teams, customers, leads, events.
-- Uses fixed UUIDs for deterministic, idempotent inserts.

-- ============================================================
-- Users
-- ============================================================
INSERT INTO users (id, external_id, email, name, role) VALUES
  ('a0000000-0000-0000-0000-000000000001', 'oid-admin-001',   'admin@pebblr.dev',   'Alex Admin',    'admin'),
  ('a0000000-0000-0000-0000-000000000002', 'oid-mgr-001',     'mgr.north@pebblr.dev', 'Morgan North', 'manager'),
  ('a0000000-0000-0000-0000-000000000003', 'oid-mgr-002',     'mgr.south@pebblr.dev', 'Sam South',    'manager'),
  ('a0000000-0000-0000-0000-000000000004', 'oid-rep-001',     'rep.alice@pebblr.dev', 'Alice Reyes',  'rep'),
  ('a0000000-0000-0000-0000-000000000005', 'oid-rep-002',     'rep.bob@pebblr.dev',   'Bob Tran',     'rep'),
  ('a0000000-0000-0000-0000-000000000006', 'oid-rep-003',     'rep.carol@pebblr.dev', 'Carol Kim',    'rep'),
  ('a0000000-0000-0000-0000-000000000007', 'oid-rep-004',     'rep.dan@pebblr.dev',   'Dan Osei',     'rep')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- Teams
-- ============================================================
INSERT INTO teams (id, name, manager_id) VALUES
  ('b0000000-0000-0000-0000-000000000001', 'North Region', 'a0000000-0000-0000-0000-000000000002'),
  ('b0000000-0000-0000-0000-000000000002', 'South Region', 'a0000000-0000-0000-0000-000000000003')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- Team Members
-- ============================================================
INSERT INTO team_members (team_id, user_id) VALUES
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000001'),
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000002'),
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000004'),
  ('b0000000-0000-0000-0000-000000000001', 'a0000000-0000-0000-0000-000000000005'),
  ('b0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000003'),
  ('b0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000006'),
  ('b0000000-0000-0000-0000-000000000002', 'a0000000-0000-0000-0000-000000000007')
ON CONFLICT DO NOTHING;

-- ============================================================
-- Customers
-- ============================================================
INSERT INTO customers (id, name, customer_type, street, city, state, country, zip, phone, email, notes) VALUES
  ('c0000000-0000-0000-0000-000000000001', 'Greenfield Market',        'retail',        '101 Oak St',        'Portland',     'OR', 'US', '97201', '503-555-0101', 'gfmarket@example.com',    'Family-owned grocery, interested in weekly delivery.'),
  ('c0000000-0000-0000-0000-000000000002', 'Pacific Wholesale Co.',    'wholesale',     '500 Industrial Blvd','Seattle',      'WA', 'US', '98101', '206-555-0202', 'pwco@example.com',        'Large distributor, price-sensitive, net-60 terms.'),
  ('c0000000-0000-0000-0000-000000000003', 'The Grand Hotel',          'hospitality',   '200 Harbor View Dr', 'San Francisco','CA', 'US', '94105', '415-555-0303', 'procurement@grandhotel.example.com', 'Upscale hotel, high-volume beverage needs.'),
  ('c0000000-0000-0000-0000-000000000004', 'Riverside College Dining', 'institutional', '10 Campus Way',      'Eugene',       'OR', 'US', '97401', '541-555-0404', 'dining@riverside.example.edu', 'University dining services, seasonal volume spikes.'),
  ('c0000000-0000-0000-0000-000000000005', 'Corner Deli & Café',       'retail',        '77 Main St',         'Bend',         'OR', 'US', '97701', '541-555-0505', 'cornerdeli@example.com',  'Local deli, steady small orders.'),
  ('c0000000-0000-0000-0000-000000000006', 'Sunrise Hospitality Group','hospitality',   '3030 Resort Rd',     'Las Vegas',    'NV', 'US', '89101', '702-555-0606', 'ops@sunrisehg.example.com','Casino resort group, bulk purchasing, quarterly reviews.'),
  ('c0000000-0000-0000-0000-000000000007', 'Northgate Grocers',        'retail',        '42 Northgate Plaza', 'Tacoma',       'WA', 'US', '98402', '253-555-0707', 'buying@northgate.example.com', 'Regional chain, 6 locations, prefers consolidated invoicing.'),
  ('c0000000-0000-0000-0000-000000000008', 'Metro Supply Partners',    'wholesale',     '900 Freight Ct',     'Sacramento',   'CA', 'US', '95814', '916-555-0808', 'orders@metrosupply.example.com', 'Mid-size distributor, expanding into OR market.')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- Leads
-- ============================================================
INSERT INTO leads (id, title, description, status, assignee_id, team_id, customer_id, customer_type, created_at) VALUES
  -- North Region leads
  ('d0000000-0000-0000-0000-000000000001',
   'Greenfield Market — Q2 restock',
   'Replenish seasonal product line ahead of summer. Rep to confirm shelf space and new SKUs.',
   'in_progress',
   'a0000000-0000-0000-0000-000000000001',
   'b0000000-0000-0000-0000-000000000001',
   'c0000000-0000-0000-0000-000000000001',
   'retail',
   NOW() - INTERVAL '14 days'),

  ('d0000000-0000-0000-0000-000000000002',
   'Pacific Wholesale — annual contract renewal',
   'Existing customer up for annual review. Discount structure needs renegotiation.',
   'visited',
   'a0000000-0000-0000-0000-000000000001',
   'b0000000-0000-0000-0000-000000000001',
   'c0000000-0000-0000-0000-000000000002',
   'wholesale',
   NOW() - INTERVAL '21 days'),

  ('d0000000-0000-0000-0000-000000000003',
   'Northgate Grocers — new locations onboarding',
   'Two new store locations opening next quarter. Coordinate with buyer for consolidated setup.',
   'assigned',
   'a0000000-0000-0000-0000-000000000001',
   'b0000000-0000-0000-0000-000000000001',
   'c0000000-0000-0000-0000-000000000007',
   'retail',
   NOW() - INTERVAL '5 days'),

  ('d0000000-0000-0000-0000-000000000004',
   'Pacific Wholesale — cold chain expansion',
   'Opportunity to upsell refrigerated line following warehouse upgrade.',
   'new',
   'a0000000-0000-0000-0000-000000000001',
   'b0000000-0000-0000-0000-000000000001',
   'c0000000-0000-0000-0000-000000000002',
   'wholesale',
   NOW() - INTERVAL '2 days'),

  ('d0000000-0000-0000-0000-000000000005',
   'Northgate Grocers — holiday promo',
   'Pitch holiday display package; decision needed by end of month.',
   'closed_won',
   'a0000000-0000-0000-0000-000000000001',
   'b0000000-0000-0000-0000-000000000001',
   'c0000000-0000-0000-0000-000000000007',
   'retail',
   NOW() - INTERVAL '45 days'),

  -- South Region leads
  ('d0000000-0000-0000-0000-000000000006',
   'The Grand Hotel — beverage program Q3',
   'Proposal for full beverage program renewal. Decision maker is F&B director.',
   'in_progress',
   'a0000000-0000-0000-0000-000000000006',
   'b0000000-0000-0000-0000-000000000002',
   'c0000000-0000-0000-0000-000000000003',
   'hospitality',
   NOW() - INTERVAL '10 days'),

  ('d0000000-0000-0000-0000-000000000007',
   'Riverside College — fall semester supply',
   'Annual ordering cycle for fall semester. Volume up ~15% vs last year.',
   'visited',
   'a0000000-0000-0000-0000-000000000007',
   'b0000000-0000-0000-0000-000000000002',
   'c0000000-0000-0000-0000-000000000004',
   'institutional',
   NOW() - INTERVAL '8 days'),

  ('d0000000-0000-0000-0000-000000000008',
   'Corner Deli — fresh line trial',
   'Pilot 4-week fresh product trial. Low volume but good brand exposure.',
   'closed_lost',
   'a0000000-0000-0000-0000-000000000006',
   'b0000000-0000-0000-0000-000000000002',
   'c0000000-0000-0000-0000-000000000005',
   'retail',
   NOW() - INTERVAL '30 days'),

  ('d0000000-0000-0000-0000-000000000009',
   'Sunrise Hospitality — Q4 bulk order',
   'Quarterly review meeting scheduled. Estimated order value $80k.',
   'assigned',
   'a0000000-0000-0000-0000-000000000007',
   'b0000000-0000-0000-0000-000000000002',
   'c0000000-0000-0000-0000-000000000006',
   'hospitality',
   NOW() - INTERVAL '3 days'),

  ('d0000000-0000-0000-0000-000000000010',
   'Metro Supply — Pacific Northwest intro',
   'New prospect expanding into PNW. Initial discovery call done; site visit TBD.',
   'new',
   NULL,
   'b0000000-0000-0000-0000-000000000002',
   'c0000000-0000-0000-0000-000000000008',
   'wholesale',
   NOW() - INTERVAL '1 day'),

  ('d0000000-0000-0000-0000-000000000011',
   'The Grand Hotel — conference catering upsell',
   'Upsell opportunity tied to a major conference booking in November.',
   'closed_won',
   'a0000000-0000-0000-0000-000000000006',
   'b0000000-0000-0000-0000-000000000002',
   'c0000000-0000-0000-0000-000000000003',
   'hospitality',
   NOW() - INTERVAL '60 days'),

  ('d0000000-0000-0000-0000-000000000012',
   'Riverside College — spring semester supply',
   'Follow-on from fall. Pre-order window opens next month.',
   'new',
   NULL,
   'b0000000-0000-0000-0000-000000000002',
   'c0000000-0000-0000-0000-000000000004',
   'institutional',
   NOW() - INTERVAL '1 day')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- Lead Events
-- ============================================================
INSERT INTO lead_events (id, lead_id, event_type, actor_id, payload) VALUES
  -- Lead 1: in_progress
  ('e0000000-0000-0000-0000-000000000001',
   'd0000000-0000-0000-0000-000000000001', 'created',
   'a0000000-0000-0000-0000-000000000002', '{"note": "Created from manager review"}'),
  ('e0000000-0000-0000-0000-000000000002',
   'd0000000-0000-0000-0000-000000000001', 'assigned',
   'a0000000-0000-0000-0000-000000000002', '{"assignee_id": "a0000000-0000-0000-0000-000000000004"}'),
  ('e0000000-0000-0000-0000-000000000003',
   'd0000000-0000-0000-0000-000000000001', 'status_changed',
   'a0000000-0000-0000-0000-000000000004', '{"from": "assigned", "to": "in_progress"}'),

  -- Lead 2: visited
  ('e0000000-0000-0000-0000-000000000004',
   'd0000000-0000-0000-0000-000000000002', 'created',
   'a0000000-0000-0000-0000-000000000002', '{"note": "Annual renewal reminder"}'),
  ('e0000000-0000-0000-0000-000000000005',
   'd0000000-0000-0000-0000-000000000002', 'assigned',
   'a0000000-0000-0000-0000-000000000002', '{"assignee_id": "a0000000-0000-0000-0000-000000000005"}'),
  ('e0000000-0000-0000-0000-000000000006',
   'd0000000-0000-0000-0000-000000000002', 'visited',
   'a0000000-0000-0000-0000-000000000005', '{"note": "Met with procurement manager. Sending revised proposal."}'),

  -- Lead 3: assigned
  ('e0000000-0000-0000-0000-000000000007',
   'd0000000-0000-0000-0000-000000000003', 'created',
   'a0000000-0000-0000-0000-000000000002', NULL),
  ('e0000000-0000-0000-0000-000000000008',
   'd0000000-0000-0000-0000-000000000003', 'assigned',
   'a0000000-0000-0000-0000-000000000002', '{"assignee_id": "a0000000-0000-0000-0000-000000000004"}'),

  -- Lead 5: closed_won
  ('e0000000-0000-0000-0000-000000000009',
   'd0000000-0000-0000-0000-000000000005', 'created',
   'a0000000-0000-0000-0000-000000000002', NULL),
  ('e0000000-0000-0000-0000-000000000010',
   'd0000000-0000-0000-0000-000000000005', 'assigned',
   'a0000000-0000-0000-0000-000000000002', '{"assignee_id": "a0000000-0000-0000-0000-000000000005"}'),
  ('e0000000-0000-0000-0000-000000000011',
   'd0000000-0000-0000-0000-000000000005', 'visited',
   'a0000000-0000-0000-0000-000000000005', '{"note": "Holiday display secured, order confirmed."}'),
  ('e0000000-0000-0000-0000-000000000012',
   'd0000000-0000-0000-0000-000000000005', 'closed',
   'a0000000-0000-0000-0000-000000000005', '{"outcome": "closed_won", "value": 4200}'),

  -- Lead 6: in_progress
  ('e0000000-0000-0000-0000-000000000013',
   'd0000000-0000-0000-0000-000000000006', 'created',
   'a0000000-0000-0000-0000-000000000003', NULL),
  ('e0000000-0000-0000-0000-000000000014',
   'd0000000-0000-0000-0000-000000000006', 'assigned',
   'a0000000-0000-0000-0000-000000000003', '{"assignee_id": "a0000000-0000-0000-0000-000000000006"}'),
  ('e0000000-0000-0000-0000-000000000015',
   'd0000000-0000-0000-0000-000000000006', 'status_changed',
   'a0000000-0000-0000-0000-000000000006', '{"from": "assigned", "to": "in_progress"}'),
  ('e0000000-0000-0000-0000-000000000016',
   'd0000000-0000-0000-0000-000000000006', 'note_added',
   'a0000000-0000-0000-0000-000000000006', '{"note": "F&B director requested formal proposal by Friday."}'),

  -- Lead 7: visited
  ('e0000000-0000-0000-0000-000000000017',
   'd0000000-0000-0000-0000-000000000007', 'created',
   'a0000000-0000-0000-0000-000000000003', NULL),
  ('e0000000-0000-0000-0000-000000000018',
   'd0000000-0000-0000-0000-000000000007', 'assigned',
   'a0000000-0000-0000-0000-000000000003', '{"assignee_id": "a0000000-0000-0000-0000-000000000007"}'),
  ('e0000000-0000-0000-0000-000000000019',
   'd0000000-0000-0000-0000-000000000007', 'visited',
   'a0000000-0000-0000-0000-000000000007', '{"note": "On-site with dining director. Volume confirmed, PO in 2 weeks."}'),

  -- Lead 8: closed_lost
  ('e0000000-0000-0000-0000-000000000020',
   'd0000000-0000-0000-0000-000000000008', 'created',
   'a0000000-0000-0000-0000-000000000003', NULL),
  ('e0000000-0000-0000-0000-000000000021',
   'd0000000-0000-0000-0000-000000000008', 'assigned',
   'a0000000-0000-0000-0000-000000000003', '{"assignee_id": "a0000000-0000-0000-0000-000000000006"}'),
  ('e0000000-0000-0000-0000-000000000022',
   'd0000000-0000-0000-0000-000000000008', 'visited',
   'a0000000-0000-0000-0000-000000000006', '{"note": "Owner decided to go with existing local supplier."}'),
  ('e0000000-0000-0000-0000-000000000023',
   'd0000000-0000-0000-0000-000000000008', 'closed',
   'a0000000-0000-0000-0000-000000000006', '{"outcome": "closed_lost", "reason": "incumbent supplier retained"}'),

  -- Lead 11: closed_won
  ('e0000000-0000-0000-0000-000000000024',
   'd0000000-0000-0000-0000-000000000011', 'created',
   'a0000000-0000-0000-0000-000000000003', NULL),
  ('e0000000-0000-0000-0000-000000000025',
   'd0000000-0000-0000-0000-000000000011', 'assigned',
   'a0000000-0000-0000-0000-000000000003', '{"assignee_id": "a0000000-0000-0000-0000-000000000006"}'),
  ('e0000000-0000-0000-0000-000000000026',
   'd0000000-0000-0000-0000-000000000011', 'visited',
   'a0000000-0000-0000-0000-000000000006', '{"note": "Conference confirmed. Catering add-on approved by events team."}'),
  ('e0000000-0000-0000-0000-000000000027',
   'd0000000-0000-0000-0000-000000000011', 'closed',
   'a0000000-0000-0000-0000-000000000006', '{"outcome": "closed_won", "value": 18500}')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- Lead enrichment (migration 003 columns)
-- ============================================================
UPDATE leads SET company = 'Greenfield Market',        industry = 'Grocery',       location = 'Portland, OR',      value_cents = 320000,  initials = 'AA' WHERE id = 'd0000000-0000-0000-0000-000000000001';
UPDATE leads SET company = 'Pacific Wholesale Co.',    industry = 'Distribution',  location = 'Seattle, WA',       value_cents = 1450000, initials = 'AA' WHERE id = 'd0000000-0000-0000-0000-000000000002';
UPDATE leads SET company = 'Northgate Grocers',        industry = 'Grocery',       location = 'Tacoma, WA',        value_cents = 560000,  initials = 'AA' WHERE id = 'd0000000-0000-0000-0000-000000000003';
UPDATE leads SET company = 'Pacific Wholesale Co.',    industry = 'Distribution',  location = 'Seattle, WA',       value_cents = 890000,  initials = 'AA' WHERE id = 'd0000000-0000-0000-0000-000000000004';
UPDATE leads SET company = 'Northgate Grocers',        industry = 'Grocery',       location = 'Tacoma, WA',        value_cents = 420000,  initials = 'AA' WHERE id = 'd0000000-0000-0000-0000-000000000005';
UPDATE leads SET company = 'The Grand Hotel',          industry = 'Hospitality',   location = 'San Francisco, CA', value_cents = 2200000, initials = 'CK' WHERE id = 'd0000000-0000-0000-0000-000000000006';
UPDATE leads SET company = 'Riverside College Dining', industry = 'Education',     location = 'Eugene, OR',        value_cents = 780000,  initials = 'DO' WHERE id = 'd0000000-0000-0000-0000-000000000007';
UPDATE leads SET company = 'Corner Deli & Café',       industry = 'Food Service',  location = 'Bend, OR',          value_cents = 85000,   initials = 'CK' WHERE id = 'd0000000-0000-0000-0000-000000000008';
UPDATE leads SET company = 'Sunrise Hospitality Group',industry = 'Hospitality',   location = 'Las Vegas, NV',     value_cents = 8000000, initials = 'DO' WHERE id = 'd0000000-0000-0000-0000-000000000009';
UPDATE leads SET company = 'Metro Supply Partners',    industry = 'Distribution',  location = 'Sacramento, CA',    value_cents = 620000,  initials = ''   WHERE id = 'd0000000-0000-0000-0000-000000000010';
UPDATE leads SET company = 'The Grand Hotel',          industry = 'Hospitality',   location = 'San Francisco, CA', value_cents = 1850000, initials = 'CK' WHERE id = 'd0000000-0000-0000-0000-000000000011';
UPDATE leads SET company = 'Riverside College Dining', industry = 'Education',     location = 'Eugene, OR',        value_cents = 750000,  initials = ''   WHERE id = 'd0000000-0000-0000-0000-000000000012';

-- ============================================================
-- User avatars and online status (migration 004 columns)
-- ============================================================
UPDATE users SET online_status = 'online'  WHERE id = 'a0000000-0000-0000-0000-000000000001';
UPDATE users SET online_status = 'online'  WHERE id = 'a0000000-0000-0000-0000-000000000002';
UPDATE users SET online_status = 'away'    WHERE id = 'a0000000-0000-0000-0000-000000000003';
UPDATE users SET online_status = 'online'  WHERE id = 'a0000000-0000-0000-0000-000000000004';
UPDATE users SET online_status = 'online'  WHERE id = 'a0000000-0000-0000-0000-000000000005';
UPDATE users SET online_status = 'offline' WHERE id = 'a0000000-0000-0000-0000-000000000006';
UPDATE users SET online_status = 'online'  WHERE id = 'a0000000-0000-0000-0000-000000000007';

-- ============================================================
-- Calendar Events
-- ============================================================
INSERT INTO calendar_events (id, title, event_type, start_time, end_time, client, creator_id, team_id) VALUES
  -- This week: field visits and calls
  ('f0000000-0000-0000-0000-000000000001',
   'Greenfield Market — shelf audit',
   'visit',
   NOW() + INTERVAL '1 day' + TIME '09:00',
   NOW() + INTERVAL '1 day' + TIME '10:30',
   'Greenfield Market',
   'a0000000-0000-0000-0000-000000000004',
   'b0000000-0000-0000-0000-000000000001'),

  ('f0000000-0000-0000-0000-000000000002',
   'Pacific Wholesale — contract review call',
   'call',
   NOW() + INTERVAL '1 day' + TIME '14:00',
   NOW() + INTERVAL '1 day' + TIME '14:45',
   'Pacific Wholesale Co.',
   'a0000000-0000-0000-0000-000000000005',
   'b0000000-0000-0000-0000-000000000001'),

  ('f0000000-0000-0000-0000-000000000003',
   'North Region — weekly sync',
   'sync',
   NOW() + INTERVAL '2 days' + TIME '08:30',
   NOW() + INTERVAL '2 days' + TIME '09:00',
   '',
   'a0000000-0000-0000-0000-000000000002',
   'b0000000-0000-0000-0000-000000000001'),

  ('f0000000-0000-0000-0000-000000000004',
   'Northgate Grocers — new store walkthrough',
   'visit',
   NOW() + INTERVAL '3 days' + TIME '10:00',
   NOW() + INTERVAL '3 days' + TIME '12:00',
   'Northgate Grocers',
   'a0000000-0000-0000-0000-000000000004',
   'b0000000-0000-0000-0000-000000000001'),

  ('f0000000-0000-0000-0000-000000000005',
   'Grand Hotel — F&B proposal presentation',
   'demo',
   NOW() + INTERVAL '2 days' + TIME '11:00',
   NOW() + INTERVAL '2 days' + TIME '12:00',
   'The Grand Hotel',
   'a0000000-0000-0000-0000-000000000006',
   'b0000000-0000-0000-0000-000000000002'),

  ('f0000000-0000-0000-0000-000000000006',
   'Riverside College — dining director check-in',
   'callback',
   NOW() + INTERVAL '3 days' + TIME '13:00',
   NOW() + INTERVAL '3 days' + TIME '13:30',
   'Riverside College Dining',
   'a0000000-0000-0000-0000-000000000007',
   'b0000000-0000-0000-0000-000000000002'),

  ('f0000000-0000-0000-0000-000000000007',
   'South Region — weekly sync',
   'sync',
   NOW() + INTERVAL '2 days' + TIME '08:30',
   NOW() + INTERVAL '2 days' + TIME '09:00',
   '',
   'a0000000-0000-0000-0000-000000000003',
   'b0000000-0000-0000-0000-000000000002'),

  ('f0000000-0000-0000-0000-000000000008',
   'Sunrise Hospitality — quarterly review',
   'review',
   NOW() + INTERVAL '5 days' + TIME '10:00',
   NOW() + INTERVAL '5 days' + TIME '11:30',
   'Sunrise Hospitality Group',
   'a0000000-0000-0000-0000-000000000007',
   'b0000000-0000-0000-0000-000000000002'),

  ('f0000000-0000-0000-0000-000000000009',
   'Metro Supply — intro lunch',
   'lunch',
   NOW() + INTERVAL '4 days' + TIME '12:00',
   NOW() + INTERVAL '4 days' + TIME '13:30',
   'Metro Supply Partners',
   'a0000000-0000-0000-0000-000000000003',
   'b0000000-0000-0000-0000-000000000002'),

  -- Past events (last week)
  ('f0000000-0000-0000-0000-000000000010',
   'Pacific Wholesale — site visit',
   'visit',
   NOW() - INTERVAL '5 days' + TIME '09:00',
   NOW() - INTERVAL '5 days' + TIME '11:00',
   'Pacific Wholesale Co.',
   'a0000000-0000-0000-0000-000000000005',
   'b0000000-0000-0000-0000-000000000001'),

  ('f0000000-0000-0000-0000-000000000011',
   'Grand Hotel — follow-up call',
   'call',
   NOW() - INTERVAL '3 days' + TIME '15:00',
   NOW() - INTERVAL '3 days' + TIME '15:30',
   'The Grand Hotel',
   'a0000000-0000-0000-0000-000000000006',
   'b0000000-0000-0000-0000-000000000002'),

  ('f0000000-0000-0000-0000-000000000012',
   'Corner Deli — product demo',
   'demo',
   NOW() - INTERVAL '7 days' + TIME '14:00',
   NOW() - INTERVAL '7 days' + TIME '15:00',
   'Corner Deli & Café',
   'a0000000-0000-0000-0000-000000000006',
   'b0000000-0000-0000-0000-000000000002')
ON CONFLICT (id) DO NOTHING;
