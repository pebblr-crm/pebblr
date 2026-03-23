import { Link } from '@tanstack/react-router'
import type { Activity } from '@/types/activity'
import type { TenantConfig } from '@/types/config'
import {
  getActivityTitle,
  getTypeCategory,
  getDurationLabel,
  getStatusDotColor,
  CATEGORY_COLORS,
} from '@/utils/config'

interface ActivityCardProps {
  activity: Activity
  config?: TenantConfig
}

export function ActivityCard({ activity, config }: ActivityCardProps) {
  const title = getActivityTitle(config, activity)
  const category = getTypeCategory(config?.activities, activity.activityType)
  const style = CATEGORY_COLORS[category] ?? CATEGORY_COLORS.field
  const statusDot = getStatusDotColor(config?.activities, activity.status)
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
        <p className="text-[10px] font-semibold truncate">{title}</p>
      </div>
      <p className="text-[8px] opacity-70 truncate">{durationLabel}</p>
    </Link>
  )
}
