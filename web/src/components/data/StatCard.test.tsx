import { render, screen } from '@testing-library/react'
import { StatCard } from './StatCard'

describe('StatCard', () => {
  it('renders label and value', () => {
    render(<StatCard label="Coverage" value="85%" />)
    expect(screen.getByText('Coverage')).toBeInTheDocument()
    expect(screen.getByText('85%')).toBeInTheDocument()
  })

  it('renders subtitle when provided', () => {
    render(<StatCard label="Visits" value={42} subtitle="this week" />)
    expect(screen.getByText('this week')).toBeInTheDocument()
  })

  it('renders numeric value', () => {
    render(<StatCard label="Count" value={7} />)
    expect(screen.getByText('7')).toBeInTheDocument()
  })
})
