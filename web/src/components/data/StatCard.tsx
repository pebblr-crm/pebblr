import { Card } from '@/components/ui/Card'

interface StatCardProps {
  readonly label: string
  readonly value: string | number
  readonly subtitle?: string
  readonly trend?: 'up' | 'down' | 'neutral'
  readonly className?: string
}

const trendColors: Record<string, string> = {
  up: 'text-emerald-600',
  down: 'text-red-600',
  neutral: 'text-slate-500',
}

export function StatCard({ label, value, subtitle, trend, className = '' }: StatCardProps) {
  const trendColor = trendColors[trend ?? 'neutral'] ?? 'text-slate-500'
  return (
    <Card className={className}>
      <p className="text-sm text-slate-500">{label}</p>
      <p className="mt-1 text-2xl font-semibold text-slate-900">{value}</p>
      {subtitle && <p className={`mt-1 text-xs ${trendColor}`}>{subtitle}</p>}
    </Card>
  )
}
