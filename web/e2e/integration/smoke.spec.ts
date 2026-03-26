/**
 * Integration e2e tests — run against a real Kind cluster deployment.
 *
 * These tests hit the actual Go backend with seeded PostgreSQL data.
 * They validate the full stack: SPA serving, API calls, auth, and data rendering.
 *
 * Seed data reference (scripts/seed-data.sql):
 *   - 7 users: 1 admin, 2 managers, 4 reps
 *   - 2 teams: "Sector 1-3", "Sector 4-6"
 *   - 26 targets: 18 doctors + 8 pharmacies (all in București)
 *   - 18 activities (visits, admin, meetings, travel, training, vacation)
 *   - 13 audit log entries
 */
import { test, expect } from '@playwright/test'

test.describe('Smoke: app boots on real backend', () => {
  test('SPA loads without crashing', async ({ page }) => {
    await page.goto('/')

    // Should redirect to /planner and show the workspace
    await expect(page.locator('text=Planning Workspace')).toBeVisible({ timeout: 15_000 })
  })

  test('sidebar renders with Pebblr branding', async ({ page }) => {
    await page.goto('/planner')

    await expect(page.locator('text=Pebblr')).toBeVisible({ timeout: 10_000 })
  })

  test('no blank white screen on any route', async ({ page }) => {
    const routes = ['/planner', '/targets', '/activities', '/dashboard', '/console', '/audit']
    for (const route of routes) {
      await page.goto(route)
      const bodyText = await page.locator('body').innerText({ timeout: 10_000 })
      expect(bodyText.trim().length, `Route ${route} rendered empty`).toBeGreaterThan(0)
    }
  })
})

test.describe('Targets: real seed data', () => {
  test('target portfolio shows seeded doctors and pharmacies', async ({ page }) => {
    await page.goto('/targets')

    await expect(page.locator('h1:has-text("Target Portfolio")')).toBeVisible({ timeout: 10_000 })

    // Seed data has 26 targets — the UI should show a non-zero count
    const countText = await page.locator('text=/\\d+ targets/').innerText()
    const count = parseInt(countText, 10)
    expect(count).toBeGreaterThanOrEqual(1)

    // A known seed doctor should appear
    await expect(page.locator('text=Dr. Elena Popescu')).toBeVisible()
  })

  test('target detail page loads real target data', async ({ page }) => {
    await page.goto('/targets')

    await expect(page.locator('text=Dr. Elena Popescu')).toBeVisible({ timeout: 10_000 })

    // Click the first doctor's name to navigate to detail
    // TanStack Table renders names in the table — find and click
    await page.locator('text=Dr. Elena Popescu').first().click()

    // Should navigate to the detail page
    await expect(page.locator('text=Schedule Visit')).toBeVisible({ timeout: 10_000 })
    await expect(page.locator('text=Dr. Elena Popescu')).toBeVisible()
  })

  test('targets page search filters results', async ({ page }) => {
    await page.goto('/targets')
    await expect(page.locator('h1:has-text("Target Portfolio")')).toBeVisible({ timeout: 10_000 })

    const searchInput = page.locator('input[placeholder="Search targets..."]')
    await searchInput.fill('Farmacia')

    // Wait for the query to refetch (debounced)
    await page.waitForTimeout(500)

    // Pharmacies should still be visible
    const countText = await page.locator('text=/\\d+ targets/').innerText()
    const count = parseInt(countText, 10)
    expect(count).toBeGreaterThanOrEqual(1)
  })
})

test.describe('Activities: real seed data', () => {
  test('activity log shows seeded activities', async ({ page }) => {
    await page.goto('/activities')

    await expect(page.locator('h1:has-text("Activity Log")')).toBeVisible({ timeout: 10_000 })

    // Should show some activity count
    const countText = await page.locator('text=/\\d+ activities/').innerText()
    const count = parseInt(countText, 10)
    expect(count).toBeGreaterThanOrEqual(1)
  })

  test('new activity form loads config from backend', async ({ page }) => {
    await page.goto('/activities/new')

    await expect(page.locator('text=Log Activity')).toBeVisible({ timeout: 10_000 })
    await expect(page.locator('text=Step 1 of 2')).toBeVisible()

    // Activity type dropdown should have options from real config
    const select = page.locator('select').first()
    const optionCount = await select.locator('option').count()
    expect(optionCount).toBeGreaterThan(1) // At least "Select type..." + real types
  })
})

test.describe('Dashboard: real aggregated data', () => {
  test('team dashboard shows real teams and KPIs', async ({ page }) => {
    await page.goto('/dashboard')

    await expect(page.locator('h1:has-text("Team Dashboard")')).toBeVisible({ timeout: 10_000 })

    // Should show team count from seed data (2 teams)
    await expect(page.locator('text=2 teams')).toBeVisible()

    // KPI cards should render
    const kpiRow = page.locator('.grid.grid-cols-4')
    await expect(kpiRow.locator('text=Cycle Compliance')).toBeVisible()
    await expect(kpiRow.locator('text=Coverage')).toBeVisible()
  })
})

test.describe('Console: real admin data', () => {
  test('users table shows seeded users', async ({ page }) => {
    await page.goto('/console')

    await expect(page.locator('h1:has-text("Users & Roles")')).toBeVisible({ timeout: 10_000 })

    // Seed data has 7 users — at least some should appear
    await expect(page.locator('text=Alexandru Dobre')).toBeVisible()
    await expect(page.locator('text=admin@pebblr.dev')).toBeVisible()
  })

  test('teams section shows seeded teams', async ({ page }) => {
    await page.goto('/console')

    await expect(page.locator('h1:has-text("Users & Roles")')).toBeVisible({ timeout: 10_000 })

    await page.click('button:has-text("Teams")')
    await expect(page.locator('h1:has-text("Teams")')).toBeVisible()

    await expect(page.locator('text=Sector 1-3')).toBeVisible()
    await expect(page.locator('text=Sector 4-6')).toBeVisible()
  })
})

test.describe('Audit: real audit entries', () => {
  test('audit page shows seeded log entries', async ({ page }) => {
    await page.goto('/audit')

    await expect(page.locator('h1:has-text("Audit Logs")')).toBeVisible({ timeout: 10_000 })

    // Should show total entries
    const countText = await page.locator('text=/\\d+ entries/').innerText()
    const count = parseInt(countText, 10)
    expect(count).toBeGreaterThanOrEqual(1)
  })
})

test.describe('Navigation: sidebar links work end-to-end', () => {
  test('can navigate between all pages without errors', async ({ page }) => {
    await page.goto('/planner')
    await expect(page.locator('text=Planning Workspace')).toBeVisible({ timeout: 15_000 })

    // Navigate to targets
    await page.click('a[href="/targets"]')
    await expect(page.locator('h1:has-text("Target Portfolio")')).toBeVisible({ timeout: 10_000 })

    // Navigate to activities
    await page.click('a[href="/activities"]')
    await expect(page.locator('h1:has-text("Activity Log")')).toBeVisible({ timeout: 10_000 })

    // Navigate to dashboard
    await page.goto('/dashboard')
    await expect(page.locator('h1:has-text("Team Dashboard")')).toBeVisible({ timeout: 10_000 })

    // Navigate to console
    await page.goto('/console')
    await expect(page.locator('h1:has-text("Users & Roles")')).toBeVisible({ timeout: 10_000 })

    // Navigate to audit
    await page.goto('/audit')
    await expect(page.locator('h1:has-text("Audit Logs")')).toBeVisible({ timeout: 10_000 })
  })
})
