import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryError } from './QueryError'

describe('QueryError', () => {
  it('renders default message', () => {
    render(<QueryError />)
    expect(screen.getByText('Failed to load data')).toBeInTheDocument()
  })

  it('renders custom message', () => {
    render(<QueryError message="Targets failed to load" />)
    expect(screen.getByText('Targets failed to load')).toBeInTheDocument()
  })

  it('renders retry button when onRetry provided', () => {
    render(<QueryError onRetry={() => {}} />)
    expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument()
  })

  it('does not render retry button when onRetry omitted', () => {
    render(<QueryError />)
    expect(screen.queryByRole('button', { name: /retry/i })).not.toBeInTheDocument()
  })

  it('calls onRetry when retry button is clicked', async () => {
    const onRetry = vi.fn()
    render(<QueryError onRetry={onRetry} />)
    const user = userEvent.setup()
    await user.click(screen.getByRole('button', { name: /retry/i }))
    expect(onRetry).toHaveBeenCalledOnce()
  })
})
