import { renderHook } from '@testing-library/react'
import { useAuth } from './context'

describe('useAuth', () => {
  it('throws when used outside AuthProvider', () => {
    // Suppress console.error for expected error
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {})
    expect(() => renderHook(() => useAuth())).toThrow('useAuth must be used within AuthProvider')
    spy.mockRestore()
  })
})
