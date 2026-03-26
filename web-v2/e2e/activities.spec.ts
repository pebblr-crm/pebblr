import { test, expect } from '@playwright/test'
import { mockApi, ACTIVITIES } from './fixtures'

test.describe('Activities page', () => {
  test.beforeEach(async ({ page }) => {
    await mockApi(page)
  })

  test('renders activity log with header and log button', async ({ page }) => {
    await page.goto('/activities')

    await expect(page.locator('text=Activity Log')).toBeVisible()
    await expect(page.locator('text=Log Activity')).toBeVisible()
  })

  test('shows recovery balance card', async ({ page }) => {
    await page.goto('/activities')

    await expect(page.locator('text=Recovery Balance')).toBeVisible()
    await expect(page.locator('text=3 days')).toBeVisible()
    await expect(page.locator('text=5 earned, 2 taken')).toBeVisible()
  })

  test('shows activity count and filter dropdowns', async ({ page }) => {
    await page.goto('/activities')

    await expect(page.locator(`text=${ACTIVITIES.items.length} activities`)).toBeVisible()

    // Type filter
    const typeSelect = page.locator('select').first()
    await expect(typeSelect).toBeVisible()

    // Status filter
    const statusSelect = page.locator('select').nth(1)
    await expect(statusSelect).toBeVisible()
  })

  test('renders activity cards grouped by date', async ({ page }) => {
    await page.goto('/activities')

    // Activities from our fixtures should appear
    await expect(page.locator('text=Dr. Elena Popescu')).toBeVisible()
    await expect(page.locator('text=Farmacia Central')).toBeVisible()
  })

  test('displays status badges on activities', async ({ page }) => {
    await page.goto('/activities')

    await expect(page.locator('text=realizat').first()).toBeVisible()
    await expect(page.locator('text=planificat').first()).toBeVisible()
  })

  test('"Log Activity" button links to new activity form', async ({ page }) => {
    await page.goto('/activities')

    await page.click('text=Log Activity')
    await expect(page).toHaveURL(/\/activities\/new/)
  })
})

test.describe('New activity form', () => {
  test.beforeEach(async ({ page }) => {
    await mockApi(page)
  })

  test('renders step 1 with activity type, tags, and outcomes', async ({ page }) => {
    await page.goto('/activities/new')

    await expect(page.locator('text=Log Activity')).toBeVisible()
    await expect(page.locator('text=Step 1 of 2')).toBeVisible()

    // Activity type select
    await expect(page.locator('text=Activity Type')).toBeVisible()

    // Quick tags
    await expect(page.locator('text=Left Samples')).toBeVisible()
    await expect(page.locator('text=Follow-up Required')).toBeVisible()

    // Outcome buttons
    await expect(page.locator('text=Completed')).toBeVisible()
    await expect(page.locator('text=Rescheduled')).toBeVisible()
    await expect(page.locator('text=No Show')).toBeVisible()
    await expect(page.locator('text=Cancelled')).toBeVisible()
  })

  test('continue button is disabled until outcome is selected', async ({ page }) => {
    await page.goto('/activities/new')

    const continueBtn = page.locator('button:has-text("Continue")')
    await expect(continueBtn).toBeDisabled()

    // Select an outcome
    await page.click('text=Completed')
    await expect(continueBtn).toBeEnabled()
  })

  test('advances to step 2 and shows notes + submit', async ({ page }) => {
    await page.goto('/activities/new')

    // Complete step 1
    await page.click('text=Left Samples')
    await page.click('text=Completed')
    await page.click('button:has-text("Continue")')

    // Step 2
    await expect(page.locator('text=Step 2 of 2')).toBeVisible()
    await expect(page.locator('text=Visit Notes')).toBeVisible()
    await expect(page.locator('text=Schedule next visit?')).toBeVisible()
    await expect(page.locator('button:has-text("Submit Activity")')).toBeVisible()
    await expect(page.locator('button:has-text("Back")')).toBeVisible()
  })

  test('submits activity and navigates to activities page', async ({ page }) => {
    await page.goto('/activities/new')

    // Step 1
    await page.click('text=Completed')
    await page.click('button:has-text("Continue")')

    // Step 2 — fill notes and submit
    await page.locator('textarea').fill('Test visit notes')
    await page.click('button:has-text("Submit Activity")')

    await expect(page).toHaveURL(/\/activities/, { timeout: 5000 })
  })

  test('back button returns to step 1', async ({ page }) => {
    await page.goto('/activities/new')

    await page.click('text=Completed')
    await page.click('button:has-text("Continue")')
    await expect(page.locator('text=Step 2 of 2')).toBeVisible()

    await page.click('button:has-text("Back")')
    await expect(page.locator('text=Step 1 of 2')).toBeVisible()
  })
})
