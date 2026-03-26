import { render, screen, act } from '@testing-library/react'
import { renderHook } from '@testing-library/react'
import { vi } from 'vitest'
import { useToast } from './Toast'

describe('useToast', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('shows a toast message', () => {
    const { result } = renderHook(() => useToast())
    const { ToastContainer, showToast } = result.current

    const { rerender } = render(<ToastContainer />)

    act(() => {
      showToast('Hello world')
    })

    rerender(<result.current.ToastContainer />)
    expect(screen.getByText('Hello world')).toBeInTheDocument()
  })

  it('removes toast after timeout', () => {
    const { result } = renderHook(() => useToast())

    const { rerender } = render(<result.current.ToastContainer />)

    act(() => {
      result.current.showToast('Temporary')
    })
    rerender(<result.current.ToastContainer />)
    expect(screen.getByText('Temporary')).toBeInTheDocument()

    // Advance past show + exit animation
    act(() => {
      vi.advanceTimersByTime(3000)
    })
    rerender(<result.current.ToastContainer />)
    expect(screen.queryByText('Temporary')).not.toBeInTheDocument()
  })

  it('applies warning variant styling', () => {
    const { result } = renderHook(() => useToast())

    const { rerender } = render(<result.current.ToastContainer />)

    act(() => {
      result.current.showToast('Warning!', 'warning')
    })

    rerender(<result.current.ToastContainer />)
    const toast = screen.getByText('Warning!')
    expect(toast.className).toContain('bg-amber-600')
  })

  it('applies error variant styling', () => {
    const { result } = renderHook(() => useToast())

    const { rerender } = render(<result.current.ToastContainer />)

    act(() => {
      result.current.showToast('Error!', 'error')
    })

    rerender(<result.current.ToastContainer />)
    const toast = screen.getByText('Error!')
    expect(toast.className).toContain('bg-red-600')
  })

  it('can show multiple toasts', () => {
    const { result } = renderHook(() => useToast())

    const { rerender } = render(<result.current.ToastContainer />)

    act(() => {
      result.current.showToast('First')
      result.current.showToast('Second')
    })

    rerender(<result.current.ToastContainer />)
    expect(screen.getByText('First')).toBeInTheDocument()
    expect(screen.getByText('Second')).toBeInTheDocument()
  })
})
