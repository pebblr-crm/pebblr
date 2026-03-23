import { createRoute, useNavigate } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { ArrowLeft } from 'lucide-react'
import { Link } from '@tanstack/react-router'
import { Route as rootRoute } from '../__root'
import { ActivityForm } from '../../components/ActivityForm'
import { useCreateActivity } from '../../services/activities'
import { useConfig } from '../../services/config'
import { useToast } from '../../hooks/useToast'
import { formatValidationToast } from '../../utils/fieldLabels'
import type { ValidationFieldError } from '../../types/activity'
import type { ApiError } from '../../types/api'
import { useState } from 'react'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/activities/new',
  component: NewActivityPage,
})

export function NewActivityPage() {
  const navigate = useNavigate()
  const createMutation = useCreateActivity()
  const { data: config } = useConfig()
  const { showToast } = useToast()
  const [serverErrors, setServerErrors] = useState<ValidationFieldError[]>([])

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="p-8 max-w-4xl mx-auto w-full space-y-6"
    >
      <Link
        to="/"
        className="inline-flex items-center gap-2 text-sm font-medium text-on-surface-variant hover:text-primary transition-colors no-underline"
      >
        <ArrowLeft className="w-4 h-4" />
        Back to dashboard
      </Link>

      <ActivityForm
        onSubmit={(data) => {
          setServerErrors([])
          createMutation.mutate(data, {
            onSuccess: () => {
              void navigate({ to: '/' })
            },
            onError: (err) => {
              const apiErr = err as ApiError
              if (apiErr.status === 422 && apiErr.fields) {
                setServerErrors(apiErr.fields)
                showToast(formatValidationToast(config, data.activityType, apiErr.fields))
              } else {
                showToast(apiErr.message || 'Failed to create activity')
              }
            },
          })
        }}
        onCancel={() => void navigate({ to: '/' })}
        isSubmitting={createMutation.isPending}
        serverErrors={serverErrors}
      />
    </motion.div>
  )
}
