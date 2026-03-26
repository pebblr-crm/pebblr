import { useTranslation } from 'react-i18next'
import type { FrequencyItem } from '@/types/dashboard'

interface FrequencyTableProps {
  items: FrequencyItem[]
}

function complianceColor(compliance: number): string {
  if (compliance >= 80) return 'text-tertiary-container'
  if (compliance >= 50) return 'text-amber-500'
  return 'text-error'
}

function complianceBarColor(compliance: number): string {
  if (compliance >= 80) return 'bg-tertiary-container'
  if (compliance >= 50) return 'bg-amber-400'
  return 'bg-error'
}

export function FrequencyTable({ items }: Readonly<FrequencyTableProps>) {
  const { t } = useTranslation()

  if (items.length === 0) return null

  return (
    <div className="bg-surface-container-lowest p-6 rounded-xl shadow-sm border border-slate-50">
      <p className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider mb-4">
        {t('dashboardCards.frequencyCompliance')}
      </p>
      <div className="space-y-3">
        {items.map((item) => {
          const pct = Math.round(item.compliance)
          return (
            <div key={item.classification}>
              <div className="flex items-center justify-between mb-1">
                <span className="text-sm font-semibold text-on-surface uppercase">
                  {item.classification}
                </span>
                <span className={`text-sm font-bold ${complianceColor(item.compliance)}`}>
                  {pct}%
                </span>
              </div>
              <div className="flex items-center space-x-3">
                <div className="flex-1 h-1.5 bg-slate-100 rounded-full overflow-hidden">
                  <div
                    className={`h-full rounded-full ${complianceBarColor(item.compliance)}`}
                    style={{ width: `${pct}%` }}
                  />
                </div>
                <span className="text-xs text-on-surface-variant whitespace-nowrap">
                  {item.totalVisits} / {item.required * item.targetCount} {t('dashboardCards.visits')}
                </span>
              </div>
              <p className="text-xs text-on-surface-variant mt-0.5">
                {t('dashboardCards.targets_count', { count: item.targetCount, required: item.required })}
              </p>
            </div>
          )
        })}
      </div>
    </div>
  )
}
