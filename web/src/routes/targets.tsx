import { useState, useMemo } from 'react'
import { createRoute } from '@tanstack/react-router'
import { createColumnHelper } from '@tanstack/react-table'
import { Route as rootRoute } from './__root'
import { useTargets, useTargetFrequencyStatus } from '@/hooks/useTargets'
import { DataTable } from '@/components/data/DataTable'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { MapContainer } from '@/components/map/MapContainer'
import { TargetMarker } from '@/components/map/TargetMarker'
import { getLat, getLng, getClassification, getCity } from '@/lib/target-fields'
import { Search, Filter, ExternalLink } from 'lucide-react'
import { useNavigate } from '@tanstack/react-router'
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

function TargetsPage() {
  const navigate = useNavigate()
  const [search, setSearch] = useState('')
  const [typeFilter, setTypeFilter] = useState('')
  const { data, isLoading } = useTargets({ q: search || undefined, type: typeFilter || undefined, limit: 200 })
  const { data: freqData } = useTargetFrequencyStatus()
  const targets = useMemo(() => data?.items ?? [], [data])

  const freqMap = useMemo(() => {
    const m = new Map<string, number>()
    freqData?.items?.forEach((f) => m.set(f.targetId, f.compliance))
    return m
  }, [freqData])

  const columns = useMemo(
    () => [
      columnHelper.accessor('name', {
        header: 'Name',
        cell: (info) => (
          <a
            href={`/targets/${info.row.original.id}`}
            className="font-medium text-slate-900 hover:text-teal-600 hover:underline inline-flex items-center gap-1"
            onClick={(e) => { e.preventDefault(); navigate({ to: '/targets/$id', params: { id: info.row.original.id } }) }}
          >
            {info.getValue()}
            <ExternalLink size={12} className="text-slate-400" />
          </a>
        ),
      }),
      columnHelper.accessor((row) => getClassification(row.fields), {
        id: 'priority',
        header: 'Priority',
        cell: (info) => (
          <Badge variant={priorityVariant[info.getValue()] ?? 'default'}>
            {info.getValue().toUpperCase()}
          </Badge>
        ),
      }),
      columnHelper.accessor((row) => getCity(row.fields), {
        id: 'city',
        header: 'City',
      }),
      columnHelper.accessor('targetType', {
        header: 'Type',
        cell: (info) => <span className="capitalize">{info.getValue()}</span>,
      }),
      columnHelper.accessor((row) => freqMap.get(row.id), {
        id: 'compliance',
        header: 'Compliance',
        cell: (info) => {
          const v = info.getValue()
          if (v == null) return <span className="text-slate-400">-</span>
          const pct = Math.round(v)
          const color = pct >= 80 ? 'text-emerald-600' : pct >= 50 ? 'text-amber-600' : 'text-red-600'
          return <span className={`font-medium ${color}`}>{pct}%</span>
        },
      }),
    ],
    [freqMap, navigate],
  )

  const geoTargets = useMemo(
    () =>
      targets.filter((t) => getLat(t.fields) != null && getLng(t.fields) != null),
    [targets],
  )

  if (isLoading) return <Spinner />

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
