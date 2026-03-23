import { useState, type FormEvent } from 'react'
import { useConfig } from '../services/config'
import { useTeamMembers } from '../services/teams'
import { useTargets } from '../services/targets'
import type { Activity, CreateActivityInput } from '../types/activity'
import type { FieldConfig, OptionDef, TenantConfig } from '../types/config'
import { LoadingSpinner } from './LoadingSpinner'
import { extractDate } from '@/utils/date'

interface ActivityFormProps {
  initialData?: Activity
  onSubmit: (data: CreateActivityInput) => void
  onCancel: () => void
  isSubmitting: boolean
  serverErrors?: Array<{ field: string; message: string }>
}

export function ActivityForm(props: ActivityFormProps) {
  const { data: config, isLoading: configLoading } = useConfig()

  if (configLoading || !config) {
    return (
      <div className="flex items-center justify-center h-64">
        <LoadingSpinner size="lg" label="Loading configuration..." />
      </div>
    )
  }

  return <ActivityFormInner {...props} config={config} />
}

interface InnerProps extends ActivityFormProps {
  config: TenantConfig
}

function ActivityFormInner({
  initialData,
  onSubmit,
  onCancel,
  isSubmitting,
  serverErrors,
  config,
}: InnerProps) {
  const initialStatus = initialData?.status
    ?? config.activities.statuses.find((s) => s.initial)?.key
    ?? ''

  const [activityType, setActivityType] = useState(initialData?.activityType ?? '')
  const [status, setStatus] = useState(initialStatus)
  const [dueDate, setDueDate] = useState(initialData?.dueDate ? extractDate(initialData.dueDate) : '')
  const [duration, setDuration] = useState(initialData?.duration ?? '')
  const [targetId, setTargetId] = useState(initialData?.targetId ?? '')
  const [fields, setFields] = useState<Record<string, unknown>>(() => {
    const f = { ...(initialData?.fields ?? {}) }
    // Seed routing into fields so the dynamic form pre-populates it.
    if (initialData?.routing) f.routing = initialData.routing
    return f
  })
  const [targetSearch, setTargetSearch] = useState('')

  const { data: targetsResult } = useTargets({ q: targetSearch, limit: 20 })
  const { data: membersResult } = useTeamMembers()

  const activityTypes = config.activities.types
  const statuses = config.activities.statuses
  const durations = config.activities.durations
  const selectedType = activityTypes.find((t) => t.key === activityType)
  const isFieldActivity = selectedType?.category === 'field'
  const isEditing = Boolean(initialData)
  const isLocked = Boolean(initialData?.submittedAt)

  function getFieldError(key: string): string | undefined {
    return serverErrors?.find((e) => e.field === key)?.message
  }

  function setFieldValue(key: string, value: unknown) {
    setFields((prev) => ({ ...prev, [key]: value }))
  }

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    // Extract routing from dynamic fields → top-level property for backend.
    const { routing: routingValue, ...restFields } = fields as Record<string, unknown> & { routing?: string }
    onSubmit({
      activityType,
      status: isEditing ? status : '',
      dueDate,
      duration,
      routing: (routingValue as string) || undefined,
      fields: restFields,
      targetId: targetId || undefined,
    })
  }

  function resolveOptions(fieldDef: FieldConfig): OptionDef[] {
    if (fieldDef.options_ref) {
      // Check top-level options map first, then special-case refs
      // that live outside the options map (mirrors backend ResolveOptions).
      if (config.options[fieldDef.options_ref]) {
        return config.options[fieldDef.options_ref]
      }
      if (fieldDef.options_ref === 'durations') {
        return config.activities.durations
      }
      return []
    }
    if (fieldDef.options) {
      return fieldDef.options.map((o) => ({ key: o, label: o }))
    }
    return []
  }

  function renderDynamicField(fieldDef: FieldConfig) {
    const value = fields[fieldDef.key]
    const error = getFieldError(fieldDef.key)
    const editable = fieldDef.editable !== false && !isLocked

    const labelEl = (
      <label className="block text-xs font-bold uppercase tracking-widest text-slate-400 mb-1">
        {fieldDef.label ?? fieldDef.key.replace(/_/g, ' ')}
        {fieldDef.required && <span className="text-error ml-1">*</span>}
      </label>
    )

    switch (fieldDef.type) {
      case 'text':
        return (
          <div key={fieldDef.key}>
            {labelEl}
            <input
              type="text"
              value={(value as string) ?? ''}
              onChange={(e) => setFieldValue(fieldDef.key, e.target.value)}
              disabled={!editable}
              className={inputClass(error)}
              data-testid={`field-${fieldDef.key}`}
            />
            {error && <p className="text-xs text-error mt-1">{error}</p>}
          </div>
        )

      case 'select': {
        const opts = resolveOptions(fieldDef)
        return (
          <div key={fieldDef.key}>
            {labelEl}
            <select
              value={(value as string) ?? ''}
              onChange={(e) => setFieldValue(fieldDef.key, e.target.value)}
              disabled={!editable}
              className={inputClass(error)}
              data-testid={`field-${fieldDef.key}`}
            >
              <option value="">— Select —</option>
              {opts.map((o) => (
                <option key={o.key} value={o.key}>{o.label}</option>
              ))}
            </select>
            {error && <p className="text-xs text-error mt-1">{error}</p>}
          </div>
        )
      }

      case 'multi_select': {
        const opts = resolveOptions(fieldDef)
        const selected = Array.isArray(value) ? (value as string[]) : []
        return (
          <div key={fieldDef.key}>
            {labelEl}
            <div className="flex flex-wrap gap-2" data-testid={`field-${fieldDef.key}`}>
              {opts.map((o) => {
                const isSelected = selected.includes(o.key)
                return (
                  <button
                    key={o.key}
                    type="button"
                    disabled={!editable}
                    onClick={() => {
                      setFieldValue(
                        fieldDef.key,
                        isSelected
                          ? selected.filter((s) => s !== o.key)
                          : [...selected, o.key],
                      )
                    }}
                    className={`px-3 py-1 rounded-full text-xs font-medium border transition-colors ${
                      isSelected
                        ? 'bg-primary text-white border-primary'
                        : 'bg-white text-slate-600 border-slate-200 hover:border-primary'
                    }`}
                  >
                    {o.label}
                  </button>
                )
              })}
            </div>
            {error && <p className="text-xs text-error mt-1">{error}</p>}
          </div>
        )
      }

      case 'relation': {
        if (fieldDef.options_ref === 'users') {
          const users = membersResult?.items ?? []
          return (
            <div key={fieldDef.key}>
              {labelEl}
              <select
                value={(value as string) ?? ''}
                onChange={(e) => setFieldValue(fieldDef.key, e.target.value || null)}
                disabled={!editable}
                className={inputClass(error)}
                data-testid={`field-${fieldDef.key}`}
              >
                <option value="">— Select —</option>
                {users.map((u) => (
                  <option key={u.id} value={u.id}>{u.name}</option>
                ))}
              </select>
              {error && <p className="text-xs text-error mt-1">{error}</p>}
            </div>
          )
        }
        return null
      }

      case 'date':
        return (
          <div key={fieldDef.key}>
            {labelEl}
            <input
              type="date"
              value={(value as string) ?? ''}
              onChange={(e) => setFieldValue(fieldDef.key, e.target.value)}
              disabled={!editable}
              className={inputClass(error)}
              data-testid={`field-${fieldDef.key}`}
            />
            {error && <p className="text-xs text-error mt-1">{error}</p>}
          </div>
        )

      default:
        return (
          <div key={fieldDef.key}>
            {labelEl}
            <input
              type="text"
              value={(value as string) ?? ''}
              onChange={(e) => setFieldValue(fieldDef.key, e.target.value)}
              disabled={!editable}
              className={inputClass(error)}
              data-testid={`field-${fieldDef.key}`}
            />
            {error && <p className="text-xs text-error mt-1">{error}</p>}
          </div>
        )
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6" data-testid="activity-form">
      {/* Core fields */}
      <div className="bg-surface-container-lowest p-8 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
        <h2 className="text-lg font-bold text-on-surface mb-6 font-headline">
          {isEditing ? 'Edit Activity' : 'New Activity'}
        </h2>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
          {/* Activity type */}
          <div>
            <label className="block text-xs font-bold uppercase tracking-widest text-slate-400 mb-1">
              Activity type <span className="text-error">*</span>
            </label>
            <select
              value={activityType}
              onChange={(e) => {
                setActivityType(e.target.value)
                setFields({})
                setTargetId('')
              }}
              disabled={isEditing || isLocked}
              className={inputClass(getFieldError('activityType'))}
              data-testid="activity-type-select"
              required
            >
              <option value="">— Select type —</option>
              {activityTypes.map((t) => (
                <option key={t.key} value={t.key}>{t.label}</option>
              ))}
            </select>
            {getFieldError('activityType') && (
              <p className="text-xs text-error mt-1">{getFieldError('activityType')}</p>
            )}
          </div>

          {/* Status — only shown when editing; on create the backend defaults to initial status */}
          {isEditing && (
            <div>
              <label className="block text-xs font-bold uppercase tracking-widest text-slate-400 mb-1">
                Status <span className="text-error">*</span>
              </label>
              <select
                value={status}
                onChange={(e) => setStatus(e.target.value)}
                disabled={isLocked}
                className={inputClass(getFieldError('status'))}
                data-testid="status-select"
                required
              >
                <option value="">— Select status —</option>
                {statuses.map((s) => (
                  <option key={s.key} value={s.key}>{s.label}</option>
                ))}
              </select>
            </div>
          )}

          {/* Due date */}
          <div>
            <label className="block text-xs font-bold uppercase tracking-widest text-slate-400 mb-1">
              Date <span className="text-error">*</span>
            </label>
            <input
              type="date"
              value={dueDate}
              onChange={(e) => setDueDate(e.target.value)}
              disabled={isLocked}
              className={inputClass(getFieldError('dueDate'))}
              data-testid="due-date-input"
              required
            />
            {getFieldError('dueDate') && (
              <p className="text-xs text-error mt-1">{getFieldError('dueDate')}</p>
            )}
          </div>

          {/* Duration */}
          <div>
            <label className="block text-xs font-bold uppercase tracking-widest text-slate-400 mb-1">
              Duration <span className="text-error">*</span>
            </label>
            <select
              value={duration}
              onChange={(e) => setDuration(e.target.value)}
              disabled={isLocked}
              className={inputClass(getFieldError('duration'))}
              data-testid="duration-select"
              required
            >
              <option value="">— Select duration —</option>
              {durations.map((d) => (
                <option key={d.key} value={d.key}>{d.label}</option>
              ))}
            </select>
          </div>

          {/* Target (for field activities) */}
          {isFieldActivity && (
            <div className="sm:col-span-2">
              <label className="block text-xs font-bold uppercase tracking-widest text-slate-400 mb-1">
                Target <span className="text-error">*</span>
              </label>
              <input
                type="text"
                placeholder="Search targets..."
                value={targetSearch}
                onChange={(e) => setTargetSearch(e.target.value)}
                disabled={isLocked}
                className={inputClass(getFieldError('targetId'))}
                data-testid="target-search"
              />
              {targetsResult && targetsResult.items.length > 0 && targetSearch && (
                <ul className="border border-slate-200 rounded-lg mt-1 max-h-40 overflow-y-auto bg-white">
                  {targetsResult.items.map((t) => (
                    <li key={t.id}>
                      <button
                        type="button"
                        onClick={() => {
                          setTargetId(t.id)
                          setTargetSearch(t.name)
                        }}
                        className={`w-full text-left px-3 py-2 text-sm hover:bg-slate-50 ${
                          t.id === targetId ? 'bg-primary-fixed text-primary font-medium' : ''
                        }`}
                      >
                        {t.name}
                        <span className="text-xs text-slate-400 ml-2">{t.targetType}</span>
                      </button>
                    </li>
                  ))}
                </ul>
              )}
              {targetId && (
                <p className="text-xs text-slate-500 mt-1">
                  Selected: {targetsResult?.items.find((t) => t.id === targetId)?.name ?? targetId}
                </p>
              )}
              {getFieldError('targetId') && (
                <p className="text-xs text-error mt-1">{getFieldError('targetId')}</p>
              )}
            </div>
          )}

        </div>
      </div>

      {/* Dynamic fields based on activity type */}
      {selectedType && selectedType.fields.filter((f) => !['duration', 'account_id'].includes(f.key)).length > 0 && (
        <div className="bg-surface-container-lowest p-8 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
          <h2 className="text-lg font-bold text-on-surface mb-6 font-headline">
            {selectedType.label} Details
          </h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
            {selectedType.fields
              .filter((f) => !['duration', 'account_id'].includes(f.key))
              .map((f) => renderDynamicField(f))}
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="flex items-center gap-4">
        <button
          type="submit"
          disabled={isSubmitting || isLocked}
          className="px-6 py-3 bg-primary text-white rounded-xl font-headline text-sm font-bold hover:bg-primary/90 transition-colors disabled:opacity-50"
          data-testid="submit-button"
        >
          {isSubmitting ? 'Saving...' : isEditing ? 'Update Activity' : 'Create Activity'}
        </button>
        <button
          type="button"
          onClick={onCancel}
          className="px-6 py-3 text-slate-500 rounded-xl font-headline text-sm font-medium hover:bg-slate-50 transition-colors"
        >
          Cancel
        </button>
      </div>
    </form>
  )
}

function inputClass(error?: string): string {
  return `w-full px-3 py-2 rounded-lg border text-sm transition-colors focus:outline-none focus:ring-2 focus:ring-primary/20 ${
    error
      ? 'border-error focus:border-error'
      : 'border-slate-200 focus:border-primary'
  } disabled:bg-slate-50 disabled:text-slate-400`
}
