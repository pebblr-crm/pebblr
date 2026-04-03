import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createElement, type ReactNode } from 'react'
import { targetKeys, useTargets, useTarget, useTargetVisitStatus } from './useTargets'
import { api } from '@/api/client'

vi.mock('@/api/client', () => ({
  api: { get: vi.fn(), post: vi.fn() },
}))

const mockedApi = vi.mocked(api)

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children)
}

describe('targetKeys', () => {
  it('builds correct key hierarchy', () => {
    expect(targetKeys.all).toEqual(['targets'])
    expect(targetKeys.lists()).toEqual(['targets', 'list'])
    expect(targetKeys.detail('t1')).toEqual(['targets', 'detail', 't1'])
    expect(targetKeys.visitStatus()).toEqual(['targets', 'visit-status'])
  })
})

describe('useTargets', () => {
  beforeEach(() => vi.clearAllMocks())

  it('fetches targets list', async () => {
    const fakeData = { items: [{ id: 't1', name: 'Pharmacy A' }], total: 1, page: 1, limit: 20 }
    mockedApi.get.mockResolvedValueOnce(fakeData)
    const { result } = renderHook(() => useTargets({ limit: 20 }), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(fakeData)
  })

  it('handles fetch error', async () => {
    mockedApi.get.mockRejectedValueOnce(new Error('network'))
    const { result } = renderHook(() => useTargets(), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isError).toBe(true))
  })
})

describe('useTarget', () => {
  beforeEach(() => vi.clearAllMocks())

  it('fetches a single target and unwraps', async () => {
    const fakeTarget = { id: 't1', name: 'Pharmacy A' }
    mockedApi.get.mockResolvedValueOnce({ target: fakeTarget })
    const { result } = renderHook(() => useTarget('t1'), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(fakeTarget)
  })

  it('does not fetch when id is empty', () => {
    const { result } = renderHook(() => useTarget(''), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe('idle')
  })
})

describe('useTargetVisitStatus', () => {
  beforeEach(() => vi.clearAllMocks())

  it('fetches visit status', async () => {
    const fakeData = { items: [{ targetId: 't1', lastVisitDate: '2026-03-20' }] }
    mockedApi.get.mockResolvedValueOnce(fakeData)
    const { result } = renderHook(() => useTargetVisitStatus(), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(fakeData)
  })
})
