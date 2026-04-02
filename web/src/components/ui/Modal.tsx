import { useEffect, type ReactNode } from 'react'
import { X } from 'lucide-react'

interface ModalProps {
  readonly open: boolean
  readonly onClose: () => void
  readonly title: string
  readonly children: ReactNode
  readonly footer?: ReactNode
}

export function Modal({ open, onClose, title, children, footer }: ModalProps) {
  useEffect(() => {
    if (!open) return
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    document.addEventListener('keydown', handleKey)
    document.body.style.overflow = 'hidden'
    return () => {
      document.removeEventListener('keydown', handleKey)
      document.body.style.overflow = ''
    }
  }, [open, onClose])

  if (!open) return null

  return (
    <dialog open aria-label={title} className="fixed inset-0 z-50 flex items-end justify-center sm:items-center bg-transparent m-0 p-0 w-full h-full max-w-none max-h-none border-none">
      <button type="button" className="fixed inset-0 bg-black/40 border-none cursor-default w-full h-full" onClick={onClose} aria-label="Close dialog" />
      <div className="relative flex max-h-[90vh] w-full flex-col rounded-t-2xl bg-white shadow-xl sm:max-w-lg sm:rounded-2xl">
        <div className="flex items-center justify-between border-b border-slate-200 px-4 py-3">
          <h2 className="text-base font-semibold text-slate-900">{title}</h2>
          <button
            onClick={onClose}
            className="rounded-lg p-1.5 text-slate-400 hover:bg-slate-100"
            aria-label="Close"
          >
            <X size={18} />
          </button>
        </div>
        <div className="flex-1 overflow-auto px-4 py-4">
          {children}
        </div>
        {footer && (
          <div className="border-t border-slate-200 px-4 py-3">
            {footer}
          </div>
        )}
      </div>
    </dialog>
  )
}
