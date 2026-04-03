import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createElement, type ReactNode } from 'react'
import { teamKeys, useTeams, useTeam } from './useTeams'
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

describe('teamKeys', () => {
  it('builds correct key hierarchy', () => {
    expect(teamKeys.all).toEqual(['teams'])
    expect(teamKeys.lists()).toEqual(['teams', 'list'])
    expect(teamKeys.detail('t1')).toEqual(['teams', 'detail', 't1'])
  })
})

describe('useTeams', () => {
  beforeEach(() => vi.clearAllMocks())

  it('fetches teams list', async () => {
    const fakeData = { items: [{ id: 't1', name: 'Alpha', managerId: 'u1' }], total: 1 }
    mockedApi.get.mockResolvedValueOnce(fakeData)
    const { result } = renderHook(() => useTeams(), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(fakeData)
    expect(mockedApi.get).toHaveBeenCalledWith('/teams')
  })
})

describe('useTeam', () => {
  beforeEach(() => vi.clearAllMocks())

  it('fetches a single team', async () => {
    const fakeTeam = { team: { id: 't1', name: 'Alpha', managerId: 'u1' }, members: [] }
    mockedApi.get.mockResolvedValueOnce(fakeTeam)
    const { result } = renderHook(() => useTeam('t1'), { wrapper: createWrapper() })
    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(fakeTeam)
  })

  it('does not fetch when id is empty', () => {
    const { result } = renderHook(() => useTeam(''), { wrapper: createWrapper() })
    expect(result.current.fetchStatus).toBe('idle')
  })
})
