/**
 * Playwright global teardown: convert V8 coverage JSON files to LCOV
 * using istanbul libraries already available via @vitest/coverage-v8.
 */
import { existsSync, readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import libCoverage from 'istanbul-lib-coverage'
import libReport from 'istanbul-lib-report'
import reports from 'istanbul-reports'
import { V8toIstanbul } from 'ast-v8-to-istanbul'

export default async function globalTeardown() {
  const webDir = join(import.meta.dirname, '..')
  const v8Dir = join(webDir, '.v8-coverage')
  if (!existsSync(v8Dir)) return

  const files = readdirSync(v8Dir).filter((f) => f.endsWith('.json'))
  if (files.length === 0) return

  const coverageMap = libCoverage.createCoverageMap({})

  for (const file of files) {
    const entries = JSON.parse(readFileSync(join(v8Dir, file), 'utf-8'))

    for (const entry of entries) {
      // Convert Vite dev server URL to local file path
      // e.g. http://localhost:5174/src/App.tsx → /absolute/path/web/src/App.tsx
      const url = new URL(entry.url)
      const relativePath = url.pathname.replace(/^\//, '')
      const absolutePath = join(webDir, relativePath)

      if (!existsSync(absolutePath)) continue

      const source = readFileSync(absolutePath, 'utf-8')
      const converter = new V8toIstanbul(absolutePath, 0, { source })
      await converter.load()
      converter.applyCoverage(entry.functions)
      const istanbulData = converter.toIstanbul()
      coverageMap.merge(istanbulData)
    }
  }

  const context = libReport.createContext({
    dir: join(webDir, 'coverage-e2e'),
    coverageMap,
  })

  reports.create('lcov').execute(context)
}
