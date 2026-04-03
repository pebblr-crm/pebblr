import { render } from '@testing-library/react'
import { vi } from 'vitest'

const mockShowToast = vi.fn()
const MockToastContainer = () => <div data-testid="toast-container" />
const mockOnToast = vi.fn()

vi.mock('./Toast', () => ({
  useToast: () => ({ showToast: mockShowToast, ToastContainer: MockToastContainer }),
}))

vi.mock('@/lib/toast-store', () => ({
  onToast: (fn: (message: string, variant: string) => void) => {
    mockOnToast(fn)
    return vi.fn() // unsubscribe
  },
}))

import { GlobalToast } from './GlobalToast'

describe('GlobalToast', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the ToastContainer', () => {
    const { getByTestId } = render(<GlobalToast />)
    expect(getByTestId('toast-container')).toBeInTheDocument()
  })

  it('subscribes to toast-store on mount', () => {
    render(<GlobalToast />)
    expect(mockOnToast).toHaveBeenCalledWith(expect.any(Function))
  })

  it('forwards store events to showToast', () => {
    render(<GlobalToast />)
    const storeCallback = mockOnToast.mock.calls[0][0]
    storeCallback('Error happened', 'error')
    expect(mockShowToast).toHaveBeenCalledWith('Error happened', 'error')
  })
})
