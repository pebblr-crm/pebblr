import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import type { Target } from '../../types/target'
import type { TenantConfig } from '../../types/config'

vi.mock('@tanstack/react-router', async () => {
  const actual = await vi.importActual('@tanstack/react-router')
  return {
    ...actual,
    Link: ({ children, className, ...rest }: { children?: React.ReactNode; className?: string; to?: string }) => (
      <a className={className} href={rest.to}>{children}</a>
    ),
  }
})

vi.mock('../../services/targets', () => ({
  useTarget: vi.fn(),
}))

vi.mock('../../services/config', () => ({
  useConfig: vi.fn(),
}))

import { useTarget } from '../../services/targets'
import { useConfig } from '../../services/config'
const mockUseTarget = vi.mocked(useTarget)
const mockUseConfig = vi.mocked(useConfig)

const testConfig: TenantConfig = {
  tenant: { name: 'Test', locale: 'en' },
  accounts: {
    types: [
      {
        key: 'doctor',
        label: 'Doctor',
        fields: [
          { key: 'name', type: 'text', required: true },
          { key: 'specialty', type: 'select', required: false, options_ref: 'specialties' },
          { key: 'city', type: 'text', required: false },
          { key: 'county', type: 'text', required: false },
          { key: 'address', type: 'text', required: false },
        ],
      },
    ],
  },
  options: {
    specialties: [
      { key: 'cardiology', label: 'Cardiology' },
    ],
  },
}

function makeTarget(overrides: Partial<Target> = {}): Target {
  return {
    id: 'target-1',
    targetType: 'doctor',
    name: 'Dr. Smith',
    fields: { specialty: 'cardiology', city: 'Bucharest', county: 'Ilfov', address: 'Str. Victoriei 10' },
    assigneeId: 'user-1',
    teamId: 'team-1',
    createdAt: '2026-01-15T10:00:00Z',
    updatedAt: '2026-01-15T10:00:00Z',
    ...overrides,
  }
}

function setupConfig() {
  mockUseConfig.mockReturnValue({
    data: testConfig,
    isLoading: false,
    isError: false,
    error: null,
  } as ReturnType<typeof useConfig>)
}

// Lazy-import the component after mocks are set up, and override Route.useParams
async function importComponent() {
  const mod = await import('./$targetId')
  // Override the Route object's useParams to avoid needing RouterProvider
  ;(mod.Route as { useParams: () => { targetId: string } }).useParams = () => ({ targetId: 'target-1' })
  return mod.TargetDetailPage
}

describe('TargetDetailPage', () => {
  let TargetDetailPage: Awaited<ReturnType<typeof importComponent>>

  beforeEach(async () => {
    vi.clearAllMocks()
    setupConfig()
    TargetDetailPage = await importComponent()
  })

  it('shows loading spinner while fetching', () => {
    mockUseTarget.mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      error: null,
    } as ReturnType<typeof useTarget>)

    render(<TargetDetailPage />)

    expect(screen.getByLabelText('Loading target...')).toBeInTheDocument()
  })

  it('shows error state when fetch fails', () => {
    mockUseTarget.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: true,
      error: new Error('Not found'),
    } as ReturnType<typeof useTarget>)

    render(<TargetDetailPage />)

    expect(screen.getByTestId('error-state')).toBeInTheDocument()
    expect(screen.getByText('Not found')).toBeInTheDocument()
  })

  it('shows not found when target is null', () => {
    mockUseTarget.mockReturnValue({
      data: undefined,
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useTarget>)

    render(<TargetDetailPage />)

    expect(screen.getByTestId('not-found')).toBeInTheDocument()
    expect(screen.getByText('Target not found.')).toBeInTheDocument()
  })

  it('renders target name and type badge', () => {
    mockUseTarget.mockReturnValue({
      data: makeTarget(),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useTarget>)

    render(<TargetDetailPage />)

    expect(screen.getByText('Dr. Smith')).toBeInTheDocument()
    expect(screen.getByText('Doctor')).toBeInTheDocument()
  })

  it('resolves option labels from config', () => {
    mockUseTarget.mockReturnValue({
      data: makeTarget(),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useTarget>)

    render(<TargetDetailPage />)

    expect(screen.getByText('Cardiology')).toBeInTheDocument()
  })

  it('shows location from address, city, county', () => {
    mockUseTarget.mockReturnValue({
      data: makeTarget(),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useTarget>)

    render(<TargetDetailPage />)

    expect(screen.getByText('Str. Victoriei 10, Bucharest, Ilfov')).toBeInTheDocument()
  })

  it('shows back to targets link', () => {
    mockUseTarget.mockReturnValue({
      data: makeTarget(),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useTarget>)

    render(<TargetDetailPage />)

    expect(screen.getByText('Back to targets')).toBeInTheDocument()
  })

  it('shows activities placeholder section', () => {
    mockUseTarget.mockReturnValue({
      data: makeTarget(),
      isLoading: false,
      isError: false,
      error: null,
    } as ReturnType<typeof useTarget>)

    render(<TargetDetailPage />)

    expect(screen.getByText('Activities')).toBeInTheDocument()
    expect(screen.getByText(/Activity tracking will be available/)).toBeInTheDocument()
  })
})
