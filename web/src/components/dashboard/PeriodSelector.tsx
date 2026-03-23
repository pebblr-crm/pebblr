import { ChevronLeft, ChevronRight } from 'lucide-react'

const MONTH_NAMES = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
]

interface PeriodSelectorProps {
  /** Current period in YYYY-MM format */
  period: string
  onPeriodChange: (period: string) => void
}

function parsePeriod(period: string): { year: number; month: number } {
  const [y, m] = period.split('-').map(Number)
  return { year: y, month: m }
}

function formatPeriod(year: number, month: number): string {
  return `${year}-${String(month).padStart(2, '0')}`
}

export function PeriodSelector({ period, onPeriodChange }: PeriodSelectorProps) {
  const { year, month } = parsePeriod(period)

  function prev() {
    if (month === 1) {
      onPeriodChange(formatPeriod(year - 1, 12))
    } else {
      onPeriodChange(formatPeriod(year, month - 1))
    }
  }

  function next() {
    if (month === 12) {
      onPeriodChange(formatPeriod(year + 1, 1))
    } else {
      onPeriodChange(formatPeriod(year, month + 1))
    }
  }

  function goToday() {
    const now = new Date()
    onPeriodChange(formatPeriod(now.getFullYear(), now.getMonth() + 1))
  }

  return (
    <div className="flex items-center space-x-3">
      <button
        onClick={prev}
        className="p-1.5 rounded-lg hover:bg-surface-container-high transition-colors"
        aria-label="Previous month"
      >
        <ChevronLeft className="w-5 h-5 text-on-surface-variant" />
      </button>
      <span className="text-lg font-bold text-on-surface min-w-[180px] text-center">
        {MONTH_NAMES[month - 1]} {year}
      </span>
      <button
        onClick={next}
        className="p-1.5 rounded-lg hover:bg-surface-container-high transition-colors"
        aria-label="Next month"
      >
        <ChevronRight className="w-5 h-5 text-on-surface-variant" />
      </button>
      <button
        onClick={goToday}
        className="px-3 py-1 text-xs font-semibold bg-surface-container-high text-on-surface rounded-lg hover:bg-surface-container-highest transition-colors"
      >
        Today
      </button>
    </div>
  )
}
