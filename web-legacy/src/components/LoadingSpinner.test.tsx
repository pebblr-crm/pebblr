import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { LoadingSpinner } from './LoadingSpinner'

describe('LoadingSpinner', () => {
  it('renders with default label', () => {
    render(<LoadingSpinner />)
    expect(screen.getByRole('status')).toBeInTheDocument()
    expect(screen.getByLabelText('Loading...')).toBeInTheDocument()
  })

  it('renders with custom label', () => {
    render(<LoadingSpinner label="Fetching leads" />)
    expect(screen.getByLabelText('Fetching leads')).toBeInTheDocument()
  })

  it('renders without crashing for each size', () => {
    const { rerender } = render(<LoadingSpinner size="sm" />)
    expect(screen.getByRole('status')).toBeInTheDocument()
    rerender(<LoadingSpinner size="md" />)
    expect(screen.getByRole('status')).toBeInTheDocument()
    rerender(<LoadingSpinner size="lg" />)
    expect(screen.getByRole('status')).toBeInTheDocument()
  })
})
