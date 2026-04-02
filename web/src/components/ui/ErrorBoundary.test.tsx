import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ErrorBoundary } from './ErrorBoundary'

function ThrowingChild({ shouldThrow }: { shouldThrow: boolean }) {
  if (shouldThrow) throw new Error('Test explosion')
  return <div>All good</div>
}

describe('ErrorBoundary', () => {
  // Suppress React error boundary console noise during tests
  const originalError = console.error
  beforeAll(() => {
    console.error = (...args: unknown[]) => {
      const msg = typeof args[0] === 'string' ? args[0] : ''
      if (msg.includes('ErrorBoundary') || msg.includes('The above error')) return
      originalError.call(console, ...args)
    }
  })
  afterAll(() => {
    console.error = originalError
  })

  it('renders children when there is no error', () => {
    render(
      <ErrorBoundary>
        <ThrowingChild shouldThrow={false} />
      </ErrorBoundary>,
    )
    expect(screen.getByText('All good')).toBeInTheDocument()
  })

  it('renders fallback UI when a child throws', () => {
    render(
      <ErrorBoundary>
        <ThrowingChild shouldThrow={true} />
      </ErrorBoundary>,
    )
    expect(screen.getByText('Something went wrong')).toBeInTheDocument()
    expect(screen.getByText('Test explosion')).toBeInTheDocument()
  })

  it('renders custom fallback when provided', () => {
    render(
      <ErrorBoundary fallback={<div>Custom fallback</div>}>
        <ThrowingChild shouldThrow={true} />
      </ErrorBoundary>,
    )
    expect(screen.getByText('Custom fallback')).toBeInTheDocument()
  })

  it('recovers when Try again is clicked', async () => {
    let shouldThrow = true
    function Toggler() {
      if (shouldThrow) throw new Error('Boom')
      return <div>Recovered</div>
    }

    const { rerender } = render(
      <ErrorBoundary>
        <Toggler />
      </ErrorBoundary>,
    )
    expect(screen.getByText('Something went wrong')).toBeInTheDocument()

    shouldThrow = false
    const user = userEvent.setup()
    await user.click(screen.getByRole('button', { name: /try again/i }))

    rerender(
      <ErrorBoundary>
        <Toggler />
      </ErrorBoundary>,
    )
    expect(screen.getByText('Recovered')).toBeInTheDocument()
  })
})
