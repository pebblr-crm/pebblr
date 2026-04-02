import { useState, useMemo } from 'react'
import { useCreateActivity } from '@/hooks/useActivities'
import { useTargets } from '@/hooks/useTargets'
import { useRecoveryBalance } from '@/hooks/useDashboard'
import { useConfig } from '@/hooks/useConfig'
import { Button } from '@/components/ui/Button'
import { Modal } from '@/components/ui/Modal'
import { Check } from 'lucide-react'

function MultiSelectOption({ option, selected, onToggle }: {
  option: { key: string; label: string }
  selected: boolean
  onToggle: () => void
}) {
  return (
    <button
      type="button"
      onClick={onToggle}
      className={`rounded-full border px-3 py-1.5 text-xs font-medium transition-colors ${
        selected
          ? 'border-teal-500 bg-teal-50 text-teal-700'
          : 'border-slate-200 bg-white text-slate-600 hover:border-slate-300'
      }`}
    >
      {selected && <Check size={12} className="mr-1 inline" />}
      {option.label}
    </button>
  )
}

const CORE_WIDGET_FIELDS = new Set(['duration', 'account_id'])
const inputCls = 'w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500'

export function CreateActivityModal({ open, onClose }: Readonly<{ open: boolean; onClose: () => void }>) {
  const { data: config } = useConfig()
  const createActivity = useCreateActivity()
  const { data: recovery } = useRecoveryBalance({})

  const [activityType, setActivityType] = useState('')
  const [dueDate, setDueDate] = useState(() => new Date().toISOString().slice(0, 10))
  const [duration, setDuration] = useState('')
  const [targetId, setTargetId] = useState('')
  const [targetSearch, setTargetSearch] = useState('')
  const [fields, setFields] = useState<Record<string, unknown>>({})

  const { data: targetsResult } = useTargets({ q: targetSearch || undefined, limit: 20 })

  const allActivityTypes = useMemo(() => config?.activities.types ?? [], [config])
  const selectedType = useMemo(() => allActivityTypes.find((t) => t.key === activityType), [allActivityTypes, activityType])
  const isFieldActivity = selectedType?.category === 'field'
  const hasDuration = selectedType?.fields.some((f) => f.key === 'duration') ?? false
  const dynamicFields = useMemo(
    () => selectedType?.fields.filter((f) => !CORE_WIDGET_FIELDS.has(f.key)) ?? [],
    [selectedType],
  )

  const recoveryType = config?.recovery?.recovery_type ?? 'recovery'
  const isRecovery = activityType === recoveryType
  const recoveryEnabled = config?.recovery?.weekend_activity_flag === true

  const initialStatus = useMemo(
    () => config?.activities.statuses.find((s) => s.initial)?.key ?? 'planificat',
    [config],
  )
  const durations = useMemo(() => config?.activities.durations ?? [], [config])

  const resolveOptions = (fieldDef: { options?: string[]; options_ref?: string }) => {
    if (fieldDef.options) return fieldDef.options.map((o) => ({ key: o, label: o }))
    if (fieldDef.options_ref && config) return config.options[fieldDef.options_ref] ?? []
    return []
  }

  const setFieldValue = (key: string, value: unknown) => {
    setFields((prev) => ({ ...prev, [key]: value }))
  }

  const reset = () => {
    setActivityType('')
    setDueDate(new Date().toISOString().slice(0, 10))
    setDuration('')
    setTargetId('')
    setTargetSearch('')
    setFields({})
  }

  const handleTypeChange = (newType: string) => {
    setActivityType(newType)
    setFields({})
    setTargetId('')
    setTargetSearch('')
    setDuration('')
  }

  const handleSubmit = () => {
    const submitFields = { ...fields }
    // Extract routing from fields if present (goes to top-level)
    const routing = (submitFields.routing as string) ?? undefined
    delete submitFields.routing

    createActivity.mutate(
      {
        activityType: activityType,
        status: initialStatus,
        dueDate,
        duration: duration || (hasDuration ? durations[0]?.key ?? '' : ''),
        routing,
        fields: submitFields,
        targetId: isFieldActivity ? targetId : undefined,
      },
      {
        onSuccess: () => {
          reset()
          onClose()
        },
      },
    )
  }

  const canSubmit = activityType
    && dueDate
    && (!isFieldActivity || targetId)
    && !(isRecovery && recovery?.balance === 0)

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Log Activity"
      footer={
        <Button
          onClick={handleSubmit}
          disabled={createActivity.isPending || !canSubmit}
          className="w-full"
          size="lg"
        >
          {createActivity.isPending ? 'Creating...' : 'Create Activity'}
        </Button>
      }
    >
      <div className="space-y-4">
        {/* Activity type */}
        <div>
          <label htmlFor="field-activity-type" className="mb-1.5 block text-sm font-medium text-slate-700">Activity Type</label>
          <select
            id="field-activity-type"
            value={activityType}
            onChange={(e) => handleTypeChange(e.target.value)}
            className={inputCls}
          >
            <option value="">Select type...</option>
            {allActivityTypes.map((t) => (
              <option key={t.key} value={t.key}>{t.label}</option>
            ))}
          </select>
        </div>

        {/* Due date */}
        <div>
          <label htmlFor="field-due-date" className="mb-1.5 block text-sm font-medium text-slate-700">Date</label>
          <input
            id="field-due-date"
            type="date"
            value={dueDate}
            onChange={(e) => setDueDate(e.target.value)}
            className={inputCls}
            required
          />
        </div>

        {/* Duration (for types that have it) */}
        {hasDuration && (
          <div>
            <label htmlFor="field-duration" className="mb-1.5 block text-sm font-medium text-slate-700">Duration</label>
            <select id="field-duration" value={duration} onChange={(e) => setDuration(e.target.value)} className={inputCls}>
              <option value="">Select duration...</option>
              {durations.map((d) => (
                <option key={d.key} value={d.key}>{d.label}</option>
              ))}
            </select>
          </div>
        )}

        {/* Target search (for field activities) */}
        {isFieldActivity && (
          <div>
            <label htmlFor="field-target" className="mb-1.5 block text-sm font-medium text-slate-700">Target</label>
            <input
              id="field-target"
              type="text"
              value={targetSearch}
              onChange={(e) => { setTargetSearch(e.target.value); setTargetId('') }}
              placeholder="Search targets..."
              className={inputCls}
            />
            {targetsResult && targetsResult.items.length > 0 && targetSearch && !targetId && (
              <ul className="border border-slate-200 rounded-lg mt-1 max-h-40 overflow-y-auto bg-white shadow-sm">
                {targetsResult.items.map((t) => (
                  <li key={t.id}>
                    <button
                      type="button"
                      onClick={() => { setTargetId(t.id); setTargetSearch(t.name) }}
                      className="w-full text-left px-3 py-2 text-sm hover:bg-slate-50"
                    >
                      <span className="font-medium text-slate-900">{t.name}</span>
                      <span className="text-xs text-slate-400 ml-2 capitalize">{t.targetType}</span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
            {targetId && (
              <p className="text-xs text-teal-600 mt-1 font-medium">
                Selected: {targetsResult?.items.find((t) => t.id === targetId)?.name ?? targetId}
              </p>
            )}
          </div>
        )}

        {/* Recovery balance info */}
        {isRecovery && recoveryEnabled && recovery && (
          <div className={`rounded-lg border p-3 ${recovery.balance > 0 ? 'border-teal-200 bg-teal-50' : 'border-red-200 bg-red-50'}`}>
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-slate-700">Recovery Balance</span>
              <span className={`text-lg font-semibold ${recovery.balance > 0 ? 'text-teal-700' : 'text-red-700'}`}>
                {recovery.balance} day{recovery.balance === 1 ? '' : 's'}
              </span>
            </div>
            <p className="text-xs text-slate-500 mt-1">{recovery.earned} earned, {recovery.taken} taken</p>
            {recovery.balance === 0 && (
              <p className="text-xs text-red-600 font-medium mt-1.5">No recovery days available to claim.</p>
            )}
            {recovery.intervals?.filter((iv: { claimed: boolean }) => !iv.claimed).length > 0 && (
              <div className="mt-2 space-y-1">
                <p className="text-[10px] font-medium text-slate-500 uppercase">Unclaimed windows</p>
                {recovery.intervals.filter((iv) => !iv.claimed).map((iv) => (
                  <div key={iv.weekendDate} className="flex items-center justify-between text-xs text-slate-600">
                    <span>Weekend {new Date(iv.weekendDate).toLocaleDateString('en-GB', { day: 'numeric', month: 'short' })}</span>
                    <span className="text-slate-400">Claim by {new Date(iv.claimBy).toLocaleDateString('en-GB', { day: 'numeric', month: 'short' })}</span>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {/* Dynamic fields from config */}
        {dynamicFields.length > 0 && (
          <div className="space-y-4 border-t border-slate-100 pt-4">
            <p className="text-xs font-medium text-slate-500 uppercase tracking-wide">Details</p>
            {dynamicFields.map((fieldDef) => {
              const value = fields[fieldDef.key]
              const label = fieldDef.label ?? fieldDef.key
              const required = fieldDef.required

              switch (fieldDef.type) {
                case 'select': {
                  const opts = resolveOptions(fieldDef)
                  return (
                    <div key={fieldDef.key}>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">
                        {label} {required && <span className="text-red-400">*</span>}
                      </label>
                      <select
                        value={(value as string) ?? ''}
                        onChange={(e) => setFieldValue(fieldDef.key, e.target.value)}
                        className={inputCls}
                      >
                        <option value="">Select...</option>
                        {opts.map((o) => (
                          <option key={o.key} value={o.key}>{o.label}</option>
                        ))}
                      </select>
                    </div>
                  )
                }

                case 'multi_select': {
                  const opts = resolveOptions(fieldDef)
                  const selected = Array.isArray(value) ? (value as string[]) : []
                  return (
                    <div key={fieldDef.key}>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">
                        {label} {required && <span className="text-red-400">*</span>}
                      </label>
                      <div className="flex flex-wrap gap-2">
                        {opts.map((o) => (
                          <MultiSelectOption
                            key={o.key}
                            option={o}
                            selected={selected.includes(o.key)}
                            onToggle={() => setFieldValue(
                              fieldDef.key,
                              selected.includes(o.key) ? selected.filter((s) => s !== o.key) : [...selected, o.key],
                            )}
                          />
                        ))}
                      </div>
                    </div>
                  )
                }

                case 'text':
                  return (
                    <div key={fieldDef.key}>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">
                        {label} {required && <span className="text-red-400">*</span>}
                      </label>
                      <textarea
                        value={(value as string) ?? ''}
                        onChange={(e) => setFieldValue(fieldDef.key, e.target.value)}
                        rows={2}
                        maxLength={500}
                        placeholder={`Enter ${label.toLowerCase()}...`}
                        className={inputCls}
                      />
                    </div>
                  )

                case 'date':
                  return (
                    <div key={fieldDef.key}>
                      <label className="mb-1.5 block text-sm font-medium text-slate-700">
                        {label} {required && <span className="text-red-400">*</span>}
                      </label>
                      <input
                        type="date"
                        value={(value as string) ?? ''}
                        onChange={(e) => setFieldValue(fieldDef.key, e.target.value)}
                        className={inputCls}
                      />
                    </div>
                  )

                default:
                  return null
              }
            })}
          </div>
        )}
      </div>
    </Modal>
  )
}
