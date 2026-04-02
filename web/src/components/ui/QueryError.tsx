import { AlertTriangle, RotateCcw } from 'lucide-react'

interface QueryErrorProps {
  readonly message?: string
  readonly onRetry?: () => void
}

export function QueryError({ message = 'Failed to load data', onRetry }: QueryErrorProps) {
  return (
    <div className="flex items-center justify-center gap-2 p-8">
      <div className="text-center">
        <div className="mx-auto mb-3 flex h-10 w-10 items-center justify-center rounded-full bg-red-50">
          <AlertTriangle size={20} className="text-red-500" />
        </div>
        <p className="text-sm font-medium text-slate-700">{message}</p>
        <p className="mt-1 text-xs text-slate-400">Check your connection and try again.</p>
        {onRetry && (
          <button
            onClick={onRetry}
            className="mt-3 inline-flex items-center gap-1.5 rounded-lg border border-slate-200 px-3 py-1.5 text-xs font-medium text-slate-600 hover:bg-slate-50 transition-colors"
          >
            <RotateCcw size={12} />
            Retry
          </button>
        )}
      </div>
    </div>
  )
}
