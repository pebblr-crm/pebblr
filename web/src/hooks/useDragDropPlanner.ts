import { useState, useCallback } from 'react'
import type { Activity } from '@/types/activity'

export interface DragPending {
  sourceDate: string
  targetId: string
}

export interface DragDropState {
  dragTargetId: string | null
  dragActivityId: string | null
  dragPending: DragPending | null
  isDragging: boolean
  dayAssignments: Record<string, string[]>
}

export interface DragDropHandlers {
  setDragTargetId: (id: string | null) => void
  setDragActivityId: (id: string | null) => void
  setDragPending: (p: DragPending | null) => void
  handleDrop: (dateStr: string) => void
  removeFromDay: (dateStr: string, targetId: string) => void
  setDayAssignments: React.Dispatch<React.SetStateAction<Record<string, string[]>>>
}

interface UseDragDropPlannerOptions {
  activities: readonly Activity[]
  selectedTargetIds: Set<string>
  clearSelection: () => void
  patchActivity: { mutate: (args: { id: string; dueDate: string }) => void }
  showToast: (message: string, variant: 'info' | 'warning' | 'error') => void
}

export function useDragDropPlanner({
  activities,
  selectedTargetIds,
  clearSelection,
  patchActivity,
  showToast,
}: UseDragDropPlannerOptions): DragDropState & DragDropHandlers {
  const [dragTargetId, setDragTargetId] = useState<string | null>(null)
  const [dragActivityId, setDragActivityId] = useState<string | null>(null)
  const [dragPending, setDragPending] = useState<DragPending | null>(null)
  const [dayAssignments, setDayAssignments] = useState<Record<string, string[]>>({})

  const isDragging = dragTargetId != null || dragActivityId != null || dragPending != null

  const handleDropPending = useCallback((dateStr: string) => {
    if (!dragPending) return
    if (dragPending.sourceDate === dateStr) {
      setDragPending(null)
      return
    }
    const existingOnDay = activities.some(
      (a) => a.dueDate.slice(0, 10) === dateStr && a.targetId === dragPending.targetId,
    )
    if (existingOnDay) {
      showToast('Target already has a visit on this day', 'warning')
      setDragPending(null)
      return
    }
    setDayAssignments((prev) => {
      const srcArr = (prev[dragPending.sourceDate] ?? []).filter((id) => id !== dragPending.targetId)
      const dstArr = prev[dateStr] ?? []
      if (dstArr.includes(dragPending.targetId)) {
        showToast('Target is already on this day', 'warning')
        return prev
      }
      const next = { ...prev, [dateStr]: [...dstArr, dragPending.targetId] }
      if (srcArr.length === 0) delete next[dragPending.sourceDate]
      else next[dragPending.sourceDate] = srcArr
      return next
    })
    setDragPending(null)
  }, [dragPending, activities, showToast])

  const handleDropTargets = useCallback((dateStr: string) => {
    const ids = new Set(selectedTargetIds)
    if (dragTargetId) ids.add(dragTargetId)
    if (ids.size === 0) return

    const pendingOnDay = new Set(dayAssignments[dateStr] ?? [])
    const existingOnDay = new Set(
      activities.filter((a) => a.dueDate.slice(0, 10) === dateStr && a.targetId).map((a) => a.targetId!),
    )
    const toAdd = Array.from(ids).filter((id) => !pendingOnDay.has(id) && !existingOnDay.has(id))
    const dupCount = ids.size - toAdd.length

    if (toAdd.length > 0) {
      setDayAssignments((prev) => ({
        ...prev,
        [dateStr]: [...(prev[dateStr] ?? []), ...toAdd],
      }))
    }
    if (dupCount > 0) {
      let msg: string
      if (dupCount === ids.size) {
        const subject = dupCount === 1 ? 'target is' : `all ${dupCount} targets are`
        msg = `Already scheduled — ${subject} already on this day`
      } else {
        msg = `${toAdd.length} added, ${dupCount} already on this day`
      }
      showToast(msg, 'warning')
    }
    setDragTargetId(null)
    clearSelection()
  }, [dragTargetId, selectedTargetIds, dayAssignments, activities, showToast, clearSelection])

  const handleDrop = useCallback((dateStr: string) => {
    if (dragPending) {
      handleDropPending(dateStr)
      return
    }
    if (dragActivityId) {
      patchActivity.mutate({ id: dragActivityId, dueDate: dateStr })
      setDragActivityId(null)
      return
    }
    if (dragTargetId || selectedTargetIds.size > 0) {
      handleDropTargets(dateStr)
    }
  }, [dragPending, dragActivityId, dragTargetId, selectedTargetIds, handleDropPending, handleDropTargets, patchActivity])

  const removeFromDay = useCallback((dateStr: string, targetId: string) => {
    setDayAssignments((prev) => {
      const arr = (prev[dateStr] ?? []).filter((id) => id !== targetId)
      const next = { ...prev }
      if (arr.length === 0) delete next[dateStr]
      else next[dateStr] = arr
      return next
    })
  }, [])

  return {
    dragTargetId,
    dragActivityId,
    dragPending,
    isDragging,
    dayAssignments,
    setDragTargetId,
    setDragActivityId,
    setDragPending,
    handleDrop,
    removeFromDay,
    setDayAssignments,
  }
}
