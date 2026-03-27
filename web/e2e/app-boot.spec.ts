import { test, expect } from './coverage-fixture'
import { mockApi } from './fixtures'

test.describe('App boot', () => {
  test.beforeEach(async ({ page }) => {
    await mockApi(page)
  })

  test('loads without crashing and redirects to /planner', async ({ page }) => {
    await page.goto('/')
    // The index route redirects to /planner
    await expect(page).toHaveURL(/\/planner/)
    await expect(page.locator('text=Planning Workspace')).toBeVisible()
  })

  test('renders the sidebar with Pebblr branding', async ({ page }) => {
    await page.goto('/planner')
    const sidebar = page.locator('aside')
    await expect(sidebar.locator('text=Pebblr')).toBeVisible()
    await expect(sidebar.locator('text=v2')).toBeVisible()
  })

  test('does not show a blank white screen on any route', async ({ page }) => {
    const routes = ['/planner', '/targets', '/activities', '/dashboard', '/console', '/audit']
    for (const route of routes) {
      await page.goto(route)
      // The body should have meaningful content — not just an empty div
      const bodyText = await page.locator('body').innerText()
      expect(bodyText.trim().length, `Route ${route} rendered empty`).toBeGreaterThan(0)
    }
  })

  test('sign-in page renders without the sidebar', async ({ page }) => {
    await page.goto('/sign-in')
    // Sign-in page should NOT show the sidebar navigation
    await expect(page.locator('aside')).not.toBeVisible()
    await expect(page.locator('text=Sign in with Microsoft')).toBeVisible()
  })
})
