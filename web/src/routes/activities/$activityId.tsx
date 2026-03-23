import { createRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { ArrowLeft, Lock, Send } from 'lucide-react'
import { useState, useEffect, useRef } from 'react'
import { Drawer } from 'vaul'
import { Route as rootRoute } from '../__root'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { SaveStateIndicator, SaveSuccessFlash } from '../../components/SaveStateIndicator'
import { useActivity, usePatchActivityStatus } from '../../services/activities'
import { useTargets } from '../../services/targets'
import { useConfig } from '../../services/config'
import { useInlineActivityEditor } from '../../hooks/useInlineActivityEditor'
import { extractDate } from '@/utils/date'
import {
  getTypeConfig,
  getTypeLabel,
  getStatusLabel,
  STATUS_BADGE_COLORS,
} from '@/utils/config'
import type { FieldConfig, OptionDef, TenantConfig } from '@/types/config'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/activities/$activityId',
  component: ActivityDetailPage,
})

export function ActivityDetailPage() {
  const { activityId } = Route.useParams()
  const { data: activity, isLoading, isError, error } = useActivity(activityId)
  const { data: config } = useConfig()

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <LoadingSpinner size="lg" label="Loading activity..." />
      </div>
    )
  }

  if (isError) {
    return (
      <div data-testid="error-state" className="p-8 text-center text-error">
        {error instanceof Error ? error.message : 'Failed to load activity.'}
      </div>
    )
  }

  if (!activity) {
    return (
      <div data-testid="not-found" className="p-8 text-center text-on-surface-variant">
        Activity not found.
      </div>
    )
  }

  if (!config) {
    return (
      <div className="flex items-center justify-center h-64">
        <LoadingSpinner size="lg" label="Loading configuration..." />
      </div>
    )
  }

  return <ActivityDetailInner activityId={activityId} config={config} />
}

// ── Inner component — rendered only when both activity + config are loaded ──

interface InnerProps {
  activityId: string
  config: TenantConfig
}

