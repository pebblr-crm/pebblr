import { test, expect } from '@playwright/test'
import { mockApi } from './fixtures'

test.describe('Sidebar navigation', () => {
  test.beforeEach(async ({ page }) => {
    await mockApi(page)
  })

  test('navigating between rep pages updates content', async ({ page }) => {
    await page.goto('/planner')
    await expect(page.locator('text=Planning Workspace')).toBeVisible()

    await page.click('a[href="/targets"]')
    await expect(page).toHaveURL(/\/targets/)
    await expect(page.locator('text=Target Portfolio')).toBeVisible()

    await page.click('a[href="/activities"]')
    await expect(page).toHaveURL(/\/activities/)
    await expect(page.locator('text=Activity Log')).toBeVisible()
  })

  test('navigating to manager pages renders correctly', async ({ page }) => {
    await page.goto('/dashboard')
    await expect(page.locator('text=Team Dashboard')).toBeVisible()

    await page.goto('/coverage')
    await expect(page.locator('.md\\:static.w-72').locator('text=Filters')).toBeVisible()
  })

  test('navigating to admin pages renders correctly', async ({ page }) => {
    await page.goto('/console')
    await expect(page.locator('h1:has-text("Users & Roles")')).toBeVisible()

    await page.goto('/audit')
    await expect(page.locator('h1:has-text("Audit Logs")')).toBeVisible()
  })

  test('sidebar shows "Switch to v1" link', async ({ page }) => {
    await page.goto('/planner')
    await expect(page.locator('a[href="/?ui=v1"]')).toBeVisible()
  })
})
