import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createElement, type ReactNode } from 'react'
import { useUsers, useUser, userKeys } from './useUsers'
import { api } from '@/api/client'
import type { User } from '@/types/user'

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

const fakeUser: User = {
  id: 'u1',
  name: 'Alice',
  displayName: 'Alice R.',
  email: 'alice@example.com',
  role: 'rep',
}

describe('userKeys', () => {
  it('builds all key', () => {
    expect(userKeys.all).toEqual(['users'])
  })

  it('builds detail key', () => {
    expect(userKeys.detail('u1')).toEqual(['users', 'u1'])
  })
})

describe('useUsers', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches user list from /users', async () => {
    const response = { items: [fakeUser], total: 1 }
    mockedApi.get.mockResolvedValueOnce(response)

    const { result } = renderHook(() => useUsers(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(response)
    expect(mockedApi.get).toHaveBeenCalledWith('/users')
  })

  it('handles errors', async () => {
    mockedApi.get.mockRejectedValueOnce(new Error('forbidden'))

    const { result } = renderHook(() => useUsers(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isError).toBe(true))
  })
})

describe('useUser', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('fetches a single user by id', async () => {
    mockedApi.get.mockResolvedValueOnce(fakeUser)

    const { result } = renderHook(() => useUser('u1'), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))
    expect(result.current.data).toEqual(fakeUser)
    expect(mockedApi.get).toHaveBeenCalledWith('/users/u1')
  })

  it('does not fetch when id is empty', () => {
    const { result } = renderHook(() => useUser(''), { wrapper: createWrapper() })

    // With enabled: !!id, the query should stay idle
    expect(result.current.fetchStatus).toBe('idle')
    expect(mockedApi.get).not.toHaveBeenCalled()
  })
})
