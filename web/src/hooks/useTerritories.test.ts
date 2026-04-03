import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createElement, type ReactNode } from 'react'
import { territoryKeys, useTerritories, useTerritory } from './useTerritories'
import { api } from '@/api/client'

vi.mock('@/api/client', () => ({
  api: { get: vi.fn(), post: vi.fn(), delete: vi.fn() },
}))

const mockedApi = vi.mocked(api)

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children)
}

describe('territoryKeys', () => {
  it('builds correct key hierarchy', () => {
    expect(territoryKeys.all).toEqual(['territories'])
    expect(territoryKeys.lists()).toEqual(['territories', 'list'])
    expect(territoryKeys.detail('t1')).toEqual(['territories', 'detail', 't1'])
  })
})

describe('useTerritories', () => {
  beforeEach(() => vi.clearAllMocks())

  it('fetches territories list', async () => {
    const fakeData = { items: [{ id: 't1', name: 'North' }], total: 1 }
    mockedApi.get.mockResolvedValueOnce(fakeData)
    const { result } = renderHook(() => useTerritories(), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(fakeData)
    expect(mockedApi.get).toHaveBeenCalledWith('/territories')
  })

  it('handles fetch error', async () => {
    mockedApi.get.mockRejectedValueOnce(new Error('network'))
    const { result } = renderHook(() => useTerritories(), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isError).toBe(true))
  })
})

describe('useTerritory', () => {
  beforeEach(() => vi.clearAllMocks())

  it('fetches a single territory', async () => {
    const fakeTerritory = { id: 't1', name: 'North', teamId: 'tm1' }
    mockedApi.get.mockResolvedValueOnce(fakeTerritory)
    const { result } = renderHook(() => useTerritory('t1'), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(fakeTerritory)
  })

  it('does not fetch when id is empty', () => {
    const { result } = renderHook(() => useTerritory(''), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe('idle')
  })
})
