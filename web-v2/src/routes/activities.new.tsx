import { useState, useMemo } from 'react'
import { createRoute, useNavigate } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { useConfig } from '@/hooks/useConfig'
import { useCreateActivity } from '@/hooks/useActivities'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Spinner } from '@/components/ui/Spinner'
import { ArrowLeft, ArrowRight, Check } from 'lucide-react'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/activities/new',
  component: NewActivityPage,
})

const OUTCOMES = [
  { key: 'completed', label: 'Completed', icon: '✓' },
  { key: 'rescheduled', label: 'Rescheduled', icon: '↻' },
  { key: 'no_show', label: 'No Show', icon: '✗' },
  { key: 'cancelled', label: 'Cancelled', icon: '—' },
] as const

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

  const [step, setStep] = useState<1 | 2>(1)
  const [activityType, setActivityType] = useState('')
  const [outcome, setOutcome] = useState('')
  const [tags, setTags] = useState<string[]>([])
  const [notes, setNotes] = useState('')
  const [scheduleNext, setScheduleNext] = useState(true)

  const fieldActivityTypes = useMemo(
    () => config?.activities.types.filter((t) => t.category === 'field') ?? [],
    [config],
  )

  const toggleTag = (tag: string) => {
    setTags((prev) => (prev.includes(tag) ? prev.filter((t) => t !== tag) : [...prev, tag]))
  }

  const handleSubmit = () => {
    const today = new Date().toISOString().slice(0, 10)
    createActivity.mutate(
      {
        activityType: activityType || fieldActivityTypes[0]?.key || 'visit',
        status: outcome === 'completed' ? 'realizat' : 'planificat',
        dueDate: today,
        duration: 'full_day',
        fields: { tags, notes, outcome },
      },
      {
        onSuccess: () => navigate({ to: '/activities' }),
      },
    )
  }

  if (isLoading) return <Spinner />

  return (
    <div className="mx-auto max-w-lg p-6">
      {/* Progress */}
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-lg font-semibold text-slate-900">Log Activity</h1>
        <span className="text-sm text-slate-500">Step {step} of 2</span>
      </div>

      <div className="mb-6 flex gap-1">
        <div className={`h-1 flex-1 rounded-full ${step >= 1 ? 'bg-teal-600' : 'bg-slate-200'}`} />
        <div className={`h-1 flex-1 rounded-full ${step >= 2 ? 'bg-teal-600' : 'bg-slate-200'}`} />
      </div>

      {step === 1 && (
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

          {/* Outcome */}
          <div>
            <label className="mb-2 block text-sm font-medium text-slate-700">Outcome</label>
            <div className="grid grid-cols-2 gap-3">
              {OUTCOMES.map((o) => (
                <button
                  key={o.key}
                  onClick={() => setOutcome(o.key)}
                  className={`flex items-center gap-2 rounded-xl border-2 p-4 text-left transition-colors ${
                    outcome === o.key
                      ? 'border-teal-500 bg-teal-50'
                      : 'border-slate-200 hover:border-slate-300'
                  }`}
                >
                  <span className="text-lg">{o.icon}</span>
                  <span className="text-sm font-medium">{o.label}</span>
                </button>
              ))}
            </div>
          </div>

          <Button
            onClick={() => setStep(2)}
            disabled={!outcome}
            className="w-full"
          >
            Continue
            <ArrowRight size={16} />
          </Button>
        </div>
      )}

      {step === 2 && (
        <div className="space-y-6">
          {/* Context */}
          <Card>
            <div className="flex items-center gap-3">
              <Badge variant={outcome === 'completed' ? 'success' : 'warning'}>
                {outcome}
              </Badge>
              {tags.map((tag) => (
                <Badge key={tag} variant="primary">{tag}</Badge>
              ))}
            </div>
          </Card>

          {/* Notes */}
          <div>
            <label className="mb-2 block text-sm font-medium text-slate-700">
              Visit Notes <span className="text-slate-400">(optional)</span>
            </label>
            <textarea
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              maxLength={250}
              rows={4}
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

          <div className="flex gap-3">
            <Button variant="secondary" onClick={() => setStep(1)} className="flex-1">
              <ArrowLeft size={16} />
              Back
            </Button>
            <Button
              onClick={handleSubmit}
              disabled={createActivity.isPending}
              className="flex-1"
            >
              {createActivity.isPending ? 'Submitting...' : 'Submit Activity'}
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
