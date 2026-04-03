import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { PlannerHeader, type PlannerHeaderProps } from './PlannerHeader'

function defaultProps(overrides: Partial<PlannerHeaderProps> = {}): PlannerHeaderProps {
  return {
    weekStart: new Date('2026-03-30'),
    weekEnd: new Date('2026-04-05'),
    totalAssigned: 0,
    onPrevWeek: vi.fn(),
    onNextWeek: vi.fn(),
    onGoToday: vi.fn(),
    onCloneWeek: vi.fn(),
    onCreateActivities: vi.fn(),
    cloneWeekPending: false,
    batchCreatePending: false,
    ...overrides,
  }
}

describe('PlannerHeader', () => {
  it('renders week date range', () => {
    render(<PlannerHeader {...defaultProps()} />)
    expect(screen.getByText(/30 Mar/)).toBeInTheDocument()
    expect(screen.getByText(/5 Apr/)).toBeInTheDocument()
  })

  it('previous week button calls onPrevWeek', () => {
    const onPrevWeek = vi.fn()
    render(<PlannerHeader {...defaultProps({ onPrevWeek })} />)
    fireEvent.click(screen.getByLabelText('Previous week'))
    expect(onPrevWeek).toHaveBeenCalledOnce()
  })

  it('next week button calls onNextWeek', () => {
    const onNextWeek = vi.fn()
    render(<PlannerHeader {...defaultProps({ onNextWeek })} />)
    fireEvent.click(screen.getByLabelText('Next week'))
    expect(onNextWeek).toHaveBeenCalledOnce()
  })

  it('today button calls goToday', () => {
    const onGoToday = vi.fn()
    render(<PlannerHeader {...defaultProps({ onGoToday })} />)
    fireEvent.click(screen.getByText('Today'))
    expect(onGoToday).toHaveBeenCalledOnce()
  })

  it('clone week button calls onCloneWeek', () => {
    const onCloneWeek = vi.fn()
    render(<PlannerHeader {...defaultProps({ onCloneWeek })} />)
    fireEvent.click(screen.getByText('Clone Week'))
    expect(onCloneWeek).toHaveBeenCalledOnce()
  })

  it('shows Create Visits button when totalAssigned > 0', () => {
    render(<PlannerHeader {...defaultProps({ totalAssigned: 3 })} />)
    expect(screen.getByText(/Create 3 Visits/)).toBeInTheDocument()
  })

  it('hides Create Visits button when totalAssigned = 0', () => {
    render(<PlannerHeader {...defaultProps({ totalAssigned: 0 })} />)
    expect(screen.queryByText(/Create.*Visit/)).not.toBeInTheDocument()
  })

  it('disables Clone Week button when cloneWeekPending', () => {
    render(<PlannerHeader {...defaultProps({ cloneWeekPending: true })} />)
    expect(screen.getByText('Clone Week').closest('button')).toBeDisabled()
  })
})
