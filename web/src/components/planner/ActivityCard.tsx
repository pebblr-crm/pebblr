import { Link } from '@tanstack/react-router'
import type { Activity } from '@/types/activity'
import type { TenantConfig } from '@/types/config'

interface ActivityCardProps {
  activity: Activity
  config?: TenantConfig
}

const categoryColors: Record<string, string> = {
  field: 'bg-amber-100 border-amber-500 text-amber-900',
  non_field: 'bg-blue-100 border-blue-400 text-blue-900',
}

const statusDots: Record<string, string> = {
  planificat: 'bg-amber-500',
  realizat: 'bg-emerald-500',
  anulat: 'bg-red-400',
}

export function ActivityCard({ activity, config }: ActivityCardProps) {
  const typeConfig = config?.activities.types.find((t) => t.key === activity.activityType)
  const typeLabel = typeConfig?.label ?? activity.activityType
  const category = typeConfig?.category ?? 'field'
  const style = categoryColors[category] ?? categoryColors.field
  const statusDot = statusDots[activity.status] ?? 'bg-slate-400'
  const durationLabel = config?.activities.durations.find((d) => d.key === activity.duration)?.label ?? activity.duration

  return (
    <Link
      to="/activities/$activityId"
      params={{ activityId: activity.id }}
      className={`block p-1.5 rounded-lg border-l-4 ${style} no-underline hover:opacity-80 transition-opacity`}
      data-testid="activity-card"
    >
      <div className="flex items-center gap-1">
        <span className={`w-1.5 h-1.5 rounded-full shrink-0 ${statusDot}`} />
        <p className="text-[10px] font-semibold truncate">{typeLabel}</p>
      </div>
      <p className="text-[8px] opacity-70 truncate">{durationLabel}</p>
    </Link>
  )
}
