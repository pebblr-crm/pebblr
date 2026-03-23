import type { ActivitiesConfig, ActivityTypeConfig } from '@/types/config'
import type { Activity } from '@/types/activity'

// ── Lookup helpers ──────────────────────────────────────────────────────────
// All helpers accept ActivitiesConfig | undefined.
// Callers that hold a full TenantConfig should pass config?.activities.

export function getTypeConfig(
  config: ActivitiesConfig | undefined,
  typeKey: string,
): ActivityTypeConfig | undefined {
  return config?.types.find((t) => t.key === typeKey)
}

export function getTypeLabel(config: ActivitiesConfig | undefined, typeKey: string): string {
  return getTypeConfig(config, typeKey)?.label ?? typeKey
}

export function getTypeCategory(
  config: ActivitiesConfig | undefined,
  typeKey: string,
): 'field' | 'non_field' {
  return getTypeConfig(config, typeKey)?.category ?? 'field'
}

export function getStatusLabel(config: ActivitiesConfig | undefined, statusKey: string): string {
  return config?.statuses.find((s) => s.key === statusKey)?.label ?? statusKey
}

export function getDurationLabel(config: ActivitiesConfig | undefined, durationKey: string): string {
  return config?.durations.find((d) => d.key === durationKey)?.label ?? durationKey
}

// ── Style constants ─────────────────────────────────────────────────────────

/** Canonical category color map. Single source of truth — do not redefine inline. */
export const CATEGORY_COLORS: Record<string, string> = {
  field: 'bg-amber-50 border-amber-500 text-amber-900',
  non_field: 'bg-blue-50 border-blue-400 text-blue-900',
}

/**
 * Status color styles. The initial status gets amber (pending), the last
 * status gets red (negative/cancelled), and all others get emerald (positive).
 * This adapts to any number of statuses without relying on array position.
 */
const STATUS_STYLES = {
  initial: { badge: 'bg-amber-100 text-amber-700', dot: 'bg-amber-500' },
  positive: { badge: 'bg-emerald-100 text-emerald-700', dot: 'bg-emerald-500' },
  negative: { badge: 'bg-red-100 text-red-700', dot: 'bg-red-400' },
  fallback: { badge: 'bg-slate-100 text-slate-600', dot: 'bg-slate-400' },
}

function resolveStatusStyle(config: ActivitiesConfig | undefined, statusKey: string) {
  if (!config || config.statuses.length === 0) return STATUS_STYLES.fallback
  const statuses = config.statuses
  const status = statuses.find((s) => s.key === statusKey)
  if (!status) return STATUS_STYLES.fallback
  if (status.initial) return STATUS_STYLES.initial
  // Last status is treated as the negative/cancelled state
  if (statuses[statuses.length - 1].key === statusKey) return STATUS_STYLES.negative
  return STATUS_STYLES.positive
}

/** Returns badge classes for a status key. */
export function getStatusBadgeColor(config: ActivitiesConfig | undefined, statusKey: string): string {
  return resolveStatusStyle(config, statusKey).badge
}

/** Returns dot classes for a status key. */
export function getStatusDotColor(config: ActivitiesConfig | undefined, statusKey: string): string {
  return resolveStatusStyle(config, statusKey).dot
}

/**
 * @deprecated Use getStatusBadgeColor() instead. Kept for existing references.
 */
export const STATUS_BADGE_COLORS: Record<string, string> = {
  planned: 'bg-amber-100 text-amber-700',
  completed: 'bg-emerald-100 text-emerald-700',
  cancelled: 'bg-red-100 text-red-700',
  // Legacy keys for backwards compatibility with tests
  planificat: 'bg-amber-100 text-amber-700',
  realizat: 'bg-emerald-100 text-emerald-700',
  anulat: 'bg-red-100 text-red-700',
}

// ── Activity title ─────────────────────────────────────────────────────────

/**
 * Returns a human-readable title for an activity.
 * Priority: activity.label (user override) → computed title.
 * For visits: "F2f Visit" / "Remote Visit" (visit_type + type label).
 * For everything else: the activity type label.
 */
export function getActivityTitle(config: ActivitiesConfig | undefined, activity: Activity): string {
  if (activity.label) return activity.label

  const typeLabel = getTypeLabel(config, activity.activityType)

  if (activity.activityType === 'visit') {
    const visitType = activity.fields?.visit_type as string | undefined
    const vtLabel = visitType ? (visitType === 'f2f' ? 'F2F' : visitType.charAt(0).toUpperCase() + visitType.slice(1)) : null
    const parts = [vtLabel, typeLabel].filter(Boolean)
    if (activity.targetName) parts.push(`— ${activity.targetName}`)
    return parts.join(' ')
  }

  return typeLabel
}

/** Month names indexed 0–11. */
export const MONTH_NAMES = [
  'January',
  'February',
  'March',
  'April',
  'May',
  'June',
  'July',
  'August',
  'September',
  'October',
  'November',
  'December',
]
