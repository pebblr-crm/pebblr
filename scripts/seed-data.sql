-- Seed data for local development
-- Pharmaceutical field-sales CRM: users, teams, targets (doctors/pharmacies), activities.
-- All targets are in Bucharest, Romania. Uses fixed UUIDs for deterministic, idempotent inserts.

-- ============================================================
-- Users
-- ============================================================
INSERT INTO users (id, external_id, email, name, role, avatar, online_status) VALUES
  ('a0000000-0000-0000-0000-000000000001', 'oid-admin-001',   'admin@pebblr.dev',     'Alexandru Dobre',  'admin',   '', 'online'),
  ('a0000000-0000-0000-0000-000000000002', 'oid-mgr-001',     'mgr.north@pebblr.dev', 'Ioana Marinescu',  'manager', '', 'online'),
  ('a0000000-0000-0000-0000-000000000003', 'oid-mgr-002',     'mgr.south@pebblr.dev', 'Cristian Barbu',   'manager', '', 'away'),
  ('a0000000-0000-0000-0000-000000000004', 'oid-rep-001',     'rep.alice@pebblr.dev', 'Alina Popa',       'rep',     '', 'online'),
  ('a0000000-0000-0000-0000-000000000005', 'oid-rep-002',     'rep.bob@pebblr.dev',   'Bogdan Toma',      'rep',     '', 'online'),
  ('a0000000-0000-0000-0000-000000000006', 'oid-rep-003',     'rep.carol@pebblr.dev', 'Camelia Radu',     'rep',     '', 'offline'),
  ('a0000000-0000-0000-0000-000000000007', 'oid-rep-004',     'rep.dan@pebblr.dev',   'Daniel Nistor',    'rep',     '', 'online')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- Teams
