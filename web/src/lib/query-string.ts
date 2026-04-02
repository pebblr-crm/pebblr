/**
 * Build a URL query string from a params object, omitting undefined/null/empty-string values.
 * Returns the path with query string appended, or the bare path if no params are set.
 */
export function buildQueryString(basePath: string, params: Record<string, string | number | undefined | null>): string {
  const qs = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value != null && value !== '') {
      qs.set(key, String(value))
    }
  }
  const q = qs.toString()
  return q ? `${basePath}?${q}` : basePath
}
