import { Info } from 'lucide-react'

export interface PlannerNudgeBannerProps {
  readonly overdueA: number
  readonly completionRate: number
  readonly coveragePct: number
}

export function PlannerNudgeBanner({ overdueA, completionRate, coveragePct }: PlannerNudgeBannerProps) {
  return (
    <div className="bg-indigo-50 border-b border-indigo-100 px-4 py-2 flex items-center justify-between shrink-0 md:px-6">
      <div className="flex items-center gap-3">
        <Info size={16} className="text-indigo-500 shrink-0" />
        <span className="text-sm text-indigo-900">
          {overdueA > 0 ? (
            <>You have <strong>{overdueA} A-priority targets</strong> that need visits.</>
          ) : (
            <>All A-priority targets are covered.</>
          )}
        </span>
      </div>
      <div className="hidden sm:flex items-center gap-4 text-xs font-medium shrink-0">
        <span className="text-slate-600">
          Completed: <span className={completionRate >= 80 ? 'text-emerald-600' : 'text-amber-600'}>{completionRate}%</span>
        </span>
        <span className="text-slate-600">
          Coverage: <span className={coveragePct >= 80 ? 'text-emerald-600' : 'text-amber-600'}>{coveragePct}%</span>
        </span>
      </div>
    </div>
  )
}
