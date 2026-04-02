import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createElement, type ReactNode } from 'react'
import { useConfig, configKeys } from './useConfig'
import { api } from '@/api/client'
import type { TenantConfig } from '@/types/config'

vi.mock('@/api/client', () => ({
  api: {
    get: vi.fn(),
  },
}))

const mockedApi = vi.mocked(api)

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children)
}

const fakeConfig: TenantConfig = {
  tenant: { name: 'DrMax', locale: 'ro' },
  accounts: { types: [] },
  activities: {
    statuses: [],
    status_transitions: {},
    durations: [],
    types: [],
    routing_options: [],
  },
  options: {},
  rules: {
    frequency: { A: 4, B: 2, C: 1 },
    max_activities_per_day: 10,
    default_visit_duration_minutes: { visit: 30 },
    visit_duration_step_minutes: 15,
  },
}

describe('configKeys', () => {
  it('has a stable all key', () => {
    expect(configKeys.all).toEqual(['config'])
  })
})

describe('useConfig', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches tenant config from /config', async () => {
    mockedApi.get.mockResolvedValueOnce(fakeConfig)

    const { result } = renderHook(() => useConfig(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(fakeConfig)
    expect(mockedApi.get).toHaveBeenCalledWith('/config')
  })

  it('calls the /config endpoint exactly once per mount', async () => {
    mockedApi.get.mockResolvedValueOnce(fakeConfig)

    const { result } = renderHook(() => useConfig(), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(mockedApi.get).toHaveBeenCalledTimes(1)
    expect(mockedApi.get).toHaveBeenCalledWith('/config')
  })

  it('handles fetch failure', async () => {
    mockedApi.get.mockRejectedValueOnce(new Error('network error'))

    const { result } = renderHook(() => useConfig(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isError).toBe(true))
    expect(result.current.error?.message).toBe('network error')
  })
})
