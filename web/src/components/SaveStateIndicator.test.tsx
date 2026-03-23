import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import userEvent from '@testing-library/user-event'
import { SaveStateIndicator } from './SaveStateIndicator'

describe('SaveStateIndicator', () => {
  it('renders nothing when idle', () => {
    const { container } = render(<SaveStateIndicator saveState="idle" />)
    expect(container.firstChild).toBeNull()
  })

  it('shows pulsing dot and Saving… when dirty', () => {
    render(<SaveStateIndicator saveState="dirty" />)
    expect(screen.getByText('Saving…')).toBeInTheDocument()
  })

  it('shows spinner and Saving… when saving', () => {
    render(<SaveStateIndicator saveState="saving" />)
    expect(screen.getByText('Saving…')).toBeInTheDocument()
  })

  it('shows retry button when error', () => {
    render(<SaveStateIndicator saveState="error" />)
    expect(screen.getByTestId('save-state-retry')).toBeInTheDocument()
    expect(screen.getByText(/Not saved/)).toBeInTheDocument()
  })

  it('calls onRetry when retry button is clicked', async () => {
    const onRetry = vi.fn()
    render(<SaveStateIndicator saveState="error" onRetry={onRetry} />)
    await userEvent.click(screen.getByTestId('save-state-retry'))
    expect(onRetry).toHaveBeenCalledOnce()
  })
})
