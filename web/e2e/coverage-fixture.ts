/**
 * Playwright fixture that collects V8 code coverage from the browser
 * after each test and writes raw V8 JSON to .v8-coverage/ for later conversion.
 *
 * Uses Chromium's CDP-based coverage — no instrumentation overhead.
 */
import { test as base } from '@playwright/test'
import { writeFile, mkdir } from 'node:fs/promises'
import { join } from 'node:path'
import { randomUUID } from 'node:crypto'

const V8_OUTPUT = join(import.meta.dirname, '..', '.v8-coverage')

export const test = base.extend({
  page: async ({ page }, use) => {
    await page.coverage.startJSCoverage({ resetOnNavigation: false })
    await use(page)
    const entries = await page.coverage.stopJSCoverage()

    // Filter to only app source files (skip node_modules, Vite internals)
    const appEntries = entries.filter(
      (e) => e.url.includes('/src/') && !e.url.includes('node_modules'),
    )
    if (appEntries.length > 0) {
      await mkdir(V8_OUTPUT, { recursive: true })
      const file = join(V8_OUTPUT, `${randomUUID()}.json`)
      await writeFile(file, JSON.stringify(appEntries))
    }
  },
})

export { expect } from '@playwright/test'
