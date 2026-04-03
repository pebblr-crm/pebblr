import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createElement, type ReactNode } from 'react'
import { auditKeys, useAuditLog, useUpdateAuditStatus } from './useAudit'
import { api } from '@/api/client'

vi.mock('@/api/client', () => ({
  api: { get: vi.fn(), patch: vi.fn() },
}))

const mockedApi = vi.mocked(api)

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children)
}

describe('auditKeys', () => {
  it('builds correct key hierarchy', () => {
    expect(auditKeys.all).toEqual(['audit'])
    expect(auditKeys.lists()).toEqual(['audit', 'list'])
    expect(auditKeys.list({ page: 1 })).toEqual(['audit', 'list', { page: 1 }])
  })
})

describe('useAuditLog', () => {
  beforeEach(() => vi.clearAllMocks())

  it('fetches audit entries', async () => {
    const fakeData = { items: [{ id: 'a1', entityType: 'activity', eventType: 'create', actorId: 'u1', status: 'pending', createdAt: '2026-01-01T00:00:00Z' }], total: 1, page: 1, limit: 20 }
    mockedApi.get.mockResolvedValueOnce(fakeData)
    const { result } = renderHook(() => useAuditLog({ limit: 20 }), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(fakeData)
  })
})

describe('useUpdateAuditStatus', () => {
  beforeEach(() => vi.clearAllMocks())

  it('patches audit status', async () => {
    mockedApi.patch.mockResolvedValueOnce(undefined)
    const { result } = renderHook(() => useUpdateAuditStatus(), { wrapper: createWrapper() })
    result.current.mutate({ id: 'a1', status: 'accepted' })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(mockedApi.patch).toHaveBeenCalledWith('/audit/a1/status', { status: 'accepted' })
  })
})
