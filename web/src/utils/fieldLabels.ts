import type { TenantConfig } from '@/types/config'

/** Well-known top-level activity fields that aren't in the config's fields array. */
const TOP_LEVEL_LABELS: Record<string, string> = {
  activityType: 'Activity Type',
  activity_type: 'Activity Type',
  dueDate: 'Date',
  due_date: 'Date',
  duration: 'Duration',
  status: 'Status',
  targetId: 'Target',
  target_id: 'Target',
  routing: 'Routing',
}

/**
 * Resolves a human-readable label for a field key.
 * Checks: top-level fields → config field labels → fallback formatting.
 */
export function getFieldLabel(
  config: TenantConfig | undefined,
  activityType: string | undefined,
  fieldKey: string,
): string {
  if (TOP_LEVEL_LABELS[fieldKey]) return TOP_LEVEL_LABELS[fieldKey]

  if (config && activityType) {
    const typeConfig = config.activities.types.find((t) => t.key === activityType)
    const fieldDef = typeConfig?.fields.find((f) => f.key === fieldKey)
    if (fieldDef?.label) return fieldDef.label
  }

  return fieldKey.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())
}

/**
 * Formats an array of validation field errors into a single toast message.
 */
export function formatValidationToast(
  config: TenantConfig | undefined,
  activityType: string | undefined,
  errors: Array<{ field: string; message: string }>,
): string {
  if (errors.length === 0) return 'Validation failed'
  const labels = errors.map((e) => getFieldLabel(config, activityType, e.field))
  return `Required fields missing: ${labels.join(', ')}`
}
