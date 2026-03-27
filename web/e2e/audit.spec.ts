import { test, expect } from './coverage-fixture'
import { mockApi } from './fixtures'

test.describe('Audit page', () => {
  test.beforeEach(async ({ page }) => {
    await mockApi(page)
  })

  test('renders audit logs header and description', async ({ page }) => {
    await page.goto('/audit')

    await expect(page.locator('h1:has-text("Audit Logs")')).toBeVisible()
    await expect(page.locator('text=Immutable change history and review workflow')).toBeVisible()
  })

  test('shows pending review badge when entries are pending', async ({ page }) => {
    await page.goto('/audit')

    await expect(page.locator('text=1 pending review')).toBeVisible()
  })

  test('renders audit table with correct columns', async ({ page }) => {
    await page.goto('/audit')

    // Table headers
    await expect(page.locator('th:has-text("Timestamp")')).toBeVisible()
    await expect(page.locator('th:has-text("Actor")')).toBeVisible()
    await expect(page.locator('th:has-text("Entity")')).toBeVisible()
    await expect(page.locator('th:has-text("Action")')).toBeVisible()
    await expect(page.locator('th:has-text("Status")')).toBeVisible()
    await expect(page.locator('th:has-text("Review")')).toBeVisible()
  })

  test('shows entry count from filters', async ({ page }) => {
    await page.goto('/audit')

    await expect(page.locator('text=3 entries')).toBeVisible()
  })

  test('filter dropdowns are present', async ({ page }) => {
    await page.goto('/audit')

    // Entity type filter
    const entitySelect = page.locator('select').first()
    const entityOptions = entitySelect.locator('option')
    await expect(entityOptions).toHaveCount(4) // All + Activity + Target + User

    // Status filter
    const statusSelect = page.locator('select').nth(1)
    const statusOptions = statusSelect.locator('option')
    await expect(statusOptions).toHaveCount(4) // All + Pending + Accepted + False Positive
  })

  test('export button is present and disabled', async ({ page }) => {
    await page.goto('/audit')

    const exportBtn = page.locator('button:has-text("Export Logs")')
    await expect(exportBtn).toBeVisible()
    await expect(exportBtn).toBeDisabled()
  })
})
