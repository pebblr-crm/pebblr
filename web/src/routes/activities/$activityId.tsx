import { createRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { useTranslation } from 'react-i18next'
import { ArrowLeft, Lock, Send, Users } from 'lucide-react'
import { useState, useEffect, useRef } from 'react'
import { Route as rootRoute } from '../__root'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { SaveStateIndicator, SaveSuccessFlash } from '../../components/SaveStateIndicator'
import { useActivity, usePatchActivityStatus } from '../../services/activities'
import { useTarget, useTargets } from '../../services/targets'
import { useTeamMembers } from '../../services/teams'
import { useConfig } from '../../services/config'
import { useToast } from '../../hooks/useToast'
import { formatValidationToast } from '../../utils/fieldLabels'
import { useInlineActivityEditor } from '../../hooks/useInlineActivityEditor'
import { extractDate } from '@/utils/date'
import {
  getTypeConfig,
  getTypeLabel,
  getActivityTitle,
  getStatusLabel,
  getStatusBadgeColor,
  translateConfigLabel,
  getConfigFieldLabel,
  getOptionLabel,
} from '@/utils/config'
import type { FieldConfig, OptionDef, TenantConfig } from '@/types/config'
import { usePlannerState } from '@/contexts/planner'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/activities/$activityId',
  component: ActivityDetailPage,
})

export function ActivityDetailPage() {
  const { t } = useTranslation()
  const { activityId } = Route.useParams()
  const { state: { from } } = usePlannerState()
  const { data: activity, isLoading, isError, error } = useActivity(activityId)
  const { data: config } = useConfig()

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <LoadingSpinner size="lg" label={t('activity.loading')} />
      </div>
    )
  }

  if (isError) {
    return (
      <div data-testid="error-state" className="p-8 text-center text-error">
        {error instanceof Error ? error.message : t('activity.failedToLoad')}
      </div>
    )
  }

  if (!activity) {
    return (
      <div data-testid="not-found" className="p-8 text-center text-on-surface-variant">
        {t('activity.notFound')}
      </div>
    )
  }

  if (!config) {
    return (
      <div className="flex items-center justify-center h-64">
        <LoadingSpinner size="lg" label={t('activity.loadingConfig')} />
      </div>
    )
  }

  return <ActivityDetailInner activityId={activityId} config={config} from={from ?? undefined} />
}

// ── Inner component — rendered only when both activity + config are loaded ──

interface InnerProps {
  activityId: string
  config: TenantConfig
  from?: string
}

