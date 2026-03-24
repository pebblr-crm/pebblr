import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { ChevronLeft, ChevronRight, Lock } from 'lucide-react'
import { useActivities } from '../../services/activities'
import { useConfig } from '../../services/config'
import { LoadingSpinner } from '../LoadingSpinner'
import {
  getActivityDisplayName,
  getStatusLabel,
  getDurationLabel,
  getTypeCategory,
  getStatusBadgeColor,
  CATEGORY_COLORS,
} from '@/utils/config'
import type { Activity } from '@/types/activity'

const PAGE_SIZE = 20

export function ActivityList() {
  const { t } = useTranslation()
  const { data: config } = useConfig()
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState('')
  const [typeFilter, setTypeFilter] = useState('')

  const { data, isLoading } = useActivities({
    page,
    limit: PAGE_SIZE,
    status: statusFilter || undefined,
    activityType: typeFilter || undefined,
  })

  const activities = data?.items ?? []
  const total = data?.total ?? 0
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  const statuses = config?.activities.statuses ?? []
  const types = config?.activities.types ?? []

  function onFilterChange() {
    setPage(1)
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <LoadingSpinner size="lg" label={t('activityList.loadingActivities')} />
      </div>
    )
  }

  return (
    <div className="space-y-4">
      {/* Filters */}
      <div className="flex flex-wrap items-center gap-3">
        <select
          value={statusFilter}
          onChange={(e) => { setStatusFilter(e.target.value); onFilterChange() }}
          className="px-3 py-2 text-sm border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary"
          data-testid="status-filter"
        >
          <option value="">{t('activityList.allStatuses')}</option>
          {statuses.map((s) => (
            <option key={s.key} value={s.key}>{s.label}</option>
          ))}
        </select>
        <select
          value={typeFilter}
          onChange={(e) => { setTypeFilter(e.target.value); onFilterChange() }}
          className="px-3 py-2 text-sm border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary"
          data-testid="type-filter"
        >
          <option value="">{t('activityList.allTypes')}</option>
          {types.map((t) => (
            <option key={t.key} value={t.key}>{t.label}</option>
          ))}
        </select>
        <span className="text-xs text-on-surface-variant ml-auto">
          {t('activityList.activity', { count: total })}
        </span>
      </div>

      {/* Table (desktop) / Cards (mobile) */}
      {activities.length === 0 ? (
        <div className="text-center py-16 text-on-surface-variant text-sm">
          {t('activityList.noActivities')}
        </div>
      ) : (
        <>
          {/* Desktop table */}
          <div className="hidden sm:block bg-surface-container-lowest rounded-xl shadow-sm border border-slate-100 overflow-hidden">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-slate-100 text-left">
                  <th className="px-4 py-3 text-xs font-bold uppercase tracking-widest text-slate-400">{t('activityList.activityCol')}</th>
                  <th className="px-4 py-3 text-xs font-bold uppercase tracking-widest text-slate-400">{t('activityList.durationCol')}</th>
                  <th className="px-4 py-3 text-xs font-bold uppercase tracking-widest text-slate-400">{t('activityList.statusCol')}</th>
                  <th className="px-4 py-3 text-xs font-bold uppercase tracking-widest text-slate-400 w-8"></th>
                </tr>
              </thead>
              <tbody>
                {activities.map((a) => (
                  <ActivityRow key={a.id} activity={a} config={config} />
                ))}
              </tbody>
            </table>
          </div>

          {/* Mobile cards */}
          <div className="sm:hidden space-y-2">
            {activities.map((a) => (
              <ActivityCardRow key={a.id} activity={a} config={config} />
            ))}
          </div>
        </>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-4 pt-2">
          <button
            onClick={() => setPage((p) => Math.max(1, p - 1))}
            disabled={page <= 1}
            className="p-2 rounded-lg hover:bg-slate-50 disabled:opacity-30 transition-colors"
            aria-label="Previous page"
          >
            <ChevronLeft className="w-4 h-4" />
          </button>
          <span className="text-xs text-on-surface-variant">
            {t('common.page')} {page} {t('common.of')} {totalPages}
          </span>
          <button
            onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
            disabled={page >= totalPages}
            className="p-2 rounded-lg hover:bg-slate-50 disabled:opacity-30 transition-colors"
            aria-label="Next page"
          >
            <ChevronRight className="w-4 h-4" />
          </button>
        </div>
      )}
    </div>
  )
}

// ── Table row (desktop) ────────────────────────────────────────────────────

interface RowProps {
  activity: Activity
  config: import('@/types/config').TenantConfig | undefined
}

function ActivityRow({ activity: a, config }: RowProps) {
  const ac = config?.activities
  const displayName = getActivityDisplayName(config, a)
  const catClass = CATEGORY_COLORS[getTypeCategory(ac, a.activityType)] ?? ''
  const statusColor = getStatusBadgeColor(ac, a.status)

  return (
    <tr className="border-b border-slate-50 hover:bg-slate-50/50 transition-colors">
      <td className="px-4 py-3">
        <Link
          to="/activities/$activityId"
          params={{ activityId: a.id }}
          className="text-primary font-medium no-underline hover:underline"
        >
          {displayName}
        </Link>
      </td>
      <td className="px-4 py-3">
        <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium border-l-2 ${catClass}`}>
          {getDurationLabel(ac, a.duration)}
        </span>
      </td>
      <td className="px-4 py-3">
        <span className={`inline-block px-2 py-0.5 rounded-full text-xs font-bold ${statusColor}`}>
          {getStatusLabel(ac, a.status)}
        </span>
      </td>
      <td className="px-4 py-3">
        {a.submittedAt && <Lock className="w-3.5 h-3.5 text-slate-400" />}
      </td>
    </tr>
  )
}

// ── Card row (mobile) ──────────────────────────────────────────────────────

function ActivityCardRow({ activity: a, config }: RowProps) {
  const ac = config?.activities
  const displayName = getActivityDisplayName(config, a)
  const catClass = CATEGORY_COLORS[getTypeCategory(ac, a.activityType)] ?? ''
  const statusColor = getStatusBadgeColor(ac, a.status)

  return (
    <Link
      to="/activities/$activityId"
      params={{ activityId: a.id }}
      className="block bg-surface-container-lowest p-4 rounded-xl shadow-sm border border-slate-100 no-underline"
    >
      <div className="flex items-center justify-between gap-2">
        <span className={`inline-block px-2 py-0.5 rounded text-xs font-medium border-l-2 ${catClass}`}>
          {displayName}
        </span>
        <div className="flex items-center gap-2">
          <span className={`inline-block px-2 py-0.5 rounded-full text-xs font-bold ${statusColor}`}>
            {getStatusLabel(ac, a.status)}
          </span>
          {a.submittedAt && <Lock className="w-3.5 h-3.5 text-slate-400" />}
        </div>
      </div>
      <div className="mt-2 flex items-center gap-3 text-xs text-on-surface-variant">
        <span>{getDurationLabel(ac, a.duration)}</span>
      </div>
    </Link>
  )
}
