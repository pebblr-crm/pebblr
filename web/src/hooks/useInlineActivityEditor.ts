import { useState, useRef, useCallback, useEffect } from 'react'
import { useNavigate } from '@tanstack/react-router'
import {
  usePatchActivity,
  usePatchActivityStatus,
  useSubmitActivity,
} from '../services/activities'
import type { Activity, UpdateActivityInput, ValidationFieldError } from '../types/activity'

export type SaveState = 'idle' | 'dirty' | 'saving' | 'error'

export interface InlineActivityEditor {
  localData: Partial<UpdateActivityInput>
  saveState: SaveState
  fieldErrors: ValidationFieldError[]
  handleFieldChange: (key: string, value: unknown) => void
  handleFieldBlur: () => void
  handleStatusChange: (newStatus: string) => void
  handleSubmit: () => Promise<void>
  retrySave: () => void
  isSubmitting: boolean
}

const DEBOUNCE_MS = 1500

export function useInlineActivityEditor(activity: Activity): InlineActivityEditor {
  const navigate = useNavigate()
  const patchMutation = usePatchActivity()
  const statusMutation = usePatchActivityStatus()
  const submitMutation = useSubmitActivity()

  const [localData, setLocalData] = useState<Partial<UpdateActivityInput>>(() => ({
    activityType: activity.activityType,
    status: activity.status,
    dueDate: activity.dueDate,
    duration: activity.duration,
    routing: activity.routing,
    fields: { ...activity.fields },
    targetId: activity.targetId,
    jointVisitUserId: activity.jointVisitUserId,
  }))
  const [saveState, setSaveState] = useState<SaveState>('idle')
  const [fieldErrors, setFieldErrors] = useState<ValidationFieldError[]>([])

  // Keep a ref to localData for use inside debounce callback without stale closure.
  const localDataRef = useRef(localData)
  useEffect(() => {
    localDataRef.current = localData
  }, [localData])

  const debounceTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const executeSave = useCallback((): Promise<void> => {
    return new Promise((resolve, reject) => {
      setSaveState('saving')
      patchMutation.mutate(
        { ...localDataRef.current, id: activity.id },
        {
          onSuccess: (updated) => {
            // If the server has set submittedAt from another session, the query
            // cache is already updated by usePatchActivity's onSuccess.
            setSaveState('idle')
            // Keep localData status in sync if status PATCH raced and won.
            if (updated.status !== localDataRef.current.status) {
              setLocalData((prev) => ({ ...prev, status: updated.status }))
            }
            resolve()
          },
          onError: (err) => {
            const apiErr = err as Error & { status?: number; code?: string }
            if (apiErr.status === 404) {
              void navigate({ to: '/' })
            }
            setSaveState('error')
            reject(err)
          },
        },
      )
    })
  }, [activity.id, navigate, patchMutation])

  const scheduleSave = useCallback(() => {
    if (debounceTimerRef.current) clearTimeout(debounceTimerRef.current)
    debounceTimerRef.current = setTimeout(() => {
      void executeSave()
    }, DEBOUNCE_MS)
  }, [executeSave])

  const flushSave = useCallback((): Promise<void> => {
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current)
      debounceTimerRef.current = null
    }
    if (saveState === 'dirty') {
      return executeSave()
    }
    return Promise.resolve()
  }, [saveState, executeSave])

  const handleFieldChange = useCallback(
    (key: string, value: unknown) => {
      setLocalData((prev) => {
        // Top-level fields vs nested fields
        if (['activityType', 'status', 'dueDate', 'duration', 'routing', 'targetId', 'jointVisitUserId'].includes(key)) {
          return { ...prev, [key]: value }
        }
        return { ...prev, fields: { ...(prev.fields ?? {}), [key]: value } }
      })
      setSaveState('dirty')
      scheduleSave()
    },
    [scheduleSave],
  )

  const handleFieldBlur = useCallback(() => {
    if (saveState === 'dirty') {
      void flushSave()
    }
  }, [saveState, flushSave])

  const handleStatusChange = useCallback(
    (newStatus: string) => {
      // Status changes fire immediately via PATCH /status — separate from full update.
      statusMutation.mutate(
        { id: activity.id, status: newStatus },
        {
          onSuccess: (updated) => {
            // Keep localData in sync so a pending full-update doesn't overwrite.
            setLocalData((prev) => ({ ...prev, status: updated.status }))
          },
        },
      )
    },
    [activity.id, statusMutation],
  )

  const handleSubmit = useCallback(async (): Promise<void> => {
    if (saveState === 'error') return // blocked — caller shows retry message
    // Pre-flight: flush any pending save.
    await flushSave()
    setFieldErrors([])
    return new Promise((resolve, reject) => {
      submitMutation.mutate(activity.id, {
        onSuccess: () => resolve(),
        onError: (err) => {
          const apiErr = err as Error & { status?: number }
          if (apiErr.status === 422) {
            try {
              const body = JSON.parse(apiErr.message) as { fields?: ValidationFieldError[] }
              if (body.fields) setFieldErrors(body.fields)
            } catch {
              // non-JSON error body
            }
          }
          reject(err)
        },
      })
    })
  }, [activity.id, flushSave, saveState, submitMutation])

  const retrySave = useCallback(() => {
    void executeSave()
  }, [executeSave])

  return {
    localData,
    saveState,
    fieldErrors,
    handleFieldChange,
    handleFieldBlur,
    handleStatusChange,
    handleSubmit,
    retrySave,
    isSubmitting: submitMutation.isPending,
  }
}
