/**
 * Playwright fixture that collects Istanbul coverage from the browser
 * after each test and writes it to .nyc_output/ for later merging.
 */
import { test as base } from '@playwright/test'
import { writeFile, mkdir } from 'node:fs/promises'
import { join } from 'node:path'
import { randomUUID } from 'node:crypto'

const NYC_OUTPUT = join(import.meta.dirname, '..', '.nyc_output')

export const test = base.extend({
  page: async ({ page }, use) => {
    await use(page)

    // After each test, grab the Istanbul coverage object from the page
    const coverage = await page.evaluate(() =>
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      (window as any).__coverage__ ?? null,
    )
    if (coverage) {
      await mkdir(NYC_OUTPUT, { recursive: true })
      const file = join(NYC_OUTPUT, `${randomUUID()}.json`)
      await writeFile(file, JSON.stringify(coverage))
    }
  },
})

export { expect } from '@playwright/test'
