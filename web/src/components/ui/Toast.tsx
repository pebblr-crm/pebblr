import { useState, useCallback, useRef } from 'react'

interface Toast {
  id: number
  message: string
  variant: 'info' | 'warning' | 'error'
  leaving: boolean
}

const variantStyles: Record<Toast['variant'], string> = {
  info: 'bg-slate-800 text-white',
  warning: 'bg-amber-600 text-white',
  error: 'bg-red-600 text-white',
}

const SHOW_MS = 2500
const EXIT_MS = 300

export function useToast() {
  const [toasts, setToasts] = useState<Toast[]>([])
  const nextId = useRef(0)

  const showToast = useCallback((message: string, variant: Toast['variant'] = 'info') => {
    const id = nextId.current++
    setToasts((prev) => [...prev, { id, message, variant, leaving: false }])

    // Start exit animation
    setTimeout(() => {
      setToasts((prev) => prev.map((t) => (t.id === id ? { ...t, leaving: true } : t)))
    }, SHOW_MS)

    // Remove from DOM after animation completes
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id))
    }, SHOW_MS + EXIT_MS)
  }, [])

  const ToastContainer = useCallback(() => {
    if (toasts.length === 0) return null
    return (
      <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-[100] flex flex-col gap-2 items-center pointer-events-none">
        {toasts.map((t) => (
          <div
            key={t.id}
            style={{
              animation: t.leaving
                ? `toast-exit ${EXIT_MS}ms ease-in forwards`
                : `toast-enter 300ms ease-out`,
            }}
            className={`${variantStyles[t.variant]} px-4 py-2.5 rounded-lg shadow-lg text-sm font-medium pointer-events-auto`}
          >
            {t.message}
          </div>
        ))}
        <style>{`
          @keyframes toast-enter {
            from { opacity: 0; transform: translateY(16px) scale(0.95); }
            to   { opacity: 1; transform: translateY(0) scale(1); }
          }
          @keyframes toast-exit {
            from { opacity: 1; transform: translateY(0) scale(1); }
            to   { opacity: 0; transform: translateY(8px) scale(0.95); }
          }
        `}</style>
      </div>
    )
  }, [toasts])

  return { showToast, ToastContainer }
}
