export interface ApiErrorDetail {
  code: string
  message: string
}

export interface ValidationFieldError {
  field: string
  message: string
}

export interface ApiErrorResponse {
  error: ApiErrorDetail
  fields?: ValidationFieldError[]
}

export class ApiError extends Error {
  public readonly fields?: ValidationFieldError[]

  constructor(
    public readonly code: string,
    message: string,
    public readonly status: number,
    fields?: ValidationFieldError[],
  ) {
    super(message)
    this.name = 'ApiError'
    this.fields = fields
  }
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  limit: number
}

export interface ListParams {
  page?: number
  limit?: number
}
