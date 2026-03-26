/** Safely coerce an unknown value to a string, returning '' for non-strings. */
export function str(v: unknown): string {
  return typeof v === 'string' ? v : ''
}

/** Return the number of whole days between a date string and now. */
export function daysAgo(dateStr: string): number {
  return Math.floor((Date.now() - new Date(dateStr).getTime()) / (1000 * 60 * 60 * 24))
}
