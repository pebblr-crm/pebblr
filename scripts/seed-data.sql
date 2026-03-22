-- Seed data for local development
-- Pharma field-sales CRM: users, teams, doctor targets, pharmacy targets.
-- Uses fixed UUIDs for deterministic, idempotent inserts.

-- ============================================================
-- Users
-- ============================================================
INSERT INTO users (id, external_id, email, name, role) VALUES
  ('a0000000-0000-0000-0000-000000000001', 'oid-admin-001',   'admin@pebblr.dev',     'Alex Admin',    'admin'),
  ('a0000000-0000-0000-0000-000000000002', 'oid-mgr-001',     'mgr.north@pebblr.dev', 'Morgan North',  'manager'),
  ('a0000000-0000-0000-0000-000000000003', 'oid-mgr-002',     'mgr.south@pebblr.dev', 'Sam South',     'manager'),
  ('a0000000-0000-0000-0000-000000000004', 'oid-rep-001',     'rep.alice@pebblr.dev', 'Alice Reyes',   'rep'),
  ('a0000000-0000-0000-0000-000000000005', 'oid-rep-002',     'rep.bob@pebblr.dev',   'Bob Tran',      'rep'),
  ('a0000000-0000-0000-0000-000000000006', 'oid-rep-003',     'rep.carol@pebblr.dev', 'Carol Kim',     'rep'),
  ('a0000000-0000-0000-0000-000000000007', 'oid-rep-004',     'rep.dan@pebblr.dev',   'Dan Osei',      'rep')
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
-- Targets: Doctors
-- Specialties and classifications match config/tenant.json options.
-- ============================================================
INSERT INTO targets (id, target_type, name, external_id, fields, assignee_id, team_id, imported_at) VALUES
  -- North Region doctors (assigned to Alice and Bob)
  ('c0000000-0000-0000-0000-000000000001', 'doctor',
   'Dr. Elena Vasquez',
   'DOC-001',
   '{"specialty": "cardiology", "potential": "a", "city": "Prague 2", "county": "Prague", "address": "Legerova 42"}',
   'a0000000-0000-0000-0000-000000000004',
   'b0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '30 days'),

  ('c0000000-0000-0000-0000-000000000002', 'doctor',
   'Dr. Martin Horak',
   'DOC-002',
   '{"specialty": "internal_medicine", "potential": "a", "city": "Prague 3", "county": "Prague", "address": "Vinohradska 120"}',
   'a0000000-0000-0000-0000-000000000004',
   'b0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '30 days'),

  ('c0000000-0000-0000-0000-000000000003', 'doctor',
   'Dr. Petra Novakova',
   'DOC-003',
   '{"specialty": "family_medicine", "potential": "b", "city": "Prague 5", "county": "Prague", "address": "Plzenska 87"}',
   'a0000000-0000-0000-0000-000000000004',
   'b0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '30 days'),

  ('c0000000-0000-0000-0000-000000000004', 'doctor',
   'Dr. Jan Kowalski',
   'DOC-004',
   '{"specialty": "gastroenterology", "potential": "a", "city": "Prague 4", "county": "Prague", "address": "Budejovicka 15"}',
   'a0000000-0000-0000-0000-000000000005',
   'b0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '30 days'),

  ('c0000000-0000-0000-0000-000000000005', 'doctor',
   'Dr. Lucie Svobodova',
   'DOC-005',
   '{"specialty": "neurology", "potential": "b", "city": "Prague 8", "county": "Prague", "address": "Sokolovska 200"}',
   'a0000000-0000-0000-0000-000000000005',
   'b0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '30 days'),

  ('c0000000-0000-0000-0000-000000000006', 'doctor',
   'Dr. Tomas Cerny',
   'DOC-006',
   '{"specialty": "cardiology", "potential": "c", "city": "Prague 10", "county": "Prague", "address": "Vrsovicka 68"}',
   'a0000000-0000-0000-0000-000000000005',
   'b0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '30 days'),

  -- South Region doctors (assigned to Carol and Dan)
  ('c0000000-0000-0000-0000-000000000007', 'doctor',
   'Dr. Ivana Kralova',
   'DOC-007',
   '{"specialty": "pulmonology", "potential": "a", "city": "Brno", "county": "South Moravia", "address": "Masarykova 31"}',
   'a0000000-0000-0000-0000-000000000006',
   'b0000000-0000-0000-0000-000000000002',
   NOW() - INTERVAL '30 days'),

  ('c0000000-0000-0000-0000-000000000008', 'doctor',
   'Dr. Pavel Nemec',
   'DOC-008',
   '{"specialty": "internal_medicine", "potential": "b", "city": "Brno", "county": "South Moravia", "address": "Cejl 55"}',
   'a0000000-0000-0000-0000-000000000006',
   'b0000000-0000-0000-0000-000000000002',
   NOW() - INTERVAL '30 days'),

  ('c0000000-0000-0000-0000-000000000009', 'doctor',
   'Dr. Katerina Dvorakova',
   'DOC-009',
   '{"specialty": "geriatrics", "potential": "a", "city": "Olomouc", "county": "Olomouc", "address": "Horni namesti 3"}',
   'a0000000-0000-0000-0000-000000000007',
   'b0000000-0000-0000-0000-000000000002',
   NOW() - INTERVAL '30 days'),

  ('c0000000-0000-0000-0000-000000000010', 'doctor',
   'Dr. Radek Pokorny',
   'DOC-010',
   '{"specialty": "family_medicine", "potential": "c", "city": "Ostrava", "county": "Moravia-Silesia", "address": "Namesti Republiky 1"}',
   'a0000000-0000-0000-0000-000000000007',
   'b0000000-0000-0000-0000-000000000002',
   NOW() - INTERVAL '30 days'),

  ('c0000000-0000-0000-0000-000000000011', 'doctor',
   'Dr. Hana Prochazkova',
   'DOC-011',
   '{"specialty": "pediatrics", "potential": "b", "city": "Zlin", "county": "Zlin", "address": "Trida Tomase Bati 18"}',
   'a0000000-0000-0000-0000-000000000006',
   'b0000000-0000-0000-0000-000000000002',
   NOW() - INTERVAL '30 days'),

  ('c0000000-0000-0000-0000-000000000012', 'doctor',
   'Dr. Michal Beran',
   'DOC-012',
   '{"specialty": "emergency_medicine", "potential": "b", "city": "Brno", "county": "South Moravia", "address": "Jihlavska 20"}',
   'a0000000-0000-0000-0000-000000000007',
   'b0000000-0000-0000-0000-000000000002',
   NOW() - INTERVAL '30 days')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- Targets: Pharmacies
-- Pharmacy types match config/tenant.json options.
-- ============================================================
INSERT INTO targets (id, target_type, name, external_id, fields, assignee_id, team_id, imported_at) VALUES
  -- North Region pharmacies (assigned to Alice and Bob)
  ('c0000000-0000-0000-0000-000000000101', 'pharmacy',
   'Dr.Max Lekarna — Wenceslas Square',
   'PHR-001',
   '{"pharmacy_type": "chain", "city": "Prague 1", "county": "Prague", "address": "Vaclavske namesti 21"}',
   'a0000000-0000-0000-0000-000000000004',
   'b0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '30 days'),

  ('c0000000-0000-0000-0000-000000000102', 'pharmacy',
   'Dr.Max Lekarna — Andel',
   'PHR-002',
   '{"pharmacy_type": "chain", "city": "Prague 5", "county": "Prague", "address": "Nadrazni 25"}',
   'a0000000-0000-0000-0000-000000000004',
   'b0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '30 days'),

  ('c0000000-0000-0000-0000-000000000103', 'pharmacy',
   'Dr.Max Lekarna — Flora',
   'PHR-003',
   '{"pharmacy_type": "chain", "city": "Prague 3", "county": "Prague", "address": "Jicinska 8"}',
   'a0000000-0000-0000-0000-000000000005',
   'b0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '30 days'),

  ('c0000000-0000-0000-0000-000000000104', 'pharmacy',
   'Dr.Max Lekarna — Chodov',
   'PHR-004',
   '{"pharmacy_type": "chain", "city": "Prague 4", "county": "Prague", "address": "Roztylska 2321"}',
   'a0000000-0000-0000-0000-000000000005',
   'b0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '30 days'),

  -- South Region pharmacies (assigned to Carol and Dan)
  ('c0000000-0000-0000-0000-000000000105', 'pharmacy',
   'Dr.Max Lekarna — Brno Galerie',
   'PHR-005',
   '{"pharmacy_type": "chain", "city": "Brno", "county": "South Moravia", "address": "Vesela 7"}',
   'a0000000-0000-0000-0000-000000000006',
   'b0000000-0000-0000-0000-000000000002',
   NOW() - INTERVAL '30 days'),

  ('c0000000-0000-0000-0000-000000000106', 'pharmacy',
   'Dr.Max Lekarna — Olomouc',
   'PHR-006',
   '{"pharmacy_type": "chain", "city": "Olomouc", "county": "Olomouc", "address": "Masarykova trida 10"}',
   'a0000000-0000-0000-0000-000000000007',
   'b0000000-0000-0000-0000-000000000002',
   NOW() - INTERVAL '30 days'),

  ('c0000000-0000-0000-0000-000000000107', 'pharmacy',
   'Dr.Max Lekarna — Ostrava Forum',
   'PHR-007',
   '{"pharmacy_type": "chain", "city": "Ostrava", "county": "Moravia-Silesia", "address": "28. rijna 3346"}',
   'a0000000-0000-0000-0000-000000000007',
   'b0000000-0000-0000-0000-000000000002',
   NOW() - INTERVAL '30 days'),

  ('c0000000-0000-0000-0000-000000000108', 'pharmacy',
   'Dr.Max Lekarna — Zlin Centrum',
   'PHR-008',
   '{"pharmacy_type": "chain", "city": "Zlin", "county": "Zlin", "address": "Namesti Miru 174"}',
   'a0000000-0000-0000-0000-000000000006',
   'b0000000-0000-0000-0000-000000000002',
   NOW() - INTERVAL '30 days')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- Calendar Events (kept for now — replaced by Activities in Phase 2)
-- ============================================================
INSERT INTO calendar_events (id, title, event_type, start_time, end_time, client, creator_id, team_id) VALUES
  -- This week: field visits
  ('f0000000-0000-0000-0000-000000000001',
   'Dr. Vasquez — cardiology follow-up',
   'visit',
   NOW() + INTERVAL '1 day' + TIME '09:00',
   NOW() + INTERVAL '1 day' + TIME '09:30',
   'Dr. Elena Vasquez',
   'a0000000-0000-0000-0000-000000000004',
   'b0000000-0000-0000-0000-000000000001'),

  ('f0000000-0000-0000-0000-000000000002',
   'Dr.Max Wenceslas — shelf check',
   'visit',
   NOW() + INTERVAL '1 day' + TIME '10:00',
   NOW() + INTERVAL '1 day' + TIME '10:15',
   'Dr.Max Lekarna — Wenceslas Square',
   'a0000000-0000-0000-0000-000000000004',
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
   'Dr. Kowalski — GI product presentation',
   'demo',
   NOW() + INTERVAL '3 days' + TIME '10:00',
   NOW() + INTERVAL '3 days' + TIME '10:30',
   'Dr. Jan Kowalski',
   'a0000000-0000-0000-0000-000000000005',
   'b0000000-0000-0000-0000-000000000001'),

  ('f0000000-0000-0000-0000-000000000005',
   'Dr. Kralova — pulmonology visit',
   'visit',
   NOW() + INTERVAL '2 days' + TIME '11:00',
   NOW() + INTERVAL '2 days' + TIME '11:30',
   'Dr. Ivana Kralova',
   'a0000000-0000-0000-0000-000000000006',
   'b0000000-0000-0000-0000-000000000002'),

  ('f0000000-0000-0000-0000-000000000006',
   'Dr.Max Brno Galerie — inventory review',
   'visit',
   NOW() + INTERVAL '3 days' + TIME '13:00',
   NOW() + INTERVAL '3 days' + TIME '13:15',
   'Dr.Max Lekarna — Brno Galerie',
   'a0000000-0000-0000-0000-000000000006',
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
   'Dr. Dvorakova — geriatrics follow-up',
   'visit',
   NOW() + INTERVAL '5 days' + TIME '10:00',
   NOW() + INTERVAL '5 days' + TIME '10:30',
   'Dr. Katerina Dvorakova',
   'a0000000-0000-0000-0000-000000000007',
   'b0000000-0000-0000-0000-000000000002'),

  -- Past events (last week)
  ('f0000000-0000-0000-0000-000000000010',
   'Dr. Horak — internal medicine visit',
   'visit',
   NOW() - INTERVAL '5 days' + TIME '09:00',
   NOW() - INTERVAL '5 days' + TIME '09:30',
   'Dr. Martin Horak',
   'a0000000-0000-0000-0000-000000000004',
   'b0000000-0000-0000-0000-000000000001'),

  ('f0000000-0000-0000-0000-000000000011',
   'Dr. Nemec — callback',
   'callback',
   NOW() - INTERVAL '3 days' + TIME '15:00',
   NOW() - INTERVAL '3 days' + TIME '15:30',
   'Dr. Pavel Nemec',
   'a0000000-0000-0000-0000-000000000006',
   'b0000000-0000-0000-0000-000000000002'),

  ('f0000000-0000-0000-0000-000000000012',
   'Dr.Max Olomouc — pharmacy check',
   'visit',
   NOW() - INTERVAL '7 days' + TIME '14:00',
   NOW() - INTERVAL '7 days' + TIME '14:15',
   'Dr.Max Lekarna — Olomouc',
   'a0000000-0000-0000-0000-000000000007',
   'b0000000-0000-0000-0000-000000000002')
ON CONFLICT (id) DO NOTHING;
