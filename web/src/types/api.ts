

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
  constructor(
    public readonly code: string,
    message: string,
    public readonly status: number,
  ) {
    super(message)
    this.name = 'ApiError'
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
