import i18n from '@/i18n'
import type { ActivitiesConfig, ActivityTypeConfig, TenantConfig } from '@/types/config'
import type { Activity } from '@/types/activity'

// ── i18n config label helper ────────────────────────────────────────────────

/**
 * Translates a config-driven label via i18n.
 * Looks up `configLabels.<key>` — if a translation exists, returns it;
 * otherwise falls back to the raw config label.
 */
export function translateConfigLabel(i18nKey: string, fallback: string): string {
  const fullKey = `configLabels.${i18nKey}`
  const translated = i18n.t(fullKey)
  // i18next returns the key itself when no translation is found
  return translated === fullKey ? fallback : translated
}

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
  return translateConfigLabel(`type.${typeKey}`, getTypeConfig(config, typeKey)?.label ?? typeKey)
}

export function getTypeCategory(
  config: ActivitiesConfig | undefined,
  typeKey: string,
): 'field' | 'non_field' {
  return getTypeConfig(config, typeKey)?.category ?? 'field'
}

export function getStatusLabel(config: ActivitiesConfig | undefined, statusKey: string): string {
  return translateConfigLabel(`status.${statusKey}`, config?.statuses.find((s) => s.key === statusKey)?.label ?? statusKey)
}

export function getDurationLabel(config: ActivitiesConfig | undefined, durationKey: string): string {
  return translateConfigLabel(`duration.${durationKey}`, config?.durations.find((d) => d.key === durationKey)?.label ?? durationKey)
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

// ── Activity title ─────────────────────────────────────────────────────────

/**
 * Returns a human-readable title for an activity.
 * Priority: activity.label (user override) → computed title.
 * When the activity type defines a title_field, the field's display label
 * is prepended to the type label (e.g. "F2F Visit — Dr. Smith").
 * For everything else: the activity type label.
 */
export function getActivityTitle(config: TenantConfig | undefined, activity: Activity): string {
  if (activity.label) return activity.label

  const ac = config?.activities
  const typeLabel = getTypeLabel(ac, activity.activityType)
  const typeCfg = getTypeConfig(ac, activity.activityType)

  const parts: string[] = []

  if (typeCfg?.title_field) {
    const fieldValue = activity.fields?.[typeCfg.title_field] as string | undefined
    const prefix = fieldValue ? resolveOptionLabel(config, typeCfg, typeCfg.title_field, fieldValue) : null
    if (prefix) parts.push(prefix)
  }

  parts.push(typeLabel)

  if (activity.targetName) {
    parts.push(`— ${activity.targetName}`)
  }

  return parts.join(' ')
}

/**
 * Returns a full display name including the date, for contexts where the
 * date isn't already shown (e.g. activity list, search results).
 * Format: "Visit — Dr. Popescu — Mar 24" or "Training — Mar 24"
 */
export function getActivityDisplayName(config: TenantConfig | undefined, activity: Activity): string {
  const title = getActivityTitle(config, activity)
  const date = new Date(activity.dueDate)
  const dateStr = date.toLocaleDateString(getDateLocale(), { day: 'numeric', month: 'short' })
  return `${title} — ${dateStr}`
}

/**
 * Resolves the display label for an option value on a given field.
 * Checks options_ref (via the tenant options map), then inline options.
 * Falls back to the raw value.
 */
function resolveOptionLabel(
  config: TenantConfig | undefined,
  typeCfg: ActivityTypeConfig,
  fieldKey: string,
  value: string,
): string {
  const fieldCfg = typeCfg.fields.find((f) => f.key === fieldKey)
  if (!fieldCfg) return value

  // options_ref: look up in the tenant options map, then special refs.
  if (fieldCfg.options_ref && config) {
    const opts = config.options[fieldCfg.options_ref]
      ?? (fieldCfg.options_ref === 'durations' ? config.activities.durations : undefined)
    const match = opts?.find((o) => o.key === value)
    if (match) return translateConfigLabel(`option.${fieldCfg.options_ref}.${value}`, match.label)
  }

  // Inline string options have no labels — return the raw value.
  return value
}

/** Returns the BCP 47 locale tag matching the current i18n language. */
export function getDateLocale(): string {
  const lang = i18n.language
  if (lang === 'ro') return 'ro-RO'
  return 'en-GB'
}

/**
 * Translates a config option label.
 * Used by components rendering option dropdowns from config data.
 */
export function getOptionLabel(optionsRef: string, key: string, fallback: string): string {
  return translateConfigLabel(`option.${optionsRef}.${key}`, fallback)
}

/**
 * Translates a config field label.
 * Used by components rendering dynamic field labels from config data.
 */
export function getConfigFieldLabel(fieldKey: string, fallback: string): string {
  return translateConfigLabel(`field.${fieldKey}`, fallback)
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
