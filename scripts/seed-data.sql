-- Seed data for local development
-- Pharmaceutical field-sales CRM: users, teams, targets (doctors/pharmacies), activities.
-- Uses fixed UUIDs for deterministic, idempotent inserts.

-- ============================================================
-- Users
-- ============================================================
INSERT INTO users (id, external_id, email, name, role, avatar, online_status) VALUES
  ('a0000000-0000-0000-0000-000000000001', 'oid-admin-001',   'admin@pebblr.dev',     'Alex Admin',    'admin',   '', 'online'),
  ('a0000000-0000-0000-0000-000000000002', 'oid-mgr-001',     'mgr.north@pebblr.dev', 'Morgan North',  'manager', '', 'online'),
  ('a0000000-0000-0000-0000-000000000003', 'oid-mgr-002',     'mgr.south@pebblr.dev', 'Sam South',     'manager', '', 'away'),
  ('a0000000-0000-0000-0000-000000000004', 'oid-rep-001',     'rep.alice@pebblr.dev', 'Alice Reyes',   'rep',     '', 'online'),
  ('a0000000-0000-0000-0000-000000000005', 'oid-rep-002',     'rep.bob@pebblr.dev',   'Bob Tran',      'rep',     '', 'online'),
  ('a0000000-0000-0000-0000-000000000006', 'oid-rep-003',     'rep.carol@pebblr.dev', 'Carol Kim',     'rep',     '', 'offline'),
  ('a0000000-0000-0000-0000-000000000007', 'oid-rep-004',     'rep.dan@pebblr.dev',   'Dan Osei',      'rep',     '', 'online')
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
-- Targets: Doctors
-- ============================================================
INSERT INTO targets (id, target_type, name, external_id, fields, assignee_id, team_id) VALUES
  -- North Region doctors
  ('c0000000-0000-0000-0000-000000000001', 'doctor', 'Dr. Elena Popescu',   'DOC-001',
   '{"specialty": "cardiology", "potential": "a", "city": "Portland", "county": "Multnomah", "address": "123 Heart Ave"}',
   'a0000000-0000-0000-0000-000000000004', 'b0000000-0000-0000-0000-000000000001'),

  ('c0000000-0000-0000-0000-000000000002', 'doctor', 'Dr. Mihai Ionescu',   'DOC-002',
   '{"specialty": "internal_medicine", "potential": "b", "city": "Seattle", "county": "King", "address": "456 Clinic Rd"}',
   'a0000000-0000-0000-0000-000000000004', 'b0000000-0000-0000-0000-000000000001'),

  ('c0000000-0000-0000-0000-000000000003', 'doctor', 'Dr. Ana Dumitrescu',  'DOC-003',
   '{"specialty": "family_medicine", "potential": "a", "city": "Tacoma", "county": "Pierce", "address": "789 Family Blvd"}',
   'a0000000-0000-0000-0000-000000000005', 'b0000000-0000-0000-0000-000000000001'),

  ('c0000000-0000-0000-0000-000000000004', 'doctor', 'Dr. Radu Constantinescu', 'DOC-004',
   '{"specialty": "gastroenterology", "potential": "c", "city": "Portland", "county": "Multnomah", "address": "321 Gastro St"}',
   'a0000000-0000-0000-0000-000000000005', 'b0000000-0000-0000-0000-000000000001'),

  -- South Region doctors
  ('c0000000-0000-0000-0000-000000000005', 'doctor', 'Dr. Maria Stanescu',  'DOC-005',
   '{"specialty": "neurology", "potential": "a", "city": "San Francisco", "county": "San Francisco", "address": "100 Neuro Way"}',
   'a0000000-0000-0000-0000-000000000006', 'b0000000-0000-0000-0000-000000000002'),

  ('c0000000-0000-0000-0000-000000000006', 'doctor', 'Dr. Andrei Georgescu','DOC-006',
   '{"specialty": "pulmonology", "potential": "b", "city": "Eugene", "county": "Lane", "address": "200 Lung Ct"}',
   'a0000000-0000-0000-0000-000000000007', 'b0000000-0000-0000-0000-000000000002'),

  ('c0000000-0000-0000-0000-000000000007', 'doctor', 'Dr. Cristina Moldovan','DOC-007',
   '{"specialty": "pediatrics", "potential": "b", "city": "Bend", "county": "Deschutes", "address": "50 Kids Ln"}',
   'a0000000-0000-0000-0000-000000000006', 'b0000000-0000-0000-0000-000000000002'),

  ('c0000000-0000-0000-0000-000000000008', 'doctor', 'Dr. Ion Petrescu',    'DOC-008',
   '{"specialty": "geriatrics", "potential": "c", "city": "Las Vegas", "county": "Clark", "address": "88 Senior Dr"}',
   'a0000000-0000-0000-0000-000000000007', 'b0000000-0000-0000-0000-000000000002')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- Targets: Pharmacies
