import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { formatPeriod } from '@/utils/date'
import { getDateLocale } from '@/utils/config'

interface PeriodSelectorProps {
  /** Current period in YYYY-MM format */
  period: string
  onPeriodChange: (period: string) => void
}

function formatPeriodLabel(period: string): string {
  const [year, month] = period.split('-')
  const date = new Date(Number(year), Number(month) - 1)
  return date.toLocaleDateString(getDateLocale(), { month: 'long', year: 'numeric' })
}

function shiftPeriod(period: string, delta: number): string {
  const [year, month] = period.split('-')
  return formatPeriod(new Date(Number(year), Number(month) - 1 + delta))
}

export function PeriodSelector({ period, onPeriodChange }: PeriodSelectorProps) {
  const { t } = useTranslation()

  return (
    <div className="flex items-center space-x-3">
      <button
        onClick={() => onPeriodChange(shiftPeriod(period, -1))}
        className="p-1.5 rounded-lg hover:bg-surface-container-high transition-colors"
        aria-label={t('pagination.previousMonth')}
      >
        <ChevronLeft className="w-5 h-5 text-on-surface-variant" />
      </button>
      <span className="text-sm font-semibold text-on-surface min-w-[140px] text-center">
        {formatPeriodLabel(period)}
      </span>
      <button
        onClick={() => onPeriodChange(shiftPeriod(period, 1))}
        className="p-1.5 rounded-lg hover:bg-surface-container-high transition-colors"
        aria-label={t('pagination.nextMonth')}
      >
        <ChevronRight className="w-5 h-5 text-on-surface-variant" />
      </button>
    </div>
  )
}
