import { api, setTokenProvider } from './client'
import { ApiError } from '@/types/api'

const originalFetch = globalThis.fetch

beforeEach(() => {
  vi.restoreAllMocks()
  setTokenProvider(null as unknown as () => string | null)
})

afterAll(() => {
  globalThis.fetch = originalFetch
})

function mockFetch(response: Partial<Response>) {
  const fn = vi.fn().mockResolvedValue({
    ok: true,
    status: 200,
    json: () => Promise.resolve({}),
    ...response,
  })
  globalThis.fetch = fn
  return fn
}

describe('api.get', () => {
  it('sends GET to /api/v1 prefixed path', async () => {
    const fetchMock = mockFetch({ ok: true, status: 200, json: () => Promise.resolve({ items: [] }) })
    const result = await api.get('/targets')
    expect(fetchMock).toHaveBeenCalledWith('/api/v1/targets', expect.objectContaining({ headers: expect.any(Headers) }))
    expect(result).toEqual({ items: [] })
  })

  it('includes Authorization header when token provider is set', async () => {
    setTokenProvider(() => 'test-token-123')
    const fetchMock = mockFetch({ ok: true, status: 200, json: () => Promise.resolve({}) })
    await api.get('/targets')
    const headers: Headers = fetchMock.mock.calls[0][1].headers
    expect(headers.get('Authorization')).toBe('Bearer test-token-123')
  })

  it('omits Authorization header when token is null', async () => {
    setTokenProvider(() => null)
    const fetchMock = mockFetch({ ok: true, status: 200, json: () => Promise.resolve({}) })
    await api.get('/targets')
    const headers: Headers = fetchMock.mock.calls[0][1].headers
    expect(headers.get('Authorization')).toBeNull()
  })
})

describe('api.post', () => {
  it('sends POST with JSON body', async () => {
    const fetchMock = mockFetch({ ok: true, status: 200, json: () => Promise.resolve({ id: '1' }) })
    const result = await api.post('/targets', { name: 'Test' })
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/targets',
      expect.objectContaining({ method: 'POST', body: '{"name":"Test"}' }),
    )
    expect(result).toEqual({ id: '1' })
  })
})

describe('api.patch', () => {
  it('sends PATCH with JSON body', async () => {
    const fetchMock = mockFetch({ ok: true, status: 200, json: () => Promise.resolve({}) })
    await api.patch('/targets/1', { name: 'Updated' })
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/targets/1',
      expect.objectContaining({ method: 'PATCH', body: '{"name":"Updated"}' }),
    )
  })
})

describe('api.put', () => {
  it('sends PUT with JSON body', async () => {
    const fetchMock = mockFetch({ ok: true, status: 200, json: () => Promise.resolve({}) })
    await api.put('/targets/1', { name: 'Replaced' })
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/targets/1',
      expect.objectContaining({ method: 'PUT', body: '{"name":"Replaced"}' }),
    )
  })
})

describe('api.delete', () => {
  it('sends DELETE and returns undefined for 204', async () => {
    mockFetch({ ok: true, status: 204, json: () => Promise.reject(new Error('no body')) })
    const result = await api.delete('/targets/1')
    expect(result).toBeUndefined()
  })
})

describe('error handling', () => {
  it('throws ApiError with structured error from response', async () => {
    mockFetch({
      ok: false,
      status: 404,
      json: () => Promise.resolve({ error: { code: 'NOT_FOUND', message: 'lead not found' } }),
    })

    try {
      await api.get('/leads/999')
      expect.fail('should have thrown')
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError)
      const apiErr = err as ApiError
      expect(apiErr.code).toBe('NOT_FOUND')
      expect(apiErr.message).toBe('lead not found')
      expect(apiErr.status).toBe(404)
    }
  })

  it('falls back to HTTP status when response is not JSON', async () => {
    mockFetch({
      ok: false,
      status: 500,
      json: () => Promise.reject(new Error('not json')),
    })

    try {
      await api.get('/broken')
      expect.fail('should have thrown')
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError)
      const apiErr = err as ApiError
      expect(apiErr.code).toBe('UNKNOWN')
      expect(apiErr.message).toBe('HTTP 500')
    }
  })

  it('includes field errors when present', async () => {
    const fields = [{ field: 'name', message: 'required' }]
    mockFetch({
      ok: false,
      status: 422,
      json: () => Promise.resolve({ error: { code: 'VALIDATION', message: 'invalid input' }, fields }),
    })

    try {
      await api.post('/targets', {})
      expect.fail('should have thrown')
    } catch (err) {
      const apiErr = err as ApiError
      expect(apiErr.fields).toEqual(fields)
    }
  })
})

describe('ApiError', () => {
  it('has correct name', () => {
    const err = new ApiError('NOT_FOUND', 'not found', 404)
    expect(err.name).toBe('ApiError')
  })

  it('extends Error', () => {
    const err = new ApiError('ERR', 'msg', 500)
    expect(err).toBeInstanceOf(Error)
  })
})