-- ============================================================
INSERT INTO teams (id, name, manager_id) VALUES
  ('b0000000-0000-0000-0000-000000000001', 'Sector 1-3', 'a0000000-0000-0000-0000-000000000002'),
  ('b0000000-0000-0000-0000-000000000002', 'Sector 4-6', 'a0000000-0000-0000-0000-000000000003')
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
-- Targets: Doctors (18)
-- ============================================================
INSERT INTO targets (id, target_type, name, external_id, fields, assignee_id, team_id) VALUES
  -- Sector 1-3: Alina's doctors
  ('c0000000-0000-0000-0000-000000000001', 'doctor', 'Dr. Elena Popescu', 'DOC-001',
   '{"specialty": "cardiology", "potential": "a", "city": "București", "county": "București", "address": "Bd. Aviatorilor 34, Sector 1"}',
   'a0000000-0000-0000-0000-000000000004', 'b0000000-0000-0000-0000-000000000001'),

  ('c0000000-0000-0000-0000-000000000002', 'doctor', 'Dr. Mihai Ionescu', 'DOC-002',
   '{"specialty": "internal_medicine", "potential": "b", "city": "București", "county": "București", "address": "Str. Știrbei Vodă 128, Sector 1"}',
   'a0000000-0000-0000-0000-000000000004', 'b0000000-0000-0000-0000-000000000001'),

  ('c0000000-0000-0000-0000-000000000003', 'doctor', 'Dr. Adriana Vlad', 'DOC-003',
   '{"specialty": "neurology", "potential": "a", "city": "București", "county": "București", "address": "Calea Dorobanți 77, Sector 1"}',
   'a0000000-0000-0000-0000-000000000004', 'b0000000-0000-0000-0000-000000000001'),

  ('c0000000-0000-0000-0000-000000000004', 'doctor', 'Dr. Sorin Munteanu', 'DOC-004',
   '{"specialty": "gastroenterology", "potential": "c", "city": "București", "county": "București", "address": "Str. Polonă 52, Sector 1"}',
   'a0000000-0000-0000-0000-000000000004', 'b0000000-0000-0000-0000-000000000001'),

  -- Sector 1-3: Bogdan's doctors
  ('c0000000-0000-0000-0000-000000000005', 'doctor', 'Dr. Ana Dumitrescu', 'DOC-005',
   '{"specialty": "family_medicine", "potential": "a", "city": "București", "county": "București", "address": "Bd. Magheru 28, Sector 2"}',
   'a0000000-0000-0000-0000-000000000005', 'b0000000-0000-0000-0000-000000000001'),

  ('c0000000-0000-0000-0000-000000000006', 'doctor', 'Dr. Radu Constantinescu', 'DOC-006',
   '{"specialty": "pulmonology", "potential": "b", "city": "București", "county": "București", "address": "Str. Barbu Văcărescu 164, Sector 2"}',
   'a0000000-0000-0000-0000-000000000005', 'b0000000-0000-0000-0000-000000000001'),

  ('c0000000-0000-0000-0000-000000000007', 'doctor', 'Dr. Florina Neagu', 'DOC-007',
   '{"specialty": "endocrinology", "potential": "a", "city": "București", "county": "București", "address": "Bd. Dacia 65, Sector 2"}',
   'a0000000-0000-0000-0000-000000000005', 'b0000000-0000-0000-0000-000000000001'),

  ('c0000000-0000-0000-0000-000000000008', 'doctor', 'Dr. Victor Preda', 'DOC-008',
   '{"specialty": "cardiology", "potential": "b", "city": "București", "county": "București", "address": "Str. Traian 200, Sector 3"}',
   'a0000000-0000-0000-0000-000000000005', 'b0000000-0000-0000-0000-000000000001'),

  ('c0000000-0000-0000-0000-000000000009', 'doctor', 'Dr. Luminița Stoian', 'DOC-009',
   '{"specialty": "dermatology", "potential": "c", "city": "București", "county": "București", "address": "Calea Călărașilor 306, Sector 3"}',
   'a0000000-0000-0000-0000-000000000005', 'b0000000-0000-0000-0000-000000000001'),

  -- Sector 4-6: Camelia's doctors
  ('c0000000-0000-0000-0000-000000000010', 'doctor', 'Dr. Maria Stanescu', 'DOC-010',
   '{"specialty": "neurology", "potential": "a", "city": "București", "county": "București", "address": "Bd. Tineretului 1, Sector 4"}',
   'a0000000-0000-0000-0000-000000000006', 'b0000000-0000-0000-0000-000000000002'),

  ('c0000000-0000-0000-0000-000000000011', 'doctor', 'Dr. Cristina Moldovan', 'DOC-011',
   '{"specialty": "pediatrics", "potential": "b", "city": "București", "county": "București", "address": "Șos. Olteniței 40, Sector 4"}',
   'a0000000-0000-0000-0000-000000000006', 'b0000000-0000-0000-0000-000000000002'),

  ('c0000000-0000-0000-0000-000000000012', 'doctor', 'Dr. Gheorghe Petre', 'DOC-012',
   '{"specialty": "orthopedics", "potential": "a", "city": "București", "county": "București", "address": "Calea Văcărești 280, Sector 4"}',
   'a0000000-0000-0000-0000-000000000006', 'b0000000-0000-0000-0000-000000000002'),

  ('c0000000-0000-0000-0000-000000000013', 'doctor', 'Dr. Daniela Luca', 'DOC-013',
   '{"specialty": "internal_medicine", "potential": "b", "city": "București", "county": "București", "address": "Str. Fabricii 12, Sector 6"}',
   'a0000000-0000-0000-0000-000000000006', 'b0000000-0000-0000-0000-000000000002'),

  -- Sector 4-6: Daniel's doctors
  ('c0000000-0000-0000-0000-000000000014', 'doctor', 'Dr. Andrei Georgescu', 'DOC-014',
   '{"specialty": "pulmonology", "potential": "b", "city": "București", "county": "București", "address": "Calea 13 Septembrie 106, Sector 5"}',
   'a0000000-0000-0000-0000-000000000007', 'b0000000-0000-0000-0000-000000000002'),

  ('c0000000-0000-0000-0000-000000000015', 'doctor', 'Dr. Ion Petrescu', 'DOC-015',
   '{"specialty": "geriatrics", "potential": "c", "city": "București", "county": "București", "address": "Bd. Libertății 18, Sector 5"}',
   'a0000000-0000-0000-0000-000000000007', 'b0000000-0000-0000-0000-000000000002'),

  ('c0000000-0000-0000-0000-000000000016', 'doctor', 'Dr. Oana Dragomir', 'DOC-016',
   '{"specialty": "family_medicine", "potential": "a", "city": "București", "county": "București", "address": "Bd. Iuliu Maniu 63, Sector 6"}',
   'a0000000-0000-0000-0000-000000000007', 'b0000000-0000-0000-0000-000000000002'),

  ('c0000000-0000-0000-0000-000000000017', 'doctor', 'Dr. Bogdan Enescu', 'DOC-017',
   '{"specialty": "gastroenterology", "potential": "b", "city": "București", "county": "București", "address": "Str. Brașov 21, Sector 6"}',
   'a0000000-0000-0000-0000-000000000007', 'b0000000-0000-0000-0000-000000000002'),

  ('c0000000-0000-0000-0000-000000000018', 'doctor', 'Dr. Raluca Matei', 'DOC-018',
   '{"specialty": "cardiology", "potential": "a", "city": "București", "county": "București", "address": "Bd. Timișoara 44, Sector 6"}',
   'a0000000-0000-0000-0000-000000000007', 'b0000000-0000-0000-0000-000000000002')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- Targets: Pharmacies (8)
