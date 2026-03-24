import { Loader2, Check, AlertCircle } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import type { SaveState } from '../hooks/useInlineActivityEditor'

interface SaveStateIndicatorProps {
  saveState: SaveState
  onRetry?: () => void
}

export function SaveStateIndicator({ saveState, onRetry }: SaveStateIndicatorProps) {
  const { t } = useTranslation()

  if (saveState === 'idle') return null

  if (saveState === 'dirty') {
    return (
      <span className="inline-flex items-center gap-1 text-xs text-slate-400">
        <span className="w-1.5 h-1.5 rounded-full bg-slate-400 animate-pulse" />
        {t('saveState.saving')}
      </span>
    )
  }

  if (saveState === 'saving') {
    return (
      <span className="inline-flex items-center gap-1 text-xs text-slate-400">
        <Loader2 className="w-3 h-3 animate-spin" />
        {t('saveState.saving')}
      </span>
    )
  }

  if (saveState === 'error') {
    return (
      <button
        type="button"
        onClick={onRetry}
        className="inline-flex items-center gap-1 text-xs text-red-500 hover:text-red-600 transition-colors"
        data-testid="save-state-retry"
      >
        <AlertCircle className="w-3 h-3" />
        {t('saveState.notSaved')}
      </button>
    )
  }

  return null
}

/** Flash the checkmark briefly after save succeeds. */
export function SaveSuccessFlash() {
  const { t } = useTranslation()
  return (
    <span className="inline-flex items-center gap-1 text-xs text-emerald-500">
      <Check className="w-3 h-3" />
      {t('saveState.saved')}
    </span>
  )
}
