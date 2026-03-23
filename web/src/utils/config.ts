import type { ActivitiesConfig, ActivityTypeConfig } from '@/types/config'

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

/** Canonical status badge color map. Single source of truth — do not redefine inline. */
export const STATUS_BADGE_COLORS: Record<string, string> = {
  planificat: 'bg-amber-100 text-amber-700',
  realizat: 'bg-emerald-100 text-emerald-700',
  anulat: 'bg-red-100 text-red-700',
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