-- ============================================================
INSERT INTO targets (id, target_type, name, external_id, fields, assignee_id, team_id) VALUES
  -- Sector 1-3 pharmacies
  ('c0000000-0000-0000-0000-000000000019', 'pharmacy', 'Farmacia Dona — Victoriei',      'PHR-001',
   '{"pharmacy_type": "chain", "city": "București", "county": "București", "address": "Calea Victoriei 155, Sector 1"}',
   'a0000000-0000-0000-0000-000000000004', 'b0000000-0000-0000-0000-000000000001'),

  ('c0000000-0000-0000-0000-000000000020', 'pharmacy', 'Farmacia Sensiblu — Floreasca',  'PHR-002',
   '{"pharmacy_type": "chain", "city": "București", "county": "București", "address": "Calea Floreasca 169, Sector 1"}',
   'a0000000-0000-0000-0000-000000000004', 'b0000000-0000-0000-0000-000000000001'),

  ('c0000000-0000-0000-0000-000000000021', 'pharmacy', 'Farmacia Catena — Obor',         'PHR-003',
   '{"pharmacy_type": "chain", "city": "București", "county": "București", "address": "Șos. Colentina 2, Sector 2"}',
   'a0000000-0000-0000-0000-000000000005', 'b0000000-0000-0000-0000-000000000001'),

  ('c0000000-0000-0000-0000-000000000022', 'pharmacy', 'Farmacia Help Net — Universitate','PHR-004',
   '{"pharmacy_type": "chain", "city": "București", "county": "București", "address": "Bd. Regina Elisabeta 15, Sector 3"}',
   'a0000000-0000-0000-0000-000000000005', 'b0000000-0000-0000-0000-000000000001'),

  -- Sector 4-6 pharmacies
  ('c0000000-0000-0000-0000-000000000023', 'pharmacy', 'Farmacia Dr. Max — Brâncoveanu',  'PHR-005',
   '{"pharmacy_type": "chain", "city": "București", "county": "București", "address": "Bd. Constantin Brâncoveanu 114, Sector 4"}',
   'a0000000-0000-0000-0000-000000000006', 'b0000000-0000-0000-0000-000000000002'),

  ('c0000000-0000-0000-0000-000000000024', 'pharmacy', 'Farmacia Ropharma — Rahova',      'PHR-006',
   '{"pharmacy_type": "independent", "city": "București", "county": "București", "address": "Calea Rahovei 266, Sector 5"}',
   'a0000000-0000-0000-0000-000000000006', 'b0000000-0000-0000-0000-000000000002'),

  ('c0000000-0000-0000-0000-000000000025', 'pharmacy', 'Farmacia Dona — Crângași',        'PHR-007',
   '{"pharmacy_type": "chain", "city": "București", "county": "București", "address": "Calea Crângași 6, Sector 6"}',
   'a0000000-0000-0000-0000-000000000007', 'b0000000-0000-0000-0000-000000000002'),

  ('c0000000-0000-0000-0000-000000000026', 'pharmacy', 'Farmacia Belladonna — Drumul Taberei', 'PHR-008',
   '{"pharmacy_type": "independent", "city": "București", "county": "București", "address": "Bd. Drumul Taberei 34, Sector 6"}',
   'a0000000-0000-0000-0000-000000000007', 'b0000000-0000-0000-0000-000000000002')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- Activities: Visits (field activities linked to targets)
