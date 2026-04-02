/** Extract latitude from a target's fields bag, returning null if missing or non-numeric. */
export function getLat(fields: Record<string, unknown>): number | null {
  const v = fields.lat
  return typeof v === 'number' ? v : null
}

/** Extract longitude from a target's fields bag, returning null if missing or non-numeric. */
export function getLng(fields: Record<string, unknown>): number | null {
  const v = fields.lng
  return typeof v === 'number' ? v : null
}

/** Extract classification (a/b/c) from a target's fields bag, defaulting to 'c'. */
export function getClassification(fields: Record<string, unknown>): string {
  return ((fields.potential as string) ?? 'c').toLowerCase()
}

/** Extract city from a target's fields bag. */
export function getCity(fields: Record<string, unknown>): string {
  return (fields.city as string) ?? ''
}

/** Check whether a target has valid geo coordinates. */
export function hasGeoCoords(fields: Record<string, unknown>): boolean {
  return getLat(fields) != null && getLng(fields) != null
}
