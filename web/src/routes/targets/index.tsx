import { useState, useMemo } from 'react'
import { createRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { useTranslation } from 'react-i18next'
import { Target, Download, ChevronLeft, ChevronRight, MapPin } from 'lucide-react'
import { Route as rootRoute } from '../__root'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { useTargets, useTargetFrequencyStatus } from '../../services/targets'
import { useConfig } from '../../services/config'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/targets',
  component: TargetsPage,
})

const PAGE_SIZE = 20

const INITIALS_COLORS = [
  'bg-primary/5 text-primary',
  'bg-amber-50 text-amber-700',
  'bg-emerald-50 text-tertiary-container',
  'bg-slate-100 text-slate-600',
  'bg-violet-50 text-violet-700',
  'bg-rose-50 text-rose-600',
]

function colorForInitials(s: string): string {
  let hash = 0
  for (let i = 0; i < s.length; i++) hash = s.charCodeAt(i) + ((hash << 5) - hash)
  return INITIALS_COLORS[Math.abs(hash) % INITIALS_COLORS.length]
}

function TypeBadge({ targetType, label }: { targetType: string; label?: string }) {
  const style =
    targetType === 'doctor'
      ? 'bg-primary-fixed text-primary'
      : targetType === 'pharmacy'
        ? 'bg-emerald-100 text-emerald-700'
        : 'bg-slate-200 text-slate-600'
  return (
    <span className={`px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-tight ${style}`}>
      {label ?? targetType.replace(/_/g, ' ')}
    </span>
  )
}

