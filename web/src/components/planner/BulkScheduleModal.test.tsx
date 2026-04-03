import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { BulkScheduleModal, type BulkScheduleModalProps } from './BulkScheduleModal'
import type { Target } from '@/types/target'

vi.mock('@/components/ui/Modal', () => ({
  Modal: ({
    open,
    children,
    title,
    footer,
  }: {
    open: boolean
    children: React.ReactNode
    title: string
    footer?: React.ReactNode
    onClose: () => void
  }) =>
    open ? (
      <div data-testid="modal">
        <h2>{title}</h2>
        {children}
        {footer && <div data-testid="modal-footer">{footer}</div>}
      </div>
    ) : null,
}))

vi.mock('@/components/ui/Button', () => ({
  Button: ({
    children,
    onClick,
    disabled,
    ...props
  }: {
    children: React.ReactNode
    onClick?: () => void
    disabled?: boolean
    variant?: string
    size?: string
  }) => (
    <button onClick={onClick} disabled={disabled} {...props}>
      {children}
    </button>
  ),
}))

vi.mock('@/lib/styles', () => ({
  priorityDot: { a: 'dot-a', b: 'dot-b', c: 'dot-c' },
  classificationBadge: { a: 'badge-a', b: 'badge-b', c: 'badge-c' },
}))

vi.mock('@/lib/target-fields', () => ({
  getClassification: () => 'a',
}))

function makeTarget(id: string, name: string): Target {
  return {
    id,
    targetType: 'pharmacy',
    name,
    fields: { classification: 'a' },
    assigneeId: 'user-1',
    teamId: 'team-1',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  }
}

function defaultProps(overrides: Partial<BulkScheduleModalProps> = {}): BulkScheduleModalProps {
  const t1 = makeTarget('t1', 'Pharmacy Alpha')
  const t2 = makeTarget('t2', 'Pharmacy Beta')
  return {
    open: true,
    onClose: vi.fn(),
    selectedTargetIds: new Set(['t1', 't2']),
    targetMap: new Map([
      ['t1', t1],
      ['t2', t2],
    ]),
    initialDate: '2026-04-01',
    onSchedule: vi.fn<(date: string, visitType: 'f2f' | 'remote') => Promise<void>>().mockResolvedValue(undefined),
    isPending: false,
    ...overrides,
  }
}

describe('BulkScheduleModal', () => {
  it('renders nothing when closed', () => {
    render(<BulkScheduleModal {...defaultProps({ open: false })} />)
    expect(screen.queryByTestId('modal')).not.toBeInTheDocument()
  })

  it('shows target count in description', () => {
    render(<BulkScheduleModal {...defaultProps()} />)
    const description = screen.getByText((_content, element) =>
      element?.tagName === 'P' && /schedule\s+2\s+selected target/i.test(element.textContent ?? ''),
    )
    expect(description).toBeInTheDocument()
  })

  it('date input changes date', () => {
    render(<BulkScheduleModal {...defaultProps()} />)
    const input = screen.getByDisplayValue('2026-04-01')
    fireEvent.change(input, { target: { value: '2026-04-10' } })
    expect(input).toHaveValue('2026-04-10')
  })

  it('visit type toggle switches between f2f and remote', () => {
    render(<BulkScheduleModal {...defaultProps()} />)
    const remoteBtn = screen.getByText('Remote')
    fireEvent.click(remoteBtn)
    // After clicking Remote, it should have the active class
    expect(remoteBtn).toHaveClass('bg-teal-600')
  })

  it('lists selected targets with names', () => {
    render(<BulkScheduleModal {...defaultProps()} />)
    expect(screen.getByText('Pharmacy Alpha')).toBeInTheDocument()
    expect(screen.getByText('Pharmacy Beta')).toBeInTheDocument()
  })

  it('schedule button calls onSchedule with date and visit type', () => {
    const onSchedule = vi.fn<(date: string, visitType: 'f2f' | 'remote') => Promise<void>>().mockResolvedValue(undefined)
    render(<BulkScheduleModal {...defaultProps({ onSchedule })} />)
    fireEvent.click(screen.getByText(/Schedule 2 Targets/))
    expect(onSchedule).toHaveBeenCalledWith('2026-04-01', 'f2f')
  })

  it('cancel button calls onClose', () => {
    const onClose = vi.fn()
    render(<BulkScheduleModal {...defaultProps({ onClose })} />)
    fireEvent.click(screen.getByText('Cancel'))
    expect(onClose).toHaveBeenCalledOnce()
  })
})
