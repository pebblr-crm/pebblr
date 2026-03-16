/**
 * API error structure matching backend convention:
 * {"error": {"code": "NOT_FOUND", "message": "..."}}
 */
export interface ApiErrorDetail {
  code: string
  message: string
}

export interface ApiErrorResponse {
  error: ApiErrorDetail
}

export class ApiError extends Error {
  readonly code: string
  readonly status: number

  constructor(code: string, message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
  }
}

/**
 * Paginated list response wrapper for collection endpoints.
 */
export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  limit: number
}

/**
 * Common query params for list endpoints.
 */
export interface ListParams {
  page?: number
  limit?: number
}
