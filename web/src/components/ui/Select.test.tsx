import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Select, selectStyles } from './Select'

describe('Select', () => {
  it('renders a select element with default styles', () => {
    render(
      <Select data-testid="select">
        <option value="a">A</option>
        <option value="b">B</option>
      </Select>,
    )
    const el = screen.getByTestId('select')
    expect(el.tagName).toBe('SELECT')
    expect(el).toHaveClass('rounded-lg')
    expect(el).toHaveClass('border-slate-300')
  })

  it('renders children options', () => {
    render(
      <Select data-testid="select">
        <option value="one">One</option>
        <option value="two">Two</option>
      </Select>,
    )
    const options = screen.getAllByRole('option')
    expect(options).toHaveLength(2)
    expect(options[0]).toHaveTextContent('One')
    expect(options[1]).toHaveTextContent('Two')
  })

  it('merges additional className', () => {
    render(
      <Select data-testid="select" className="w-auto">
        <option>X</option>
      </Select>,
    )
    const el = screen.getByTestId('select')
    expect(el).toHaveClass('w-auto')
    expect(el).toHaveClass('rounded-lg')
  })

  it('handles selection changes', async () => {
    const user = userEvent.setup()
    render(
      <Select data-testid="select">
        <option value="a">A</option>
        <option value="b">B</option>
      </Select>,
    )
    const el = screen.getByTestId('select')
    await user.selectOptions(el, 'b')
    expect(el).toHaveValue('b')
  })

  it('exports selectStyles for reuse', () => {
    expect(selectStyles).toContain('rounded-lg')
    expect(selectStyles).toContain('border-slate-300')
    expect(selectStyles).toContain('bg-white')
  })
})
