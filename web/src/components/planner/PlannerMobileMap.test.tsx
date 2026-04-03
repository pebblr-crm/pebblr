import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import { PlannerMobileMap } from './PlannerMobileMap'
import type { Target } from '@/types/target'

vi.mock('@/components/map/MapContainer', () => ({
  MapContainer: ({ children, ...props }: { children?: React.ReactNode; className?: string }) => (
    <div data-testid="map-container" {...props}>{children}</div>
  ),
}))

vi.mock('@/components/map/TargetMarker', () => ({
  TargetMarker: ({ name, onClick }: { name: string; onClick: () => void }) => (
    <div data-testid="target-marker" onClick={onClick}>{name}</div>
  ),
}))

vi.mock('@/lib/target-fields', () => ({
  getLat: () => 44.4,
  getLng: () => 26.1,
  getClassification: () => 'a',
}))

function makeTarget(id: string, name: string): Target {
  return {
    id,
    targetType: 'pharmacy',
    name,
    fields: { lat: 44.4, lng: 26.1, classification: 'a' },
    assigneeId: 'user-1',
    teamId: 'team-1',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  }
}

describe('PlannerMobileMap', () => {
  it('renders map container', () => {
    render(
      <PlannerMobileMap
        geoTargets={[]}
        selectedTargetIds={new Set()}
        onToggleTarget={vi.fn()}
        onClose={vi.fn()}
      />,
    )
    expect(screen.getByTestId('map-container')).toBeInTheDocument()
  })

  it('renders markers for each geoTarget', () => {
    const targets = [makeTarget('t1', 'Pharmacy Alpha'), makeTarget('t2', 'Pharmacy Beta')]
    render(
      <PlannerMobileMap
        geoTargets={targets}
        selectedTargetIds={new Set()}
        onToggleTarget={vi.fn()}
        onClose={vi.fn()}
      />,
    )
    const markers = screen.getAllByTestId('target-marker')
    expect(markers).toHaveLength(2)
    expect(screen.getByText('Pharmacy Alpha')).toBeInTheDocument()
    expect(screen.getByText('Pharmacy Beta')).toBeInTheDocument()
  })

  it('close button calls onClose', () => {
    const onClose = vi.fn()
    render(
      <PlannerMobileMap
        geoTargets={[]}
        selectedTargetIds={new Set()}
        onToggleTarget={vi.fn()}
        onClose={onClose}
      />,
    )
    fireEvent.click(screen.getByLabelText('Close map'))
    expect(onClose).toHaveBeenCalledOnce()
  })

  it('marker click calls onToggleTarget', () => {
    const onToggleTarget = vi.fn()
    const targets = [makeTarget('t1', 'Pharmacy Alpha')]
    render(
      <PlannerMobileMap
        geoTargets={targets}
        selectedTargetIds={new Set()}
        onToggleTarget={onToggleTarget}
        onClose={vi.fn()}
      />,
    )
    fireEvent.click(screen.getByText('Pharmacy Alpha'))
    expect(onToggleTarget).toHaveBeenCalledWith('t1')
  })
})
