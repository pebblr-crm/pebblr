import { useMemo } from 'react'
import { createRoute, useParams } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { useTarget } from '@/hooks/useTargets'
import { useActivities } from '@/hooks/useActivities'
import { useConfig } from '@/hooks/useConfig'
import { Badge } from '@/components/ui/Badge'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { MapContainer } from '@/components/map/MapContainer'
import { TargetMarker } from '@/components/map/TargetMarker'
import { ArrowLeft, MapPin, Phone, Calendar, Clock } from 'lucide-react'

function str(v: unknown): string {
  return typeof v === 'string' ? v : ''
}

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/targets/$id',
  component: TargetDetailPage,
})

const priorityVariant: Record<string, 'danger' | 'warning' | 'default'> = {
  A: 'danger',
  B: 'warning',
  C: 'default',
}

const statusColor: Record<string, string> = {
  realizat: 'bg-emerald-500',
  planificat: 'bg-blue-500',
  anulat: 'bg-red-500',
}

function TargetDetailPage() {
  const { id } = useParams({ from: '/targets/$id' })
  const { data: target, isLoading } = useTarget(id)
  const { data: activityData } = useActivities({ targetId: id, limit: 50 })
  const { data: config } = useConfig()

  const activities = useMemo(() => activityData?.items ?? [], [activityData])
  const classification = (target?.fields?.classification as string) ?? 'C'
  const lat = typeof target?.fields?.latitude === 'number' ? target.fields.latitude : null
  const lng = typeof target?.fields?.longitude === 'number' ? target.fields.longitude : null

  const accountType = useMemo(
    () => config?.accounts.types.find((t) => t.key === target?.targetType),
    [config, target],
  )

  if (isLoading || !target) return <Spinner />

  return (
    <div className="flex h-full flex-col">
      {/* Header */}
      <div className="border-b border-slate-200 bg-white px-6 py-4">
        <div className="flex items-center gap-3 mb-3">
          <a href="/targets" className="text-slate-400 hover:text-slate-600">
            <ArrowLeft size={20} />
          </a>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-bold text-slate-900">{target.name}</h1>
              <Badge variant={priorityVariant[classification] ?? 'default'}>{classification}</Badge>
              <Badge variant="primary">{target.targetType}</Badge>
            </div>
            {str(target.fields.address) && (
              <p className="mt-1 flex items-center gap-1 text-sm text-slate-500">
                <MapPin size={14} />
                {str(target.fields.address)}
                {str(target.fields.city) && `, ${str(target.fields.city)}`}
              </p>
            )}
          </div>
        </div>

        <div className="flex gap-2">
          <Button variant="primary" size="sm">
            <Calendar size={14} />
            Schedule Visit
          </Button>
          <Button variant="secondary" size="sm">
            <Phone size={14} />
            Start Navigation
          </Button>
        </div>
      </div>

      {/* Content: 3 columns */}
      <div className="flex flex-1 min-h-0">
        {/* Left: Target fields */}
        <div className="w-1/3 overflow-auto border-r border-slate-200 p-6 space-y-4">
          <Card>
            <h3 className="text-sm font-semibold text-slate-900 mb-3">Target Details</h3>
            <dl className="space-y-2">
              {accountType?.fields.map((field) => {
                const value = target.fields[field.key]
                if (value == null || value === '') return null
                return (
                  <div key={field.key}>
                    <dt className="text-xs text-slate-500">{field.label ?? field.key}</dt>
                    <dd className="text-sm text-slate-900">{String(value)}</dd>
                  </div>
                )
              })}
            </dl>
          </Card>

          {str(target.fields.key_contact) && (
            <Card>
              <h3 className="text-sm font-semibold text-slate-900 mb-2">Key Contact</h3>
              <p className="text-sm text-slate-700">{str(target.fields.key_contact)}</p>
            </Card>
          )}
        </div>

        {/* Center: Activity timeline */}
        <div className="w-1/3 overflow-auto border-r border-slate-200 p-6">
          <h3 className="text-sm font-semibold text-slate-900 mb-4">
            Visit History ({activities.length})
          </h3>
          {activities.length === 0 ? (
            <p className="text-sm text-slate-400">No visits recorded yet.</p>
          ) : (
            <div className="relative pl-6">
              <div className="absolute left-2 top-2 bottom-2 w-px bg-slate-200" />
              {activities.map((activity) => (
                <div key={activity.id} className="relative mb-4 pb-4 border-b border-slate-100 last:border-0">
                  <div className={`absolute -left-4 top-1.5 h-3 w-3 rounded-full border-2 border-white ${statusColor[activity.status] ?? 'bg-slate-400'}`} />
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-xs font-medium text-slate-500">
                      {new Date(activity.dueDate).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })}
                    </span>
                    <Badge variant={activity.status === 'realizat' ? 'success' : activity.status === 'anulat' ? 'danger' : 'primary'}>
                      {activity.status}
                    </Badge>
                  </div>
                  <p className="text-sm text-slate-700">{activity.activityType}</p>
                  {str(activity.fields?.notes) && (
                    <p className="mt-1 text-xs text-slate-500">{str(activity.fields.notes)}</p>
                  )}
                  {Array.isArray(activity.fields?.tags) && (
                    <div className="mt-1 flex flex-wrap gap-1">
                      {(activity.fields.tags as string[]).map((tag) => (
                        <Badge key={tag} variant="default">{tag}</Badge>
                      ))}
                    </div>
                  )}
                  <div className="mt-1 flex items-center gap-1 text-xs text-slate-400">
                    <Clock size={12} />
                    {activity.duration}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Right: Map */}
        <div className="w-1/3">
          {lat != null && lng != null ? (
            <MapContainer
              className="h-full"
              center={[lng as number, lat as number]}
              zoom={14}
            >
              {(map) => (
                <TargetMarker
                  map={map}
                  lat={lat as number}
                  lng={lng as number}
                  name={target.name}
                  priority={classification}
                />
              )}
            </MapContainer>
          ) : (
            <div className="flex h-full items-center justify-center bg-slate-100 text-sm text-slate-400">
              No coordinates available
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