export function ActivityDetailInner({ activityId, config }: InnerProps) {
  const { data: activity } = useActivity(activityId)

  // activity is guaranteed here since ActivityDetailPage already checked
  const act = activity!

  const editor = useInlineActivityEditor(act)
  const { localData, saveState, fieldErrors, handleFieldChange, handleFieldBlur,
          handleStatusChange, handleSubmit, retrySave, isSubmitting } = editor

  const isSubmitted = Boolean(act.submittedAt)
  const typeConfig = getTypeConfig(config?.activities, act.activityType)
  const typeLabel = getTypeLabel(config?.activities, act.activityType)
  const statusLabel = getStatusLabel(config?.activities, localData.status ?? act.status)
  const statusColor = STATUS_BADGE_COLORS[localData.status ?? act.status] ?? 'bg-primary-fixed text-primary'
  const allowedTransitions = config?.activities.status_transitions[localData.status ?? act.status] ?? []

  // Flash "Saved" for 500ms after transition from saving→idle
  const [showSavedFlash, setShowSavedFlash] = useState(false)
  const prevSaveState = useRef(saveState)
  useEffect(() => {
    let cleanup: (() => void) | undefined
    if (prevSaveState.current === 'saving' && saveState === 'idle') {
      // Defer setState to avoid synchronous setState in effect body.
      const tOn = setTimeout(() => {
        setShowSavedFlash(true)
        const tOff = setTimeout(() => setShowSavedFlash(false), 500)
        cleanup = () => clearTimeout(tOff)
      }, 0)
      cleanup = () => clearTimeout(tOn)
    }
    prevSaveState.current = saveState
    return cleanup
  }, [saveState])

  // Status bottom sheet (mobile) / inline popover (desktop)
  const [statusSheetOpen, setStatusSheetOpen] = useState(false)
  const statusMutation = usePatchActivityStatus()

  function onStatusSelect(toStatus: string) {
    handleStatusChange(toStatus)
    setStatusSheetOpen(false)
  }

  function getFieldError(field: string): string | undefined {
    return fieldErrors.find((e) => e.field === field)?.message
  }

  async function onSubmitClick() {
    if (saveState === 'error') {
      retrySave()
      return
    }
    await handleSubmit()
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="p-4 sm:p-8 max-w-4xl mx-auto w-full space-y-6 pb-28 sm:pb-8"
    >
      {/* Back link */}
      <Link
        to="/"
        className="inline-flex items-center gap-2 text-sm font-medium text-on-surface-variant hover:text-primary transition-colors no-underline"
      >
        <ArrowLeft className="w-4 h-4" />
        Back to dashboard
      </Link>

      {/* Header card */}
      <div className="bg-surface-container-lowest p-6 sm:p-8 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0">
            <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight text-primary font-headline">
              {typeLabel}
            </h1>
            <div className="flex flex-wrap items-center gap-2 mt-2">
              {/* Tappable status badge */}
              {!isSubmitted ? (
                <button
                  type="button"
                  onClick={() => setStatusSheetOpen(true)}
                  disabled={statusMutation.isPending}
                  className={`px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-tight min-h-[36px] min-w-[44px] transition-opacity disabled:opacity-50 ${statusColor}`}
                  data-testid="status-badge"
                  aria-label={`Status: ${statusLabel}. Tap to change.`}
                >
                  {statusLabel}
                </button>
              ) : (
                <span className={`px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-tight ${statusColor}`}>
                  {statusLabel}
                </span>
              )}

              {isSubmitted && (
                <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-tight bg-slate-200 text-slate-600">
                  <Lock className="w-3 h-3" />
                  Submitted
                </span>
              )}
            </div>
          </div>

          {/* Save state indicator */}
          <div className="flex-shrink-0 mt-1">
            {showSavedFlash ? (
              <SaveSuccessFlash />
            ) : (
              <SaveStateIndicator saveState={saveState} onRetry={retrySave} />
            )}
          </div>
        </div>
      </div>

      {/* Core editable fields */}
      <div className="bg-surface-container-lowest p-6 sm:p-8 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
        <h2 className="text-lg font-bold text-on-surface mb-6 font-headline">Details</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
          {/* Activity type — read-only, cannot change after creation */}
          <div>
            <label className={labelClass}>Type</label>
            <p className="text-sm text-on-surface py-2">{typeLabel}</p>
          </div>

          {/* Due date */}
          <div>
            <label className={labelClass}>
              Date <span className="text-error">*</span>
            </label>
            <input
              type="date"
              value={localData.dueDate ? extractDate(localData.dueDate) : ''}
              onChange={(e) => handleFieldChange('dueDate', e.target.value)}
              onBlur={handleFieldBlur}
              disabled={isSubmitted}
              className={inputClass(getFieldError('dueDate'))}
              data-testid="due-date-input"
            />
            {getFieldError('dueDate') && (
              <p className="text-xs text-error mt-1">{getFieldError('dueDate')}</p>
            )}
          </div>

          {/* Duration */}
          <div>
            <label className={labelClass}>
              Duration <span className="text-error">*</span>
            </label>
            <select
              value={localData.duration ?? ''}
              onChange={(e) => handleFieldChange('duration', e.target.value)}
              onBlur={handleFieldBlur}
              disabled={isSubmitted}
              className={inputClass(getFieldError('duration'))}
              data-testid="duration-select"
            >
              <option value="">— Select duration —</option>
              {config.activities.durations.map((d) => (
                <option key={d.key} value={d.key}>{d.label}</option>
              ))}
            </select>
          </div>

          {/* Target (field activities) */}
          {getTypeConfig(config.activities, act.activityType)?.category === 'field' && (
            <div className="sm:col-span-2">
              <label className={labelClass}>
                Target <span className="text-error">*</span>
              </label>
              <TargetField
                value={localData.targetId ?? ''}
                onChange={(id) => handleFieldChange('targetId', id)}
                onBlur={handleFieldBlur}
                disabled={isSubmitted}
                error={getFieldError('targetId')}
              />
            </div>
          )}

          {/* Joint visit user */}
          <div>
            <label className={labelClass}>Joint visit user</label>
            <input
              type="text"
              value={localData.jointVisitUserId ?? ''}
              onChange={(e) => handleFieldChange('jointVisitUserId', e.target.value)}
              onBlur={handleFieldBlur}
              disabled={isSubmitted}
              placeholder="User ID (optional)"
              className={inputClass()}
              data-testid="joint-visit-input"
            />
          </div>
        </div>
      </div>

      {/* Dynamic fields per activity type */}
      {typeConfig && typeConfig.fields.length > 0 && (
        <div className="bg-surface-container-lowest p-6 sm:p-8 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
          <h2 className="text-lg font-bold text-on-surface mb-6 font-headline">
            {typeLabel} Fields
          </h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
            {typeConfig.fields
              .filter((f) => !['duration', 'account_id', 'joint_visit_user_id'].includes(f.key))
              .map((f) => (
                <DynamicField
                  key={f.key}
                  fieldDef={f}
                  value={localData.fields?.[f.key]}
                  onChange={(v) => handleFieldChange(f.key, v)}
                  onBlur={handleFieldBlur}
                  disabled={isSubmitted || f.editable === false}
                  error={getFieldError(f.key)}
                  config={config}
                />
              ))}
          </div>
        </div>
      )}

      {/* Status transition bottom sheet (mobile) */}
      <Drawer.Root open={statusSheetOpen} onOpenChange={setStatusSheetOpen}>
        <Drawer.Portal>
          <Drawer.Overlay className="fixed inset-0 bg-black/40 z-40" />
          <Drawer.Content className="fixed bottom-0 left-0 right-0 z-50 bg-white rounded-t-2xl pb-safe focus:outline-none">
            <div className="p-4">
              <div className="mx-auto w-12 h-1.5 rounded-full bg-slate-200 mb-6" />
              <h3 className="text-base font-bold text-on-surface mb-4 font-headline px-2">
                Change Status
              </h3>
              <div className="space-y-1">
                {allowedTransitions.map((toStatus) => {
                  const label = getStatusLabel(config?.activities, toStatus)
                  return (
                    <button
                      key={toStatus}
                      type="button"
                      onClick={() => onStatusSelect(toStatus)}
                      className="w-full text-left px-4 py-4 min-h-[56px] text-sm font-medium rounded-xl hover:bg-slate-50 transition-colors"
                    >
                      {label}
                    </button>
                  )
                })}
                <button
                  type="button"
                  onClick={() => setStatusSheetOpen(false)}
                  className="w-full text-left px-4 py-4 min-h-[56px] text-sm text-slate-400 rounded-xl hover:bg-slate-50 transition-colors"
                >
                  Cancel
                </button>
              </div>
            </div>
          </Drawer.Content>
        </Drawer.Portal>
      </Drawer.Root>

      {/* Sticky submit bar (mobile) / inline (desktop) */}
      {!isSubmitted && (
        <>
          {/* Mobile sticky bar */}
          <div className="fixed bottom-0 left-0 right-0 sm:hidden bg-white border-t border-slate-100 px-4 pt-3 pb-safe z-30">
            <SubmitButton
              saveState={saveState}
              isSubmitting={isSubmitting}
              isStatusPending={statusMutation.isPending}
              onClick={() => void onSubmitClick()}
            />
          </div>
          {/* Desktop inline */}
          <div className="hidden sm:flex justify-end">
            <SubmitButton
              saveState={saveState}
              isSubmitting={isSubmitting}
              isStatusPending={statusMutation.isPending}
              onClick={() => void onSubmitClick()}
            />
          </div>
        </>
      )}
    </motion.div>
  )
}

