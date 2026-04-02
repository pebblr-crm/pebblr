import { ApiError, type ApiErrorResponse } from '@/types/api'

const API_BASE = '/api/v1'

let getAccessToken: (() => string | null) | null = null

export function setTokenProvider(provider: () => string | null): void {
  getAccessToken = provider
}

function buildHeaders(extra?: HeadersInit): Headers {
  const headers = new Headers(extra)
  headers.set('Content-Type', 'application/json')
  const token = getAccessToken?.()
  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  return headers
}

async function parseError(response: Response): Promise<ApiError> {
  let body: ApiErrorResponse | null = null
  try {
    body = (await response.json()) as ApiErrorResponse
  } catch {
    // response not JSON
  }
  const code = body?.error?.code ?? 'UNKNOWN'
  const message = body?.error?.message ?? `HTTP ${response.status}`
  return new ApiError(code, message, response.status, body?.fields)
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE}${path}`
  const headers = buildHeaders(options?.headers)
  const response = await fetch(url, { ...options, headers })

  if (!response.ok) {
    throw await parseError(response)
  }
  if (response.status === 204) {
    // Callers that expect void should type the generic as `void`.
    // The cast is safe only when T is void; callers returning data
    // from a 204 have a logic bug regardless.
    return undefined as unknown as T
  }
  return response.json() as Promise<T>
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body: unknown) =>
    request<T>(path, { method: 'POST', body: JSON.stringify(body) }),
  patch: <T>(path: string, body: unknown) =>
    request<T>(path, { method: 'PATCH', body: JSON.stringify(body) }),
  put: <T>(path: string, body: unknown) =>
    request<T>(path, { method: 'PUT', body: JSON.stringify(body) }),
  delete: <T>(path: string) => request<T>(path, { method: 'DELETE' }),
}
