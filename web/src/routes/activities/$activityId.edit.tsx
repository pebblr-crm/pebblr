import { createRoute, useNavigate } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { ArrowLeft } from 'lucide-react'
import { Link } from '@tanstack/react-router'
import { useState } from 'react'
import { Route as rootRoute } from '../__root'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { ActivityForm } from '../../components/ActivityForm'
import { useActivity, useUpdateActivity } from '../../services/activities'
import type { ValidationFieldError } from '../../types/activity'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/activities/$activityId/edit',
  component: EditActivityPage,
})

export function EditActivityPage() {
  const { activityId } = Route.useParams()
  const navigate = useNavigate()
  const { data: activity, isLoading, isError, error } = useActivity(activityId)
  const updateMutation = useUpdateActivity()
  const [serverErrors, setServerErrors] = useState<ValidationFieldError[]>([])

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <LoadingSpinner size="lg" label="Loading activity..." />
      </div>
    )
  }

  if (isError) {
    return (
      <div data-testid="error-state" className="p-8 text-center text-error">
        {error instanceof Error ? error.message : 'Failed to load activity.'}
      </div>
    )
  }

  if (!activity) {
    return (
      <div data-testid="not-found" className="p-8 text-center text-on-surface-variant">
        Activity not found.
      </div>
    )
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="p-8 max-w-4xl mx-auto w-full space-y-6"
    >
      <Link
        to="/activities/$activityId"
        params={{ activityId }}
        className="inline-flex items-center gap-2 text-sm font-medium text-on-surface-variant hover:text-primary transition-colors no-underline"
      >
        <ArrowLeft className="w-4 h-4" />
        Back to activity
      </Link>

      <ActivityForm
        initialData={activity}
        onSubmit={(data) => {
          setServerErrors([])
          updateMutation.mutate(
            { ...data, id: activityId },
            {
              onSuccess: () => {
                void navigate({
                  to: '/activities/$activityId',
                  params: { activityId },
                })
              },
              onError: (err) => {
                const apiErr = err as Error & { status?: number }
                if (apiErr.status === 422) {
                  try {
                    const body = JSON.parse(apiErr.message) as { fields?: ValidationFieldError[] }
                    if (body.fields) setServerErrors(body.fields)
                  } catch {
                    // non-JSON error
                  }
                }
              },
            },
          )
        }}
        onCancel={() =>
          void navigate({
            to: '/activities/$activityId',
            params: { activityId },
          })
        }
        isSubmitting={updateMutation.isPending}
        serverErrors={serverErrors}
      />
    </motion.div>
  )
}