// ── Sub-components ──────────────────────────────────────────────────────────

interface SubmitButtonProps {
  saveState: string
  isSubmitting: boolean
  isStatusPending: boolean
  onClick: () => void
}

function SubmitButton({ saveState, isSubmitting, isStatusPending, onClick }: SubmitButtonProps) {
  const disabled = isSubmitting || isStatusPending
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-6 py-3 min-h-[44px] text-sm font-medium text-white bg-primary rounded-xl hover:bg-primary/90 transition-colors disabled:opacity-50"
      data-testid="submit-report-button"
    >
      <Send className="w-4 h-4" />
      {isSubmitting ? 'Submitting…' : saveState === 'error' ? 'Retry Save' : 'Submit Report'}
    </button>
  )
}

// ── Target field with full-screen search on mobile ──────────────────────────

interface TargetFieldProps {
  value: string
  onChange: (id: string) => void
  onBlur: () => void
  disabled: boolean
  error?: string
}

function TargetField({ value, onChange, onBlur, disabled, error }: TargetFieldProps) {
  const [search, setSearch] = useState('')
  const [modalOpen, setModalOpen] = useState(false)
  const { data: targetsResult } = useTargets({ q: search, limit: 20 })
  const { data: selectedTargetsResult } = useTargets({ q: value, limit: 5 })

  const selectedName = selectedTargetsResult?.items.find((t) => t.id === value)?.name ?? value

  if (disabled) {
    return <p className="text-sm text-on-surface py-2">{selectedName || '—'}</p>
  }

  return (
    <>
      <button
        type="button"
        onClick={() => setModalOpen(true)}
        onBlur={onBlur}
        className={`w-full text-left ${inputClass(error)}`}
        data-testid="target-search-trigger"
      >
        {selectedName || <span className="text-slate-400">Search targets…</span>}
      </button>
      {error && <p className="text-xs text-error mt-1">{error}</p>}

      {/* Full-screen search modal (preferred on mobile, also used on desktop for simplicity) */}
      {modalOpen && (
        <div className="fixed inset-0 z-50 bg-white flex flex-col" data-testid="target-search-modal">
          <div className="flex items-center gap-3 p-4 border-b border-slate-100">
            <button
              type="button"
              onClick={() => { setModalOpen(false); setSearch('') }}
              className="p-2 rounded-lg hover:bg-slate-50 transition-colors"
              aria-label="Close"
            >
              <ArrowLeft className="w-5 h-5" />
            </button>
            <input
              type="text"
              autoFocus
              placeholder="Search targets…"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="flex-1 text-sm border-none focus:outline-none"
              data-testid="target-search-input"
            />
          </div>
          <ul className="flex-1 overflow-y-auto">
            {targetsResult?.items.map((t) => (
              <li key={t.id}>
                <button
                  type="button"
                  onClick={() => {
                    onChange(t.id)
                    setSearch('')
                    setModalOpen(false)
                  }}
                  className={`w-full text-left px-4 py-4 min-h-[44px] text-sm hover:bg-slate-50 transition-colors ${
                    t.id === value ? 'bg-primary-fixed text-primary font-medium' : ''
                  }`}
                >
                  {t.name}
                  <span className="text-xs text-slate-400 ml-2">{t.targetType}</span>
                </button>
              </li>
            ))}
          </ul>
        </div>
      )}
    </>
  )
}

