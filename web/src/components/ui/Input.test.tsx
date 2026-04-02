import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Input, inputStyles } from './Input'

describe('Input', () => {
  it('renders an input element with default styles', () => {
    render(<Input data-testid="input" />)
    const el = screen.getByTestId('input')
    expect(el.tagName).toBe('INPUT')
    expect(el).toHaveClass('rounded-lg')
    expect(el).toHaveClass('border-slate-300')
  })

  it('merges additional className', () => {
    render(<Input data-testid="input" className="mt-2" />)
    const el = screen.getByTestId('input')
    expect(el).toHaveClass('mt-2')
    expect(el).toHaveClass('rounded-lg')
  })

  it('forwards native props', () => {
    render(<Input placeholder="Enter name" type="email" data-testid="input" />)
    const el = screen.getByTestId('input')
    expect(el).toHaveAttribute('placeholder', 'Enter name')
    expect(el).toHaveAttribute('type', 'email')
  })

  it('handles user input', async () => {
    const user = userEvent.setup()
    render(<Input data-testid="input" />)
    const el = screen.getByTestId('input')
    await user.type(el, 'hello')
    expect(el).toHaveValue('hello')
  })

  it('exports inputStyles for reuse', () => {
    expect(inputStyles).toContain('rounded-lg')
    expect(inputStyles).toContain('border-slate-300')
    expect(inputStyles).toContain('focus:border-teal-500')
  })
})
