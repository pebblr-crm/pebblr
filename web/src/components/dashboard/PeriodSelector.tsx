import { ChevronLeft, ChevronRight } from 'lucide-react'

interface PeriodSelectorProps {
  /** Current period in YYYY-MM format */
  period: string
  onPeriodChange: (period: string) => void
}

function formatPeriodLabel(period: string): string {
  const [year, month] = period.split('-')
  const date = new Date(Number(year), Number(month) - 1)
  return date.toLocaleDateString('en-US', { month: 'long', year: 'numeric' })
}

function shiftPeriod(period: string, delta: number): string {
  const [year, month] = period.split('-')
  const date = new Date(Number(year), Number(month) - 1 + delta)
  const y = date.getFullYear()
  const m = String(date.getMonth() + 1).padStart(2, '0')
  return `${y}-${m}`
}

export function PeriodSelector({ period, onPeriodChange }: PeriodSelectorProps) {
  return (
    <div className="flex items-center space-x-3">
      <button
        onClick={() => onPeriodChange(shiftPeriod(period, -1))}
        className="p-1.5 rounded-lg hover:bg-surface-container-high transition-colors"
        aria-label="Previous month"
      >
        <ChevronLeft className="w-5 h-5 text-on-surface-variant" />
      </button>
      <span className="text-sm font-semibold text-on-surface min-w-[140px] text-center">
        {formatPeriodLabel(period)}
      </span>
      <button
        onClick={() => onPeriodChange(shiftPeriod(period, 1))}
        className="p-1.5 rounded-lg hover:bg-surface-container-high transition-colors"
        aria-label="Next month"
      >
        <ChevronRight className="w-5 h-5 text-on-surface-variant" />
      </button>
    </div>
  )
}