// ── Dynamic field renderer ──────────────────────────────────────────────────

interface DynamicFieldProps {
  fieldDef: FieldConfig
  value: unknown
  onChange: (v: unknown) => void
  onBlur: () => void
  disabled: boolean
  error?: string
  config: TenantConfig
}

function DynamicField({ fieldDef, value, onChange, onBlur, disabled, error, config }: DynamicFieldProps) {
  function resolveOptions(): OptionDef[] {
    if (fieldDef.options_ref) {
      if (config.options[fieldDef.options_ref]) return config.options[fieldDef.options_ref]
      if (fieldDef.options_ref === 'durations') return config.activities.durations
      return []
    }
    if (fieldDef.options) return fieldDef.options.map((o) => ({ key: o, label: o }))
    return []
  }

  const labelEl = (
    <label className={labelClass}>
      {fieldDef.key.replace(/_/g, ' ')}
      {fieldDef.required && <span className="text-error ml-1">*</span>}
    </label>
  )

  switch (fieldDef.type) {
    case 'select': {
      const opts = resolveOptions()
      return (
        <div>
          {labelEl}
          <select
            value={(value as string) ?? ''}
            onChange={(e) => onChange(e.target.value)}
            onBlur={onBlur}
            disabled={disabled}
            className={inputClass(error)}
            data-testid={`field-${fieldDef.key}`}
          >
            <option value="">— Select —</option>
            {opts.map((o) => <option key={o.key} value={o.key}>{o.label}</option>)}
          </select>
          {error && <p className="text-xs text-error mt-1">{error}</p>}
        </div>
      )
    }

    case 'multi_select': {
      const opts = resolveOptions()
      const selected = Array.isArray(value) ? (value as string[]) : []
      return (
        <div>
          {labelEl}
          <div className="flex flex-wrap gap-2" data-testid={`field-${fieldDef.key}`}>
            {opts.map((o) => {
              const isSelected = selected.includes(o.key)
              return (
                <button
                  key={o.key}
                  type="button"
                  disabled={disabled}
                  onClick={() => {
                    onChange(isSelected ? selected.filter((s) => s !== o.key) : [...selected, o.key])
                  }}
                  className={`px-2 rounded-full text-xs font-medium border transition-colors min-h-[36px] px-[8px] ${
                    isSelected
                      ? 'bg-primary text-white border-primary'
                      : 'bg-white text-slate-600 border-slate-200 hover:border-primary'
                  } disabled:opacity-50`}
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

    case 'date':
      return (
        <div>
          {labelEl}
          <input
            type="date"
            value={(value as string) ?? ''}
            onChange={(e) => onChange(e.target.value)}
            onBlur={onBlur}
            disabled={disabled}
            className={`${inputClass(error)} min-h-[44px]`}
            data-testid={`field-${fieldDef.key}`}
          />
          {error && <p className="text-xs text-error mt-1">{error}</p>}
        </div>
      )

    default:
      return (
        <div>
          {labelEl}
          <input
            type="text"
            value={(value as string) ?? ''}
            onChange={(e) => onChange(e.target.value)}
            onBlur={onBlur}
            disabled={disabled}
            className={inputClass(error)}
            data-testid={`field-${fieldDef.key}`}
          />
          {error && <p className="text-xs text-error mt-1">{error}</p>}
        </div>
      )
  }
}

// ── Style helpers ───────────────────────────────────────────────────────────

const labelClass = 'block text-xs font-bold uppercase tracking-widest text-slate-400 mb-1'

function inputClass(error?: string): string {
  return `w-full px-3 py-2 min-h-[44px] rounded-lg border text-sm transition-colors focus:outline-none focus:ring-2 focus:ring-primary/20 ${
    error
      ? 'border-error focus:border-error'
      : 'border-slate-200 focus:border-primary'
  } disabled:bg-slate-50 disabled:text-slate-400`
}
