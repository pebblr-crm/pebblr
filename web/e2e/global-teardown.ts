/**
 * Playwright global teardown: convert V8 coverage JSON files to LCOV
 * using istanbul libraries already available via @vitest/coverage-v8.
 */
import { existsSync, readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import { convert } from 'ast-v8-to-istanbul'
import { parse } from 'acorn'
import libCoverage from 'istanbul-lib-coverage'
import libReport from 'istanbul-lib-report'
import reports from 'istanbul-reports'

function extractInlineSourceMap(code: string) {
  const match = code.match(
    /\/\/[#@] sourceMappingURL=data:application\/json;(?:charset=utf-8;)?base64,(.+)$/m,
  )
  if (!match) return undefined
  try {
    return JSON.parse(Buffer.from(match[1], 'base64').toString('utf-8'))
  } catch {
    return undefined
  }
}

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
      if (!entry.source || !entry.functions?.length) continue

      try {
        const sourceMap = extractInlineSourceMap(entry.source)
        const code = entry.source.replace(
          /\/\/[#@] sourceMappingURL=data:[^\n]+$/m,
          '',
        )

        const data = await convert({
          ast: parse(code, {
            ecmaVersion: 'latest',
            sourceType: 'module',
          }),
          code,
          wrapperLength: 0,
          coverage: {
            scriptId: entry.scriptId || '0',
            url: entry.url,
            functions: entry.functions,
          },
          sourceMap,
        })

        coverageMap.merge(data)
      } catch {
        // Skip entries that fail to convert (e.g. dynamic imports, eval'd code)
      }
    }
  }

  const context = libReport.createContext({
    dir: join(webDir, 'coverage-e2e'),
    coverageMap,
  })

  reports.create('lcov').execute(context)
}
