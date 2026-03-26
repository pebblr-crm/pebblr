/**
 * Tests for the DemoGate account picker (VITE_DEMO_MODE=true).
 *
 * These tests run against a Vite dev server started with demo mode enabled
 * and NO static token, so the DemoGate component is exercised.
 *
 * All /demo/* and /api/v1/* endpoints are mocked via route interception.
 */
import { test, expect, type Page, type Route } from '@playwright/test'

const DEMO_ACCOUNTS = [
  { id: 'u-rep', name: 'Riley Rep', email: 'rep@demo.pebblr.com', role: 'rep' },
  { id: 'u-mgr', name: 'Morgan Manager', email: 'mgr@demo.pebblr.com', role: 'manager' },
  { id: 'u-adm', name: 'Alex Admin', email: 'adm@demo.pebblr.com', role: 'admin' },
]

async function mockDemoEndpoints(page: Page) {
  // Mock /demo/accounts — returns bare array (matches Go handler)
  await page.route('**/demo/accounts', (route: Route) => {
    return route.fulfill({ json: DEMO_ACCOUNTS })
  })

  // Mock /demo/token — returns JWT + account
  await page.route('**/demo/token', async (route: Route) => {
    const body = JSON.parse(route.request().postData() ?? '{}')
    const acct = DEMO_ACCOUNTS.find((a) => a.id === body.user_id) ?? DEMO_ACCOUNTS[0]
    return route.fulfill({
      json: {
        token: 'demo-jwt-token-for-testing',
        account: acct,
      },
    })
  })

  // Mock all API endpoints so the app works after login
  await page.route('**/api/v1/**', (route: Route) => {
    const url = new URL(route.request().url())
    const path = url.pathname.replace('/api/v1', '')

    if (path.startsWith('/targets')) return route.fulfill({ json: { items: [], total: 0, page: 1 } })
    if (path.startsWith('/activities')) return route.fulfill({ json: { items: [], total: 0, page: 1 } })
    if (path.startsWith('/dashboard')) return route.fulfill({ json: { total: 0, byStatus: {}, byCategory: {} } })
    if (path.startsWith('/users')) return route.fulfill({ json: { items: [], total: 0, page: 1 } })
    if (path.startsWith('/teams')) return route.fulfill({ json: { items: [], total: 0, page: 1 } })
    if (path.startsWith('/territories')) return route.fulfill({ json: { items: [], total: 0, page: 1 } })
    if (path.startsWith('/config')) {
      return route.fulfill({
        json: {
          tenant: { name: 'Demo', locale: 'en' },
          accounts: { types: [] },
          options: {},
          rules: { frequency: {}, max_activities_per_day: 8, default_visit_duration_minutes: {}, visit_duration_step_minutes: 15 },
          activities: { types: [], statuses: [], status_transitions: {}, durations: [], routing_options: [] },
        },
      })
    }
    if (path.startsWith('/audit')) return route.fulfill({ json: { items: [], total: 0, page: 1 } })
    if (path === '/me') {
      return route.fulfill({
        json: { id: 'u-adm', displayName: 'Alex Admin', email: 'adm@demo.pebblr.com', role: 'admin' },
      })
    }
    return route.fulfill({ json: {} })
  })
}

test.describe('Demo account picker', () => {
  test.beforeEach(async ({ page }) => {
    await mockDemoEndpoints(page)
  })

  test('shows account picker with all demo accounts', async ({ page }) => {
    await page.goto('/')

    await expect(page.locator('text=Pebblr v2 Demo')).toBeVisible()
    await expect(page.locator('text=Select a demo account to continue.')).toBeVisible()

    // All 3 accounts should be listed
    for (const acct of DEMO_ACCOUNTS) {
      await expect(page.locator(`text=${acct.name}`)).toBeVisible()
      await expect(page.locator(`text=${acct.email}`)).toBeVisible()
    }
  })

  test('does not show "Loading accounts..." when accounts load', async ({ page }) => {
    await page.goto('/')

    await expect(page.locator('text=Pebblr v2 Demo')).toBeVisible()
    // Wait for accounts to load — "Loading accounts..." should disappear
    await expect(page.locator('text=Riley Rep')).toBeVisible()
    await expect(page.locator('text=Loading accounts...')).not.toBeVisible()
  })

  test('clicking an account logs in and shows the app', async ({ page }) => {
    await page.goto('/')

    await expect(page.locator('text=Alex Admin')).toBeVisible()
    await page.click('text=Alex Admin')

    // After login, the app shell should appear (redirects to /planner)
    await expect(page.locator('text=Pebblr')).toBeVisible({ timeout: 10_000 })
  })

  test('handles /demo/accounts returning 500 gracefully', async ({ page }) => {
    // Override the accounts mock to return an error
    await page.route('**/demo/accounts', (route: Route) => {
      return route.fulfill({ status: 500, json: { error: { code: 'LIST_ERROR', message: 'db down' } } })
    })

    await page.goto('/')

    await expect(page.locator('text=Pebblr v2 Demo')).toBeVisible()
    // Should show fallback, not crash
    await expect(page.locator('text=Loading accounts...')).toBeVisible()
  })

  test('handles /demo/accounts returning unexpected shape gracefully', async ({ page }) => {
    // Override to return the old wrong shape {accounts: [...]}
    await page.route('**/demo/accounts', (route: Route) => {
      return route.fulfill({ json: { accounts: DEMO_ACCOUNTS } })
    })

    await page.goto('/')

    await expect(page.locator('text=Pebblr v2 Demo')).toBeVisible()
    // Should not crash — falls back to empty array
    await expect(page.locator('text=Loading accounts...')).toBeVisible()
  })
})
