/**
 * Shared API mock fixtures for web-v2 e2e tests.
 *
 * All tests mock /api/v1/* so they run without a backend.
 * The static-token auth flow is used (VITE_STATIC_TOKEN env var),
 * so the app boots directly into the authenticated shell.
 */
import { type Page, type Route } from '@playwright/test'

// ── Seed data ────────────────────────────────────────────────────────────────

export const TARGETS = {
  items: [
    {
      id: 't-001',
      name: 'Dr. Elena Popescu',
      targetType: 'doctor',
      fields: { classification: 'A', city: 'Bucharest', address: '12 Strada Victoriei', lat: 44.4268, lng: 26.1025, key_contact: 'Elena Popescu' },
    },
    {
      id: 't-002',
      name: 'Farmacia Central',
      targetType: 'pharmacy',
      fields: { classification: 'B', city: 'Cluj-Napoca', address: '5 Bulevardul Eroilor', lat: 46.7712, lng: 23.6236 },
    },
    {
      id: 't-003',
      name: 'Spitalul Judetean',
      targetType: 'hospital',
      fields: { classification: 'C', city: 'Timisoara', address: '1 Piata Unirii' },
    },
  ],
  total: 3,
  page: 1,
}

export const ACTIVITIES = {
  items: [
    {
      id: 'a-001',
      activityType: 'visit',
      status: 'realizat',
      dueDate: '2026-03-25',
      duration: 'full_day',
      targetName: 'Dr. Elena Popescu',
      label: null,
      creatorId: 'u-001',
      submittedAt: '2026-03-25T18:00:00Z',
      createdAt: '2026-03-25T08:00:00Z',
      updatedAt: '2026-03-25T18:00:00Z',
      fields: { notes: 'Good discussion about new treatment.', tags: ['Left Samples'] },
    },
    {
      id: 'a-002',
      activityType: 'visit',
      status: 'planificat',
      dueDate: '2026-03-26',
      duration: 'half_day',
      targetName: 'Farmacia Central',
      label: null,
      creatorId: 'u-001',
      submittedAt: null,
      createdAt: '2026-03-26T07:00:00Z',
      updatedAt: '2026-03-26T07:00:00Z',
      fields: { tags: [] },
    },
    {
      id: 'a-003',
      activityType: 'administrative',
      status: 'anulat',
      dueDate: '2026-03-24',
      duration: 'full_day',
      targetName: null,
      label: 'Team meeting',
      creatorId: 'u-001',
      submittedAt: null,
      createdAt: '2026-03-24T09:00:00Z',
      updatedAt: '2026-03-24T09:00:00Z',
      fields: {},
    },
  ],
  total: 3,
  page: 1,
}

export const ACTIVITY_STATS = {
  total: 12,
  byStatus: { realizat: 8, planificat: 3, anulat: 1 },
  byCategory: { field: 10, other: 2 },
}

export const COVERAGE = {
  percentage: 72,
  visitedTargets: 18,
  totalTargets: 25,
}

export const FREQUENCY = {
  items: [
    { classification: 'A', targetCount: 5, totalVisits: 20, required: 25, compliance: 80 },
    { classification: 'B', targetCount: 10, totalVisits: 15, required: 20, compliance: 75 },
    { classification: 'C', targetCount: 10, totalVisits: 5, required: 10, compliance: 50 },
  ],
}

export const RECOVERY_BALANCE = {
  balance: 3,
  earned: 5,
  taken: 2,
}

export const USERS = {
  items: [
    { id: 'u-001', displayName: 'Ana Ionescu', email: 'ana@pebblr.dev', role: 'rep' },
    { id: 'u-002', displayName: 'Mihai Radu', email: 'mihai@pebblr.dev', role: 'manager' },
    { id: 'u-003', displayName: 'Admin Dev', email: 'admin@pebblr.dev', role: 'admin' },
  ],
  total: 3,
  page: 1,
}

export const TEAMS = {
  items: [
    { id: 'team-1', name: 'Alpha', managerId: 'u-002' },
    { id: 'team-2', name: 'Bravo', managerId: 'u-002' },
  ],
  total: 2,
  page: 1,
}

export const TERRITORIES = {
  items: [
    { id: 'ter-1', name: 'Bucharest Metro', region: 'South', teamId: 'team-1' },
    { id: 'ter-2', name: 'Transylvania', region: 'North', teamId: 'team-2' },
  ],
  total: 2,
  page: 1,
}

export const CONFIG = {
  tenant: { name: 'Pebblr Demo', locale: 'en' },
  accounts: {
    types: [
      {
        key: 'doctor',
        label: 'Doctor',
        fields: [
          { key: 'specialty', label: 'Specialty', type: 'text', required: false },
          { key: 'classification', label: 'Priority', type: 'select', required: true, options: ['A', 'B', 'C'] },
        ],
      },
      {
        key: 'pharmacy',
        label: 'Pharmacy',
        fields: [
          { key: 'chain', label: 'Chain', type: 'text', required: false },
          { key: 'classification', label: 'Priority', type: 'select', required: true, options: ['A', 'B', 'C'] },
        ],
      },
      {
        key: 'hospital',
        label: 'Hospital',
        fields: [
          { key: 'classification', label: 'Priority', type: 'select', required: true, options: ['A', 'B', 'C'] },
        ],
      },
    ],
  },
  options: {},
  rules: {
    frequency: { A: 5, B: 3, C: 1 },
    max_activities_per_day: 8,
    visit_cadence_days: 14,
    default_visit_duration_minutes: { doctor: 30, pharmacy: 20, hospital: 45 },
    visit_duration_step_minutes: 15,
  },
  activities: {
    types: [
      { key: 'visit', label: 'Visit', category: 'field', fields: [], blocks_field_activities: false },
      { key: 'joint_visit', label: 'Joint Visit', category: 'field', fields: [], blocks_field_activities: false },
      { key: 'recovery', label: 'Recovery', category: 'other', fields: [], blocks_field_activities: true },
    ],
    statuses: [
      { key: 'planificat', label: 'Planned', initial: true, submittable: false },
      { key: 'realizat', label: 'Completed', initial: false, submittable: true },
      { key: 'anulat', label: 'Cancelled', initial: false, submittable: false },
    ],
    status_transitions: {
      planificat: ['realizat', 'anulat'],
      realizat: [],
      anulat: [],
    },
    durations: [
      { key: 'full_day', label: 'Full Day' },
      { key: 'half_day', label: 'Half Day' },
    ],
    routing_options: [],
  },
}

