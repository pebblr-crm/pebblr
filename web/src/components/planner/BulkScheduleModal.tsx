import { useState, useCallback } from 'react'
import { Modal } from '@/components/ui/Modal'
import { Button } from '@/components/ui/Button'
import { priorityDot, classificationBadge } from '@/lib/styles'
import { getClassification } from '@/lib/target-fields'
import { Info, CalendarPlus } from 'lucide-react'
import type { Target } from '@/types/target'

export interface BulkScheduleModalProps {
  readonly open: boolean
  readonly onClose: () => void
  readonly selectedTargetIds: Set<string>
  readonly targetMap: Map<string, Target>
  readonly initialDate: string
  readonly onSchedule: (date: string, visitType: 'f2f' | 'remote') => Promise<void>
  readonly isPending: boolean
}

export function BulkScheduleModal({
  open,
  onClose,
  selectedTargetIds,
  targetMap,
  initialDate,
  onSchedule,
  isPending,
}: BulkScheduleModalProps) {
  const [bulkDate, setBulkDate] = useState(initialDate)
  const [bulkVisitType, setBulkVisitType] = useState<'f2f' | 'remote'>('f2f')

  const handleSchedule = useCallback(() => {
    onSchedule(bulkDate, bulkVisitType)
  }, [onSchedule, bulkDate, bulkVisitType])

  const targetPlural = selectedTargetIds.size === 1 ? '' : 's'
  const buttonLabel = isPending
    ? 'Creating...'
    : `Schedule ${selectedTargetIds.size} Target${targetPlural}`

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Bulk Schedule"
      footer={
        <div className="flex items-center justify-between">
          <span className="text-[11px] text-slate-400">
            <Info size={12} className="inline text-slate-400 mr-1" />You can also drag and drop targets onto the calendar
          </span>
          <div className="flex items-center gap-2">
            <Button variant="secondary" size="sm" onClick={onClose}>
              Cancel
            </Button>
            <Button variant="primary" size="sm" onClick={handleSchedule} disabled={isPending}>
              <CalendarPlus size={14} />
              {buttonLabel}
            </Button>
          </div>
        </div>
      }
    >
      <div className="space-y-4">
        <p className="text-sm text-slate-600">
          Pick a date to schedule <strong>{selectedTargetIds.size}</strong> selected target{selectedTargetIds.size === 1 ? '' : 's'}.
        </p>
        <input
          type="date"
          value={bulkDate}
          onChange={(e) => setBulkDate(e.target.value)}
          className="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:border-teal-500 focus:outline-none focus:ring-1 focus:ring-teal-500"
        />
        <div>
          <fieldset>
            <legend className="mb-1.5 block text-sm font-medium text-slate-700">Visit type</legend>
            <div className="flex rounded-lg border border-slate-200 overflow-hidden">
              <button
                type="button"
                onClick={() => setBulkVisitType('f2f')}
                className={`flex-1 px-3 py-2 text-sm font-medium transition-colors ${
                  bulkVisitType === 'f2f'
                    ? 'bg-teal-600 text-white'
                    : 'bg-white text-slate-600 hover:bg-slate-50'
                }`}
              >
                Face to face
              </button>
              <button
                type="button"
                onClick={() => setBulkVisitType('remote')}
                className={`flex-1 px-3 py-2 text-sm font-medium border-l border-slate-200 transition-colors ${
                  bulkVisitType === 'remote'
                    ? 'bg-teal-600 text-white'
                    : 'bg-white text-slate-600 hover:bg-slate-50'
                }`}
              >
                Remote
              </button>
            </div>
          </fieldset>
        </div>
        <div className="max-h-48 overflow-y-auto rounded-lg border border-slate-200">
          <ul className="divide-y divide-slate-100">
            {Array.from(selectedTargetIds).map((id) => {
              const t = targetMap.get(id)
              if (!t) return null
              const p = getClassification(t.fields)
              const badge = classificationBadge[p] ?? classificationBadge.c
              return (
                <li key={id} className="flex items-center gap-2 px-3 py-2">
                  <span className={`w-2 h-2 rounded-full shrink-0 ${priorityDot[p] ?? priorityDot.c}`} />
                  <span className="text-sm text-slate-800 truncate flex-1">{t.name}</span>
                  <span className={`text-[9px] font-bold uppercase px-1.5 py-0.5 rounded shrink-0 ${badge}`}>{p.toUpperCase()}</span>
                </li>
              )
            })}
          </ul>
        </div>
      </div>
    </Modal>
  )
}
