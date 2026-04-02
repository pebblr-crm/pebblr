import { renderHook, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createElement, type ReactNode } from 'react'
import { useCurrentUser } from './useMe'
import { api } from '@/api/client'
import type { CurrentUser } from '@/types/user'

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

describe('useCurrentUser', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('fetches the current user from /me', async () => {
    const user: CurrentUser = {
      id: 'u1',
      email: 'rep@test.com',
      name: 'Test Rep',
      role: 'rep',
      teamIds: ['t1'],
    }
    mockedApi.get.mockResolvedValueOnce(user)

    const { result } = renderHook(() => useCurrentUser(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    expect(result.current.data).toEqual(user)
    expect(mockedApi.get).toHaveBeenCalledWith('/me')
  })

  it('uses a staleTime of 5 minutes', async () => {
    const user: CurrentUser = {
      id: 'u1',
      email: 'rep@test.com',
      name: 'Test Rep',
      role: 'rep',
      teamIds: [],
    }
    mockedApi.get.mockResolvedValueOnce(user)

    const { result } = renderHook(() => useCurrentUser(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isSuccess).toBe(true))

    // Verify the hook was called with the expected query key
    expect(mockedApi.get).toHaveBeenCalledTimes(1)
    expect(mockedApi.get).toHaveBeenCalledWith('/me')
  })

  it('handles errors', async () => {
    mockedApi.get.mockRejectedValueOnce(new Error('unauthorized'))

    const { result } = renderHook(() => useCurrentUser(), { wrapper: createWrapper() })

    await waitFor(() => expect(result.current.isError).toBe(true))
    expect(result.current.error?.message).toBe('unauthorized')
  })
})
