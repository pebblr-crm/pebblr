import { useTranslation } from 'react-i18next'
import type { CoverageResponse } from '@/types/dashboard'

interface CoverageCardProps {
  data: CoverageResponse
}

export function CoverageCard({ data }: Readonly<CoverageCardProps>) {
  const { t } = useTranslation()
  const pct = Math.round(data.percentage)

  return (
    <div className="bg-surface-container-lowest p-6 rounded-xl shadow-sm border border-slate-50">
      <p className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider mb-2">
        {t('dashboardCards.targetCoverage')}
      </p>
      <div className="flex items-baseline space-x-2">
        <h2 className="text-3xl font-extrabold text-primary font-headline">{pct}%</h2>
        <span className="text-sm text-on-surface-variant">
          {data.visitedTargets} / {data.totalTargets}
        </span>
      </div>
      <div className="mt-4 h-1.5 w-full bg-slate-100 rounded-full overflow-hidden">
        <div className="h-full bg-primary rounded-full" style={{ width: `${pct}%` }} />
      </div>
    </div>
  )
}