export const AUDIT_ENTRIES = {
  items: [
    { id: 'aud-1', createdAt: '2026-03-25T10:30:00Z', actorId: 'u-001-xxxx-xxxx-xxxx', entityType: 'activity', eventType: 'activity_created', status: 'pending' },
    { id: 'aud-2', createdAt: '2026-03-24T14:00:00Z', actorId: 'u-002-xxxx-xxxx-xxxx', entityType: 'target', eventType: 'target_updated', status: 'accepted' },
    { id: 'aud-3', createdAt: '2026-03-23T09:15:00Z', actorId: 'u-003-xxxx-xxxx-xxxx', entityType: 'user', eventType: 'role_changed', status: 'false_positive' },
  ],
  total: 3,
  page: 1,
}

export const VISIT_STATUS = {
  items: [
    { targetId: 't-001', lastVisitDate: '2026-03-25' },
    { targetId: 't-002', lastVisitDate: '2026-03-20' },
  ],
}

export const FREQUENCY_STATUS = {
  items: [
    { targetId: 't-001', compliance: 90 },
    { targetId: 't-002', compliance: 60 },
    { targetId: 't-003', compliance: 30 },
  ],
}

// ── Route handler ────────────────────────────────────────────────────────────

/**
 * Intercept all /api/v1/* requests and return mock JSON.
 * Call this in beforeEach to give every test a working backend.
 */
export async function mockApi(page: Page) {
  await page.route('**/api/v1/**', (route: Route) => {
    const url = new URL(route.request().url())
    const path = url.pathname.replace('/api/v1', '')
    const method = route.request().method()

    // POST endpoints
    if (method === 'POST') {
      if (path === '/activities') {
        return route.fulfill({ status: 201, json: { id: 'a-new', ...ACTIVITIES.items[1] } })
      }
      if (path === '/activities/clone-week') {
        return route.fulfill({ status: 200, json: { cloned: 5 } })
      }
      return route.fulfill({ status: 200, json: {} })
    }

    // PATCH endpoints
    if (method === 'PATCH') {
      if (path.match(/^\/audit\/[^/]+\/status$/)) {
        return route.fulfill({ status: 200, json: { status: 'accepted' } })
      }
      return route.fulfill({ status: 200, json: {} })
    }

    // GET endpoints
    if (path === '/targets' || path === '/targets/') {
      const id = url.searchParams.get('id')
      if (id) {
        const t = TARGETS.items.find((t) => t.id === id)
        return route.fulfill({ json: t ?? TARGETS.items[0] })
      }
      return route.fulfill({ json: TARGETS })
    }
    if (path.match(/^\/targets\/[^/]+$/) && !path.includes('visit-status') && !path.includes('frequency-status') && !path.includes('coverage')) {
      const id = path.split('/').pop()
      const t = TARGETS.items.find((t) => t.id === id)
      return route.fulfill({ json: { target: t ?? TARGETS.items[0] } })
    }
    if (path === '/targets/visit-status') return route.fulfill({ json: VISIT_STATUS })
    if (path === '/targets/frequency-status') return route.fulfill({ json: FREQUENCY_STATUS })
    if (path === '/targets/coverage') return route.fulfill({ json: COVERAGE })
    if (path === '/activities' || path === '/activities/') return route.fulfill({ json: ACTIVITIES })
    if (path.match(/^\/activities\/[^/]+$/) && !path.includes('stats') && !path.includes('frequency') && !path.includes('recovery') && !path.includes('clone')) {
      const id = path.split('/').pop()
      const a = ACTIVITIES.items.find((a) => a.id === id)
      return route.fulfill({ json: { activity: a ?? ACTIVITIES.items[0] } })
    }
    if (path === '/activities/stats') return route.fulfill({ json: ACTIVITY_STATS })
    if (path === '/activities/frequency') return route.fulfill({ json: FREQUENCY })
    if (path === '/activities/recovery-balance') return route.fulfill({ json: RECOVERY_BALANCE })
    // Dashboard endpoints (used by useDashboard hooks)
    if (path === '/dashboard/activities') return route.fulfill({ json: ACTIVITY_STATS })
    if (path === '/dashboard/coverage') return route.fulfill({ json: COVERAGE })
    if (path === '/dashboard/frequency') return route.fulfill({ json: FREQUENCY })
    if (path === '/dashboard/recovery') return route.fulfill({ json: RECOVERY_BALANCE })
    if (path === '/users' || path === '/users/') return route.fulfill({ json: USERS })
    if (path === '/teams' || path === '/teams/') return route.fulfill({ json: TEAMS })
    if (path === '/territories' || path === '/territories/') return route.fulfill({ json: TERRITORIES })
    if (path === '/config' || path === '/config/') return route.fulfill({ json: CONFIG })
    if (path === '/audit' || path === '/audit/') return route.fulfill({ json: AUDIT_ENTRIES })

    // Fallback: 404
    return route.fulfill({ status: 404, json: { error: { code: 'NOT_FOUND', message: `no mock for ${path}` } } })
  })
}
