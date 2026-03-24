import { useTranslation } from 'react-i18next'
import type { ActivityStatsResponse } from '@/types/dashboard'
import type { ActivitiesConfig } from '@/types/config'
import { getStatusLabel } from '@/utils/config'

interface ActivityStatsCardProps {
  data: ActivityStatsResponse
  activitiesConfig?: ActivitiesConfig
}

export function ActivityStatsCard({ data, activitiesConfig }: ActivityStatsCardProps) {
  const { t } = useTranslation()
  const statusEntries = Object.entries(data.byStatus).sort(([, a], [, b]) => b - a)
  const fieldCount = data.byCategory['field'] ?? 0
  const nonFieldCount = data.byCategory['non_field'] ?? 0
  const fieldPct = data.total > 0 ? Math.round((fieldCount / data.total) * 100) : 0

  return (
    <div className="bg-surface-container-lowest p-6 rounded-xl shadow-sm border border-slate-50">
      <p className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider mb-2">
        {t('dashboardCards.activities')}
      </p>
      <div className="flex items-baseline space-x-2 mb-4">
        <h2 className="text-3xl font-extrabold text-primary font-headline">{data.total}</h2>
        <span className="text-sm text-on-surface-variant">{t('dashboardCards.total')}</span>
      </div>

      {/* By status */}
      <div className="space-y-2 mb-4">
        {statusEntries.map(([status, count]) => {
          const pct = data.total > 0 ? Math.round((count / data.total) * 100) : 0
          return (
            <div key={status} className="flex items-center justify-between text-sm">
              <span className="text-on-surface-variant">{getStatusLabel(activitiesConfig, status)}</span>
              <span className="font-semibold text-on-surface">{count} <span className="text-on-surface-variant font-normal">({pct}%)</span></span>
            </div>
          )
        })}
      </div>

      {/* Field vs non-field bar */}
      <p className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider mb-1">
        {t('dashboardCards.fieldVsNonField')}
      </p>
      <div className="flex h-2 rounded-full overflow-hidden bg-slate-100">
        {fieldCount > 0 && (
          <div
            className="bg-primary h-full"
            style={{ width: `${fieldPct}%` }}
            title={`${t('dashboardCards.field')}: ${fieldCount}`}
          />
        )}
        {nonFieldCount > 0 && (
          <div
            className="bg-tertiary-container h-full"
            style={{ width: `${100 - fieldPct}%` }}
            title={`${t('dashboardCards.nonField')}: ${nonFieldCount}`}
          />
        )}
      </div>
      <div className="flex justify-between text-xs text-on-surface-variant mt-1">
        <span>{t('dashboardCards.field')}: {fieldCount}</span>
        <span>{t('dashboardCards.nonField')}: {nonFieldCount}</span>
      </div>
    </div>
  )
}
