import { Card } from '@/components/ui/Card'

interface StatCardProps {
  label: string
  value: string | number
  subtitle?: string
  trend?: 'up' | 'down' | 'neutral'
  className?: string
}

export function StatCard({ label, value, subtitle, trend, className = '' }: StatCardProps) {
  const trendColor = trend === 'up' ? 'text-emerald-600' : trend === 'down' ? 'text-red-600' : 'text-slate-500'
  return (
    <Card className={className}>
      <p className="text-sm text-slate-500">{label}</p>
      <p className="mt-1 text-2xl font-semibold text-slate-900">{value}</p>
      {subtitle && <p className={`mt-1 text-xs ${trendColor}`}>{subtitle}</p>}
    </Card>
  )
}
