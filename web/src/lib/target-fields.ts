/**
 * Shared helpers for extracting typed values from target field maps.
 *
 * These were previously copy-pasted across planner.tsx, coverage.tsx,
 * reps.$id.tsx, and targets.tsx.
 */

/** Extract latitude from a target's fields map. */
export function getLat(fields: Record<string, unknown>): number | null {
  const v = fields.lat
  return typeof v === 'number' ? v : null
}

/** Extract longitude from a target's fields map. */
export function getLng(fields: Record<string, unknown>): number | null {
  const v = fields.lng
  return typeof v === 'number' ? v : null
}

/** Extract the classification/priority tier (a/b/c), defaulting to 'c'. */
export function getClassification(fields: Record<string, unknown>): string {
  return ((fields.potential as string) ?? 'c').toLowerCase()
}

/** Extract city from a target's fields map. */
export function getCity(fields: Record<string, unknown>): string {
  return (fields.city as string) ?? ''
}

/** Returns true when both lat and lng are present in the fields map. */
export function hasGeoCoords(fields: Record<string, unknown>): boolean {
  return getLat(fields) != null && getLng(fields) != null
}
