import { ApiError } from '@/types/api'

describe('ApiError', () => {
  it('has correct name', () => {
    const err = new ApiError('NOT_FOUND', 'not found', 404)
    expect(err.name).toBe('ApiError')
  })

  it('stores code and status', () => {
    const err = new ApiError('FORBIDDEN', 'access denied', 403)
    expect(err.code).toBe('FORBIDDEN')
    expect(err.status).toBe(403)
    expect(err.message).toBe('access denied')
  })

  it('stores field errors', () => {
    const fields = [{ field: 'name', message: 'required' }]
    const err = new ApiError('VALIDATION', 'invalid', 422, fields)
    expect(err.fields).toEqual(fields)
  })

  it('extends Error', () => {
    const err = new ApiError('ERR', 'msg', 500)
    expect(err).toBeInstanceOf(Error)
  })
})
