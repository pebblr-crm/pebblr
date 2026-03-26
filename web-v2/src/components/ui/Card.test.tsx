import { render, screen } from '@testing-library/react'
import { Card } from './Card'

describe('Card', () => {
  it('renders children', () => {
    render(<Card><p>Content</p></Card>)
    expect(screen.getByText('Content')).toBeInTheDocument()
  })

  it('applies border and shadow classes', () => {
    const { container } = render(<Card>Test</Card>)
    expect(container.firstChild).toHaveClass('rounded-xl', 'border', 'shadow-sm')
  })

  it('merges custom className', () => {
    const { container } = render(<Card className="mt-4">Test</Card>)
    expect(container.firstChild).toHaveClass('mt-4')
  })
})
