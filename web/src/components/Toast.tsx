import { useCallback, useMemo, useState, type ReactNode } from 'react'
import { AnimatePresence, motion } from 'motion/react'
import { AlertCircle, CheckCircle, X } from 'lucide-react'
import { ToastContext } from '../contexts/toast'

// ── Types ────────────────────────────────────────────────────────────────────

type ToastVariant = 'error' | 'success'

const VARIANT_STYLES: Record<ToastVariant, { container: string; icon: string; dismiss: string }> = {
  error: {
    container: 'bg-red-50 border-red-200 text-red-800',
    icon: 'text-red-500',
    dismiss: 'hover:bg-red-100',
  },
  success: {
    container: 'bg-emerald-50 border-emerald-200 text-emerald-800',
    icon: 'text-emerald-500',
    dismiss: 'hover:bg-emerald-100',
  },
}

interface Toast {
  id: number
  message: string
  variant: ToastVariant
}

// ── Provider ─────────────────────────────────────────────────────────────────

let nextId = 0
const AUTO_DISMISS_MS = 5000

export function ToastProvider({ children }: Readonly<{ children: ReactNode }>) {
  const [toasts, setToasts] = useState<Toast[]>([])

  const dismiss = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id))
  }, [])

  const showToast = useCallback(
    (message: string, variant: ToastVariant = 'error') => {
      const id = nextId++
      setToasts((prev) => [...prev, { id, message, variant }])
      setTimeout(() => dismiss(id), AUTO_DISMISS_MS)
    },
    [dismiss],
  )

  const toastContextValue = useMemo(() => ({ showToast }), [showToast])

  return (
    <ToastContext.Provider value={toastContextValue}>
      {children}

      {/* Toast container — fixed top-right */}
      <div className="fixed top-4 right-4 z-[100] flex flex-col gap-2 pointer-events-none max-w-sm w-full">
        <AnimatePresence>
          {toasts.map((t) => {
            const style = VARIANT_STYLES[t.variant]
            const Icon = t.variant === 'success' ? CheckCircle : AlertCircle
            return (
              <motion.div
                key={t.id}
                initial={{ opacity: 0, y: -12, scale: 0.95 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, y: -12, scale: 0.95 }}
                transition={{ duration: 0.2 }}
                className={`pointer-events-auto flex items-start gap-3 px-4 py-3 rounded-xl shadow-lg border text-sm ${style.container}`}
                role="alert"
              >
                <Icon className={`w-4 h-4 mt-0.5 flex-shrink-0 ${style.icon}`} />
                <p className="flex-1">{t.message}</p>
                <button
                  type="button"
                  onClick={() => dismiss(t.id)}
                  className={`flex-shrink-0 p-0.5 rounded transition-colors ${style.dismiss}`}
                  aria-label="Dismiss"
                >
                  <X className="w-3.5 h-3.5" />
                </button>
              </motion.div>
            )
          })}
        </AnimatePresence>
      </div>
    </ToastContext.Provider>
  )
}