export function ActivityDetailInner({ activityId, config, from }: InnerProps) {
  const { t } = useTranslation()
  const { data: activity } = useActivity(activityId)
  const { showToast } = useToast()

  // activity is guaranteed here since ActivityDetailPage already checked
  const act = activity!

  const editor = useInlineActivityEditor(act)
  const { localData, saveState, fieldErrors, handleFieldChange, handleFieldBlur,
          handleStatusChange, handleSubmit, retrySave, isSubmitting } = editor

  const { data: membersResult } = useTeamMembers()

  const isSubmitted = Boolean(act.submittedAt)
  const typeConfig = getTypeConfig(config?.activities, act.activityType)
  const typeLabel = getTypeLabel(config?.activities, act.activityType)
  const hasDuration = typeConfig?.fields.some((f) => f.key === 'duration') ?? false
  const jointVisitUserId = act.jointVisitUserId
    ?? (act.fields?.joint_visit_user_id as string | undefined)
  const jointVisitorName = jointVisitUserId
    ? (membersResult?.items.find((u) => u.id === jointVisitUserId)?.name ?? '?')
    : undefined
  const activityTitle = getActivityTitle(config, act)
  const statusLabel = getStatusLabel(config?.activities, localData.status ?? act.status)
  const statusColor = getStatusBadgeColor(config?.activities, localData.status ?? act.status)
  const currentStatus = localData.status ?? act.status
  const allowedTransitions = config?.activities.status_transitions[currentStatus] ?? []
  const isSubmittable = config?.activities.statuses.find((s) => s.key === currentStatus)?.submittable ?? false

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

  const statusMutation = usePatchActivityStatus()

  function getFieldError(field: string): string | undefined {
    return fieldErrors.find((e) => e.field === field)?.message
  }

  // Show toast when save errors occur
  const prevSaveStateForToast = useRef(saveState)
  useEffect(() => {
    if (prevSaveStateForToast.current !== 'error' && saveState === 'error') {
      showToast(t('activityDetail.failedToSave'))
    }
    prevSaveStateForToast.current = saveState
  }, [saveState, showToast, t])

  // Show toast when field validation errors arrive
  const prevFieldErrorCount = useRef(0)
  useEffect(() => {
    if (fieldErrors.length > 0 && fieldErrors.length !== prevFieldErrorCount.current) {
      showToast(formatValidationToast(config, act.activityType, fieldErrors))
    }
    prevFieldErrorCount.current = fieldErrors.length
  }, [fieldErrors, config, act.activityType, showToast])

  // Confirmation modal for irreversible actions
  const [confirmAction, setConfirmAction] = useState<{
    title: string
    message: string
    onConfirm: () => void
  } | null>(null)

  function isTerminalStatus(status: string): boolean {
    const transitions = config?.activities.status_transitions[status] ?? []
    return transitions.length === 0
  }

  function requestStatusChange(toStatus: string) {
    if (isTerminalStatus(toStatus)) {
      const label = getStatusLabel(config?.activities, toStatus)
      setConfirmAction({
        title: t('activityDetail.markAs', { status: label }),
        message: t('activityDetail.markAsMessage', { status: label }),
        onConfirm: () => { handleStatusChange(toStatus); setConfirmAction(null) },
      })
    } else {
      handleStatusChange(toStatus)
    }
  }

  function requestSubmit() {
    if (saveState === 'error') {
      retrySave()
      return
    }
    if (!isSubmittable) {
      showToast(t('activityDetail.setStatusFirst'))
      return
    }
    setConfirmAction({
      title: t('activityDetail.submitReportQuestion'),
      message: t('activityDetail.submitReportMessage'),
      onConfirm: () => { setConfirmAction(null); void handleSubmit() },
    })
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="p-4 sm:p-8 max-w-4xl mx-auto w-full space-y-6 pb-28 sm:pb-8"
    >
      {/* Back link */}
      <Link
        to={from === 'planner' ? '/planner'
          : from === 'daily' ? '/planner/daily'
          : from === 'map' ? '/planner/map'
          : '/'}
        className="inline-flex items-center gap-2 text-sm font-medium text-on-surface-variant hover:text-primary transition-colors no-underline"
      >
        <ArrowLeft className="w-4 h-4" />
        {from === 'planner' ? t('activityDetail.backToPlanner')
          : from === 'daily' ? t('activityDetail.backToDaily')
          : from === 'map' ? t('activityDetail.backToMap')
          : t('activityDetail.backToDashboard')}
      </Link>

      {/* Header card */}
      <div className="bg-surface-container-lowest p-6 sm:p-8 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0">
            <h1 className="text-2xl sm:text-3xl font-extrabold tracking-tight text-primary font-headline">
              {activityTitle}
            </h1>
            <div className="flex flex-wrap items-center gap-2 mt-2">
              {/* Status badge (read-only indicator) */}
              <span
                className={`px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-tight ${statusColor}`}
                data-testid="status-badge"
              >
                {statusLabel}
              </span>

              {isSubmitted && (
                <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-tight bg-slate-200 text-slate-600">
                  <Lock className="w-3 h-3" />
                  {t('activityDetail.submitted')}
                </span>
              )}

              {jointVisitorName && (
                <span
                  className="inline-flex items-center gap-1 px-3 py-1 rounded-full text-[10px] font-bold tracking-tight bg-blue-50 text-blue-700"
                  data-testid="joint-visit-badge"
                >
                  <Users className="w-3 h-3" />
                  {t('activityDetail.jointVisit', { name: jointVisitorName })}
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
        <h2 className="text-lg font-bold text-on-surface mb-6 font-headline">{t('activity.details')}</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
          {/* Activity type — read-only, cannot change after creation */}
          <div>
            <label className={labelClass}>{t('activity.type')}</label>
            <p className="text-sm text-on-surface py-2">{typeLabel}</p>
          </div>

          {/* Due date */}
          <div>
            <label className={labelClass}>
              {t('activity.date')} <span className="text-error">*</span>
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

          {/* Duration (only for activity types with a duration field) */}
          {hasDuration && (
            <div>
              <label className={labelClass}>
                {t('activity.duration')} <span className="text-error">*</span>
              </label>
              <select
                value={localData.duration ?? ''}
                onChange={(e) => handleFieldChange('duration', e.target.value)}
                onBlur={handleFieldBlur}
                disabled={isSubmitted}
                className={inputClass(getFieldError('duration'))}
                data-testid="duration-select"
              >
                <option value="">{t('activity.selectDuration')}</option>
                {config.activities.durations.map((d) => (
                  <option key={d.key} value={d.key}>{translateConfigLabel(`duration.${d.key}`, d.label)}</option>
                ))}
              </select>
            </div>
          )}

          {/* Target (field activities) */}
          {getTypeConfig(config.activities, act.activityType)?.category === 'field' && (
            <div className="sm:col-span-2">
              <label className={labelClass}>
                {t('activity.target')} <span className="text-error">*</span>
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

        </div>
      </div>

      {/* Dynamic fields per activity type */}
      {typeConfig && typeConfig.fields.length > 0 && (
        <div className="bg-surface-container-lowest p-6 sm:p-8 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
          <h2 className="text-lg font-bold text-on-surface mb-6 font-headline">
            {t('activity.fields', { type: typeLabel })}
          </h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-6">
            {typeConfig.fields
              .filter((f) => !['duration', 'account_id'].includes(f.key))
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

      {/* Action bar — status transitions + submit */}
      {!isSubmitted && (
        <>
          {/* Mobile sticky bar */}
          <div className="fixed bottom-0 left-0 right-0 sm:hidden bg-white border-t border-slate-100 px-4 pt-3 pb-safe z-30">
            <div className="flex items-center gap-2">
              {allowedTransitions.map((toStatus) => (
                <StatusTransitionButton
                  key={toStatus}
                  toStatus={toStatus}
                  config={config}
                  isPending={statusMutation.isPending}
                  onClick={() => requestStatusChange(toStatus)}
                />
              ))}
              <div className="flex-1" />
              <SubmitButton
                saveState={saveState}
                isSubmitting={isSubmitting}
                isStatusPending={statusMutation.isPending}
                onClick={requestSubmit}
              />
            </div>
          </div>
          {/* Desktop inline */}
          <div className="hidden sm:flex items-center gap-3 justify-end">
            {allowedTransitions.map((toStatus) => (
              <StatusTransitionButton
                key={toStatus}
                toStatus={toStatus}
                config={config}
                isPending={statusMutation.isPending}
                onClick={() => requestStatusChange(toStatus)}
              />
            ))}
            <SubmitButton
              saveState={saveState}
              isSubmitting={isSubmitting}
              isStatusPending={statusMutation.isPending}
              onClick={requestSubmit}
            />
          </div>
        </>
      )}

      {/* Confirmation modal */}
      {confirmAction && (
        <ConfirmModal
          title={confirmAction.title}
          message={confirmAction.message}
          onConfirm={confirmAction.onConfirm}
          onCancel={() => setConfirmAction(null)}
        />
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
  const { t } = useTranslation()
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
      {isSubmitting ? t('activityDetail.submitting') : saveState === 'error' ? t('activityDetail.retrySave') : t('activityDetail.submitReport')}
    </button>
  )
}

interface StatusTransitionButtonProps {
  toStatus: string
  config: TenantConfig
  isPending: boolean
  onClick: () => void
}

function StatusTransitionButton({ toStatus, config, isPending, onClick }: StatusTransitionButtonProps) {
  const label = getStatusLabel(config?.activities, toStatus)
  const badgeColor = getStatusBadgeColor(config?.activities, toStatus)

  return (
    <button
      type="button"
      onClick={onClick}
      disabled={isPending}
      className={`inline-flex items-center justify-center px-4 py-2.5 min-h-[44px] text-sm font-bold rounded-xl border transition-colors disabled:opacity-50 ${badgeColor} border-current/20 hover:opacity-80`}
      data-testid={`status-transition-${toStatus}`}
    >
      {label}
    </button>
  )
}

// ── Confirmation modal ──────────────────────────────────────────────────────

interface ConfirmModalProps {
  title: string
  message: string
  onConfirm: () => void
  onCancel: () => void
}

function ConfirmModal({ title, message, onConfirm, onCancel }: ConfirmModalProps) {
  const { t } = useTranslation()
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="fixed inset-0 bg-black/40" onClick={onCancel} />
      <div className="relative bg-white rounded-2xl shadow-xl max-w-sm w-full p-6 space-y-4">
        <h3 className="text-lg font-bold text-on-surface font-headline">{title}</h3>
        <p className="text-sm text-on-surface-variant">{message}</p>
        <div className="flex items-center gap-3 justify-end pt-2">
          <button
            type="button"
            onClick={onCancel}
            className="px-4 py-2.5 min-h-[44px] text-sm font-medium text-slate-500 rounded-xl hover:bg-slate-50 transition-colors"
          >
            {t('common.goBack')}
          </button>
          <button
            type="button"
            onClick={onConfirm}
            className="px-4 py-2.5 min-h-[44px] text-sm font-bold text-white bg-primary rounded-xl hover:bg-primary/90 transition-colors"
            data-testid="confirm-action"
          >
            {t('common.confirm')}
          </button>
        </div>
      </div>
    </div>
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
  const { t } = useTranslation()
  const [search, setSearch] = useState('')
  const [modalOpen, setModalOpen] = useState(false)
  const { data: targetsResult } = useTargets({ q: search, limit: 20 })
  const { data: selectedTarget } = useTarget(value)

  const selectedName = selectedTarget?.name ?? value

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
        {selectedName || <span className="text-slate-400">{t('activity.searchTargetsEllipsis')}</span>}
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
              placeholder={t('activity.searchTargetsEllipsis')}
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
  const { t } = useTranslation()
  const { data: membersResult } = useTeamMembers()

  function resolveOptions(): OptionDef[] {
    if (fieldDef.options_ref) {
      let opts: OptionDef[] = []
      if (config.options[fieldDef.options_ref]) opts = config.options[fieldDef.options_ref]
      else if (fieldDef.options_ref === 'durations') opts = config.activities.durations
      const ref = fieldDef.options_ref
      return opts.map((o) => ({ key: o.key, label: getOptionLabel(ref, o.key, o.label) }))
    }
    if (fieldDef.options) return fieldDef.options.map((o) => ({ key: o, label: o }))
    return []
  }

  const labelEl = (
    <label className={labelClass}>
      {getConfigFieldLabel(fieldDef.key, fieldDef.label ?? fieldDef.key.replace(/_/g, ' '))}
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
            <option value="">{t('common.select')}</option>
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

    case 'relation': {
      if (fieldDef.options_ref === 'users') {
        const users = membersResult?.items ?? []
        const selectedUser = users.find((u) => u.id === value)
        if (disabled) {
          return (
            <div>
              {labelEl}
              <p className="text-sm text-on-surface py-2">{selectedUser?.name || (value as string) || '—'}</p>
            </div>
          )
        }
        return (
          <div>
            {labelEl}
            <select
              value={(value as string) ?? ''}
              onChange={(e) => onChange(e.target.value || null)}
              onBlur={onBlur}
              disabled={disabled}
              className={inputClass(error)}
              data-testid={`field-${fieldDef.key}`}
            >
              <option value="">{t('common.select')}</option>
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
