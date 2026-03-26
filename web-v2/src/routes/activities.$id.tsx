import { useMemo } from 'react'
import { createRoute, useParams, Link } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { useActivity, usePatchActivityStatus, useSubmitActivity } from '@/hooks/useActivities'
import { useConfig } from '@/hooks/useConfig'
import { Badge } from '@/components/ui/Badge'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { ArrowLeft, Send } from 'lucide-react'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/activities/$id',
  component: ActivityDetailPage,
})

const statusVariant: Record<string, 'primary' | 'success' | 'danger' | 'default'> = {
  planificat: 'primary',
  realizat: 'success',
  anulat: 'danger',
}

function str(v: unknown): string {
  return typeof v === 'string' ? v : ''
}

function ActivityDetailPage() {
  const { id } = useParams({ from: '/activities/$id' })
  const { data: activity, isLoading } = useActivity(id)
  const { data: config } = useConfig()
  const patchStatus = usePatchActivityStatus()
  const submitActivity = useSubmitActivity()

  const transitions = useMemo(() => {
    if (!config || !activity) return []
    return config.activities.status_transitions[activity.status] ?? []
  }, [config, activity])

  const statusLabel = useMemo(() => {
    if (!config || !activity) return activity?.status ?? ''
    return config.activities.statuses.find((s) => s.key === activity.status)?.label ?? activity.status
  }, [config, activity])

  const isSubmittable = useMemo(() => {
    if (!config || !activity) return false
    const statusDef = config.activities.statuses.find((s) => s.key === activity.status)
    return statusDef?.submittable === true && !activity.submittedAt
  }, [config, activity])

  if (isLoading || !activity) return <Spinner />

  return (
    <div className="mx-auto max-w-2xl p-4 md:p-6">
      {/* Back link */}
      <Link to="/activities" className="mb-4 inline-flex items-center gap-1 text-sm text-slate-500 hover:text-slate-700">
        <ArrowLeft size={16} />
        Back to activities
      </Link>

      {/* Header */}
      <div className="mb-6">
        <div className="flex flex-wrap items-center gap-2">
          <h1 className="text-xl font-bold text-slate-900">
            {activity.targetName ?? activity.label ?? activity.activityType}
          </h1>
          <Badge variant={statusVariant[activity.status] ?? 'default'}>
            {statusLabel}
          </Badge>
          {activity.submittedAt && <Badge variant="success">Submitted</Badge>}
        </div>
        <p className="mt-1 text-sm text-slate-500 capitalize">
          {activity.activityType} &middot; {activity.duration} &middot; {new Date(activity.dueDate).toLocaleDateString('en-GB', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}
        </p>
      </div>

      <div className="space-y-4">
        {/* Target info */}
        {activity.targetSummary && (
          <Card>
            <h3 className="text-sm font-semibold text-slate-900 mb-2">Target</h3>
            <Link
              to="/targets/$id"
              params={{ id: activity.targetSummary.id }}
              className="text-sm font-medium text-teal-700 hover:underline"
            >
              {activity.targetSummary.name}
            </Link>
            <p className="text-xs text-slate-500 capitalize">{activity.targetSummary.targetType}</p>
          </Card>
        )}

        {/* Notes */}
        {str(activity.fields?.notes) && (
          <Card>
            <h3 className="text-sm font-semibold text-slate-900 mb-2">Notes</h3>
            <p className="text-sm text-slate-700 whitespace-pre-wrap">{str(activity.fields.notes)}</p>
          </Card>
        )}

        {/* Tags */}
        {Array.isArray(activity.fields?.tags) && (activity.fields.tags as string[]).length > 0 && (
          <Card>
            <h3 className="text-sm font-semibold text-slate-900 mb-2">Tags</h3>
            <div className="flex flex-wrap gap-1">
              {(activity.fields.tags as string[]).map((tag) => (
                <Badge key={tag}>{tag}</Badge>
              ))}
            </div>
          </Card>
        )}

        {/* Metadata */}
        <Card>
          <h3 className="text-sm font-semibold text-slate-900 mb-2">Details</h3>
          <dl className="space-y-2 text-sm">
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
        </Card>

        {/* Actions */}
        {(transitions.length > 0 || isSubmittable) && (
          <div className="flex flex-wrap gap-2">
            {transitions.map((nextStatus) => {
              const label = config?.activities.statuses.find((s) => s.key === nextStatus)?.label ?? nextStatus
              return (
                <Button
                  key={nextStatus}
                  variant="secondary"
                  size="sm"
                  disabled={patchStatus.isPending}
                  onClick={() => patchStatus.mutate({ id: activity.id, status: nextStatus })}
                >
                  {label}
                </Button>
              )
            })}
            {isSubmittable && (
              <Button
                size="sm"
                disabled={submitActivity.isPending}
                onClick={() => submitActivity.mutate(activity.id)}
              >
                <Send size={14} />
                Submit
              </Button>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
