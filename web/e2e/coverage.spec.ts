import { test, expect } from './coverage-fixture'
import { mockApi } from './fixtures'

test.describe('Coverage page', () => {
  test.beforeEach(async ({ page }) => {
    await mockApi(page)
  })

  test('renders filter sidebar with team and priority filters', async ({ page }) => {
    await page.goto('/coverage')

    const filterSidebar = page.locator('.md\\:static.w-72')
    await expect(filterSidebar.locator('text=Filters')).toBeVisible()
    await expect(filterSidebar.locator('text=Reset')).toBeVisible()

    // Team names from fixtures
    await expect(filterSidebar.locator('text=Alpha')).toBeVisible()
    await expect(filterSidebar.locator('text=Bravo')).toBeVisible()

    // Priority filter buttons
    await expect(filterSidebar.locator('button:has-text("A")')).toBeVisible()
    await expect(filterSidebar.locator('button:has-text("B")')).toBeVisible()
    await expect(filterSidebar.locator('button:has-text("C")')).toBeVisible()
  })

  test('shows coverage summary card with percentage', async ({ page }) => {
    await page.goto('/coverage')

    const sidebar = page.locator('.w-72')
    await expect(sidebar.locator('text=72%')).toBeVisible()
    await expect(sidebar.locator('text=18 of 25 visited')).toBeVisible()
  })

  test('lists territories', async ({ page }) => {
    await page.goto('/coverage')

    await expect(page.locator('text=Bucharest Metro')).toBeVisible()
    await expect(page.locator('text=Transylvania')).toBeVisible()
  })

  test('shows pin count', async ({ page }) => {
    await page.goto('/coverage')

    // 2 targets have coordinates in fixtures
    const filterSidebar = page.locator('.md\\:static.w-72')
    await expect(filterSidebar.locator('text=2 pins')).toBeVisible()
  })
})
