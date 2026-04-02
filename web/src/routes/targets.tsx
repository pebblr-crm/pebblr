import { useState, useMemo, useCallback } from 'react'
import { createRoute, useNavigate } from '@tanstack/react-router'
import { createColumnHelper } from '@tanstack/react-table'
import { Route as rootRoute } from './__root'
import { useTargets, useTargetFrequencyStatus } from '@/hooks/useTargets'
import { DataTable } from '@/components/data/DataTable'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { QueryError } from '@/components/ui/QueryError'
import { MapContainer } from '@/components/map/MapContainer'
import { TargetMarker } from '@/components/map/TargetMarker'
import { getLat, getLng, getClassification, getCity } from '@/lib/target-fields'
import { Search, Filter, ExternalLink } from 'lucide-react'
import type { Target } from '@/types/target'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/targets',
  component: TargetsPage,
})

const columnHelper = createColumnHelper<Target>()

const priorityVariant: Record<string, 'danger' | 'warning' | 'default'> = {
  a: 'danger',
  b: 'warning',
  c: 'default',
}

function TargetNameCell({ name, id, onNavigate }: Readonly<{ name: string; id: string; onNavigate: (id: string) => void }>) {
  return (
    <a
      href={`/targets/${id}`}
      className="font-medium text-slate-900 hover:text-teal-600 hover:underline inline-flex items-center gap-1"
      onClick={(e) => { e.preventDefault(); onNavigate(id) }}
    >
      {name}
      <ExternalLink size={12} className="text-slate-400" />
    </a>
  )
}

function PriorityCell({ getValue }: Readonly<{ getValue: () => string }>) {
  return (
    <Badge variant={priorityVariant[getValue()] ?? 'default'}>
      {getValue().toUpperCase()}
    </Badge>
  )
}

function TargetTypeCell({ getValue }: Readonly<{ getValue: () => string }>) {
  return <span className="capitalize">{getValue()}</span>
}

function complianceColor(pct: number): string {
  if (pct >= 80) return 'text-emerald-600'
  if (pct >= 50) return 'text-amber-600'
  return 'text-red-600'
}

function TargetComplianceCell({ getValue }: Readonly<{ getValue: () => number | undefined }>) {
  const v = getValue()
  if (v == null) return <span className="text-slate-400">-</span>
  const pct = Math.round(v)
  return <span className={`font-medium ${complianceColor(pct)}`}>{pct}%</span>
}

function renderPriorityCell(info: { getValue: () => unknown }) { return <PriorityCell getValue={() => String(info.getValue())} /> }
function renderTargetTypeCell(info: { getValue: () => string }) { return <TargetTypeCell getValue={info.getValue} /> }
function renderTargetComplianceCell(info: { getValue: () => unknown }) { return <TargetComplianceCell getValue={() => info.getValue() as number | undefined} /> }

function TargetsPage() {
  const navigate = useNavigate()
  const [search, setSearch] = useState('')
  const [typeFilter, setTypeFilter] = useState('')
  const { data, isLoading, isError, refetch } = useTargets({ q: search || undefined, type: typeFilter || undefined, limit: 200 })
  const { data: freqData } = useTargetFrequencyStatus()
  const targets = useMemo(() => data?.items ?? [], [data])

  const freqMap = useMemo(() => {
    const m = new Map<string, number>()
    freqData?.items?.forEach((f) => m.set(f.targetId, f.compliance))
    return m
  }, [freqData])

  const renderNameCell = useCallback(
    (info: { getValue: () => string; row: { original: { id: string } } }) => (
      <TargetNameCell
        name={info.getValue()}
        id={info.row.original.id}
        onNavigate={(id) => navigate({ to: '/targets/$id', params: { id } })}
      />
    ),
    [navigate],
  )

  const columns = useMemo(
    () => [
      columnHelper.accessor('name', {
        header: 'Name',
        cell: renderNameCell,
      }),
      columnHelper.accessor((row) => getClassification(row.fields), {
        id: 'priority',
        header: 'Priority',
        cell: renderPriorityCell,
      }),
      columnHelper.accessor((row) => getCity(row.fields), {
        id: 'city',
        header: 'City',
      }),
      columnHelper.accessor('targetType', {
        header: 'Type',
        cell: renderTargetTypeCell,
      }),
      columnHelper.accessor((row) => freqMap.get(row.id), {
        id: 'compliance',
        header: 'Compliance',
        cell: renderTargetComplianceCell,
      }),
    ],
    [freqMap, renderNameCell],
  )

  const geoTargets = useMemo(
    () =>
      targets.filter((t) => getLat(t.fields) != null && getLng(t.fields) != null),
    [targets],
  )

  if (isLoading) return <Spinner />
  if (isError) return <QueryError message="Failed to load targets" onRetry={() => { refetch() }} />

  return (
    <div className="flex h-full">
      {/* Table panel */}
      <div className="flex-1 flex flex-col min-h-0">
        {/* Sticky header + filters */}
        <div className="shrink-0 p-4 md:px-6 md:pt-6 md:pb-0 space-y-4">
          <div className="flex items-center justify-between">
            <h1 className="text-2xl font-bold text-slate-900">Target Portfolio</h1>
            <div className="flex items-center gap-2">
              <span className="text-sm text-slate-500">{targets.length} targets</span>
            </div>
          </div>

          {/* Filters */}
          <div className="flex flex-wrap items-center gap-3">
            <div className="relative w-full max-w-xs sm:flex-1">
              <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
              <input
                type="text"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Search targets..."
                className="w-full rounded-lg border border-slate-300 py-2 pl-9 pr-3 text-sm focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
              />
            </div>
            <select
              value={typeFilter}
              onChange={(e) => setTypeFilter(e.target.value)}
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-teal-500 focus:outline-none"
            >
              <option value="">All types</option>
              <option value="doctor">Doctor</option>
              <option value="pharmacy">Pharmacy</option>
              <option value="hospital">Hospital</option>
            </select>
            <Button variant="secondary" size="sm">
              <Filter size={14} />
              More filters
            </Button>
          </div>
        </div>

        {/* Scrollable table area */}
        <div className="flex-1 overflow-auto p-4 md:px-6 md:pb-6">
          <DataTable data={targets} columns={columns} />
        </div>
      </div>

      {/* Map sidebar */}
      <div className="hidden w-96 border-l border-slate-200 lg:block">
        <MapContainer className="h-full">
          {geoTargets.map((t) => (
            <TargetMarker
              key={t.id}
              lat={getLat(t.fields)!}
              lng={getLng(t.fields)!}
              name={t.name}
              priority={getClassification(t.fields)}
              onClick={() => navigate({ to: '/targets/$id', params: { id: t.id } })}
            />
          ))}
        </MapContainer>
      </div>
    </div>
  )
}
