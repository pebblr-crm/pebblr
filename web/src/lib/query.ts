/**
 * Shared utility for building query strings from parameter objects.
 *
 * Replaces the per-hook buildQuery functions that were duplicated across
 * useActivities, useAudit, useTargets, and useDashboard.
 */

/**
 * Build a URL with query string from a base path and a params object.
 * Undefined/null/empty-string values are omitted from the query string.
 */
export function buildQuery(
  basePath: string,
  params: Record<string, string | number | undefined | null>,
): string {
  const qs = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== null && value !== '') {
      qs.set(key, String(value))
    }
  }
  const q = qs.toString()
  return q ? `${basePath}?${q}` : basePath
}
