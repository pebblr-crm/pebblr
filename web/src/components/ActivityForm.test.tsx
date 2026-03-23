import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ActivityForm } from './ActivityForm'
import type { TenantConfig } from '../types/config'
import type { PaginatedResponse } from '../types/api'
import type { Target } from '../types/target'

vi.mock('../services/config', () => ({
  useConfig: vi.fn(),
}))

vi.mock('../services/targets', () => ({
  useTargets: vi.fn(),
}))

vi.mock('../services/teams', () => ({
  useTeamMembers: vi.fn(),
}))

import { useConfig } from '../services/config'
import { useTargets } from '../services/targets'
import { useTeamMembers } from '../services/teams'
const mockUseConfig = vi.mocked(useConfig)
const mockUseTargets = vi.mocked(useTargets)
const mockUseTeamMembers = vi.mocked(useTeamMembers)

const testConfig: TenantConfig = {
  tenant: { name: 'Test', locale: 'en' },
  accounts: {
    types: [
      { key: 'doctor', label: 'Doctor', fields: [{ key: 'name', type: 'text', required: true }] },
    ],
  },
  activities: {
    statuses: [
      { key: 'planificat', label: 'Planned', initial: true },
      { key: 'realizat', label: 'Realized' },
    ],
    status_transitions: {
      planificat: ['realizat', 'anulat'],
      realizat: [],
    },
    durations: [
      { key: 'full_day', label: 'Full Day' },
      { key: 'half_day', label: 'Half Day' },
    ],
    types: [
      {
        key: 'visit',
        label: 'Visit',
        category: 'field',
        fields: [
          { key: 'notes', type: 'text', required: false },
          { key: 'products', type: 'multi_select', required: true, options_ref: 'products' },
        ],
      },
      {
        key: 'vacation',
        label: 'Vacation',
        category: 'non_field',
        fields: [],
        blocks_field_activities: true,
      },
    ],
    routing_options: [
      { key: 'week_1', label: 'Week 1' },
      { key: 'week_2', label: 'Week 2' },
    ],
  },
  options: {
    products: [
      { key: 'aspirin', label: 'Aspirin' },
      { key: 'ibuprofen', label: 'Ibuprofen' },
    ],
  },
  rules: {
    frequency: {},
    max_activities_per_day: 10,
    default_visit_duration_minutes: {},
    visit_duration_step_minutes: 30,
  },
}

const emptyTargets: PaginatedResponse<Target> = {
  items: [],
  total: 0,
  page: 1,
  limit: 20,
}

function queryResult<T>(data: T) {
  return {
    data,
    isLoading: false,
    isError: false,
    error: null,
  } as unknown as ReturnType<typeof useConfig>
}

function targetsResult<T>(data: T) {
  return {
    data,
    isLoading: false,
    isError: false,
    error: null,
  } as unknown as ReturnType<typeof useTargets>
}