-- ============================================================
INSERT INTO activities (id, activity_type, status, due_date, duration, routing, fields, target_id, creator_id, joint_visit_user_id, team_id, submitted_at, created_at) VALUES
  -- Sector 1-3: Alina's visits
  ('d0000000-0000-0000-0000-000000000001', 'visit', 'completed',
   CURRENT_DATE - INTERVAL '7 days', 'full_day', 'week_1',
   '{"visit_type": "f2f", "promoted_products": ["product_1"], "feedback": "Good reception, doctor interested in new cardiology line.", "details": "Discussed Q2 samples.", "duration": "full_day"}',
   'c0000000-0000-0000-0000-000000000001',
   'a0000000-0000-0000-0000-000000000004', NULL,
   'b0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '6 days',
   NOW() - INTERVAL '14 days'),

  ('d0000000-0000-0000-0000-000000000002', 'visit', 'completed',
   CURRENT_DATE - INTERVAL '5 days', 'half_day', 'week_1',
   '{"visit_type": "f2f", "promoted_products": ["product_1"], "feedback": "Needs follow-up on dosage info.", "details": "Quick check-in, left brochures.", "duration": "half_day"}',
   'c0000000-0000-0000-0000-000000000002',
   'a0000000-0000-0000-0000-000000000004', NULL,
   'b0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '4 days',
   NOW() - INTERVAL '10 days'),

  ('d0000000-0000-0000-0000-000000000003', 'visit', 'planned',
   CURRENT_DATE + INTERVAL '1 day', 'full_day', 'week_2',
   '{"visit_type": "f2f", "duration": "full_day"}',
   'c0000000-0000-0000-0000-000000000001',
   'a0000000-0000-0000-0000-000000000004', NULL,
   'b0000000-0000-0000-0000-000000000001',
   NULL,
   NOW() - INTERVAL '2 days'),

  -- Sector 1-3: Bogdan's visits
  ('d0000000-0000-0000-0000-000000000004', 'visit', 'completed',
   CURRENT_DATE - INTERVAL '3 days', 'full_day', 'week_1',
   '{"visit_type": "f2f", "promoted_products": ["product_1"], "feedback": "Family medicine doc very receptive, requested samples.", "details": "Presented new product line.", "duration": "full_day"}',
   'c0000000-0000-0000-0000-000000000005',
   'a0000000-0000-0000-0000-000000000005', NULL,
   'b0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '2 days',
   NOW() - INTERVAL '7 days'),

  ('d0000000-0000-0000-0000-000000000005', 'visit', 'planned',
   CURRENT_DATE + INTERVAL '2 days', 'half_day', 'week_2',
   '{"visit_type": "remote", "duration": "half_day"}',
   'c0000000-0000-0000-0000-000000000008',
   'a0000000-0000-0000-0000-000000000005', NULL,
   'b0000000-0000-0000-0000-000000000001',
   NULL,
   NOW() - INTERVAL '1 day'),

  ('d0000000-0000-0000-0000-000000000006', 'visit', 'planned',
   CURRENT_DATE + INTERVAL '3 days', 'full_day', 'week_2',
   '{"visit_type": "f2f", "duration": "full_day"}',
   'c0000000-0000-0000-0000-000000000021',
   'a0000000-0000-0000-0000-000000000005', NULL,
   'b0000000-0000-0000-0000-000000000001',
   NULL,
   NOW() - INTERVAL '1 day'),

  -- Sector 4-6: Camelia's visits
  ('d0000000-0000-0000-0000-000000000007', 'visit', 'completed',
   CURRENT_DATE - INTERVAL '6 days', 'full_day', 'week_1',
   '{"visit_type": "f2f", "promoted_products": ["product_1"], "feedback": "Neurology specialist interested, scheduled follow-up.", "details": "Detailed product presentation.", "duration": "full_day"}',
   'c0000000-0000-0000-0000-000000000010',
   'a0000000-0000-0000-0000-000000000006', NULL,
   'b0000000-0000-0000-0000-000000000002',
   NOW() - INTERVAL '5 days',
   NOW() - INTERVAL '12 days'),

  ('d0000000-0000-0000-0000-000000000008', 'visit', 'cancelled',
   CURRENT_DATE - INTERVAL '2 days', 'half_day', 'week_1',
   '{"visit_type": "f2f", "duration": "half_day"}',
   'c0000000-0000-0000-0000-000000000011',
   'a0000000-0000-0000-0000-000000000006', NULL,
   'b0000000-0000-0000-0000-000000000002',
   NULL,
   NOW() - INTERVAL '8 days'),

  ('d0000000-0000-0000-0000-000000000009', 'visit', 'planned',
   CURRENT_DATE + INTERVAL '2 days', 'full_day', 'week_2',
   '{"visit_type": "f2f", "duration": "full_day"}',
   'c0000000-0000-0000-0000-000000000010',
   'a0000000-0000-0000-0000-000000000006', NULL,
   'b0000000-0000-0000-0000-000000000002',
   NULL,
   NOW() - INTERVAL '1 day'),

  -- Sector 4-6: Daniel's visits
  ('d0000000-0000-0000-0000-000000000010', 'visit', 'completed',
   CURRENT_DATE - INTERVAL '4 days', 'full_day', 'week_1',
   '{"visit_type": "f2f", "promoted_products": ["product_1"], "feedback": "Discussed pulmonology treatments, positive outlook.", "details": "On-site visit with Dr. Georgescu.", "duration": "full_day"}',
   'c0000000-0000-0000-0000-000000000014',
   'a0000000-0000-0000-0000-000000000007', NULL,
   'b0000000-0000-0000-0000-000000000002',
   NOW() - INTERVAL '3 days',
   NOW() - INTERVAL '10 days'),

  ('d0000000-0000-0000-0000-000000000011', 'visit', 'planned',
   CURRENT_DATE + INTERVAL '4 days', 'half_day', 'week_2',
   '{"visit_type": "remote", "duration": "half_day"}',
   'c0000000-0000-0000-0000-000000000015',
   'a0000000-0000-0000-0000-000000000007', NULL,
   'b0000000-0000-0000-0000-000000000002',
   NULL,
   NOW() - INTERVAL '1 day'),

  -- Joint visit: Alina + manager Ioana visiting a high-potential doctor
  ('d0000000-0000-0000-0000-000000000012', 'visit', 'planned',
   CURRENT_DATE + INTERVAL '5 days', 'full_day', 'week_2',
   '{"visit_type": "f2f", "duration": "full_day"}',
   'c0000000-0000-0000-0000-000000000003',
   'a0000000-0000-0000-0000-000000000004',
   'a0000000-0000-0000-0000-000000000002',
   'b0000000-0000-0000-0000-000000000001',
   NULL,
   NOW() - INTERVAL '1 day')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- Activities: Non-field activities
