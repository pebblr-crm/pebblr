import { MapContainer } from '@/components/map/MapContainer'
import { TargetMarker } from '@/components/map/TargetMarker'
import { getLat, getLng, getClassification } from '@/lib/target-fields'
import { X } from 'lucide-react'
import type { Target } from '@/types/target'

export interface PlannerMobileMapProps {
  readonly geoTargets: Target[]
  readonly selectedTargetIds: Set<string>
  readonly onToggleTarget: (id: string) => void
  readonly onClose: () => void
}

export function PlannerMobileMap({ geoTargets, selectedTargetIds, onToggleTarget, onClose }: PlannerMobileMapProps) {
  return (
    <div className="fixed inset-0 z-50 flex flex-col bg-white lg:hidden">
      <div className="flex items-center justify-between border-b border-slate-200 px-4 py-3">
        <h2 className="text-sm font-semibold text-slate-900">Target Map</h2>
        <button
          onClick={onClose}
          className="rounded-lg p-1.5 text-slate-400 hover:bg-slate-100"
          aria-label="Close map"
        >
          <X size={20} />
        </button>
      </div>
      <div className="flex-1">
        <MapContainer className="h-full">
          {geoTargets.map((t) => (
            <TargetMarker
              key={t.id}
              lat={getLat(t.fields)!}
              lng={getLng(t.fields)!}
              name={t.name}
              priority={getClassification(t.fields)}
              selected={selectedTargetIds.has(t.id)}
              onClick={() => onToggleTarget(t.id)}
            />
          ))}
        </MapContainer>
      </div>
    </div>
  )
}
