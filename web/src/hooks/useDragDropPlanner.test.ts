import { describe, it, expect, vi } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useDragDropPlanner } from './useDragDropPlanner'
import type { Activity } from '@/types/activity'

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function makeActivity(overrides: Partial<Activity> = {}): Activity {
  return {
    id: overrides.id ?? crypto.randomUUID(),
    activityType: 'visit',
    status: 'planned',
    dueDate: '2026-04-10T00:00:00Z',
    duration: '30m',
    fields: {},
    creatorId: 'user-1',
    createdAt: '2026-04-01T00:00:00Z',
    updatedAt: '2026-04-01T00:00:00Z',
    ...overrides,
  }
}

function setup(overrides: {
  activities?: readonly Activity[]
  selectedTargetIds?: Set<string>
} = {}) {
  const clearSelection = vi.fn()
  const patchActivity = { mutate: vi.fn() }
  const showToast = vi.fn()

  const hook = renderHook(() =>
    useDragDropPlanner({
      activities: overrides.activities ?? [],
      selectedTargetIds: overrides.selectedTargetIds ?? new Set(),
      clearSelection,
      patchActivity,
      showToast,
    }),
  )

  return { hook, clearSelection, patchActivity, showToast }
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

describe('useDragDropPlanner', () => {
  // 1. Initial state
  it('returns null drag states, isDragging false, and empty dayAssignments', () => {
    const { hook } = setup()
    const s = hook.result.current

    expect(s.dragTargetId).toBeNull()
    expect(s.dragActivityId).toBeNull()
    expect(s.dragPending).toBeNull()
    expect(s.isDragging).toBe(false)
    expect(s.dayAssignments).toEqual({})
  })

  // 2. setDragTargetId sets isDragging
  it('sets isDragging to true when dragTargetId is set', () => {
    const { hook } = setup()

    act(() => hook.result.current.setDragTargetId('t-1'))

    expect(hook.result.current.dragTargetId).toBe('t-1')
    expect(hook.result.current.isDragging).toBe(true)
  })

  // 3. setDragActivityId sets isDragging
  it('sets isDragging to true when dragActivityId is set', () => {
    const { hook } = setup()

    act(() => hook.result.current.setDragActivityId('a-1'))

    expect(hook.result.current.dragActivityId).toBe('a-1')
    expect(hook.result.current.isDragging).toBe(true)
  })

  // 4. setDragPending sets isDragging
  it('sets isDragging to true when dragPending is set', () => {
    const { hook } = setup()

    act(() =>
      hook.result.current.setDragPending({ sourceDate: '2026-04-10', targetId: 't-1' }),
    )

    expect(hook.result.current.dragPending).toEqual({
      sourceDate: '2026-04-10',
      targetId: 't-1',
    })
    expect(hook.result.current.isDragging).toBe(true)
  })

  // 5. handleDrop with dragTargetId adds target to dayAssignments
  it('adds target to dayAssignments when dropping a dragTargetId', () => {
    const { hook, clearSelection } = setup()

    act(() => hook.result.current.setDragTargetId('t-1'))
    act(() => hook.result.current.handleDrop('2026-04-11'))

    expect(hook.result.current.dayAssignments).toEqual({ '2026-04-11': ['t-1'] })
    expect(hook.result.current.dragTargetId).toBeNull()
    expect(clearSelection).toHaveBeenCalled()
  })

  // 6. handleDrop with dragTargetId + selectedTargetIds adds all
  it('adds dragTargetId and selectedTargetIds together', () => {
    const { hook } = setup({
      selectedTargetIds: new Set(['t-2', 't-3']),
    })

    act(() => hook.result.current.setDragTargetId('t-1'))
    act(() => hook.result.current.handleDrop('2026-04-11'))

    const assigned = hook.result.current.dayAssignments['2026-04-11']!
    expect(assigned).toHaveLength(3)
    expect(assigned).toContain('t-1')
    expect(assigned).toContain('t-2')
    expect(assigned).toContain('t-3')
  })

  // 7. handleDrop with duplicate target on same day (pending duplicate)
  it('shows warning toast when target is already pending on the day', () => {
    const { hook, showToast } = setup()

    // First drop adds t-1 to the day
    act(() => hook.result.current.setDragTargetId('t-1'))
    act(() => hook.result.current.handleDrop('2026-04-11'))

    // Second drop of same target
    act(() => hook.result.current.setDragTargetId('t-1'))
    act(() => hook.result.current.handleDrop('2026-04-11'))

    // Should still only have one entry
    expect(hook.result.current.dayAssignments['2026-04-11']).toEqual(['t-1'])
    expect(showToast).toHaveBeenCalledWith(
      expect.stringContaining('already on this day'),
      'warning',
    )
  })

  // 8. handleDrop with dragActivityId calls patchActivity.mutate
  it('calls patchActivity.mutate and clears dragActivityId on drop', () => {
    const { hook, patchActivity } = setup()

    act(() => hook.result.current.setDragActivityId('a-1'))
    act(() => hook.result.current.handleDrop('2026-04-12'))

    expect(patchActivity.mutate).toHaveBeenCalledWith({ id: 'a-1', dueDate: '2026-04-12' })
    expect(hook.result.current.dragActivityId).toBeNull()
  })

  // 9. handleDrop with dragPending moves pending from source to dest
  it('moves pending target from source day to destination day', () => {
    const { hook } = setup()

    // Seed source day with a target
    act(() => hook.result.current.setDragTargetId('t-1'))
    act(() => hook.result.current.handleDrop('2026-04-10'))
    expect(hook.result.current.dayAssignments['2026-04-10']).toEqual(['t-1'])

    // Now drag it as pending from 04-10 to 04-11
    act(() =>
      hook.result.current.setDragPending({ sourceDate: '2026-04-10', targetId: 't-1' }),
    )
    act(() => hook.result.current.handleDrop('2026-04-11'))

    expect(hook.result.current.dayAssignments['2026-04-11']).toEqual(['t-1'])
    // Source day should be cleaned up (key deleted since no remaining targets)
    expect(hook.result.current.dayAssignments['2026-04-10']).toBeUndefined()
    expect(hook.result.current.dragPending).toBeNull()
  })

  // 10. handleDrop with dragPending to same day cancels
  it('cancels when pending is dropped on the same source day', () => {
    const { hook } = setup()

    // Seed
    act(() => hook.result.current.setDragTargetId('t-1'))
    act(() => hook.result.current.handleDrop('2026-04-10'))

    act(() =>
      hook.result.current.setDragPending({ sourceDate: '2026-04-10', targetId: 't-1' }),
    )
    act(() => hook.result.current.handleDrop('2026-04-10'))

    // Nothing should change, target stays on source day
    expect(hook.result.current.dayAssignments['2026-04-10']).toEqual(['t-1'])
    expect(hook.result.current.dragPending).toBeNull()
  })

  // 11. handleDrop with dragPending when target already exists in activities on day
  it('shows toast when pending target already has an activity on the dest day', () => {
    const existingActivity = makeActivity({
      dueDate: '2026-04-11T09:00:00Z',
      targetId: 't-1',
    })
    const { hook, showToast } = setup({ activities: [existingActivity] })

    // Seed source day
    act(() => hook.result.current.setDragTargetId('t-1'))
    act(() => hook.result.current.handleDrop('2026-04-10'))

    // Try to move pending to day that already has the target in activities
    act(() =>
      hook.result.current.setDragPending({ sourceDate: '2026-04-10', targetId: 't-1' }),
    )
    act(() => hook.result.current.handleDrop('2026-04-11'))

    expect(showToast).toHaveBeenCalledWith(
      'Target already has a visit on this day',
      'warning',
    )
    // Target should remain on source day (not moved)
    expect(hook.result.current.dayAssignments['2026-04-10']).toEqual(['t-1'])
  })

  // 12. removeFromDay removes target
  it('removes a target from dayAssignments', () => {
    const { hook } = setup()

    // Add two targets to the same day
    act(() => hook.result.current.setDragTargetId('t-1'))
    act(() => hook.result.current.handleDrop('2026-04-10'))
    act(() => hook.result.current.setDragTargetId('t-2'))
    act(() => hook.result.current.handleDrop('2026-04-10'))

    expect(hook.result.current.dayAssignments['2026-04-10']).toEqual(['t-1', 't-2'])

    act(() => hook.result.current.removeFromDay('2026-04-10', 't-1'))

    expect(hook.result.current.dayAssignments['2026-04-10']).toEqual(['t-2'])
  })

  // 13. removeFromDay deletes day key when last target removed
  it('deletes the day key when the last target is removed', () => {
    const { hook } = setup()

    act(() => hook.result.current.setDragTargetId('t-1'))
    act(() => hook.result.current.handleDrop('2026-04-10'))

    act(() => hook.result.current.removeFromDay('2026-04-10', 't-1'))

    expect(hook.result.current.dayAssignments).toEqual({})
    expect(hook.result.current.dayAssignments['2026-04-10']).toBeUndefined()
  })

  // 14. handleDrop with targets already in existing activities skips duplicates
  it('skips targets that already have activities on the day and shows toast', () => {
    const existingActivity = makeActivity({
      dueDate: '2026-04-11T09:00:00Z',
      targetId: 't-1',
    })
    const { hook, showToast } = setup({
      activities: [existingActivity],
      selectedTargetIds: new Set(['t-1', 't-2']),
    })

    act(() => hook.result.current.setDragTargetId('t-3'))
    act(() => hook.result.current.handleDrop('2026-04-11'))

    // t-1 should be skipped (already in activities), t-2 and t-3 should be added
    const assigned = hook.result.current.dayAssignments['2026-04-11']!
    expect(assigned).toHaveLength(2)
    expect(assigned).toContain('t-2')
    expect(assigned).toContain('t-3')
    expect(assigned).not.toContain('t-1')

    expect(showToast).toHaveBeenCalledWith(
      expect.stringContaining('1 already on this day'),
      'warning',
    )
  })

  // 15. handleDrop with all targets duplicate shows "all X targets are already on this day"
  it('shows "all targets are already on this day" when every target is a duplicate', () => {
    const activities = [
      makeActivity({ dueDate: '2026-04-11T09:00:00Z', targetId: 't-1' }),
      makeActivity({ dueDate: '2026-04-11T09:00:00Z', targetId: 't-2' }),
    ]
    const { hook, showToast } = setup({
      activities,
      selectedTargetIds: new Set(['t-2']),
    })

    act(() => hook.result.current.setDragTargetId('t-1'))
    act(() => hook.result.current.handleDrop('2026-04-11'))

    // No new assignments
    expect(hook.result.current.dayAssignments['2026-04-11']).toBeUndefined()

    expect(showToast).toHaveBeenCalledWith(
      expect.stringContaining('all 2 targets are already on this day'),
      'warning',
    )
  })

  it('shows singular message when single target is duplicate', () => {
    const activities = [
      makeActivity({ dueDate: '2026-04-11T09:00:00Z', targetId: 't-1' }),
    ]
    const { hook, showToast } = setup({ activities })

    act(() => hook.result.current.setDragTargetId('t-1'))
    act(() => hook.result.current.handleDrop('2026-04-11'))

    expect(showToast).toHaveBeenCalledWith(
      expect.stringContaining('target is already on this day'),
      'warning',
    )
  })
})
