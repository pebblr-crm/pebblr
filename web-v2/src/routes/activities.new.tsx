import { useState, useMemo } from 'react'
import { createRoute, useNavigate } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { useConfig } from '@/hooks/useConfig'
import { useCreateActivity } from '@/hooks/useActivities'
import { Button } from '@/components/ui/Button'
import { Badge } from '@/components/ui/Badge'
import { Spinner } from '@/components/ui/Spinner'
import { Check } from 'lucide-react'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/activities/new',
  component: NewActivityPage,
})

const QUICK_TAGS = [
  'Left Samples',
  'Follow-up Required',
  'Decision Maker Present',
  'Trial Discussed',
]

function NewActivityPage() {
  const navigate = useNavigate()
  const { data: config, isLoading } = useConfig()
  const createActivity = useCreateActivity()

  const [activityType, setActivityType] = useState('')
  const [tags, setTags] = useState<string[]>([])
  const [notes, setNotes] = useState('')
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

  const [duration, setDuration] = useState('')

  const toggleTag = (tag: string) => {
    setTags((prev) => (prev.includes(tag) ? prev.filter((t) => t !== tag) : [...prev, tag]))
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
        onSuccess: () => navigate({ to: '/activities' }),
      },
    )
  }

  if (isLoading) return <Spinner />

  return (
    <div className="mx-auto max-w-lg p-4 md:p-6">
      <div className="mb-6">
        <h1 className="text-lg font-semibold text-slate-900">Log Activity</h1>
        <p className="text-sm text-slate-500">New activities start as {config?.activities.statuses.find((s) => s.initial)?.label ?? 'Planned'}.</p>
      </div>

      <div className="space-y-6">
        {/* Activity type */}
        <div>
          <label className="mb-2 block text-sm font-medium text-slate-700">Activity Type</label>
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
            <label className="mb-2 block text-sm font-medium text-slate-700">Duration</label>
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
          <label className="mb-2 block text-sm font-medium text-slate-700">Quick Tags</label>
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
          <label className="mb-2 block text-sm font-medium text-slate-700">
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
        <label className="flex items-center gap-3 rounded-lg border border-slate-200 p-4 cursor-pointer">
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

        {/* Summary */}
        {(activityType || tags.length > 0) && (
          <div className="flex flex-wrap items-center gap-2 rounded-lg bg-slate-50 px-3 py-2">
            {activityType && (
              <Badge variant="primary">{fieldActivityTypes.find((t) => t.key === activityType)?.label ?? activityType}</Badge>
            )}
            {tags.map((tag) => (
              <Badge key={tag}>{tag}</Badge>
            ))}
          </div>
        )}

        <Button
          onClick={handleSubmit}
          disabled={createActivity.isPending || !activityType}
          className="w-full"
        >
          {createActivity.isPending ? 'Submitting...' : 'Log Activity'}
        </Button>
      </div>
    </div>
  )
}
