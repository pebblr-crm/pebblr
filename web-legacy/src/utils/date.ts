/** Returns YYYY-MM-DD for a Date object. */
export function formatDate(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

/** Extracts YYYY-MM-DD from an ISO string (replaces .split('T')[0] inline calls). */
export function extractDate(iso: string): string {
  return iso.split('T')[0]
}

/** Returns a new Date offset by `n` days from `d`. */
export function addDays(d: Date, n: number): Date {
  const result = new Date(d)
  result.setDate(result.getDate() + n)
  return result
}

/** Builds a YYYY-MM-DD string from numeric year, month (1-based), and day. */
export function formatDateStr(year: number, month: number, day: number): string {
  return `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`
}

/** Returns YYYY-MM period string for a Date object. */
export function formatPeriod(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  return `${y}-${m}`
}

/** Returns the Monday of the week containing `date`, at midnight. */
export function getMonday(date: Date): Date {
  const d = new Date(date)
  const dow = d.getDay()
  const diff = dow === 0 ? -6 : 1 - dow
  d.setDate(d.getDate() + diff)
  d.setHours(0, 0, 0, 0)
  return d
}

/** Formats an ISO date string as a locale-aware display string (e.g. "March 23, 2026"). */
export function displayDate(iso: string): string {
  return new Date(iso).toLocaleDateString()
}
