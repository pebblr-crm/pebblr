import { createRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { ArrowLeft, Lock, Send } from 'lucide-react'
import { Route as rootRoute } from '../__root'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { useActivity, useSubmitActivity, usePatchActivityStatus } from '../../services/activities'
import { useConfig } from '../../services/config'
import { displayDate } from '@/utils/date'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/activities/$activityId',
  component: ActivityDetailPage,
})

export function ActivityDetailPage() {
  const { activityId } = Route.useParams()
  const { data: activity, isLoading, isError, error } = useActivity(activityId)
  const { data: config } = useConfig()
  const submitMutation = useSubmitActivity()
  const statusMutation = usePatchActivityStatus()

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

  const typeConfig = config?.activities.types.find((t) => t.key === activity.activityType)
  const typeLabel = typeConfig?.label ?? activity.activityType
  const statusLabel = config?.activities.statuses.find((s) => s.key === activity.status)?.label ?? activity.status
  const durationLabel = config?.activities.durations.find((d) => d.key === activity.duration)?.label ?? activity.duration
  const isSubmitted = Boolean(activity.submittedAt)

  const allowedTransitions = config?.activities.status_transitions[activity.status] ?? []

  function resolveOptionLabel(ref: string, value: string): string {
    const opts = config?.options[ref]
    if (!opts) return value
    return opts.find((o) => o.key === value)?.label ?? value
  }

  function getFieldDisplay(fieldKey: string, value: unknown): string {
    if (value == null || value === '') return '—'
    const fieldDef = typeConfig?.fields.find((f) => f.key === fieldKey)
    if (fieldDef?.options_ref && typeof value === 'string') {
      return resolveOptionLabel(fieldDef.options_ref, value)
    }
    if (Array.isArray(value)) {
      const fieldDef2 = typeConfig?.fields.find((f) => f.key === fieldKey)
      if (fieldDef2?.options_ref) {
        return value.map((v) => resolveOptionLabel(fieldDef2.options_ref!, String(v))).join(', ')
      }
      return value.join(', ')
    }
    return String(value)
  }

  const statusColor = activity.status === 'realizat'
    ? 'bg-emerald-100 text-emerald-700'
    : activity.status === 'anulat'
      ? 'bg-red-100 text-red-700'
      : 'bg-primary-fixed text-primary'

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="p-8 max-w-4xl mx-auto w-full space-y-8"
    >
      {/* Back link */}
      <Link
        to="/"
        className="inline-flex items-center gap-2 text-sm font-medium text-on-surface-variant hover:text-primary transition-colors no-underline"
      >
        <ArrowLeft className="w-4 h-4" />
        Back to dashboard
      </Link>

      {/* Header */}
      <div className="bg-surface-container-lowest p-8 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-3xl font-extrabold tracking-tight text-primary font-headline">
              {typeLabel}
            </h1>
            <div className="flex items-center gap-3 mt-2">
              <span className={`px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-tight ${statusColor}`}>
                {statusLabel}
              </span>
              <span className="text-sm text-on-surface-variant">
                {displayDate(activity.dueDate)}
              </span>
              <span className="text-sm text-on-surface-variant">
                {durationLabel}
              </span>
              {isSubmitted && (
                <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-tight bg-slate-200 text-slate-600">
                  <Lock className="w-3 h-3" />
                  Submitted
                </span>
              )}
            </div>
          </div>

          <div className="flex items-center gap-2">
            {!isSubmitted && (
              <>
                <Link
                  to="/activities/$activityId/edit"
                  params={{ activityId }}
                  className="px-4 py-2 text-sm font-medium text-primary border border-primary rounded-lg hover:bg-primary-fixed transition-colors no-underline"
                >
                  Edit
                </Link>
                <button
                  onClick={() => submitMutation.mutate(activityId)}
                  disabled={submitMutation.isPending}
                  className="inline-flex items-center gap-2 px-4 py-2 text-sm font-medium text-white bg-primary rounded-lg hover:bg-primary/90 transition-colors disabled:opacity-50"
                >
                  <Send className="w-4 h-4" />
                  {submitMutation.isPending ? 'Submitting...' : 'Submit Report'}
                </button>
              </>
            )}
          </div>
        </div>
      </div>

      {/* Core details */}
      <div className="bg-surface-container-lowest p-8 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
        <h2 className="text-lg font-bold text-on-surface mb-6 font-headline">Details</h2>
        <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-4">
          <div>
            <dt className="text-xs font-bold uppercase tracking-widest text-slate-400 mb-1">Type</dt>
            <dd className="text-sm text-on-surface">{typeLabel}</dd>
          </div>
          <div>
            <dt className="text-xs font-bold uppercase tracking-widest text-slate-400 mb-1">Date</dt>
            <dd className="text-sm text-on-surface">{displayDate(activity.dueDate)}</dd>
          </div>
          <div>
            <dt className="text-xs font-bold uppercase tracking-widest text-slate-400 mb-1">Duration</dt>
            <dd className="text-sm text-on-surface">{durationLabel}</dd>
          </div>
          <div>
            <dt className="text-xs font-bold uppercase tracking-widest text-slate-400 mb-1">Status</dt>
            <dd className="text-sm text-on-surface">{statusLabel}</dd>
          </div>
          {activity.routing && (
            <div>
              <dt className="text-xs font-bold uppercase tracking-widest text-slate-400 mb-1">Routing</dt>
              <dd className="text-sm text-on-surface">{activity.routing}</dd>
            </div>
          )}
          {activity.targetId && (
            <div>
              <dt className="text-xs font-bold uppercase tracking-widest text-slate-400 mb-1">Target</dt>
              <dd className="text-sm text-on-surface">
                <Link
                  to="/targets/$targetId"
                  params={{ targetId: activity.targetId }}
                  className="text-primary hover:underline no-underline"
                >
                  {activity.targetId}
                </Link>
              </dd>
            </div>
          )}
        </dl>
      </div>

      {/* Dynamic fields */}
      {typeConfig && typeConfig.fields.length > 0 && (
        <div className="bg-surface-container-lowest p-8 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
          <h2 className="text-lg font-bold text-on-surface mb-6 font-headline">
            {typeLabel} Fields
          </h2>
          <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-4">
            {typeConfig.fields.map((f) => (
              <div key={f.key}>
                <dt className="text-xs font-bold uppercase tracking-widest text-slate-400 mb-1">
                  {f.key.replace(/_/g, ' ')}
                </dt>
                <dd className="text-sm text-on-surface">
                  {getFieldDisplay(f.key, activity.fields[f.key])}
                </dd>
              </div>
            ))}
          </dl>
        </div>
      )}

      {/* Status transitions */}
      {!isSubmitted && allowedTransitions.length > 0 && (
        <div className="bg-surface-container-lowest p-8 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
          <h2 className="text-lg font-bold text-on-surface mb-4 font-headline">Change Status</h2>
          <div className="flex flex-wrap gap-2">
            {allowedTransitions.map((toStatus) => {
              const label = config?.activities.statuses.find((s) => s.key === toStatus)?.label ?? toStatus
              return (
                <button
                  key={toStatus}
                  onClick={() => statusMutation.mutate({ id: activityId, status: toStatus })}
                  disabled={statusMutation.isPending}
                  className="px-4 py-2 text-sm font-medium border border-slate-200 rounded-lg hover:bg-slate-50 transition-colors disabled:opacity-50"
                >
                  {label}
                </button>
              )
            })}
          </div>
        </div>
      )}
    </motion.div>
  )
}
