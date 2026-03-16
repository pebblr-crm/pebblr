import { ApiError, ApiErrorResponse } from '@/types/api'

/**
 * Base URL for all API calls. Vite dev server proxies /api → Go backend.
 */
const API_BASE = '/api/v1'

/**
 * Retrieve the current bearer token. Auth service provides this;
 * the getter is a function so it is always fresh.
 */
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
  return new ApiError(code, message, response.status)
}

/**
 * Core fetch wrapper. All API calls go through here.
 * - Adds auth header
 * - Parses structured error responses
 * - Throws ApiError on non-2xx
 */
export async function apiFetch<T>(
  path: string,
  options?: RequestInit,
): Promise<T> {
  const url = `${API_BASE}${path}`
  const headers = buildHeaders(options?.headers)

  const response = await fetch(url, {
    ...options,
    headers,
  })

  if (!response.ok) {
    throw await parseError(response)
  }

  // 204 No Content
  if (response.status === 204) {
    return undefined as T
  }

  return response.json() as Promise<T>
}

export function apiGet<T>(path: string): Promise<T> {
  return apiFetch<T>(path, { method: 'GET' })
}

export function apiPost<T>(path: string, body: unknown): Promise<T> {
  return apiFetch<T>(path, {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function apiPatch<T>(path: string, body: unknown): Promise<T> {
  return apiFetch<T>(path, {
    method: 'PATCH',
    body: JSON.stringify(body),
  })
}

export function apiDelete(path: string): Promise<void> {
  return apiFetch<void>(path, { method: 'DELETE' })
}
