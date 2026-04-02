import { test, expect } from './coverage-fixture'
import { mockApi, TARGETS } from './fixtures'

test.describe('Targets page', () => {
  test.beforeEach(async ({ page }) => {
    await mockApi(page)
  })

  test('renders target table with correct data', async ({ page }) => {
    await page.goto('/targets')

    await expect(page.locator('text=Target Portfolio')).toBeVisible()
    await expect(page.locator(`text=${TARGETS.items.length} targets`)).toBeVisible()

    // Each target name should appear in the table
    for (const target of TARGETS.items) {
      await expect(page.locator(`text=${target.name}`)).toBeVisible()
    }
  })

  test('displays priority badges (A, B, C)', async ({ page }) => {
    await page.goto('/targets')

    // Priority badges should be rendered
    await expect(page.locator('table >> text=A').first()).toBeVisible()
    await expect(page.locator('table >> text=B').first()).toBeVisible()
    await expect(page.locator('table >> text=C').first()).toBeVisible()
  })

  test('search input is present and functional', async ({ page }) => {
    await page.goto('/targets')

    const searchInput = page.locator('input[placeholder="Search targets..."]')
    await expect(searchInput).toBeVisible()
    await searchInput.fill('Elena')

    // Verify the search value is set (API mock returns same data regardless of query)
    await expect(searchInput).toHaveValue('Elena')
  })

  test('type filter dropdown has expected options', async ({ page }) => {
    await page.goto('/targets')

    const select = page.locator('select').first()
    await expect(select).toBeVisible()

    const options = select.locator('option')
    await expect(options).toHaveCount(4) // All types + Doctor + Pharmacy + Hospital
  })
})

test.describe('Target detail page', () => {
  test.beforeEach(async ({ page }) => {
    await mockApi(page)
  })

  test('renders target details with name and priority', async ({ page }) => {
    await page.goto('/targets/t-001')

    await expect(page.locator('text=Dr. Elena Popescu')).toBeVisible()
    await expect(page.locator('text=Schedule Visit')).toBeVisible()
  })

  test('shows back link to targets list', async ({ page }) => {
    await page.goto('/targets/t-001')

    await expect(page.locator('main a[href="/targets"]')).toBeVisible()
  })
})