function FrequencyBadge({ freq }: { freq?: { compliance: number; visitCount: number; required: number } }) {
  if (!freq || freq.required === 0) {
    return <span className="text-sm text-slate-300">—</span>
  }
  const pct = Math.round(freq.compliance)
  const color =
    pct >= 80
      ? 'bg-emerald-100 text-emerald-700'
      : pct >= 50
        ? 'bg-amber-100 text-amber-700'
        : 'bg-red-100 text-red-700'
  return (
    <div className="flex items-center gap-2">
      <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold ${color}`}>
        {pct}%
      </span>
      <span className="text-[10px] text-slate-400">
        {freq.visitCount}/{freq.required}
      </span>
    </div>
  )
}

export function TargetsPage() {
  const { t } = useTranslation()
  const [page, setPage] = useState(1)
  const [typeFilter, setTypeFilter] = useState('')

  const { data: config } = useConfig()
  const { data, isLoading, isError, error } = useTargets({
    page,
    limit: PAGE_SIZE,
    ...(typeFilter ? { type: typeFilter } : {}),
  })

  // Current month period for frequency status
  const currentPeriod = useMemo(() => {
    const now = new Date()
    return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
  }, [])
  const { data: frequencyData } = useTargetFrequencyStatus(currentPeriod)
  const frequencyMap = useMemo(() => {
    const map = new Map<string, { compliance: number; visitCount: number; required: number }>()
    if (frequencyData) {
      for (const item of frequencyData) {
        map.set(item.targetId, {
          compliance: item.compliance,
          visitCount: item.visitCount,
          required: item.required,
        })
      }
    }
    return map
  }, [frequencyData])

  const targets = data?.items ?? []
  const total = data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))
  const hasPrev = page > 1
  const hasNext = page < totalPages

  const accountTypes = config?.accounts.types ?? []
  const typeOptions: { value: string; label: string }[] = [
    { value: '', label: t('targets.allTypes') },
    ...accountTypes.map((t) => ({ value: t.key, label: t.label })),
  ]

  // Resolve field label from config
  function resolveFieldLabel(targetType: string, fieldKey: string): string {
    const acct = accountTypes.find((a) => a.key === targetType)
    const field = acct?.fields.find((f) => f.key === fieldKey)
    if (field) return field.key.replace(/_/g, ' ')
    return fieldKey.replace(/_/g, ' ')
  }

  // Resolve option label from config
  function resolveOptionLabel(ref: string, value: string): string {
    const opts = config?.options[ref]
    if (!opts) return value
    const opt = opts.find((o) => o.key === value)
    return opt?.label ?? value
  }

  // Get display value for a dynamic field
  function getFieldDisplay(targetType: string, fieldKey: string, value: unknown): string {
    if (value == null || value === '') return '—'
    const acct = accountTypes.find((a) => a.key === targetType)
    const fieldDef = acct?.fields.find((f) => f.key === fieldKey)
    if (fieldDef?.options_ref && typeof value === 'string') {
      return resolveOptionLabel(fieldDef.options_ref, value)
    }
    return String(value)
  }

  // Determine which dynamic columns to show based on active type filter
  const dynamicColumns: { key: string; label: string }[] = []
  if (typeFilter) {
    const acct = accountTypes.find((a) => a.key === typeFilter)
    if (acct) {
      for (const f of acct.fields) {
        if (f.key === 'name') continue // name is already the main column
        dynamicColumns.push({ key: f.key, label: resolveFieldLabel(typeFilter, f.key) })
      }
    }
  }

  // Visible page numbers
  const visiblePages: number[] = []
  const maxVisible = 3
  let startPage = Math.max(1, page - Math.floor(maxVisible / 2))
  const endPage = Math.min(totalPages, startPage + maxVisible - 1)
  startPage = Math.max(1, endPage - maxVisible + 1)
  for (let i = startPage; i <= endPage; i++) visiblePages.push(i)

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="p-4 sm:p-8 max-w-7xl mx-auto w-full space-y-6 sm:space-y-8"
    >
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:justify-between sm:items-end gap-4">
        <div>
          <h1 className="text-2xl sm:text-4xl font-extrabold tracking-tight text-primary leading-tight font-headline">
            {t('targets.title')}
          </h1>
          <p className="text-on-surface-variant mt-1 font-medium text-sm sm:text-base">
            {t('targets.subtitle', { count: total.toLocaleString() })}
          </p>
        </div>
        <div className="flex gap-3 items-center">
          <label htmlFor="type-filter" className="sr-only">Type:</label>
          <select
            id="type-filter"
            value={typeFilter}
            onChange={(e) => {
              setTypeFilter(e.target.value)
              setPage(1)
            }}
            className="px-4 py-2.5 bg-surface-container-high text-on-surface rounded-xl text-sm font-semibold border-none cursor-pointer hover:bg-surface-dim transition-colors"
          >
            {typeOptions.map((opt) => (
              <option key={opt.value} value={opt.value}>{opt.label}</option>
            ))}
          </select>
          <button className="px-5 py-2.5 bg-primary text-white rounded-xl text-sm font-semibold flex items-center gap-2 shadow-sm hover:opacity-90 transition-opacity">
            <Download className="w-4 h-4" />
            {t('common.export')}
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="bg-surface-container-lowest p-6 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
          <div className="flex justify-between items-start">
            <div className="p-2 bg-blue-50 text-primary rounded-lg">
              <Target className="w-5 h-5" />
            </div>
          </div>
          <div className="mt-4">
            <div className="text-sm font-medium text-slate-500">{t('targets.totalTargets')}</div>
            <div className="text-3xl font-extrabold font-headline mt-1">{total}</div>
          </div>
        </div>
        {accountTypes.map((acct) => {
          const count = targets.filter((t) => t.targetType === acct.key).length
          return (
            <div key={acct.key} className="bg-surface-container-lowest p-6 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
              <div className="flex justify-between items-start">
                <div className={`p-2 rounded-lg ${acct.key === 'doctor' ? 'bg-primary-fixed text-primary' : 'bg-emerald-50 text-emerald-600'}`}>
                  <Target className="w-5 h-5" />
                </div>
              </div>
              <div className="mt-4">
                <div className="text-sm font-medium text-slate-500">{acct.label}s</div>
                <div className="text-3xl font-extrabold font-headline mt-1">{count}</div>
              </div>
            </div>
          )
        })}
      </div>

      {/* Main Content */}
      {isLoading ? (
        <div className="flex items-center justify-center h-64">
          <LoadingSpinner size="lg" label={t('targets.loading')} />
        </div>
      ) : isError ? (
        <div data-testid="error-state" className="p-8 text-center text-error">
          {error instanceof Error ? error.message : t('error.failedToLoadTargets')}
        </div>
      ) : (
        <>
          {/* Data Table */}
          <div className="bg-surface-container-low rounded-xl p-1 overflow-x-auto">
            <div className="bg-surface-container-lowest rounded-lg shadow-[0px_24px_48px_rgba(25,28,30,0.06)] overflow-hidden min-w-[600px]">
              <table className="w-full text-left border-separate border-spacing-0">
                <thead>
                  <tr className="bg-surface-container-low/50">
                    <th className="px-6 py-4 text-[11px] font-bold uppercase tracking-widest text-slate-400">
                      {t('targets.name')}
                    </th>
                    <th className="px-6 py-4 text-[11px] font-bold uppercase tracking-widest text-slate-400">
                      {t('targets.type')}
                    </th>
                    {dynamicColumns.map((col) => (
                      <th key={col.key} className="px-6 py-4 text-[11px] font-bold uppercase tracking-widest text-slate-400">
                        {col.label}
                      </th>
                    ))}
                    {!typeFilter && (
                      <th className="px-6 py-4 text-[11px] font-bold uppercase tracking-widest text-slate-400">
                        {t('targets.location')}
                      </th>
                    )}
                    <th className="px-6 py-4 text-[11px] font-bold uppercase tracking-widest text-slate-400">
                      {t('targets.frequency')}
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-50">
                  {targets.map((target) => (
                    <tr key={target.id} className="hover:bg-slate-50/50 transition-colors">
                      <td className="px-6 py-5">
                        <Link
                          to="/targets/$targetId"
                          params={{ targetId: target.id }}
                          className="flex items-center gap-3 no-underline"
                        >
                          <div
                            className={`w-10 h-10 rounded-lg flex items-center justify-center font-bold text-sm ${colorForInitials(target.name)}`}
                          >
                            {target.name.slice(0, 2).toUpperCase()}
                          </div>
                          <div>
                            <div className="text-sm font-bold text-primary">{target.name}</div>
                          </div>
                        </Link>
                      </td>
                      <td className="px-6 py-5">
                        <TypeBadge
                          targetType={target.targetType}
                          label={accountTypes.find((a) => a.key === target.targetType)?.label}
                        />
                      </td>
                      {dynamicColumns.map((col) => (
                        <td key={col.key} className="px-6 py-5">
                          <span className="text-sm text-slate-600">
                            {getFieldDisplay(target.targetType, col.key, target.fields[col.key])}
                          </span>
                        </td>
                      ))}
                      {!typeFilter && (
                        <td className="px-6 py-5">
                          {target.fields['city'] ? (
                            <div className="flex items-center gap-2">
                              <MapPin className="w-4 h-4 text-slate-300" />
                              <span className="text-sm text-slate-600">
                                {String(target.fields['city'])}
                                {target.fields['county'] ? `, ${String(target.fields['county'])}` : ''}
                              </span>
                            </div>
                          ) : (
                            <span className="text-sm text-slate-300">—</span>
                          )}
                        </td>
                      )}
                      <td className="px-6 py-5">
                        <FrequencyBadge freq={frequencyMap.get(target.id)} />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>

              {targets.length === 0 && (
                <div data-testid="empty-state" className="px-6 py-12 text-center text-on-surface-variant">
                  {typeFilter
                    ? t('targets.noTargetsOfType', { type: typeFilter })
                    : t('targets.noTargets')}
                </div>
              )}

              {/* Pagination */}
              <div
                data-testid="pagination"
                className="px-6 py-4 bg-surface-container-lowest border-t border-slate-50 flex items-center justify-between"
              >
                <span data-testid="result-count" className="text-xs font-medium text-slate-400">
                  {total > 0
                    ? `${(page - 1) * PAGE_SIZE + 1}–${Math.min(page * PAGE_SIZE, total)} ${t('common.of')} ${total}`
                    : t('common.noResults')}
                </span>
                <div className="flex gap-2">
                  <button
                    data-testid="prev-page"
                    onClick={() => setPage((p) => p - 1)}
                    disabled={!hasPrev}
                    className="w-8 h-8 flex items-center justify-center border border-slate-100 rounded-lg text-slate-400 hover:bg-slate-50 transition-colors disabled:opacity-40"
                  >
                    <ChevronLeft className="w-4 h-4" />
                  </button>
                  {visiblePages.map((p) => (
                    <button
                      key={p}
                      onClick={() => setPage(p)}
                      className={`w-8 h-8 flex items-center justify-center rounded-lg text-xs font-bold transition-colors ${
                        p === page
                          ? 'bg-primary text-white'
                          : 'border border-slate-100 text-slate-600 hover:bg-slate-50'
                      }`}
                    >
                      {p}
                    </button>
                  ))}
                  <button
                    data-testid="next-page"
                    onClick={() => setPage((p) => p + 1)}
                    disabled={!hasNext}
                    className="w-8 h-8 flex items-center justify-center border border-slate-100 rounded-lg text-slate-400 hover:bg-slate-50 transition-colors disabled:opacity-40"
                  >
                    <ChevronRight className="w-4 h-4" />
                  </button>
                </div>
                <span data-testid="page-indicator" className="text-xs font-medium text-slate-400">
                  {t('common.page')} {page} {t('common.of')} {totalPages}
                </span>
              </div>
            </div>
          </div>
        </>
      )}
    </motion.div>
  )
}
