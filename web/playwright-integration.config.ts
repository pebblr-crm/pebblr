/**
 * Playwright config for integration e2e tests against a real Kind cluster.
 *
 * Expects the app to be deployed in the pebblr-e2e namespace and
 * port-forwarded to a local port. The base URL is provided via:
 *
 *   E2E_BASE_URL  — e.g. http://127.0.0.1:9222
 *
 * Run with: make e2e-web-integration
 */
import { defineConfig, devices } from '@playwright/test'

const baseURL = process.env.E2E_BASE_URL
if (!baseURL) {
  throw new Error('E2E_BASE_URL must be set (e.g. http://127.0.0.1:9222)')
}

export default defineConfig({
  testDir: './e2e/integration',
  fullyParallel: false, // serial — single port-forward, shared cookie state
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: process.env.CI ? 'github' : 'list',
  timeout: 30_000,
  use: {
    baseURL,
    trace: 'on-first-retry',
    // The pebblr_ui=v2 cookie tells the server to serve the v2 SPA.
    extraHTTPHeaders: {
      Cookie: 'pebblr_ui=v2',
    },
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  // No webServer — the app is already running on the Kind cluster.
})
