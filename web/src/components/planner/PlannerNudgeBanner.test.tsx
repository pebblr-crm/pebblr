import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { PlannerNudgeBanner } from './PlannerNudgeBanner'

describe('PlannerNudgeBanner', () => {
  it('shows overdue count when overdueA > 0', () => {
    render(<PlannerNudgeBanner overdueA={5} completionRate={60} coveragePct={70} />)
    expect(screen.getByText(/5 A-priority targets/)).toBeInTheDocument()
    expect(screen.getByText(/need visits/)).toBeInTheDocument()
  })

  it('shows "All covered" when overdueA = 0', () => {
    render(<PlannerNudgeBanner overdueA={0} completionRate={90} coveragePct={85} />)
    expect(screen.getByText(/All A-priority targets are covered/)).toBeInTheDocument()
  })

  it('shows completion rate with green when >= 80%', () => {
    render(<PlannerNudgeBanner overdueA={0} completionRate={85} coveragePct={90} />)
    const rateEl = screen.getByText('85%')
    expect(rateEl).toHaveClass('text-emerald-600')
  })

  it('shows completion rate with amber when < 80%', () => {
    render(<PlannerNudgeBanner overdueA={0} completionRate={50} coveragePct={90} />)
    const rateEl = screen.getByText('50%')
    expect(rateEl).toHaveClass('text-amber-600')
  })

  it('shows coverage pct', () => {
    render(<PlannerNudgeBanner overdueA={0} completionRate={90} coveragePct={72} />)
    expect(screen.getByText('72%')).toBeInTheDocument()
  })
})
