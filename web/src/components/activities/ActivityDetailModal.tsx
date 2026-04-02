import { useState, useMemo, useEffect, useCallback } from 'react'
import { useActivity, usePatchActivity, usePatchActivityStatus, useSubmitActivity } from '@/hooks/useActivities'
import { useConfig } from '@/hooks/useConfig'
import { Badge } from '@/components/ui/Badge'
import { Spinner } from '@/components/ui/Spinner'
import { Modal } from '@/components/ui/Modal'
import { statusVariant, transitionColors } from '@/lib/styles'
import { str } from '@/lib/helpers'
import { Send, ExternalLink, AlertCircle, Check } from 'lucide-react'

interface ActivityDetailModalProps {
  readonly activityId: string | null
  readonly onClose: () => void
}

export function ActivityDetailModal({ activityId, onClose }: ActivityDetailModalProps) {
  const { data: activity, isLoading, refetch } = useActivity(activityId ?? '')
  const { data: config } = useConfig()
  const patchActivity = usePatchActivity()
  const patchStatus = usePatchActivityStatus()
  const submitActivity = useSubmitActivity()

  // Editable fields
  const [feedback, setFeedback] = useState('')
  const [promotedProducts, setPromotedProducts] = useState<string[]>([])
  const [validationErrors, setValidationErrors] = useState<Array<{ field: string; message: string }>>([])

  // Sync editable fields when activity loads
  useEffect(() => {
    if (activity) {
      setFeedback(str(activity.fields?.feedback))
      setPromotedProducts(Array.isArray(activity.fields?.promoted_products) ? (activity.fields.promoted_products as string[]) : [])
      setValidationErrors([])
    }
  }, [activity])

  const selectedType = useMemo(
    () => config?.activities.types.find((t) => t.key === activity?.activityType),
    [config, activity],
  )

  const submitRequired = useMemo(() => new Set(selectedType?.submit_required ?? []), [selectedType])

  const transitions = useMemo(() => {
    if (!config || !activity) return []
    return config.activities.status_transitions[activity.status] ?? []
  }, [config, activity])

  const statusLabel = useMemo(() => {
    if (!config || !activity) return activity?.status ?? ''
    return config.activities.statuses.find((s) => s.key === activity.status)?.label ?? activity.status
  }, [config, activity])

  const isFieldActivity = selectedType?.category === 'field'

  const isSubmittable = useMemo(() => {
    if (!config || !activity || !isFieldActivity) return false
    const statusDef = config.activities.statuses.find((s) => s.key === activity.status)
    return statusDef?.submittable === true && !activity.submittedAt
  }, [config, activity, isFieldActivity])

  const productOptions = useMemo(
    () => config?.options?.products ?? [],
    [config],
  )

  const getFieldError = useCallback(
    (field: string) => validationErrors.find((e) => e.field === field)?.message,
    [validationErrors],
  )

  if (!activityId) return null

  const title = isLoading || !activity
    ? 'Activity'
    : (activity.targetName ?? activity.label ?? activity.activityType)

  const buildFields = () => {
    const fields: Record<string, unknown> = {}
    if (feedback.trim()) fields.feedback = feedback.trim()
    if (promotedProducts.length > 0) fields.promoted_products = promotedProducts
    return Object.keys(fields).length > 0 ? fields : undefined
  }

  /** Save fields first (if edited), then execute the follow-up action. */
  const saveFieldsThen = (afterSave: () => void) => {
    if (!activity) return
    const fields = buildFields()
    if (fields) {
      patchActivity.mutate({ id: activity.id, fields }, { onSuccess: afterSave })
    } else {
      afterSave()
    }
  }

  const handleTransition = (nextStatus: string) => {
    if (!activity) return
    setValidationErrors([])
    saveFieldsThen(() => {
      patchStatus.mutate(
        { id: activity.id, status: nextStatus },
        { onSuccess: () => { refetch() } },
      )
    })
  }

  const validateSubmitFields = (): Array<{ field: string; message: string }> => {
    const errors: Array<{ field: string; message: string }> = []
    if (submitRequired.has('feedback') && !feedback.trim()) {
      errors.push({ field: 'feedback', message: 'Feedback is required before submission' })
    }
    if (submitRequired.has('promoted_products') && promotedProducts.length === 0) {
      errors.push({ field: 'promoted_products', message: 'Select at least one product' })
    }
    return errors
  }

  const handleSubmit = () => {
    if (!activity) return
    const errors = validateSubmitFields()
    if (errors.length > 0) {
      setValidationErrors(errors)
      return
    }
    setValidationErrors([])
    saveFieldsThen(() => {
      submitActivity.mutate(activity.id, {
        onSuccess: () => { refetch() },
        onError: (err) => {
          const apiErr = err as Error & { fields?: Array<{ field: string; message: string }> }
          if (apiErr.fields) setValidationErrors(apiErr.fields)
        },
      })
    })
  }

  const toggleProduct = (key: string) => {
    setPromotedProducts((prev) =>
      prev.includes(key) ? prev.filter((p) => p !== key) : [...prev, key],
    )
  }

  const canAct = transitions.length > 0 || isSubmittable
  const isLocked = !!activity?.submittedAt

  const actionFooter = activity && canAct && !isLocked ? (
    <div className="space-y-3">
      <div className="flex gap-2">
        {transitions.map((nextStatus) => {
          const label = config?.activities.statuses.find((s) => s.key === nextStatus)?.label ?? nextStatus
          const colorCls = transitionColors[nextStatus] ?? 'bg-slate-600 text-white hover:bg-slate-700'
          return (
            <button
              key={nextStatus}
              disabled={patchStatus.isPending}
              onClick={() => handleTransition(nextStatus)}
              className={`flex-1 rounded-lg px-4 py-2.5 text-sm font-medium shadow-sm transition-colors disabled:opacity-50 ${colorCls}`}
            >
              {label}
            </button>
          )
        })}
        {isSubmittable && (
          <button
            disabled={submitActivity.isPending || patchActivity.isPending}
            onClick={handleSubmit}
            className="flex-1 inline-flex items-center justify-center gap-2 rounded-lg bg-teal-600 px-4 py-2.5 text-sm font-medium text-white shadow-sm transition-colors hover:bg-teal-700 disabled:opacity-50"
          >
            <Send size={16} />
            Submit
          </button>
        )}
      </div>
    </div>
  ) : undefined

  return (
    <Modal open={!!activityId} onClose={onClose} title={title} footer={actionFooter}>
      {isLoading || !activity ? (
        <Spinner />
      ) : (
        <div className="space-y-4">
          {/* Status + date */}
          <div>
            <div className="flex flex-wrap items-center gap-2">
              <Badge variant={statusVariant[activity.status] ?? 'default'}>
                {statusLabel}
              </Badge>
              {activity.submittedAt && <Badge variant="success">Submitted</Badge>}
              {str(activity.fields?.visit_type) && (
                <span className={`text-[10px] font-medium px-1.5 py-0.5 rounded ${
                  str(activity.fields.visit_type) === 'f2f'
                    ? 'bg-amber-50 text-amber-700 border border-amber-200'
                    : 'bg-blue-50 text-blue-700 border border-blue-200'
                }`}>
                  {str(activity.fields.visit_type) === 'f2f' ? 'In person' : 'Remote'}
                </span>
              )}
            </div>
            <p className="mt-1 text-sm text-slate-500 capitalize">
              {activity.activityType} &middot; {new Date(activity.dueDate).toLocaleDateString('en-GB', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}
            </p>
          </div>

          {/* Target */}
          {activity.targetSummary && (
            <a
              href={`/targets/${activity.targetSummary.id}`}
              className="block rounded-lg border border-slate-200 p-3 hover:border-teal-300 hover:bg-teal-50/50 transition-colors group"
            >
              <p className="text-xs text-slate-500">Target</p>
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-slate-900 group-hover:text-teal-700">{activity.targetSummary.name}</p>
                  <p className="text-xs text-slate-500 capitalize">{activity.targetSummary.targetType}</p>
                </div>
                <ExternalLink size={14} className="text-slate-400 group-hover:text-teal-500" />
              </div>
            </a>
          )}

          {/* Promoted Products (editable if not locked) */}
          {selectedType?.fields.some((f) => f.key === 'promoted_products') && (
            <div>
              <label className="mb-1.5 block text-sm font-medium text-slate-700">
                Promoted Products {submitRequired.has('promoted_products') && <span className="text-red-400">*</span>}
              </label>
              {isLocked ? (
                <div className="flex flex-wrap gap-1">
                  {promotedProducts.map((p) => {
                    const label = productOptions.find((o) => o.key === p)?.label ?? p
                    return <Badge key={p}>{label}</Badge>
                  })}
                  {promotedProducts.length === 0 && <span className="text-sm text-slate-400">None</span>}
                </div>
              ) : (
                <div className="flex flex-wrap gap-2">
                  {productOptions.map((o) => {
                    const selected = promotedProducts.includes(o.key)
                    return (
                      <button
                        key={o.key}
                        type="button"
                        onClick={() => toggleProduct(o.key)}
                        className={`rounded-full border px-3 py-1.5 text-xs font-medium transition-colors ${
                          selected
                            ? 'border-teal-500 bg-teal-50 text-teal-700'
                            : 'border-slate-200 bg-white text-slate-600 hover:border-slate-300'
                        }`}
                      >
                        {selected && <Check size={12} className="mr-1 inline" />}
                        {o.label}
                      </button>
                    )
                  })}
                </div>
              )}
              {getFieldError('promoted_products') && (
                <p className="mt-1 flex items-center gap-1 text-xs text-red-600">
                  <AlertCircle size={12} /> {getFieldError('promoted_products')}
                </p>
              )}
            </div>
          )}

          {/* Feedback (editable if not locked, field activities only) */}
          {selectedType?.fields.some((f) => f.key === 'feedback') && <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">
              Feedback {submitRequired.has('feedback') && <span className="text-red-400">*</span>}
            </label>
            {isLocked ? (
              <p className="text-sm text-slate-700 whitespace-pre-wrap">{feedback || <span className="text-slate-400">No feedback</span>}</p>
            ) : (
              <textarea
                value={feedback}
                onChange={(e) => setFeedback(e.target.value)}
                maxLength={500}
                rows={3}
                placeholder="How did the visit go?"
                className={`w-full rounded-lg border px-3 py-2 text-sm focus:outline-none focus:ring-1 ${
                  getFieldError('feedback')
                    ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
                    : 'border-slate-300 focus:border-teal-500 focus:ring-teal-500'
                }`}
              />
            )}
            {getFieldError('feedback') && (
              <p className="mt-1 flex items-center gap-1 text-xs text-red-600">
                <AlertCircle size={12} /> {getFieldError('feedback')}
              </p>
            )}
          </div>}

          {/* Notes (read-only) */}
          {str(activity.fields?.notes) && (
            <div>
              <p className="mb-1 text-xs font-medium text-slate-500">Notes</p>
              <p className="text-sm text-slate-700 whitespace-pre-wrap">{str(activity.fields.notes)}</p>
            </div>
          )}

          {/* Tags */}
          {Array.isArray(activity.fields?.tags) && (activity.fields.tags as string[]).length > 0 && (
            <div className="flex flex-wrap gap-1">
              {(activity.fields.tags as string[]).map((tag) => (
                <Badge key={tag}>{tag}</Badge>
              ))}
            </div>
          )}

          {/* Metadata */}
          <dl className="space-y-1.5 text-sm border-t border-slate-100 pt-3">
            <div className="flex justify-between">
              <dt className="text-slate-500">Created</dt>
              <dd className="text-slate-900">{new Date(activity.createdAt).toLocaleString('en-GB')}</dd>
            </div>
            {activity.submittedAt && (
              <div className="flex justify-between">
                <dt className="text-slate-500">Submitted</dt>
                <dd className="text-slate-900">{new Date(activity.submittedAt).toLocaleString('en-GB')}</dd>
              </div>
            )}
            {activity.routing && (
              <div className="flex justify-between">
                <dt className="text-slate-500">Routing</dt>
                <dd className="text-slate-900">{activity.routing}</dd>
              </div>
            )}
          </dl>
        </div>
      )}
    </Modal>
  )
}
