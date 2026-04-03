import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createElement, type ReactNode } from 'react'
import { dashboardKeys, useActivityStats, useCoverage, useFrequency, useRecoveryBalance } from './useDashboard'
import { api } from '@/api/client'

vi.mock('@/api/client', () => ({
  api: { get: vi.fn() },
}))

const mockedApi = vi.mocked(api)

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children)
}

describe('dashboardKeys', () => {
  it('builds correct key hierarchy', () => {
    expect(dashboardKeys.all).toEqual(['dashboard'])
    expect(dashboardKeys.activities({})).toEqual(['dashboard', 'activities', {}])
    expect(dashboardKeys.coverage({ userId: 'u1' })).toEqual(['dashboard', 'coverage', { userId: 'u1' }])
  })
})

describe('useActivityStats', () => {
  beforeEach(() => vi.clearAllMocks())

  it('fetches activity stats', async () => {
    const fakeStats = { total: 10, byStatus: { realizat: 5 }, byCategory: {} }
    mockedApi.get.mockResolvedValueOnce(fakeStats)
    const { result } = renderHook(() => useActivityStats(), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(fakeStats)
  })
})

describe('useCoverage', () => {
  beforeEach(() => vi.clearAllMocks())

  it('fetches coverage data', async () => {
    const fakeCoverage = { totalTargets: 100, visitedTargets: 75, percentage: 75 }
    mockedApi.get.mockResolvedValueOnce(fakeCoverage)
    const { result } = renderHook(() => useCoverage(), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(fakeCoverage)
  })
})

describe('useFrequency', () => {
  beforeEach(() => vi.clearAllMocks())

  it('fetches frequency data', async () => {
    const fakeFrequency = { items: [{ classification: 'A', targetCount: 10, totalVisits: 5, required: 4, compliance: 125 }] }
    mockedApi.get.mockResolvedValueOnce(fakeFrequency)
    const { result } = renderHook(() => useFrequency(), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(fakeFrequency)
  })
})

describe('useRecoveryBalance', () => {
  beforeEach(() => vi.clearAllMocks())

  it('fetches recovery balance', async () => {
    const fakeRecovery = { earned: 5, taken: 2, balance: 3, intervals: [] }
    mockedApi.get.mockResolvedValueOnce(fakeRecovery)
    const { result } = renderHook(() => useRecoveryBalance(), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(fakeRecovery)
  })
})
