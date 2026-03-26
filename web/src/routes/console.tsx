import { useState, useMemo } from 'react'
import { createRoute } from '@tanstack/react-router'
import { createColumnHelper } from '@tanstack/react-table'
import { Route as rootRoute } from './__root'
import { useUsers } from '@/hooks/useUsers'
import { useTeams } from '@/hooks/useTeams'
import { useTerritories } from '@/hooks/useTerritories'
import { useConfig } from '@/hooks/useConfig'
import { DataTable } from '@/components/data/DataTable'
import { Badge } from '@/components/ui/Badge'
import { Card } from '@/components/ui/Card'
import { Spinner } from '@/components/ui/Spinner'
import { Users, Building2, MapPinned, Sliders } from 'lucide-react'
import type { User } from '@/types/user'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/console',
  component: ConsolePage,
})

type Section = 'users' | 'teams' | 'territories' | 'rules'

const roleVariant: Record<string, 'danger' | 'warning' | 'primary'> = {
  admin: 'danger',
  manager: 'warning',
  rep: 'primary',
}

const userColumnHelper = createColumnHelper<User>()

function ConsolePage() {
  const [section, setSection] = useState<Section>('users')
  const { data: usersData, isLoading: usersLoading } = useUsers()
  const { data: teamsData } = useTeams()
  const { data: territoriesData } = useTerritories()
  const { data: config } = useConfig()

  const users = useMemo(() => usersData?.items ?? [], [usersData])
  const teams = useMemo(() => teamsData?.items ?? [], [teamsData])
  const territories = useMemo(() => territoriesData?.items ?? [], [territoriesData])

  const userColumns = useMemo(
    () => [
      userColumnHelper.accessor('displayName', {
        header: 'Name',
        cell: (info) => (
          <div className="flex items-center gap-2">
            <div className="h-8 w-8 rounded-full bg-slate-200 flex items-center justify-center text-xs font-medium text-slate-600">
              {info.getValue()?.charAt(0) ?? '?'}
            </div>
            <span className="font-medium text-slate-900">{info.getValue()}</span>
          </div>
        ),
      }),
      userColumnHelper.accessor('email', { header: 'Email' }),
      userColumnHelper.accessor('role', {
        header: 'Role',
        cell: (info) => (
          <Badge variant={roleVariant[info.getValue()] ?? 'default'}>
            {info.getValue()}
          </Badge>
        ),
      }),
    ],
    [],
  )

  const sections: { key: Section; label: string; icon: React.ReactNode; count: number }[] = [
    { key: 'users', label: 'Users & Roles', icon: <Users size={18} />, count: users.length },
    { key: 'teams', label: 'Teams', icon: <Building2 size={18} />, count: teams.length },
    { key: 'territories', label: 'Territories', icon: <MapPinned size={18} />, count: territories.length },
    { key: 'rules', label: 'Business Rules', icon: <Sliders size={18} />, count: 0 },
  ]

  if (usersLoading) return <Spinner />

  return (
    <div className="flex h-full flex-col md:flex-row">
      {/* Mobile tab bar */}
      <div className="flex gap-1 overflow-x-auto border-b border-slate-200 bg-white px-3 py-2 md:hidden">
        {sections.map((s) => (
          <button
            key={s.key}
            onClick={() => setSection(s.key)}
            className={`flex shrink-0 items-center gap-1.5 rounded-lg px-3 py-2 text-sm transition-colors ${
              section === s.key
                ? 'bg-teal-50 text-teal-700 font-medium'
                : 'text-slate-600 hover:bg-slate-50'
            }`}
          >
            {s.icon}
            {s.label}
          </button>
        ))}
      </div>

      {/* Desktop sub-nav */}
      <div className="hidden w-56 border-r border-slate-200 bg-white p-3 space-y-1 md:block">
        <h2 className="px-3 py-2 text-xs font-semibold uppercase text-slate-400">Configuration</h2>
        {sections.map((s) => (
          <button
            key={s.key}
            onClick={() => setSection(s.key)}
            className={`flex w-full items-center gap-2 rounded-lg px-3 py-2 text-sm transition-colors ${
              section === s.key
                ? 'bg-teal-50 text-teal-700 font-medium'
                : 'text-slate-600 hover:bg-slate-50'
            }`}
          >
            {s.icon}
            {s.label}
            <span className="ml-auto text-xs text-slate-400">{s.count || ''}</span>
          </button>
        ))}
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto p-4 md:p-6">
        {section === 'users' && (
          <div>
            <h1 className="mb-4 text-xl font-bold text-slate-900">Users & Roles</h1>
            <DataTable data={users} columns={userColumns} />
          </div>
        )}

        {section === 'teams' && (
          <div>
            <h1 className="mb-4 text-xl font-bold text-slate-900">Teams</h1>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {teams.map((team) => (
                <Card key={team.id}>
                  <h3 className="font-semibold text-slate-900">{team.name}</h3>
                  <p className="mt-1 text-sm text-slate-500">Manager: {team.managerId}</p>
                </Card>
              ))}
              {teams.length === 0 && <p className="text-sm text-slate-400">No teams configured.</p>}
            </div>
          </div>
        )}

        {section === 'territories' && (
          <div>
            <h1 className="mb-4 text-xl font-bold text-slate-900">Territories</h1>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {territories.map((t) => (
                <Card key={t.id}>
                  <h3 className="font-semibold text-slate-900">{t.name}</h3>
                  <p className="mt-1 text-sm text-slate-500">
                    {t.region && <Badge>{t.region}</Badge>}
                  </p>
                  <p className="mt-1 text-xs text-slate-400">Team: {t.teamId}</p>
                </Card>
              ))}
              {territories.length === 0 && <p className="text-sm text-slate-400">No territories configured.</p>}
            </div>
          </div>
        )}

        {section === 'rules' && config && (
          <div className="space-y-6">
            <h1 className="text-xl font-bold text-slate-900">Business Rules</h1>

            <Card>
              <h3 className="mb-3 text-sm font-semibold text-slate-900">Visit Frequency Requirements</h3>
              <div className="space-y-2">
                {Object.entries(config.rules.frequency).map(([classification, required]) => (
                  <div key={classification} className="flex items-center justify-between">
                    <Badge variant={classification === 'A' ? 'danger' : classification === 'B' ? 'warning' : 'default'}>
                      Class {classification}
                    </Badge>
                    <span className="text-sm text-slate-700">{required} visits/period</span>
                  </div>
                ))}
              </div>
            </Card>

            <Card>
              <h3 className="mb-3 text-sm font-semibold text-slate-900">Activity Rules</h3>
              <dl className="space-y-2">
                <div className="flex justify-between">
                  <dt className="text-sm text-slate-500">Max activities per day</dt>
                  <dd className="text-sm font-medium text-slate-900">{config.rules.max_activities_per_day}</dd>
                </div>
                {config.rules.visit_cadence_days && (
                  <div className="flex justify-between">
                    <dt className="text-sm text-slate-500">Visit cadence (days)</dt>
                    <dd className="text-sm font-medium text-slate-900">{config.rules.visit_cadence_days}</dd>
                  </div>
                )}
              </dl>
            </Card>

            <Card>
              <h3 className="mb-3 text-sm font-semibold text-slate-900">Activity Types</h3>
              <div className="space-y-2">
                {config.activities.types.map((type) => (
                  <div key={type.key} className="flex items-center justify-between rounded-lg border border-slate-100 p-3">
                    <div>
                      <span className="text-sm font-medium text-slate-900">{type.label}</span>
                      <Badge variant={type.category === 'field' ? 'warning' : 'primary'} className="ml-2">
                        {type.category}
                      </Badge>
                    </div>
                    {type.blocks_field_activities && (
                      <Badge variant="danger">Blocks field</Badge>
                    )}
                  </div>
                ))}
              </div>
            </Card>

            <Card>
              <h3 className="mb-3 text-sm font-semibold text-slate-900">Status Workflow</h3>
              <div className="space-y-2">
                {config.activities.statuses.map((s) => (
                  <div key={s.key} className="flex items-center gap-2">
                    <span className="text-sm text-slate-700">{s.label}</span>
                    {s.initial && <Badge variant="primary">Initial</Badge>}
                    {s.submittable && <Badge variant="success">Submittable</Badge>}
                    {config.activities.status_transitions[s.key] && (
                      <span className="text-xs text-slate-400">
                        &rarr; {config.activities.status_transitions[s.key].join(', ')}
                      </span>
                    )}
                  </div>
                ))}
              </div>
            </Card>
          </div>
        )}
      </div>
    </div>
  )
}
