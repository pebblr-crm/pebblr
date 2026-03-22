interface StatCardProps {
  label: string
  value: string
  change?: string
  progress?: number
  variant?: 'default' | 'primary'
}

export function StatCard({ label, value, change, progress, variant = 'default' }: StatCardProps) {
  const isPositive = change?.startsWith('+')

  if (variant === 'primary') {
    return (
      <div className="bg-primary text-white p-6 rounded-xl shadow-xl shadow-primary/10">
        <p className="text-[10px] font-bold text-blue-200 uppercase tracking-wider mb-2">{label}</p>
        <div className="flex items-baseline space-x-2">
          <h2 className="text-3xl font-extrabold font-headline">{value}</h2>
          {change && <span className="text-[10px] font-medium text-blue-200">{change}</span>}
        </div>
        {progress !== undefined && (
          <div className="mt-4 h-1.5 w-full bg-white/20 rounded-full overflow-hidden">
            <div className="h-full bg-white/60 rounded-full" style={{ width: `${progress}%` }} />
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="bg-surface-container-lowest p-6 rounded-xl shadow-sm border border-slate-50">
      <p className="text-[10px] font-bold text-on-surface-variant uppercase tracking-wider mb-2">{label}</p>
      <div className="flex items-baseline space-x-2">
        <h2 className="text-3xl font-extrabold text-primary font-headline">{value}</h2>
        {change && (
          <span className={`text-[10px] font-bold ${isPositive ? 'text-tertiary-container' : 'text-error'}`}>
            {change}
          </span>
        )}
      </div>
      {progress !== undefined && (
        <div className="mt-4 h-1.5 w-full bg-slate-100 rounded-full overflow-hidden">
          <div className="h-full bg-primary rounded-full" style={{ width: `${progress}%` }} />
        </div>
      )}
    </div>
  )
}
