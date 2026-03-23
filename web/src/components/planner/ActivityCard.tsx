import { Link } from '@tanstack/react-router'
import { Users } from 'lucide-react'
import type { Activity } from '@/types/activity'
import type { TenantConfig } from '@/types/config'
import {
  getActivityTitle,
  getTypeCategory,
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
  const jointVisitUserId = activity.jointVisitUserId
    ?? (activity.fields?.joint_visit_user_id as string | undefined)
  const hasJointVisit = Boolean(jointVisitUserId)

  return (
    <Link
      to="/activities/$activityId"
      params={{ activityId: activity.id }}
      className={`block px-1.5 py-0.5 rounded border-l-2 ${style} no-underline hover:opacity-80 transition-opacity`}
      data-testid="activity-card"
    >
      <div className="flex items-center gap-1">
        <span className={`w-1.5 h-1.5 rounded-full shrink-0 ${statusDot}`} />
        <p className="text-[9px] font-semibold truncate">{title}</p>
        {hasJointVisit && (
          <Users className="w-2.5 h-2.5 shrink-0 opacity-60" data-testid="joint-visit-icon" />
        )}
      </div>
    </Link>
  )
}
