import { test, expect } from './coverage-fixture'
import { mockApi, TARGETS } from './fixtures'

test.describe('Planner page', () => {
  test.beforeEach(async ({ page }) => {
    await mockApi(page)
  })

  test('renders header, stats row, and week navigation', async ({ page }) => {
    await page.goto('/planner')

    await expect(page.locator('text=Planning Workspace')).toBeVisible()
    await expect(page.locator('text=Today')).toBeVisible()
    await expect(page.locator('text=Clone Week')).toBeVisible()

    // Stats row should be visible
    const statsRow = page.locator('.grid.grid-cols-2.md\\:grid-cols-4')
    await expect(statsRow.locator('text=Planned')).toBeVisible()
    await expect(statsRow.locator('text=Completed')).toBeVisible()
    await expect(statsRow.locator('text=Coverage')).toBeVisible()
    await expect(statsRow.locator('text=Overdue A')).toBeVisible()
  })

  test('shows A-priority nudge banner when overdue targets exist', async ({ page }) => {
    await page.goto('/planner')

    // We have 1 A-priority target in fixtures
    const aPriorityCount = TARGETS.items.filter(
      (t) => t.fields.classification === 'A',
    ).length

    if (aPriorityCount > 0) {
      await expect(page.locator('text=A-priority targets need attention')).toBeVisible()
    }
  })

  test('week navigation buttons change the displayed date range', async ({ page }) => {
    await page.goto('/planner')

    // Grab the initial date text
    const weekNav = page.locator('.rounded-lg.border.border-slate-200.bg-slate-50')
    const dateText = weekNav.locator('span')
    const initialText = await dateText.innerText()

    // Click next week
    await weekNav.locator('button:last-child').click()
    const nextText = await dateText.innerText()
    expect(nextText).not.toBe(initialText)

    // Click previous week (back twice to go before initial)
    await weekNav.locator('button:first-child').click()
    const backText = await dateText.innerText()
    expect(backText).toBe(initialText)
  })
})
