import { render, screen } from '@testing-library/react'
import { Badge } from './Badge'

describe('Badge', () => {
  it('renders children', () => {
    render(<Badge>Active</Badge>)
    expect(screen.getByText('Active')).toBeInTheDocument()
  })

  it('applies default variant classes', () => {
    const { container } = render(<Badge>Test</Badge>)
    expect(container.firstChild).toHaveClass('bg-slate-100')
  })

  it('applies danger variant', () => {
    const { container } = render(<Badge variant="danger">Error</Badge>)
    expect(container.firstChild).toHaveClass('bg-red-50')
  })

  it('applies success variant', () => {
    const { container } = render(<Badge variant="success">OK</Badge>)
    expect(container.firstChild).toHaveClass('bg-emerald-50')
  })

  it('merges custom className', () => {
    const { container } = render(<Badge className="ml-2">Tag</Badge>)
    expect(container.firstChild).toHaveClass('ml-2')
  })
})
