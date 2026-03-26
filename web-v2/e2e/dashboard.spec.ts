import { test, expect } from '@playwright/test'
import { mockApi } from './fixtures'

test.describe('Dashboard page', () => {
  test.beforeEach(async ({ page }) => {
    await mockApi(page)
  })

  test('renders team dashboard header with team count', async ({ page }) => {
    await page.goto('/dashboard')

    await expect(page.locator('text=Team Dashboard')).toBeVisible()
    await expect(page.locator('text=2 teams')).toBeVisible()
  })

  test('renders KPI stat cards', async ({ page }) => {
    await page.goto('/dashboard')

    const kpiRow = page.locator('.grid.grid-cols-4')
    await expect(kpiRow.locator('text=Cycle Compliance')).toBeVisible()
    await expect(kpiRow.locator('text=Coverage')).toBeVisible()
    await expect(kpiRow.locator('text=Week Progress')).toBeVisible()
    await expect(kpiRow.locator('text=Recovery Balance')).toBeVisible()
  })

  test('renders activity breakdown sections', async ({ page }) => {
    await page.goto('/dashboard')

    await expect(page.locator('text=Activity by Status')).toBeVisible()
    await expect(page.locator('text=Activity by Category')).toBeVisible()
  })

  test('renders frequency compliance table with classification badges', async ({ page }) => {
    await page.goto('/dashboard')

    await expect(page.locator('text=Frequency Compliance by Classification')).toBeVisible()

    // Table headers
    const table = page.locator('table')
    await expect(table.locator('th:has-text("Classification")')).toBeVisible()
    await expect(table.locator('th:has-text("Targets")')).toBeVisible()
    await expect(table.locator('th:has-text("Visits")')).toBeVisible()
    await expect(table.locator('th:has-text("Required")')).toBeVisible()
    await expect(table.locator('th:has-text("Compliance")')).toBeVisible()
  })
})
