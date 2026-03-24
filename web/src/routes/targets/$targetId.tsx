import { createRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { useTranslation } from 'react-i18next'
import { ArrowLeft, MapPin } from 'lucide-react'
import { Route as rootRoute } from '../__root'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { useTarget } from '../../services/targets'
import { useConfig } from '../../services/config'

function getTypeBadgeColor(targetType: string): string {
  if (targetType === 'doctor') return 'bg-primary-fixed text-primary'
  if (targetType === 'pharmacy') return 'bg-emerald-100 text-emerald-700'
  return 'bg-slate-200 text-slate-600'
}

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/targets/$targetId',
  component: TargetDetailPage,
})

export function TargetDetailPage() {
  const { t } = useTranslation()
  const { targetId } = Route.useParams()
  const { data: target, isLoading, isError, error } = useTarget(targetId)
  const { data: config } = useConfig()

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <LoadingSpinner size="lg" label={t('targets.loadingTarget')} />
      </div>
    )
  }

  if (isError) {
    return (
      <div data-testid="error-state" className="p-8 text-center text-error">
        {error instanceof Error ? error.message : t('targets.failedToLoad')}
      </div>
    )
  }

  if (!target) {
    return (
      <div data-testid="not-found" className="p-8 text-center text-on-surface-variant">
        {t('targets.notFound')}
      </div>
    )
  }

  const accountTypes = config?.accounts.types ?? []
  const acctConfig = accountTypes.find((a) => a.key === target.targetType)
  const typeLabel = acctConfig?.label ?? target.targetType
  const fields = acctConfig?.fields.filter((f) => f.key !== 'name') ?? []

  function resolveOptionLabel(ref: string, value: string): string {
    const opts = config?.options[ref]
    if (!opts) return value
    const opt = opts.find((o) => o.key === value)
    return opt?.label ?? value
  }

  function getFieldDisplay(fieldKey: string, value: unknown): string {
    if (value == null || value === '') return '—'
    const fieldDef = acctConfig?.fields.find((f) => f.key === fieldKey)
    if (fieldDef?.options_ref && typeof value === 'string') {
      return resolveOptionLabel(fieldDef.options_ref, value)
    }
    if (typeof value === 'object') return JSON.stringify(value)
    return String(value)
  }

  const city = target.fields['city']
  const county = target.fields['county']
  const address = target.fields['address']
  const location = [address, city, county].filter(Boolean).join(', ')

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="p-4 sm:p-8 max-w-4xl mx-auto w-full space-y-6 sm:space-y-8"
    >
      {/* Back link */}
      <Link
        to="/targets"
        className="inline-flex items-center gap-2 text-sm font-medium text-on-surface-variant hover:text-primary transition-colors no-underline"
      >
        <ArrowLeft className="w-4 h-4" />
        {t('targets.backToTargets')}
      </Link>

      {/* Header */}
      <div className="bg-surface-container-lowest p-4 sm:p-8 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
        <div className="flex items-start gap-4 sm:gap-6">
          <div className="w-16 h-16 rounded-xl bg-primary/5 text-primary flex items-center justify-center font-bold text-xl">
            {target.name.slice(0, 2).toUpperCase()}
          </div>
          <div className="flex-1">
            <h1 className="text-xl sm:text-3xl font-extrabold tracking-tight text-primary font-headline">
              {target.name}
            </h1>
            <div className="flex items-center gap-3 mt-2">
              <span className={`px-3 py-1 rounded-full text-[10px] font-bold uppercase tracking-tight ${getTypeBadgeColor(target.targetType)}`}>
                {typeLabel}
              </span>
              {location && (
                <div className="flex items-center gap-1 text-sm text-on-surface-variant">
                  <MapPin className="w-4 h-4" />
                  {location}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Details */}
      <div className="bg-surface-container-lowest p-8 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
        <h2 className="text-lg font-bold text-on-surface mb-6 font-headline">{t('targets.details')}</h2>
        <dl className="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-4">
          {fields.map((f) => (
            <div key={f.key}>
              <dt className="text-xs font-bold uppercase tracking-widest text-slate-400 mb-1">
                {f.key.replace(/_/g, ' ')}
              </dt>
              <dd className="text-sm text-on-surface">
                {getFieldDisplay(f.key, target.fields[f.key])}
              </dd>
            </div>
          ))}
        </dl>
      </div>

      {/* Activities placeholder */}
      <div className="bg-surface-container-lowest p-8 rounded-xl shadow-[0px_24px_48px_rgba(25,28,30,0.06)]">
        <h2 className="text-lg font-bold text-on-surface mb-4 font-headline">{t('targets.activitiesSection')}</h2>
        <p className="text-sm text-on-surface-variant">
          {t('targets.activitiesPlaceholder')}
        </p>
      </div>
    </motion.div>
  )
}
