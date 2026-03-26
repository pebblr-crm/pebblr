import { renderHook, act } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useInlineActivityEditor } from './useInlineActivityEditor'
import type { Activity } from '../types/activity'

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
}))

const mockPatchMutate = vi.fn()
const mockStatusMutate = vi.fn()
const mockSubmitMutate = vi.fn()

vi.mock('../services/activities', () => ({
  usePatchActivity: () => ({ mutate: mockPatchMutate, isPending: false }),
  usePatchActivityStatus: () => ({ mutate: mockStatusMutate, isPending: false }),
  useSubmitActivity: () => ({ mutate: mockSubmitMutate, isPending: false }),
}))

const baseActivity: Activity = {
  id: 'act-1',
  activityType: 'visit',
  status: 'planificat',
  dueDate: '2026-03-23T00:00:00Z',
  duration: 'full_day',
  fields: { notes: 'initial' },
  creatorId: 'user-1',
  createdAt: '2026-03-23T00:00:00Z',
  updatedAt: '2026-03-23T00:00:00Z',
}

describe('useInlineActivityEditor', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('initialises localData from activity', () => {
    const { result } = renderHook(() => useInlineActivityEditor(baseActivity))
    expect(result.current.localData.status).toBe('planificat')
    expect(result.current.localData.duration).toBe('full_day')
    expect(result.current.saveState).toBe('idle')
  })

  it('sets saveState to dirty and updates localData on field change', () => {
    const { result } = renderHook(() => useInlineActivityEditor(baseActivity))
    act(() => {
      result.current.handleFieldChange('duration', 'half_day')
    })
    expect(result.current.saveState).toBe('dirty')
    expect(result.current.localData.duration).toBe('half_day')
  })

  it('updates nested fields via handleFieldChange', () => {
    const { result } = renderHook(() => useInlineActivityEditor(baseActivity))
    act(() => {
      result.current.handleFieldChange('notes', 'updated notes')
    })
    expect(result.current.localData.fields?.notes).toBe('updated notes')
  })

  it('calls statusMutation on handleStatusChange', () => {
    const { result } = renderHook(() => useInlineActivityEditor(baseActivity))
    act(() => {
      result.current.handleStatusChange('realizat')
    })
    expect(mockStatusMutate).toHaveBeenCalledWith(
      { id: 'act-1', status: 'realizat' },
      expect.any(Object),
    )
  })

  it('calls patchMutation on retrySave', () => {
    const { result } = renderHook(() => useInlineActivityEditor(baseActivity))
    act(() => {
      result.current.retrySave()
    })
    expect(mockPatchMutate).toHaveBeenCalledWith(
      expect.objectContaining({ id: 'act-1' }),
      expect.any(Object),
    )
  })
})
