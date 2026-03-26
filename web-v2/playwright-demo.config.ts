/**
 * Playwright config for demo-mode tests.
 *
 * Starts a Vite dev server with VITE_DEMO_MODE=true and no static token,
 * so the DemoGate component is activated and testable.
 */
import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './e2e',
  testMatch: 'demo-gate.spec.ts',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI ? 'github' : 'list',
  use: {
    baseURL: 'http://localhost:5175',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: {
    command: 'bun run dev -- --port 5175',
    url: 'http://localhost:5175',
    reuseExistingServer: !process.env.CI,
    timeout: 30_000,
    env: {
      VITE_DEMO_MODE: 'true',
      // No VITE_STATIC_TOKEN — forces demo auth flow
    },
  },
})
