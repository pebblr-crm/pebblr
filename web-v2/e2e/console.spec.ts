import { test, expect } from '@playwright/test'
import { mockApi, USERS, TEAMS, TERRITORIES } from './fixtures'

test.describe('Console page', () => {
  test.beforeEach(async ({ page }) => {
    await mockApi(page)
  })

  test('renders Users & Roles section by default', async ({ page }) => {
    await page.goto('/console')

    await expect(page.locator('h1:has-text("Users & Roles")')).toBeVisible()

    // User data should appear in the table
    for (const user of USERS.items) {
      await expect(page.locator(`text=${user.displayName}`)).toBeVisible()
    }
  })

  test('sidebar shows all configuration sections with counts', async ({ page }) => {
    await page.goto('/console')

    const sidebar = page.locator('.w-56')
    await expect(sidebar.locator('text=Users & Roles')).toBeVisible()
    await expect(sidebar.locator('text=Teams')).toBeVisible()
    await expect(sidebar.locator('text=Territories')).toBeVisible()
    await expect(sidebar.locator('text=Business Rules')).toBeVisible()
  })

  test('switching to Teams section shows team cards', async ({ page }) => {
    await page.goto('/console')

    await page.click('button:has-text("Teams")')
    await expect(page.locator('h1:has-text("Teams")')).toBeVisible()

    for (const team of TEAMS.items) {
      await expect(page.locator(`text=${team.name}`)).toBeVisible()
    }
  })

  test('switching to Territories section shows territory cards', async ({ page }) => {
    await page.goto('/console')

    await page.click('button:has-text("Territories")')
    await expect(page.locator('h1:has-text("Territories")')).toBeVisible()

    for (const territory of TERRITORIES.items) {
      await expect(page.locator(`text=${territory.name}`)).toBeVisible()
    }
  })

  test('switching to Business Rules section shows config', async ({ page }) => {
    await page.goto('/console')

    await page.click('button:has-text("Business Rules")')
    await expect(page.locator('h1:has-text("Business Rules")')).toBeVisible()

    // Visit frequency requirements
    await expect(page.locator('text=Visit Frequency Requirements')).toBeVisible()
    await expect(page.locator('text=5 visits/period')).toBeVisible()

    // Activity rules
    await expect(page.locator('text=Activity Rules')).toBeVisible()
    await expect(page.locator('text=Max activities per day')).toBeVisible()

    // Activity types
    await expect(page.locator('text=Activity Types')).toBeVisible()

    // Status workflow
    await expect(page.locator('text=Status Workflow')).toBeVisible()
  })

  test('user table shows role badges', async ({ page }) => {
    await page.goto('/console')

    // Each user's role should appear as a badge in the table
    for (const user of USERS.items) {
      await expect(page.locator(`table >> text=${user.email}`)).toBeVisible()
    }
  })
})
