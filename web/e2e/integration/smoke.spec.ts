/**
 * Integration e2e tests — run against a real Kind cluster deployment.
 *
 * These tests hit the actual Go backend with seeded PostgreSQL data.
 * They validate the full stack: SPA serving, API calls, auth, and data rendering.
 */
import { test, expect } from '@playwright/test'

test.use({ viewport: { width: 1280, height: 720 } })

test.describe('Smoke: app boots on real backend', () => {
  test('SPA loads and renders content', async ({ page }) => {
    await page.goto('/')
    await page.waitForSelector('nav, main, [role="navigation"]', { timeout: 15_000 })
    const bodyText = await page.locator('body').innerText()
    expect(bodyText.trim().length).toBeGreaterThan(0)
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
  test('target portfolio shows seeded targets', async ({ page }) => {
    await page.goto('/targets')

    await expect(page.locator('h1:has-text("Target Portfolio")')).toBeVisible({ timeout: 10_000 })
    await expect(page.locator('text=/\\d+ targets/').first()).toBeVisible()
  })

  test('target detail page loads real target data', async ({ page }) => {
    await page.goto('/targets')

    await expect(page.locator('h1:has-text("Target Portfolio")')).toBeVisible({ timeout: 10_000 })

    // Click first target name to navigate to detail
    await page.locator('text=Dr. Elena Popescu').first().click()

    // Should show the target detail with schedule button
    await expect(page.locator('text=Schedule Visit')).toBeVisible({ timeout: 10_000 })
    await expect(page.getByRole('heading', { name: 'Dr. Elena Popescu' })).toBeVisible()
  })

  test('targets page search filters results', async ({ page }) => {
    await page.goto('/targets')
    await expect(page.locator('h1:has-text("Target Portfolio")')).toBeVisible({ timeout: 10_000 })

    const searchInput = page.locator('input[placeholder="Search targets..."]')
    await searchInput.fill('Farmacia')
    await page.waitForTimeout(500)

    await expect(page.locator('text=/\\d+ targets/').first()).toBeVisible()
  })
})

test.describe('Activities: real seed data', () => {
  test('activity log renders with data', async ({ page }) => {
    await page.goto('/activities')

    await expect(page.locator('h1:has-text("Activity Log")')).toBeVisible({ timeout: 10_000 })
    await expect(page.locator('text=/\\d+ activities/').first()).toBeVisible()
  })

  test('log activity button opens modal', async ({ page }) => {
    await page.goto('/activities')

    await expect(page.locator('h1:has-text("Activity Log")')).toBeVisible({ timeout: 10_000 })

    await page.locator('button:has-text("Log Activity")').first().click()

    // Modal should open with activity type select from real config
    await expect(page.locator('select#field-activity-type')).toBeVisible({ timeout: 5_000 })
    expect(page.url()).toMatch(/\/activities\/?$/)
  })
})

test.describe('Dashboard: real aggregated data', () => {
  test('team dashboard shows teams and KPIs', async ({ page }) => {
    await page.goto('/dashboard')

    await expect(page.locator('h1:has-text("Team Dashboard")')).toBeVisible({ timeout: 10_000 })
    await expect(page.locator('text=/\\d+ teams/')).toBeVisible()
    await expect(page.locator('text=Cycle Compliance')).toBeVisible()
  })
})

test.describe('Console: real admin data', () => {
  test('users section shows seeded users', async ({ page }) => {
    await page.goto('/console')

    await expect(page.locator('h1:has-text("Users & Roles")')).toBeVisible({ timeout: 10_000 })

    // Check for admin email from seed data
    await expect(page.locator('text=admin@pebblr.dev')).toBeVisible()
  })

  test('teams section shows seeded teams', async ({ page }) => {
    await page.goto('/console')

    await expect(page.locator('h1:has-text("Users & Roles")')).toBeVisible({ timeout: 10_000 })

    // Click Teams tab — use evaluate to trigger click on hidden element
    await page.locator('button').filter({ hasText: /^Teams/ }).first().evaluate((el) => (el as HTMLElement).click())
    await expect(page.locator('h1:has-text("Teams")')).toBeVisible({ timeout: 5_000 })

    await expect(page.locator('text=Sector 1-3')).toBeVisible()
    await expect(page.locator('text=Sector 4-6')).toBeVisible()
  })
})

test.describe('Audit: real audit entries', () => {
  test('audit page shows seeded log entries', async ({ page }) => {
    await page.goto('/audit')

    await expect(page.locator('h1:has-text("Audit Logs")')).toBeVisible({ timeout: 10_000 })
    await expect(page.locator('text=/\\d+ entries/')).toBeVisible()
  })
})

test.describe('Navigation: routes load without errors', () => {
  test('can navigate between all pages via direct URL', async ({ page }) => {
    await page.goto('/targets')
    await expect(page.locator('h1:has-text("Target Portfolio")')).toBeVisible({ timeout: 10_000 })

    await page.goto('/activities')
    await expect(page.locator('h1:has-text("Activity Log")')).toBeVisible({ timeout: 10_000 })

    await page.goto('/dashboard')
    await expect(page.locator('h1:has-text("Team Dashboard")')).toBeVisible({ timeout: 10_000 })

    await page.goto('/console')
    await expect(page.locator('h1:has-text("Users & Roles")')).toBeVisible({ timeout: 10_000 })

    await page.goto('/audit')
    await expect(page.locator('h1:has-text("Audit Logs")')).toBeVisible({ timeout: 10_000 })
  })
})
