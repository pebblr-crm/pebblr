/**
 * Playwright global teardown: convert .nyc_output/ JSON to LCOV
 * so it can be merged with vitest unit-test coverage.
 */
import { execSync } from 'node:child_process'
import { existsSync } from 'node:fs'
import { join } from 'node:path'

export default function globalTeardown() {
  const nycOutput = join(import.meta.dirname, '..', '.nyc_output')
  if (!existsSync(nycOutput)) return

  execSync(
    'npx nyc report --reporter=lcov --report-dir=coverage-e2e --temp-dir=.nyc_output',
    { cwd: join(import.meta.dirname, '..'), stdio: 'inherit' },
  )
}
