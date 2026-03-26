import { useState, useMemo, useCallback } from 'react'
import { createRoute, useParams, useNavigate } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { useTarget, useTargetVisitStatus } from '@/hooks/useTargets'
import { useUsers } from '@/hooks/useUsers'
import { useActivities, useCreateActivity } from '@/hooks/useActivities'
import { useConfig } from '@/hooks/useConfig'
import { Badge } from '@/components/ui/Badge'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { Modal } from '@/components/ui/Modal'
import { MapContainer } from '@/components/map/MapContainer'
import { TargetMarker } from '@/components/map/TargetMarker'
import { ActivityDetailModal } from '@/components/activities/ActivityDetailModal'
import { priorityLabel, priorityStyle, priorityDot, statusDot } from '@/lib/styles'
import { str, daysAgo } from '@/lib/helpers'
import {
  ArrowLeft, MapPin, ChevronRight,
  Building2, Stethoscope, AlertCircle, CalendarPlus,
} from 'lucide-react'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/targets/$id',
  component: TargetDetailPage,
})

function TargetDetailPage() {
  const { id } = useParams({ from: '/targets/$id' })
  const navigate = useNavigate()
  const { data: target, isLoading } = useTarget(id)
  const { data: activityData } = useActivities({ targetId: id, limit: 50 })
  const { data: visitStatusData } = useTargetVisitStatus()
  const { data: config } = useConfig()

  const [detailActivityId, setDetailActivityId] = useState<string | null>(null)
  const [scheduleOpen, setScheduleOpen] = useState(false)

  const { data: usersData } = useUsers()
  const activities = useMemo(() => activityData?.items ?? [], [activityData])
  const userMap = useMemo(() => {
    const m = new Map<string, string>()
    usersData?.items?.forEach((u) => m.set(u.id, u.name || u.displayName))
    return m
  }, [usersData])

  const potential = ((target?.fields?.potential as string) ?? 'c').toLowerCase()
  const lat = typeof target?.fields?.lat === 'number' ? target.fields.lat : null
  const lng = typeof target?.fields?.lng === 'number' ? target.fields.lng : null

  const accountType = useMemo(
    () => config?.accounts.types.find((t) => t.key === target?.targetType),
    [config, target],
  )

  // Last visit date
  const lastVisitDate = useMemo(() => {
    const vs = visitStatusData?.items?.find((v) => v.targetId === id)
    return vs?.lastVisitDate ?? null
  }, [visitStatusData, id])

  // Next scheduled visit (future planned activities)
  const nextVisit = useMemo(() => {
    const now = new Date().toISOString().slice(0, 10)
    return activities
      .filter((a) => a.dueDate.slice(0, 10) >= now && a.status !== 'anulat' && a.status !== 'cancelled')
      .sort((a, b) => a.dueDate.localeCompare(b.dueDate))[0] ?? null
  }, [activities])

  const isScheduledToday = nextVisit?.dueDate.slice(0, 10) === new Date().toISOString().slice(0, 10)

  // Target icon
  const TargetIcon = target?.targetType === 'doctor' ? Stethoscope : Building2

  if (isLoading || !target) return <Spinner />

  return (
    <div className="flex h-full flex-col bg-slate-50">
      {/* Breadcrumbs & actions bar */}
      <div className="px-4 py-3 bg-white border-b border-slate-200 flex flex-col sm:flex-row sm:items-center justify-between gap-3 sticky top-0 z-30 shadow-sm md:px-8">
        <div className="flex items-center gap-2 text-sm">
          <button onClick={() => navigate({ to: '/targets' })} className="text-slate-400 hover:text-slate-600">
            <ArrowLeft size={18} />
          </button>
          <button onClick={() => navigate({ to: '/targets' })} className="text-slate-500 hover:text-slate-700">Targets</button>
          <ChevronRight size={12} className="text-slate-400" />
          <span className="font-medium text-slate-800">{target.name}</span>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="primary" size="sm" onClick={() => setScheduleOpen(true)}>
            <CalendarPlus size={14} />
            Schedule Visit
          </Button>
        </div>
      </div>

      <div className="flex-1 overflow-auto p-4 md:p-8">
        <div className="max-w-7xl mx-auto space-y-6">

          {/* Target profile header card */}
          <div className="bg-white rounded-xl shadow-sm border border-slate-200 p-6 flex flex-col md:flex-row gap-6 items-start">
            <div className="w-14 h-14 rounded-lg bg-slate-100 border border-slate-200 flex items-center justify-center text-slate-400 shrink-0">
              <TargetIcon size={28} />
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex flex-wrap items-center gap-2 mb-2">
                <h1 className="text-2xl font-bold text-slate-900">{target.name}</h1>
                <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-semibold border ${priorityStyle[potential] ?? priorityStyle.c}`}>
                  <span className={`w-1.5 h-1.5 rounded-full ${priorityDot[potential] ?? priorityDot.c}`} />
                  {priorityLabel[potential] ?? 'Priority C'}
                </span>
                <Badge variant="primary">{target.targetType}</Badge>
                {isScheduledToday && (
                  <span className="inline-flex items-center px-2.5 py-1 rounded-md text-xs font-medium bg-teal-50 text-teal-700 border border-teal-100">
                    Scheduled Today
                  </span>
                )}
              </div>

              {str(target.fields.address) && (
                <p className="text-sm text-slate-600 mb-4 flex items-center gap-2">
                  <MapPin size={14} className="text-slate-400 shrink-0" />
                  {str(target.fields.address)}
                  {str(target.fields.city) && `, ${str(target.fields.city)}`}
                </p>
              )}

              {/* Quick stats row */}
              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 border-t border-slate-100 pt-4">
                <div>
                  <div className="text-xs text-slate-500 mb-1">Next Visit</div>
                  {nextVisit ? (
                    <>
                      <div className="text-sm font-medium text-slate-800">
                        {new Date(nextVisit.dueDate).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })}
                      </div>
                      <div className="text-xs text-slate-500 capitalize">{nextVisit.activityType}</div>
                    </>
                  ) : (
                    <div className="flex items-center gap-1.5">
                      <AlertCircle size={14} className="text-red-500" />
                      <span className="text-sm font-semibold text-red-600">No visit scheduled</span>
                    </div>
                  )}
                </div>
                <div>
                  <div className="text-xs text-slate-500 mb-1">Last Visit</div>
                  {lastVisitDate ? (
                    <>
                      <div className="text-sm font-medium text-slate-800">
                        {new Date(lastVisitDate).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })}
                      </div>
                      <div className="text-xs text-slate-500">{daysAgo(lastVisitDate)} days ago</div>
                    </>
                  ) : (
                    <div className="text-sm text-slate-400">Never visited</div>
                  )}
                </div>
                <div>
                  <div className="text-xs text-slate-500 mb-1">Total Visits</div>
                  <div className="text-sm font-medium text-slate-800">{activities.length}</div>
                </div>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Left column: Details + History */}
            <div className="lg:col-span-2 space-y-6">

              {/* Target details */}
              <Card className="!p-0 overflow-hidden">
                <div className="px-6 py-4 border-b border-slate-200 bg-slate-50">
                  <h2 className="text-base font-semibold text-slate-800">Target Details</h2>
                </div>
                <div className="p-6">
                  <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-3">
                    {accountType?.fields.map((field) => {
                      const value = target.fields[field.key]
                      if (value == null || value === '') return null
                      return (
                        <div key={field.key}>
                          <dt className="text-xs text-slate-500">{field.label ?? field.key}</dt>
                          <dd className="text-sm font-medium text-slate-900 capitalize">{String(value)}</dd>
                        </div>
                      )
                    })}
                  </dl>
                </div>
              </Card>

              {/* Visit history */}
              <Card className="!p-0 overflow-hidden flex flex-col max-h-[600px]">
                <div className="px-6 py-4 border-b border-slate-200 bg-slate-50 flex items-center justify-between shrink-0">
                  <h2 className="text-base font-semibold text-slate-800">Visit History</h2>
                  <span className="text-xs font-medium text-slate-500">{activities.length} total</span>
                </div>
                <div className="p-6 overflow-auto flex-1">
                  {activities.length === 0 ? (
                    <p className="text-sm text-slate-400 text-center py-6">No visits recorded yet.</p>
                  ) : (
                    <div className="relative pl-6 border-l-2 border-slate-200 space-y-5">
                      {activities.map((activity) => {
                        const dot = statusDot[activity.status] ?? 'bg-slate-400'
                        return (
                          <div key={activity.id} className="relative">
                            <div className={`absolute -left-[29px] top-1 w-4 h-4 rounded-full ${dot} border-4 border-white shadow-sm`} />
                            <button
                              onClick={() => setDetailActivityId(activity.id)}
                              className="w-full text-left bg-slate-50 rounded-lg p-4 border border-slate-200 hover:border-slate-300 hover:shadow-sm transition-all"
                            >
                              <div className="flex items-center justify-between mb-1">
                                <div className="flex items-center gap-2">
                                  <span className="text-sm font-semibold text-slate-800">
                                    {new Date(activity.dueDate).toLocaleDateString('en-GB', { weekday: 'short', day: 'numeric', month: 'short', year: 'numeric' })}
                                  </span>
                                  {str(activity.fields?.visit_type) && (
                                    <span className={`text-[10px] font-medium px-1.5 py-0.5 rounded ${
                                      str(activity.fields.visit_type) === 'f2f'
                                        ? 'bg-amber-50 text-amber-700 border border-amber-200'
                                        : 'bg-blue-50 text-blue-700 border border-blue-200'
                                    }`}>
                                      {str(activity.fields.visit_type) === 'f2f' ? 'In person' : 'Remote'}
                                    </span>
                                  )}
                                  <Badge variant={activity.status === 'realizat' || activity.status === 'completed' ? 'success' : activity.status === 'anulat' || activity.status === 'cancelled' ? 'danger' : 'primary'}>
                                    {activity.status}
                                  </Badge>
                                </div>
                                <span className="text-xs text-slate-500">
                                  {userMap.get(activity.creatorId) ?? 'Unknown'}
                                </span>
                              </div>
                              {str(activity.fields?.feedback) && (
                                <p className="text-sm text-slate-600 mb-2">{str(activity.fields.feedback)}</p>
                              )}
                              {str(activity.fields?.notes) && !str(activity.fields?.feedback) && (
                                <p className="text-sm text-slate-600 mb-2">{str(activity.fields.notes)}</p>
                              )}
                              {Array.isArray(activity.fields?.tags) && (activity.fields.tags as string[]).length > 0 && (
                                <div className="flex flex-wrap gap-1">
                                  {(activity.fields.tags as string[]).map((tag) => (
                                    <span key={tag} className="px-2 py-0.5 bg-white border border-slate-200 rounded text-[10px] font-medium text-slate-600">{tag}</span>
                                  ))}
                                </div>
                              )}
                            </button>
                          </div>
                        )
                      })}
                    </div>
                  )}
                </div>
              </Card>
            </div>

            {/* Right column: Map */}
            <div className="space-y-6">
              <Card className="!p-0 overflow-hidden h-[400px] flex flex-col">
                <div className="px-4 py-3 border-b border-slate-200 bg-slate-50 shrink-0">
                  <h2 className="text-sm font-semibold text-slate-800">Location</h2>
                </div>
                <div className="flex-1">
                  {lat != null && lng != null ? (
                    <MapContainer className="h-full" center={[lng, lat]} zoom={14}>
                      {(map) => (
                        <TargetMarker map={map} lat={lat} lng={lng} name={target.name} priority={potential} />
                      )}
                    </MapContainer>
                  ) : (
                    <div className="flex h-full items-center justify-center bg-slate-100 text-sm text-slate-400">
                      No coordinates available
                    </div>
                  )}
                </div>
              </Card>
            </div>
          </div>
        </div>
      </div>

      {/* Activity detail modal */}
      <ActivityDetailModal activityId={detailActivityId} onClose={() => setDetailActivityId(null)} />

      {/* Schedule visit modal */}
      <ScheduleVisitModal
        open={scheduleOpen}
        onClose={() => setScheduleOpen(false)}
        targetId={id}
        targetName={target.name}
      />
    </div>
  )
}

/* ── Schedule Visit Modal ── */

function ScheduleVisitModal({ open, onClose, targetId, targetName }: {
  open: boolean
  onClose: () => void
  targetId: string
  targetName: string
}) {
  const { data: config } = useConfig()
  const createActivity = useCreateActivity()
  const [dueDate, setDueDate] = useState(() => new Date().toISOString().slice(0, 10))
  const [visitType, setVisitType] = useState('f2f')

  const visitTypes = useMemo(() => {
    const vt = config?.options?.visit_types
    return vt ?? [{ key: 'f2f', label: 'Face to face' }, { key: 'remote', label: 'Remote' }]
  }, [config])

  const initialStatus = useMemo(
    () => config?.activities.statuses.find((s) => s.initial)?.key ?? 'planificat',
    [config],
  )

  const handleSubmit = useCallback(() => {
    createActivity.mutate(
      {
        activityType: 'visit',
        status: initialStatus,
        dueDate,
        fields: { visit_type: visitType },
        targetId,
      },
      {
        onSuccess: () => {
          setDueDate(new Date().toISOString().slice(0, 10))
          setVisitType('f2f')
          onClose()
        },
      },
    )
  }, [createActivity, dueDate, visitType, targetId, initialStatus, onClose])

  return (
    <Modal
      open={open}
      onClose={onClose}
      title={`Schedule Visit — ${targetName}`}
      footer={
        <div className="flex items-center justify-end gap-2">
          <Button variant="secondary" size="sm" onClick={onClose}>Cancel</Button>
          <Button variant="primary" size="sm" onClick={handleSubmit} disabled={createActivity.isPending}>
            <CalendarPlus size={14} />
            {createActivity.isPending ? 'Creating...' : 'Schedule'}
          </Button>
        </div>
      }
    >
      <div className="space-y-4">
        <div>
          <label htmlFor="field-schedule-date" className="mb-1.5 block text-sm font-medium text-slate-700">Date</label>
          <input
            id="field-schedule-date"
            type="date"
            value={dueDate}
            onChange={(e) => setDueDate(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
          />
        </div>
        <div>
          <label id="field-visit-type" className="mb-1.5 block text-sm font-medium text-slate-700">Visit Type</label>
          <div role="group" aria-labelledby="field-visit-type" className="flex rounded-lg border border-slate-200 overflow-hidden">
            {visitTypes.map((vt) => (
              <button
                key={vt.key}
                type="button"
                onClick={() => setVisitType(vt.key)}
                className={`flex-1 px-3 py-2 text-sm font-medium transition-colors ${
                  visitType === vt.key
                    ? 'bg-teal-600 text-white'
                    : 'bg-white text-slate-600 hover:bg-slate-50'
                } ${vt.key !== visitTypes[0].key ? 'border-l border-slate-200' : ''}`}
              >
                {vt.label}
              </button>
            ))}
          </div>
        </div>
      </div>
    </Modal>
  )
}
