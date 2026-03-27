import { test, expect } from './coverage-fixture'
import { mockApi, ACTIVITIES, CONFIG } from './fixtures'

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
})

test.describe('Create activity modal', () => {
  test.beforeEach(async ({ page }) => {
    await mockApi(page)
  })

  test('"Log Activity" opens create modal with form fields', async ({ page }) => {
    await page.goto('/activities')

    await page.click('text=Log Activity')

    // Modal should be open with form
    await expect(page.locator('label:has-text("Activity Type")')).toBeVisible()
    await expect(page.locator('label:has-text("Quick Tags")')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Left Samples', exact: true })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Follow-up Required', exact: true })).toBeVisible()

    // URL should NOT change — stays on /activities
    expect(page.url()).toMatch(/\/activities\/?$/)
  })

  test('submit button is disabled until activity type is selected', async ({ page }) => {
    await page.goto('/activities')
    await page.click('text=Log Activity')

    // The submit button in the modal — last "Log Activity" button
    const submitBtn = page.locator('button:has-text("Log Activity")').last()
    await expect(submitBtn).toBeDisabled()

    // Select an activity type from the modal's select
    const typeSelect = page.locator('label:has-text("Activity Type")').locator('..').locator('select')
    await typeSelect.selectOption(CONFIG.activities.types[0].key)
    await expect(submitBtn).toBeEnabled()
  })

  test('can toggle quick tags', async ({ page }) => {
    await page.goto('/activities')
    await page.click('text=Log Activity')

    const tag = page.getByRole('button', { name: 'Left Samples', exact: true })
    await tag.click()
    // Tag should have teal styling when selected
    await expect(tag).toHaveClass(/border-teal-500/)

    // Click again to deselect
    await tag.click()
    await expect(tag).toHaveClass(/border-slate-200/)
  })

  test('submitting closes modal and stays on activities page', async ({ page }) => {
    await page.goto('/activities')
    await page.click('text=Log Activity')

    // Fill form
    const typeSelect = page.locator('label:has-text("Activity Type")').locator('..').locator('select')
    await typeSelect.selectOption(CONFIG.activities.types[0].key)
    await page.locator('button:has-text("Log Activity")').last().click()

    // Modal should close
    await expect(page.locator('label:has-text("Activity Type")')).not.toBeVisible({ timeout: 5000 })
    expect(page.url()).toMatch(/\/activities\/?$/)
  })

  test('closing modal with X button dismisses it', async ({ page }) => {
    await page.goto('/activities')
    await page.click('text=Log Activity')

    await expect(page.locator('label:has-text("Activity Type")')).toBeVisible()

    // Click X button
    await page.locator('button[aria-label="Close"]').click()

    await expect(page.locator('label:has-text("Activity Type")')).not.toBeVisible()
  })
})

test.describe('Activity detail modal', () => {
  test.beforeEach(async ({ page }) => {
    await mockApi(page)
  })

  test('clicking an activity card opens detail modal', async ({ page }) => {
    await page.goto('/activities')

    // Click the first activity (Dr. Elena Popescu)
    await page.locator('button:has-text("Dr. Elena Popescu")').click()

    // Modal should show activity detail line with date
    await expect(page.getByText(/visit · full_day · /)).toBeVisible()

    // URL should NOT change
    expect(page.url()).toMatch(/\/activities\/?$/)
  })

  test('detail modal shows notes', async ({ page }) => {
    await page.goto('/activities')

    await page.locator('button:has-text("Dr. Elena Popescu")').click()

    // Notes text appears in both the list card and the modal — check count is 2
    await expect(page.getByText('Good discussion about new treatment.')).toHaveCount(2)
  })

  test('detail modal shows tags', async ({ page }) => {
    await page.goto('/activities')

    await page.locator('button:has-text("Dr. Elena Popescu")').click()

    // "Left Samples" appears once in the list card and once in the modal
    await expect(page.getByText('Left Samples')).toHaveCount(2)
  })

  test('detail modal shows submitted badge for completed activities', async ({ page }) => {
    await page.goto('/activities')

    await page.locator('button:has-text("Dr. Elena Popescu")').click()

    // Submitted badge in both list and modal
    const submittedBadges = page.getByText('Submitted')
    await expect(submittedBadges.first()).toBeVisible()
  })

  test('detail modal shows status transition buttons for planned activities', async ({ page }) => {
    await page.goto('/activities')

    // Click the planned activity (Farmacia Central)
    await page.locator('button:has-text("Farmacia Central")').click()

    // planificat can transition to realizat or anulat
    await expect(page.getByRole('button', { name: 'Completed', exact: true })).toBeVisible({ timeout: 5000 })
    await expect(page.getByRole('button', { name: 'Cancelled', exact: true })).toBeVisible()

    // feedback textarea is available in the footer
    await expect(page.locator('textarea[placeholder="How did the visit go?"]')).toBeVisible()
  })

  test('closing detail modal with X dismisses it', async ({ page }) => {
    await page.goto('/activities')

    await page.locator('button:has-text("Dr. Elena Popescu")').click()

    // Wait for modal content
    await expect(page.getByText('Created')).toBeVisible({ timeout: 5000 })

    // Close
    await page.locator('button[aria-label="Close"]').click()

    // "Created" label is only in the modal
    await expect(page.getByText('Created')).not.toBeVisible()
  })

  test('detail modal shows valid creation date, not Invalid Date', async ({ page }) => {
    await page.goto('/activities')

    await page.locator('button:has-text("Dr. Elena Popescu")').click()

    await expect(page.getByText('Created')).toBeVisible({ timeout: 5000 })
    await expect(page.getByText('Invalid Date')).not.toBeVisible()
  })
})