-- ============================================================
INSERT INTO targets (id, target_type, name, external_id, fields, assignee_id, team_id) VALUES
  -- North Region pharmacies
  ('c0000000-0000-0000-0000-000000000009', 'pharmacy', 'Greenfield Pharmacy',     'PHR-001',
   '{"pharmacy_type": "chain", "city": "Portland", "county": "Multnomah", "address": "10 Main St"}',
   'a0000000-0000-0000-0000-000000000004', 'b0000000-0000-0000-0000-000000000001'),

  ('c0000000-0000-0000-0000-000000000010', 'pharmacy', 'Pacific Health Pharmacy', 'PHR-002',
   '{"pharmacy_type": "chain", "city": "Seattle", "county": "King", "address": "22 Wellness Ave"}',
   'a0000000-0000-0000-0000-000000000005', 'b0000000-0000-0000-0000-000000000001'),

  -- South Region pharmacies
  ('c0000000-0000-0000-0000-000000000011', 'pharmacy', 'Sunrise Pharmacy',        'PHR-003',
   '{"pharmacy_type": "chain", "city": "San Francisco", "county": "San Francisco", "address": "5 Bay Rd"}',
   'a0000000-0000-0000-0000-000000000006', 'b0000000-0000-0000-0000-000000000002'),

  ('c0000000-0000-0000-0000-000000000012', 'pharmacy', 'Metro Drug Store',        'PHR-004',
   '{"pharmacy_type": "chain", "city": "Sacramento", "county": "Sacramento", "address": "99 Capitol Blvd"}',
   'a0000000-0000-0000-0000-000000000007', 'b0000000-0000-0000-0000-000000000002')
ON CONFLICT (id) DO NOTHING;

