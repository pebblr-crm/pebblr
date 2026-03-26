import { createRoute, useNavigate, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { useTranslation } from 'react-i18next'
import { ArrowLeft } from 'lucide-react'
import { Route as rootRoute } from '../__root'
import { ActivityForm } from '../../components/ActivityForm'
import { useCreateActivity } from '../../services/activities'
import { useConfig } from '../../services/config'
import { useToast } from '../../hooks/useToast'
import { formatValidationToast } from '../../utils/fieldLabels'
import { usePlannerState } from '../../contexts/planner'
import type { ValidationFieldError } from '../../types/activity'
import type { ApiError } from '../../types/api'
import { useState } from 'react'

function resolveBackPath(from?: string): '/' | '/planner' | '/planner/daily' | '/planner/map' {
  if (from === 'planner') return '/planner'
  if (from === 'daily') return '/planner/daily'
  if (from === 'map') return '/planner/map'
  return '/'
}

function resolveBackLabel(from: string | undefined, t: (key: string) => string): string {
  if (from === 'planner') return t('activityDetail.backToPlanner')
  if (from === 'daily') return t('activityDetail.backToDaily')
  if (from === 'map') return t('activityDetail.backToMap')
  return t('activityDetail.backToDashboard')
}

interface NewActivitySearch {
  date?: string
}

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/activities/new',
  component: NewActivityPage,
  validateSearch: (search: Record<string, unknown>): NewActivitySearch => ({
    date: typeof search.date === 'string' ? search.date : undefined,
  }),
})

export function NewActivityPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { date } = Route.useSearch()
  const { state: { from } } = usePlannerState()
  const createMutation = useCreateActivity()
  const { data: config } = useConfig()
  const { showToast } = useToast()
  const [serverErrors, setServerErrors] = useState<ValidationFieldError[]>([])

  const backLabel = resolveBackLabel(from ?? undefined, t)
  const backPath = resolveBackPath(from ?? undefined)

  function navigateBack() {
    navigate({ to: backPath })
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="p-4 sm:p-8 max-w-4xl mx-auto w-full space-y-6"
    >
      <Link
        to={backPath}
        className="inline-flex items-center gap-2 text-sm font-medium text-on-surface-variant hover:text-primary transition-colors no-underline"
      >
        <ArrowLeft className="w-4 h-4" />
        {backLabel}
      </Link>

      <ActivityForm
        initialDate={date}
        onSubmit={(data) => {
          setServerErrors([])
          createMutation.mutate(data, {
            onSuccess: () => {
              navigateBack()
            },
            onError: (err) => {
              const apiErr = err as ApiError
              if (apiErr.status === 422 && apiErr.fields) {
                setServerErrors(apiErr.fields)
                showToast(formatValidationToast(config, data.activityType, apiErr.fields))
              } else {
                showToast(apiErr.message || t('activity.failedToCreate'))
              }
            },
          })
        }}
        onCancel={navigateBack}
        isSubmitting={createMutation.isPending}
        serverErrors={serverErrors}
      />
    </motion.div>
  )
}