describe('ActivityForm', () => {
  const onSubmit = vi.fn()
  const onCancel = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
    mockUseConfig.mockReturnValue(queryResult(testConfig))
    mockUseTargets.mockReturnValue(targetsResult(emptyTargets))
    mockUseTeamMembers.mockReturnValue({ data: { items: [], total: 0, page: 1, limit: 20 }, isLoading: false, isError: false, error: null } as unknown as ReturnType<typeof useTeamMembers>)
  })

  it('renders the form with core fields (status hidden on create)', () => {
    render(
      <ActivityForm onSubmit={onSubmit} onCancel={onCancel} isSubmitting={false} />,
    )

    expect(screen.getByTestId('activity-form')).toBeInTheDocument()
    expect(screen.getByTestId('activity-type-select')).toBeInTheDocument()
    expect(screen.queryByTestId('status-select')).not.toBeInTheDocument()
    expect(screen.getByTestId('due-date-input')).toBeInTheDocument()
    expect(screen.getByTestId('duration-select')).toBeInTheDocument()
    expect(screen.getByText('New Activity')).toBeInTheDocument()
  })

  it('shows activity type options from config', () => {
    render(
      <ActivityForm onSubmit={onSubmit} onCancel={onCancel} isSubmitting={false} />,
    )

    const select = screen.getByTestId('activity-type-select')
    expect(select).toBeInTheDocument()
    expect(screen.getByText('Visit')).toBeInTheDocument()
    expect(screen.getByText('Vacation')).toBeInTheDocument()
  })

  it('shows status select when editing with initial value from data', () => {
    const activity = {
      id: '1',
      activityType: 'visit',
      status: 'planificat',
      dueDate: '2026-04-01',
      duration: 'full_day',
      fields: {},
      creatorId: 'user1',
      createdAt: '2026-04-01T00:00:00Z',
      updatedAt: '2026-04-01T00:00:00Z',
    }

    render(
      <ActivityForm
        initialData={activity}
        onSubmit={onSubmit}
        onCancel={onCancel}
        isSubmitting={false}
      />,
    )

    const statusSelect = screen.getByTestId('status-select') as HTMLSelectElement
    expect(statusSelect.value).toBe('planificat')
  })

  it('renders dynamic fields when activity type is selected', async () => {
    const user = userEvent.setup()
    render(
      <ActivityForm onSubmit={onSubmit} onCancel={onCancel} isSubmitting={false} />,
    )

    await user.selectOptions(screen.getByTestId('activity-type-select'), 'visit')

    expect(screen.getByTestId('field-notes')).toBeInTheDocument()
    expect(screen.getByTestId('field-products')).toBeInTheDocument()
  })

  it('shows target search for field activities', async () => {
    const user = userEvent.setup()
    render(
      <ActivityForm onSubmit={onSubmit} onCancel={onCancel} isSubmitting={false} />,
    )

    await user.selectOptions(screen.getByTestId('activity-type-select'), 'visit')
    expect(screen.getByTestId('target-search')).toBeInTheDocument()
  })

  it('does not show target search for non-field activities', async () => {
    const user = userEvent.setup()
    render(
      <ActivityForm onSubmit={onSubmit} onCancel={onCancel} isSubmitting={false} />,
    )

    await user.selectOptions(screen.getByTestId('activity-type-select'), 'vacation')
    expect(screen.queryByTestId('target-search')).not.toBeInTheDocument()
  })

  it('calls onSubmit with form data', async () => {
    const user = userEvent.setup()
    render(
      <ActivityForm onSubmit={onSubmit} onCancel={onCancel} isSubmitting={false} />,
    )

    await user.selectOptions(screen.getByTestId('activity-type-select'), 'vacation')
    await user.type(screen.getByTestId('due-date-input'), '2026-04-01')
    await user.selectOptions(screen.getByTestId('duration-select'), 'full_day')

    await user.click(screen.getByTestId('submit-button'))

    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({
        activityType: 'vacation',
        status: '',
        dueDate: '2026-04-01',
        duration: 'full_day',
        fields: {},
      }),
    )
  })

  it('calls onCancel when cancel button is clicked', async () => {
    const user = userEvent.setup()
    render(
      <ActivityForm onSubmit={onSubmit} onCancel={onCancel} isSubmitting={false} />,
    )

    await user.click(screen.getByText('Cancel'))
    expect(onCancel).toHaveBeenCalled()
  })

  it('disables submit button when isSubmitting is true', () => {
    render(
      <ActivityForm onSubmit={onSubmit} onCancel={onCancel} isSubmitting={true} />,
    )

    expect(screen.getByTestId('submit-button')).toBeDisabled()
    expect(screen.getByText('Saving...')).toBeInTheDocument()
  })

  it('shows Edit Activity title when editing', () => {
    const activity = {
      id: '1',
      activityType: 'visit',
      status: 'planificat',
      dueDate: '2026-04-01',
      duration: 'full_day',
      fields: {},
      creatorId: 'user1',
      createdAt: '2026-04-01T00:00:00Z',
      updatedAt: '2026-04-01T00:00:00Z',
    }

    render(
      <ActivityForm
        initialData={activity}
        onSubmit={onSubmit}
        onCancel={onCancel}
        isSubmitting={false}
      />,
    )

    expect(screen.getByText('Edit Activity')).toBeInTheDocument()
    expect(screen.getByText('Update Activity')).toBeInTheDocument()
  })

  it('shows loading spinner while config loads', () => {
    mockUseConfig.mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      error: null,
    } as ReturnType<typeof useConfig>)

    render(
      <ActivityForm onSubmit={onSubmit} onCancel={onCancel} isSubmitting={false} />,
    )

    expect(screen.getByText('Loading configuration...')).toBeInTheDocument()
  })

  it('displays server-side field errors', () => {
    render(
      <ActivityForm
        onSubmit={onSubmit}
        onCancel={onCancel}
        isSubmitting={false}
        serverErrors={[{ field: 'activityType', message: 'invalid type' }]}
      />,
    )

    expect(screen.getByText('invalid type')).toBeInTheDocument()
  })

  it('shows routing as a dynamic field for visit type', async () => {
    const configWithRouting: TenantConfig = {
      ...testConfig,
      activities: {
        ...testConfig.activities,
        types: [
          {
            key: 'visit',
            label: 'Visit',
            category: 'field',
            fields: [
              { key: 'routing', type: 'select', required: false, options_ref: 'routing_options' },
              { key: 'notes', type: 'text', required: false },
            ],
          },
          { key: 'vacation', label: 'Vacation', category: 'non_field', fields: [], blocks_field_activities: true },
        ],
      },
      options: {
        ...testConfig.options,
        routing_options: [
          { key: 'week_1', label: 'Week 1' },
          { key: 'week_2', label: 'Week 2' },
        ],
      },
    }
    mockUseConfig.mockReturnValue(queryResult(configWithRouting))

    const user = userEvent.setup()
    render(
      <ActivityForm onSubmit={onSubmit} onCancel={onCancel} isSubmitting={false} />,
    )

    // Routing should NOT be visible before selecting visit
    expect(screen.queryByTestId('field-routing')).not.toBeInTheDocument()

    await user.selectOptions(screen.getByTestId('activity-type-select'), 'visit')

    // Now routing should appear as a dynamic field
    expect(screen.getByTestId('field-routing')).toBeInTheDocument()
    expect(screen.getByText('Week 1')).toBeInTheDocument()
    expect(screen.getByText('Week 2')).toBeInTheDocument()
  })

  it('resolves options_ref pointing to activities.durations', async () => {
    const configWithRef: TenantConfig = {
      ...testConfig,
      activities: {
        ...testConfig.activities,
        types: [
          {
            key: 'admin',
            label: 'Administrative',
            category: 'non_field',
            fields: [
              { key: 'pace', type: 'select', required: true, options_ref: 'durations' },
            ],
          },
        ],
      },
    }
    mockUseConfig.mockReturnValue(queryResult(configWithRef))

    const user = userEvent.setup()
    render(
      <ActivityForm onSubmit={onSubmit} onCancel={onCancel} isSubmitting={false} />,
    )

    await user.selectOptions(screen.getByTestId('activity-type-select'), 'admin')

    const paceSelect = screen.getByTestId('field-pace') as HTMLSelectElement
    expect(paceSelect).toBeInTheDocument()
    // Should contain the duration options from config.activities.durations
    expect(paceSelect.querySelectorAll('option').length).toBe(3) // placeholder + 2 durations
  })

  it('does not render duration in dynamic fields (handled by core field)', async () => {
    const configWithDuration: TenantConfig = {
      ...testConfig,
      activities: {
        ...testConfig.activities,
        types: [
          {
            key: 'admin',
            label: 'Administrative',
            category: 'non_field',
            fields: [
              { key: 'duration', type: 'select', required: true, options_ref: 'durations' },
              { key: 'details', type: 'text', required: false },
            ],
          },
        ],
      },
    }
    mockUseConfig.mockReturnValue(queryResult(configWithDuration))

    const user = userEvent.setup()
    render(
      <ActivityForm onSubmit={onSubmit} onCancel={onCancel} isSubmitting={false} />,
    )

    await user.selectOptions(screen.getByTestId('activity-type-select'), 'admin')

    // duration should NOT appear in dynamic fields
    expect(screen.queryByTestId('field-duration')).not.toBeInTheDocument()
    // details should still appear
    expect(screen.getByTestId('field-details')).toBeInTheDocument()
  })

  it('renders multi-select fields with toggle buttons', async () => {
    const user = userEvent.setup()
    render(
      <ActivityForm onSubmit={onSubmit} onCancel={onCancel} isSubmitting={false} />,
    )

    await user.selectOptions(screen.getByTestId('activity-type-select'), 'visit')

    expect(screen.getByText('Aspirin')).toBeInTheDocument()
    expect(screen.getByText('Ibuprofen')).toBeInTheDocument()
  })
})
