import { Link } from '@tanstack/react-router'
import type { Activity } from '@/types/activity'
import type { TenantConfig } from '@/types/config'
import {
  getTypeLabel,
  getTypeCategory,
  getDurationLabel,
  CATEGORY_COLORS,
} from '@/utils/config'

interface ActivityCardProps {
  activity: Activity
  config?: TenantConfig
}

const statusDots: Record<string, string> = {
  planificat: 'bg-amber-500',
  realizat: 'bg-emerald-500',
  anulat: 'bg-red-400',
}

export function ActivityCard({ activity, config }: ActivityCardProps) {
  const typeLabel = getTypeLabel(config?.activities, activity.activityType)
  const category = getTypeCategory(config?.activities, activity.activityType)
  const style = CATEGORY_COLORS[category] ?? CATEGORY_COLORS.field
  const statusDot = statusDots[activity.status] ?? 'bg-slate-400'
  const durationLabel = getDurationLabel(config?.activities, activity.duration)

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
