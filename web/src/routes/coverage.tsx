import { useState, useMemo, useCallback } from 'react'
import { createRoute } from '@tanstack/react-router'
import { Route as rootRoute } from './__root'
import { useTargets } from '@/hooks/useTargets'
import { useTerritories } from '@/hooks/useTerritories'
import { useTeams } from '@/hooks/useTeams'
import { useCoverage } from '@/hooks/useDashboard'
import { MapContainer } from '@/components/map/MapContainer'
import { TargetMarker } from '@/components/map/TargetMarker'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { RotateCcw, SlidersHorizontal, X } from 'lucide-react'

export const Route = createRoute({
  getParentRoute: () => rootRoute,
  path: '/coverage',
  component: CoveragePage,
})

function getLat(fields: Record<string, unknown>): number | null {
  const v = fields.lat
  return typeof v === 'number' ? v : null
}

function getLng(fields: Record<string, unknown>): number | null {
  const v = fields.lng
  return typeof v === 'number' ? v : null
}

function getClassification(fields: Record<string, unknown>): string {
  return ((fields.potential as string) ?? 'c').toLowerCase()
}

function CoveragePage() {
  const [teamFilter, setTeamFilter] = useState('')
  const [priorityFilter, setPriorityFilter] = useState('')
  const [filtersOpen, setFiltersOpen] = useState(false)

  const { data: targetData, isLoading: targetsLoading } = useTargets({ limit: 1000 })
  const { data: territoryData } = useTerritories()
  const { data: teamsData } = useTeams()
  const { data: coverage } = useCoverage({})

  const targets = useMemo(() => targetData?.items ?? [], [targetData])
  const territories = useMemo(() => territoryData?.items ?? [], [territoryData])
  const teams = useMemo(() => teamsData?.items ?? [], [teamsData])

  const filteredTargets = useMemo(() => {
    let result = targets
    if (teamFilter) result = result.filter((t) => t.teamId === teamFilter)
    if (priorityFilter) result = result.filter((t) => getClassification(t.fields) === priorityFilter)
    return result
  }, [targets, teamFilter, priorityFilter])

  const geoTargets = useMemo(
    () => filteredTargets.filter((t) => getLat(t.fields) != null && getLng(t.fields) != null),
    [filteredTargets],
  )

  const resetFilters = useCallback(() => {
    setTeamFilter('')
    setPriorityFilter('')
  }, [])

  if (targetsLoading) return <Spinner />

  const filterContent = (
    <>
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold text-slate-900">Filters</h2>
        <div className="flex items-center gap-1">
          <Button variant="ghost" size="sm" onClick={resetFilters}>
            <RotateCcw size={12} />
            Reset
          </Button>
          <button
            onClick={() => setFiltersOpen(false)}
            className="rounded-lg p-1.5 text-slate-400 hover:bg-slate-100 md:hidden"
            aria-label="Close filters"
          >
            <X size={18} />
          </button>
        </div>
      </div>

      {/* Teams */}
      <div>
        <label className="mb-2 block text-xs font-medium text-slate-500 uppercase">Team</label>
        <div className="space-y-1">
          {teams.map((team) => (
            <label key={team.id} className="flex items-center gap-2 rounded px-2 py-1.5 text-sm hover:bg-slate-50 cursor-pointer">
              <input
                type="radio"
                name="team"
                value={team.id}
                checked={teamFilter === team.id}
                onChange={() => setTeamFilter(teamFilter === team.id ? '' : team.id)}
                className="text-teal-600"
              />
              {team.name}
            </label>
          ))}
        </div>
      </div>

      {/* Priority */}
      <div>
        <label className="mb-2 block text-xs font-medium text-slate-500 uppercase">Priority</label>
        <div className="flex gap-2">
          {['a', 'b', 'c'].map((p) => (
            <button
              key={p}
              onClick={() => setPriorityFilter(priorityFilter === p ? '' : p)}
              className={`rounded-lg border px-3 py-1.5 text-sm font-medium transition-colors ${
                priorityFilter === p
                  ? 'border-teal-500 bg-teal-50 text-teal-700'
                  : 'border-slate-200 text-slate-600 hover:border-slate-300'
              }`}
            >
              {p.toUpperCase()}
            </button>
          ))}
        </div>
      </div>

      {/* Coverage summary */}
      {coverage && (
        <Card>
          <h3 className="text-xs font-medium text-slate-500 uppercase mb-2">Coverage</h3>
          <p className="text-2xl font-semibold text-slate-900">{Math.round(coverage.percentage)}%</p>
          <p className="text-xs text-slate-500">{coverage.visitedTargets} of {coverage.totalTargets} visited</p>
          <div className="mt-2 h-2 rounded-full bg-slate-100">
            <div
              className="h-2 rounded-full bg-teal-500 transition-all"
              style={{ width: `${Math.min(100, coverage.percentage)}%` }}
            />
          </div>
        </Card>
      )}

      {/* Territories */}
      <div>
        <h3 className="mb-2 text-xs font-medium text-slate-500 uppercase">Territories ({territories.length})</h3>
        <div className="space-y-1">
          {territories.map((t) => (
            <div key={t.id} className="rounded px-2 py-1.5 text-sm text-slate-700 hover:bg-slate-50">
              <div className="font-medium">{t.name}</div>
              {t.region && <div className="text-xs text-slate-400">{t.region}</div>}
            </div>
          ))}
          {territories.length === 0 && (
            <p className="text-xs text-slate-400">No territories defined.</p>
          )}
        </div>
      </div>

      {/* Stats */}
      <Card>
        <div className="flex items-center justify-between text-sm">
          <span className="text-slate-500">Showing</span>
          <span className="font-medium text-slate-900">{geoTargets.length} pins</span>
        </div>
        <div className="mt-1 flex items-center justify-between text-sm">
          <span className="text-slate-500">Total targets</span>
          <span className="font-medium text-slate-900">{filteredTargets.length}</span>
        </div>
      </Card>
    </>
  )

  return (
    <div className="flex h-full flex-col md:flex-row">
      {/* Mobile filter toggle */}
      <div className="flex items-center gap-2 border-b border-slate-200 bg-white px-4 py-3 md:hidden">
        <Button variant="secondary" size="sm" onClick={() => setFiltersOpen(true)}>
          <SlidersHorizontal size={14} />
          Filters
        </Button>
        {coverage && (
          <span className="text-sm text-slate-500">Coverage: {Math.round(coverage.percentage)}%</span>
        )}
        <span className="ml-auto text-sm text-slate-500">{geoTargets.length} pins</span>
      </div>

      {/* Mobile filter overlay */}
      {filtersOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/40 md:hidden"
          onClick={() => setFiltersOpen(false)}
        />
      )}

      {/* Filter sidebar */}
      <div
        className={`fixed inset-y-0 left-0 z-50 w-72 transform overflow-auto border-r border-slate-200 bg-white p-4 space-y-4 transition-transform duration-200 ease-in-out md:static md:translate-x-0 ${
          filtersOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        {filterContent}
      </div>

      {/* Map */}
      <div className="flex-1">
        <MapContainer className="h-full">
          {geoTargets.map((t) => (
            <TargetMarker
              key={t.id}
              lat={getLat(t.fields)!}
              lng={getLng(t.fields)!}
              name={t.name}
              priority={getClassification(t.fields)}
            />
          ))}
        </MapContainer>
      </div>
    </div>
  )
}
