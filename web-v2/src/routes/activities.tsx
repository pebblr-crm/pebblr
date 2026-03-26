import { useState, useMemo } from 'react'
import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { useActivities, useActivity, useCreateActivity, usePatchActivityStatus, useSubmitActivity } from '@/hooks/useActivities'
import { useRecoveryBalance } from '@/hooks/useDashboard'
import { useConfig } from '@/hooks/useConfig'
import { Badge } from '@/components/ui/Badge'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { Modal } from '@/components/ui/Modal'
import { Plus, Clock, AlertTriangle, Check, Send } from 'lucide-react'
import type { Activity } from '@/types/activity'

function str(v: unknown): string {
  return typeof v === 'string' ? v : ''
}

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/activities',
  component: ActivitiesPage,
})

const statusVariant: Record<string, 'primary' | 'success' | 'danger' | 'default'> = {
  planificat: 'primary',
  realizat: 'success',
  anulat: 'danger',
}

const statusColor: Record<string, string> = {
  realizat: 'bg-emerald-500',
  planificat: 'bg-blue-500',
  anulat: 'bg-red-500',
}

function groupByDate(activities: Activity[]): Map<string, Activity[]> {
  const map = new Map<string, Activity[]>()
  for (const a of activities) {
    const list = map.get(a.dueDate) ?? []
    list.push(a)
    map.set(a.dueDate, list)
  }
  return map
}

// --- Create Activity Modal ---

const QUICK_TAGS = [
  'Left Samples',
  'Follow-up Required',
  'Decision Maker Present',
  'Trial Discussed',
]

function CreateActivityModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { data: config } = useConfig()
  const createActivity = useCreateActivity()

  const [activityType, setActivityType] = useState('')
  const [tags, setTags] = useState<string[]>([])
  const [notes, setNotes] = useState('')
  const [duration, setDuration] = useState('')
  const [scheduleNext, setScheduleNext] = useState(true)

  const fieldActivityTypes = useMemo(
    () => config?.activities.types.filter((t) => t.category === 'field') ?? [],
    [config],
  )
  const initialStatus = useMemo(
    () => config?.activities.statuses.find((s) => s.initial)?.key ?? 'planificat',
    [config],
  )
  const durations = useMemo(() => config?.activities.durations ?? [], [config])

  const toggleTag = (tag: string) => {
    setTags((prev) => (prev.includes(tag) ? prev.filter((t) => t !== tag) : [...prev, tag]))
  }

  const reset = () => {
    setActivityType('')
    setTags([])
    setNotes('')
    setDuration('')
    setScheduleNext(true)
  }

  const handleSubmit = () => {
    const today = new Date().toISOString().slice(0, 10)
    createActivity.mutate(
      {
        activityType: activityType || fieldActivityTypes[0]?.key || 'visit',
        status: initialStatus,
        dueDate: today,
        duration: duration || durations[0]?.key || 'full_day',
        fields: { tags, notes, scheduleNext },
      },
      {
        onSuccess: () => {
          reset()
          onClose()
        },
      },
    )
  }

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Log Activity"
      footer={
        <Button
          onClick={handleSubmit}
          disabled={createActivity.isPending || !activityType}
          className="w-full"
          size="lg"
        >
          {createActivity.isPending ? 'Submitting...' : 'Log Activity'}
        </Button>
      }
    >
      <div className="space-y-5">
        <p className="text-xs text-slate-500">
          New activities start as {config?.activities.statuses.find((s) => s.initial)?.label ?? 'Planned'}.
        </p>

        {/* Activity type */}
        <div>
          <label className="mb-1.5 block text-sm font-medium text-slate-700">Activity Type</label>
          <select
            value={activityType}
            onChange={(e) => setActivityType(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
          >
            <option value="">Select type...</option>
            {fieldActivityTypes.map((t) => (
              <option key={t.key} value={t.key}>{t.label}</option>
            ))}
          </select>
        </div>

        {/* Duration */}
        {durations.length > 0 && (
          <div>
            <label className="mb-1.5 block text-sm font-medium text-slate-700">Duration</label>
            <select
              value={duration}
              onChange={(e) => setDuration(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
            >
              {durations.map((d) => (
                <option key={d.key} value={d.key}>{d.label}</option>
              ))}
            </select>
          </div>
        )}

        {/* Quick tags */}
        <div>
          <label className="mb-1.5 block text-sm font-medium text-slate-700">Quick Tags</label>
          <div className="flex flex-wrap gap-2">
            {QUICK_TAGS.map((tag) => (
              <button
                key={tag}
                onClick={() => toggleTag(tag)}
                className={`rounded-full border px-3 py-1.5 text-xs font-medium transition-colors ${
                  tags.includes(tag)
                    ? 'border-teal-500 bg-teal-50 text-teal-700'
                    : 'border-slate-200 bg-white text-slate-600 hover:border-slate-300'
                }`}
              >
                {tags.includes(tag) && <Check size={12} className="mr-1 inline" />}
                {tag}
              </button>
            ))}
          </div>
        </div>

        {/* Notes */}
        <div>
          <label className="mb-1.5 block text-sm font-medium text-slate-700">
            Notes <span className="text-slate-400">(optional)</span>
          </label>
          <textarea
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            maxLength={250}
            rows={3}
            placeholder="Add notes about the visit..."
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
          />
          <p className="mt-1 text-xs text-slate-400 text-right">{notes.length}/250</p>
        </div>

        {/* Schedule next */}
        <label className="flex items-center gap-3 rounded-lg border border-slate-200 p-3 cursor-pointer">
          <input
            type="checkbox"
            checked={scheduleNext}
            onChange={(e) => setScheduleNext(e.target.checked)}
            className="h-4 w-4 rounded border-slate-300 text-teal-600 focus:ring-teal-500"
          />
          <div>
            <span className="text-sm font-medium text-slate-700">Schedule next visit?</span>
            <p className="text-xs text-slate-500">Auto-suggested based on cadence</p>
          </div>
        </label>

      </div>
    </Modal>
  )
}

// --- Activity Detail Modal ---

const transitionColors: Record<string, string> = {
  realizat: 'bg-emerald-600 text-white hover:bg-emerald-700',
  anulat: 'bg-red-600 text-white hover:bg-red-700',
}

function ActivityDetailModal({ activityId, onClose }: { activityId: string | null; onClose: () => void }) {
  const { data: activity, isLoading } = useActivity(activityId ?? '')
  const { data: config } = useConfig()
  const patchStatus = usePatchActivityStatus()
  const submitActivity = useSubmitActivity()
  const [feedback, setFeedback] = useState('')

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

  if (!activityId) return null

  const title = isLoading || !activity
    ? 'Activity'
    : (activity.targetName ?? activity.label ?? activity.activityType)

  const handleTransition = (nextStatus: string) => {
    if (!activity) return
    const fields = feedback.trim() ? { notes: feedback.trim() } : undefined
    patchStatus.mutate(
      { id: activity.id, status: nextStatus, fields },
      { onSuccess: () => { setFeedback(''); onClose() } },
    )
  }

  const canAct = transitions.length > 0 || isSubmittable
  const actionFooter = activity && canAct ? (
    <div className="space-y-3">
      <div>
        <label className="mb-1.5 block text-sm font-medium text-slate-700">
          Feedback <span className="text-slate-400">(optional)</span>
        </label>
        <textarea
          value={feedback}
          onChange={(e) => setFeedback(e.target.value)}
          maxLength={250}
          rows={2}
          placeholder="How did the visit go?"
          className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
        />
      </div>
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
            disabled={submitActivity.isPending}
            onClick={() => submitActivity.mutate(activity.id)}
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
            </div>
            <p className="mt-1 text-sm text-slate-500 capitalize">
              {activity.activityType} &middot; {activity.duration} &middot; {new Date(activity.dueDate).toLocaleDateString('en-GB', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}
            </p>
          </div>

          {/* Target */}
          {activity.targetSummary && (
            <div className="rounded-lg border border-slate-200 p-3">
              <p className="text-xs text-slate-500">Target</p>
              <p className="text-sm font-medium text-slate-900">{activity.targetSummary.name}</p>
              <p className="text-xs text-slate-500 capitalize">{activity.targetSummary.targetType}</p>
            </div>
          )}

          {/* Notes */}
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

// --- Main Page ---

function ActivitiesPage() {
  const [typeFilter, setTypeFilter] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [createOpen, setCreateOpen] = useState(false)
  const [detailId, setDetailId] = useState<string | null>(null)

  const { data, isLoading } = useActivities({
    activityType: typeFilter || undefined,
    status: statusFilter || undefined,
    limit: 100,
  })
  const { data: recovery } = useRecoveryBalance({})

  const activities = useMemo(() => data?.items ?? [], [data])
  const grouped = useMemo(() => groupByDate(activities), [activities])
  const sortedDates = useMemo(
    () => [...grouped.keys()].sort((a, b) => b.localeCompare(a)),
    [grouped],
  )

  if (isLoading) return <Spinner />

  return (
    <div className="p-4 md:p-6">
      {/* Header */}
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold text-slate-900">Activity Log</h1>
        <Button size="sm" onClick={() => setCreateOpen(true)}>
          <Plus size={14} />
          Log Activity
        </Button>
      </div>

      {/* Recovery + nudge cards */}
      <div className="mb-6 grid grid-cols-1 gap-3 sm:grid-cols-3 sm:gap-4">
        {recovery && (
          <Card>
            <div className="flex items-center gap-2">
              <Clock size={16} className="text-teal-600" />
              <span className="text-sm font-medium text-slate-700">Recovery Balance</span>
            </div>
            <p className="mt-2 text-2xl font-semibold text-slate-900">{recovery.balance} days</p>
            <p className="text-xs text-slate-500">{recovery.earned} earned, {recovery.taken} taken</p>
          </Card>
        )}
        <Card className="sm:col-span-2 flex items-center gap-3 bg-amber-50 border-amber-200">
          <AlertTriangle size={18} className="text-amber-600 shrink-0" />
          <span className="text-sm text-amber-800">
            Review your submitted activities. Overdue targets need visits scheduled.
          </span>
        </Card>
      </div>

      {/* Filters */}
      <div className="mb-4 flex flex-wrap items-center gap-3">
        <select
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value)}
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm"
        >
          <option value="">All types</option>
          <option value="visit">Visit</option>
          <option value="administrative">Administrative</option>
        </select>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm"
        >
          <option value="">All statuses</option>
          <option value="planificat">Planned</option>
          <option value="realizat">Completed</option>
          <option value="anulat">Cancelled</option>
        </select>
        <span className="text-sm text-slate-500">{activities.length} activities</span>
      </div>

      {/* Timeline */}
      {sortedDates.length === 0 ? (
        <p className="py-12 text-center text-sm text-slate-400">No activities found.</p>
      ) : (
        <div className="space-y-6">
          {sortedDates.map((date) => {
            const dayActivities = grouped.get(date) ?? []
            const d = new Date(date)
            return (
              <div key={date}>
                <h3 className="mb-3 text-sm font-semibold text-slate-500">
                  {d.toLocaleDateString('en-GB', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}
                </h3>
                <div className="relative pl-6 space-y-3">
                  <div className="absolute left-2 top-2 bottom-2 w-px bg-slate-200" />
                  {dayActivities.map((activity) => (
                    <button
                      key={activity.id}
                      type="button"
                      onClick={() => setDetailId(activity.id)}
                      className="relative block w-full rounded-lg border border-slate-200 bg-white p-4 text-left transition-colors hover:border-slate-300 hover:bg-slate-50"
                    >
                      <div className={`absolute -left-4 top-5 h-3 w-3 rounded-full border-2 border-white ${statusColor[activity.status] ?? 'bg-slate-400'}`} />
                      <div className="flex items-start justify-between">
                        <div>
                          <div className="flex items-center gap-2">
                            <span className="font-medium text-slate-900">
                              {activity.targetName ?? activity.label ?? activity.activityType}
                            </span>
                            <Badge variant={statusVariant[activity.status] ?? 'default'}>
                              {activity.status}
                            </Badge>
                          </div>
                          <p className="mt-1 text-xs text-slate-500 capitalize">{activity.activityType} &middot; {activity.duration}</p>
                        </div>
                        {activity.submittedAt && (
                          <Badge variant="success">Submitted</Badge>
                        )}
                      </div>
                      {str(activity.fields?.notes) && (
                        <p className="mt-2 text-sm text-slate-600">{str(activity.fields.notes)}</p>
                      )}
                      {Array.isArray(activity.fields?.tags) && (
                        <div className="mt-2 flex flex-wrap gap-1">
                          {(activity.fields.tags as string[]).map((tag) => (
                            <Badge key={tag}>{tag}</Badge>
                          ))}
                        </div>
                      )}
                    </button>
                  ))}
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* Modals */}
      <CreateActivityModal open={createOpen} onClose={() => setCreateOpen(false)} />
      <ActivityDetailModal activityId={detailId} onClose={() => setDetailId(null)} />
    </div>
  )
}
