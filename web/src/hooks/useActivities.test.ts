import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createElement, type ReactNode } from 'react'
import {
  activityKeys,
  useActivities,
  useActivity,
  useCreateActivity,
  usePatchActivityStatus,
  useSubmitActivity,
  useBatchCreateActivities,
  usePatchActivity,
  useCloneWeek,
} from './useActivities'
import { api } from '@/api/client'
import type { Activity } from '@/types/activity'
import type { PaginatedResponse } from '@/types/api'

vi.mock('@/api/client', () => ({
  api: { get: vi.fn(), post: vi.fn(), patch: vi.fn() },
}))

const mockedApi = vi.mocked(api)

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children)
}

const fakeActivity: Activity = {
  id: 'a1',
  activityType: 'visit',
  status: 'planificat',
  dueDate: '2026-03-23',
  duration: '30m',
  fields: {},
  creatorId: 'u1',
  createdAt: '2026-03-23T00:00:00Z',
  updatedAt: '2026-03-23T00:00:00Z',
}

const fakeList: PaginatedResponse<Activity> = {
  items: [fakeActivity],
  total: 1,
  page: 1,
  limit: 20,
}

describe('activityKeys', () => {
  it('builds correct key hierarchy', () => {
    expect(activityKeys.all).toEqual(['activities'])
    expect(activityKeys.lists()).toEqual(['activities', 'list'])
    expect(activityKeys.list({ status: 'done' })).toEqual(['activities', 'list', { status: 'done' }])
    expect(activityKeys.detail('a1')).toEqual(['activities', 'detail', 'a1'])
  })
})

describe('useActivities', () => {
  beforeEach(() => vi.clearAllMocks())

  it('fetches activities list', async () => {
    mockedApi.get.mockResolvedValueOnce(fakeList)
    const { result } = renderHook(() => useActivities({ limit: 20 }), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(fakeList)
  })
})

describe('useActivity', () => {
  beforeEach(() => vi.clearAllMocks())

  it('fetches a single activity', async () => {
    mockedApi.get.mockResolvedValueOnce({ activity: fakeActivity })
    const { result } = renderHook(() => useActivity('a1'), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(fakeActivity)
  })

  it('does not fetch when id is empty', () => {
    const { result } = renderHook(() => useActivity(''), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe('idle')
  })
})

describe('useCreateActivity', () => {
  beforeEach(() => vi.clearAllMocks())

  it('posts to /activities', async () => {
    mockedApi.post.mockResolvedValueOnce(fakeActivity)
    const { result } = renderHook(() => useCreateActivity(), { wrapper: createWrapper() })
    result.current.mutate({ activityType: 'visit', status: 'planificat', dueDate: '2026-03-23', fields: {} })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mockedApi.post).toHaveBeenCalledWith('/activities', expect.any(Object))
  })
})

describe('usePatchActivityStatus', () => {
  beforeEach(() => vi.clearAllMocks())

  it('patches activity status', async () => {
    mockedApi.patch.mockResolvedValueOnce(undefined)
    const { result } = renderHook(() => usePatchActivityStatus(), { wrapper: createWrapper() })
    result.current.mutate({ id: 'a1', status: 'realizat' })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mockedApi.patch).toHaveBeenCalledWith('/activities/a1/status', { status: 'realizat', fields: undefined })
  })
})

describe('useSubmitActivity', () => {
  beforeEach(() => vi.clearAllMocks())

  it('posts to submit endpoint', async () => {
    mockedApi.post.mockResolvedValueOnce(undefined)
    const { result } = renderHook(() => useSubmitActivity(), { wrapper: createWrapper() })
    result.current.mutate('a1')
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mockedApi.post).toHaveBeenCalledWith('/activities/a1/submit', {})
  })
})

describe('useBatchCreateActivities', () => {
  beforeEach(() => vi.clearAllMocks())

  it('posts batch create', async () => {
    mockedApi.post.mockResolvedValueOnce({ created: [fakeActivity], errors: [] })
    const { result } = renderHook(() => useBatchCreateActivities(), { wrapper: createWrapper() })
    result.current.mutate([{ targetId: 't1', dueDate: '2026-03-23' }])
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mockedApi.post).toHaveBeenCalledWith('/activities/batch', { items: [{ targetId: 't1', dueDate: '2026-03-23' }] })
  })
})

describe('usePatchActivity', () => {
  beforeEach(() => vi.clearAllMocks())

  it('patches activity fields', async () => {
    mockedApi.patch.mockResolvedValueOnce(undefined)
    const { result } = renderHook(() => usePatchActivity(), { wrapper: createWrapper() })
    result.current.mutate({ id: 'a1', dueDate: '2026-03-25' })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mockedApi.patch).toHaveBeenCalledWith('/activities/a1', { dueDate: '2026-03-25' })
  })
})

describe('useCloneWeek', () => {
  beforeEach(() => vi.clearAllMocks())

  it('posts clone-week request', async () => {
    mockedApi.post.mockResolvedValueOnce(undefined)
    const { result } = renderHook(() => useCloneWeek(), { wrapper: createWrapper() })
    result.current.mutate({ sourceWeekStart: '2026-03-23', targetWeekStart: '2026-03-30' })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mockedApi.post).toHaveBeenCalledWith('/activities/clone-week', {
      sourceWeekStart: '2026-03-23',
      targetWeekStart: '2026-03-30',
    })
  })
})