-- ============================================================
-- Activities: Visits (field activities linked to targets)
-- ============================================================
INSERT INTO activities (id, activity_type, status, due_date, duration, routing, fields, target_id, creator_id, joint_visit_user_id, team_id, submitted_at, created_at) VALUES
  -- North Region: Alice's visits
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

  -- North Region: Bob's visits
  ('d0000000-0000-0000-0000-000000000004', 'visit', 'completed',
   CURRENT_DATE - INTERVAL '3 days', 'full_day', 'week_1',
   '{"visit_type": "f2f", "promoted_products": ["product_1"], "feedback": "Family medicine doc very receptive, requested samples.", "details": "Presented new product line.", "duration": "full_day"}',
   'c0000000-0000-0000-0000-000000000003',
   'a0000000-0000-0000-0000-000000000005', NULL,
   'b0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '2 days',
   NOW() - INTERVAL '7 days'),

  ('d0000000-0000-0000-0000-000000000005', 'visit', 'planned',
   CURRENT_DATE + INTERVAL '2 days', 'half_day', 'week_2',
   '{"visit_type": "remote", "duration": "half_day"}',
   'c0000000-0000-0000-0000-000000000004',
   'a0000000-0000-0000-0000-000000000005', NULL,
   'b0000000-0000-0000-0000-000000000001',
   NULL,
   NOW() - INTERVAL '1 day'),

  ('d0000000-0000-0000-0000-000000000006', 'visit', 'planned',
   CURRENT_DATE + INTERVAL '3 days', 'full_day', 'week_2',
   '{"visit_type": "f2f", "duration": "full_day"}',
   'c0000000-0000-0000-0000-000000000010',
   'a0000000-0000-0000-0000-000000000005', NULL,
   'b0000000-0000-0000-0000-000000000001',
   NULL,
   NOW() - INTERVAL '1 day'),

  -- South Region: Carol's visits
  ('d0000000-0000-0000-0000-000000000007', 'visit', 'completed',
   CURRENT_DATE - INTERVAL '6 days', 'full_day', 'week_1',
   '{"visit_type": "f2f", "promoted_products": ["product_1"], "feedback": "Neurology specialist interested, scheduled follow-up.", "details": "Detailed product presentation.", "duration": "full_day"}',
   'c0000000-0000-0000-0000-000000000005',
   'a0000000-0000-0000-0000-000000000006', NULL,
   'b0000000-0000-0000-0000-000000000002',
   NOW() - INTERVAL '5 days',
   NOW() - INTERVAL '12 days'),

  ('d0000000-0000-0000-0000-000000000008', 'visit', 'cancelled',
   CURRENT_DATE - INTERVAL '2 days', 'half_day', 'week_1',
   '{"visit_type": "f2f", "duration": "half_day"}',
   'c0000000-0000-0000-0000-000000000007',
   'a0000000-0000-0000-0000-000000000006', NULL,
   'b0000000-0000-0000-0000-000000000002',
   NULL,
   NOW() - INTERVAL '8 days'),

  ('d0000000-0000-0000-0000-000000000009', 'visit', 'planned',
   CURRENT_DATE + INTERVAL '2 days', 'full_day', 'week_2',
   '{"visit_type": "f2f", "duration": "full_day"}',
   'c0000000-0000-0000-0000-000000000005',
   'a0000000-0000-0000-0000-000000000006', NULL,
   'b0000000-0000-0000-0000-000000000002',
   NULL,
   NOW() - INTERVAL '1 day'),

  -- South Region: Dan's visits
  ('d0000000-0000-0000-0000-000000000010', 'visit', 'completed',
   CURRENT_DATE - INTERVAL '4 days', 'full_day', 'week_1',
   '{"visit_type": "f2f", "promoted_products": ["product_1"], "feedback": "Discussed pulmonology treatments, positive outlook.", "details": "On-site visit with Dr. Georgescu.", "duration": "full_day"}',
   'c0000000-0000-0000-0000-000000000006',
   'a0000000-0000-0000-0000-000000000007', NULL,
   'b0000000-0000-0000-0000-000000000002',
   NOW() - INTERVAL '3 days',
   NOW() - INTERVAL '10 days'),

  ('d0000000-0000-0000-0000-000000000011', 'visit', 'planned',
   CURRENT_DATE + INTERVAL '4 days', 'half_day', 'week_2',
   '{"visit_type": "remote", "duration": "half_day"}',
   'c0000000-0000-0000-0000-000000000008',
   'a0000000-0000-0000-0000-000000000007', NULL,
   'b0000000-0000-0000-0000-000000000002',
   NULL,
   NOW() - INTERVAL '1 day'),

  -- Joint visit: Alice + manager Morgan visiting a high-potential doctor
  ('d0000000-0000-0000-0000-000000000012', 'visit', 'planned',
   CURRENT_DATE + INTERVAL '5 days', 'full_day', 'week_2',
   '{"visit_type": "f2f", "visit_partner": "Morgan North", "duration": "full_day"}',
   'c0000000-0000-0000-0000-000000000001',
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
   '{"duration": "half_day", "details": "North Region weekly sync."}',
   'a0000000-0000-0000-0000-000000000002',
   'b0000000-0000-0000-0000-000000000001',
   NOW() - INTERVAL '3 days'),

  ('d0000000-0000-0000-0000-000000000015', 'team_meeting', 'planned',
   CURRENT_DATE + INTERVAL '2 days', 'half_day',
   '{"duration": "half_day", "details": "South Region weekly sync."}',
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
   '{"duration": "full_day", "details": "Travel to Portland territory."}',
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

  ('e0000000-0000-0000-0000-000000000013', 'target', 'c0000000-0000-0000-0000-000000000009',
   'created', 'a0000000-0000-0000-0000-000000000001', NULL,
   '{"name": "Greenfield Pharmacy", "target_type": "pharmacy"}',
   NOW() - INTERVAL '30 days')
ON CONFLICT (id) DO NOTHING;