-- ============================================================
INSERT INTO activities (id, activity_type, status, due_date, duration, fields, creator_id, team_id, created_at) VALUES
  -- Administrative
  ('d0000000-0000-0000-0000-000000000013', 'administrative', 'completed',
   CURRENT_DATE - INTERVAL '6 days', 'half_day',
   '{"duration": "half_day", "details": "CRM data entry and report preparation."}',
   'a0000000-0000-0000-0000-000000000004',
   'b0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '10 days'),

  -- Team meeting
  ('d0000000-0000-0000-0000-000000000014', 'team_meeting', 'planned',
   CURRENT_DATE + INTERVAL '2 days', 'half_day',
   '{"duration": "half_day", "details": "Sector 1-3 weekly sync."}',
   'a0000000-0000-0000-0000-000000000002',
   'b0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '3 days'),

  ('d0000000-0000-0000-0000-000000000015', 'team_meeting', 'planned',
   CURRENT_DATE + INTERVAL '2 days', 'half_day',
   '{"duration": "half_day", "details": "Sector 4-6 weekly sync."}',
   'a0000000-0000-0000-0000-000000000003',
   'b0000000-0000-0000-0000-000000000002',
   NOW() - INTERVAL '3 days'),

  -- Training
  ('d0000000-0000-0000-0000-000000000016', 'training', 'planned',
   CURRENT_DATE + INTERVAL '5 days', 'full_day',
   '{"duration": "full_day", "details": "New product launch training — all reps."}',
   'a0000000-0000-0000-0000-000000000002',
   'b0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '5 days'),

  -- Business travel
  ('d0000000-0000-0000-0000-000000000017', 'business_travel', 'completed',
   CURRENT_DATE - INTERVAL '8 days', 'full_day',
   '{"duration": "full_day", "details": "Travel to Sector 3 territory."}',
   'a0000000-0000-0000-0000-000000000005',
   'b0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '12 days'),

  -- Vacation
  ('d0000000-0000-0000-0000-000000000018', 'vacation', 'planned',
   CURRENT_DATE + INTERVAL '14 days', 'full_day',
   '{"duration": "full_day"}',
   'a0000000-0000-0000-0000-000000000006',
   'b0000000-0000-0000-0000-000000000002',
   NOW() - INTERVAL '7 days')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- Audit Log (sample entries for completed activities)
-- ============================================================
INSERT INTO audit_log (id, entity_type, entity_id, event_type, actor_id, old_value, new_value, created_at) VALUES
  -- Activity d...001: created then completed
  ('e0000000-0000-0000-0000-000000000001', 'activity', 'd0000000-0000-0000-0000-000000000001',
   'created', 'a0000000-0000-0000-0000-000000000004', NULL,
   '{"activity_type": "visit", "status": "planned", "target": "Dr. Elena Popescu"}',
   NOW() - INTERVAL '14 days'),

  ('e0000000-0000-0000-0000-000000000002', 'activity', 'd0000000-0000-0000-0000-000000000001',
   'status_changed', 'a0000000-0000-0000-0000-000000000004',
   '{"status": "planned"}', '{"status": "completed"}',
   NOW() - INTERVAL '7 days'),

  ('e0000000-0000-0000-0000-000000000003', 'activity', 'd0000000-0000-0000-0000-000000000001',
   'submitted', 'a0000000-0000-0000-0000-000000000004', NULL,
   '{"feedback": "Good reception, doctor interested in new cardiology line."}',
   NOW() - INTERVAL '6 days'),

  -- Activity d...004: created then completed
  ('e0000000-0000-0000-0000-000000000004', 'activity', 'd0000000-0000-0000-0000-000000000004',
   'created', 'a0000000-0000-0000-0000-000000000005', NULL,
   '{"activity_type": "visit", "status": "planned", "target": "Dr. Ana Dumitrescu"}',
   NOW() - INTERVAL '7 days'),

  ('e0000000-0000-0000-0000-000000000005', 'activity', 'd0000000-0000-0000-0000-000000000004',
   'status_changed', 'a0000000-0000-0000-0000-000000000005',
   '{"status": "planned"}', '{"status": "completed"}',
   NOW() - INTERVAL '3 days'),

  -- Activity d...007: created then completed
  ('e0000000-0000-0000-0000-000000000006', 'activity', 'd0000000-0000-0000-0000-000000000007',
   'created', 'a0000000-0000-0000-0000-000000000006', NULL,
   '{"activity_type": "visit", "status": "planned", "target": "Dr. Maria Stanescu"}',
   NOW() - INTERVAL '12 days'),

  ('e0000000-0000-0000-0000-000000000007', 'activity', 'd0000000-0000-0000-0000-000000000007',
   'status_changed', 'a0000000-0000-0000-0000-000000000006',
   '{"status": "planned"}', '{"status": "completed"}',
   NOW() - INTERVAL '6 days'),

  -- Activity d...008: created then cancelled
  ('e0000000-0000-0000-0000-000000000008', 'activity', 'd0000000-0000-0000-0000-000000000008',
   'created', 'a0000000-0000-0000-0000-000000000006', NULL,
   '{"activity_type": "visit", "status": "planned", "target": "Dr. Cristina Moldovan"}',
   NOW() - INTERVAL '8 days'),

  ('e0000000-0000-0000-0000-000000000009', 'activity', 'd0000000-0000-0000-0000-000000000008',
   'status_changed', 'a0000000-0000-0000-0000-000000000006',
   '{"status": "planned"}', '{"status": "cancelled"}',
   NOW() - INTERVAL '2 days'),

  -- Activity d...010: created then completed
  ('e0000000-0000-0000-0000-000000000010', 'activity', 'd0000000-0000-0000-0000-000000000010',
   'created', 'a0000000-0000-0000-0000-000000000007', NULL,
   '{"activity_type": "visit", "status": "planned", "target": "Dr. Andrei Georgescu"}',
   NOW() - INTERVAL '10 days'),

  ('e0000000-0000-0000-0000-000000000011', 'activity', 'd0000000-0000-0000-0000-000000000010',
   'status_changed', 'a0000000-0000-0000-0000-000000000007',
   '{"status": "planned"}', '{"status": "completed"}',
   NOW() - INTERVAL '4 days'),

  -- Target creation audit entries
  ('e0000000-0000-0000-0000-000000000012', 'target', 'c0000000-0000-0000-0000-000000000001',
   'created', 'a0000000-0000-0000-0000-000000000001', NULL,
   '{"name": "Dr. Elena Popescu", "target_type": "doctor"}',
   NOW() - INTERVAL '30 days'),

  ('e0000000-0000-0000-0000-000000000013', 'target', 'c0000000-0000-0000-0000-000000000019',
   'created', 'a0000000-0000-0000-0000-000000000001', NULL,
   '{"name": "Farmacia Dona — Victoriei", "target_type": "pharmacy"}',
   NOW() - INTERVAL '30 days')
ON CONFLICT (id) DO NOTHING;
