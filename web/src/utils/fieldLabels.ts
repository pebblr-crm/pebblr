import i18n from '@/i18n'
import type { TenantConfig } from '@/types/config'

/** Well-known top-level activity fields that aren't in the config's fields array. */
const TOP_LEVEL_LABEL_KEYS: Record<string, string> = {
  activityType: 'fieldLabels.activityType',
  activity_type: 'fieldLabels.activityType',
  dueDate: 'fieldLabels.date',
  due_date: 'fieldLabels.date',
  duration: 'fieldLabels.duration',
  status: 'fieldLabels.status',
  targetId: 'fieldLabels.target',
  target_id: 'fieldLabels.target',
  routing: 'fieldLabels.routing',
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
  const i18nKey = TOP_LEVEL_LABEL_KEYS[fieldKey]
  if (i18nKey) return i18n.t(i18nKey)

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
  if (errors.length === 0) return i18n.t('fieldLabels.validationFailed')
  const labels = errors.map((e) => getFieldLabel(config, activityType, e.field))
  return i18n.t('fieldLabels.requiredFieldsMissing', { labels: labels.join(', ') })
}
