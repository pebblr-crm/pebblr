import { createContext } from 'react'

type ToastVariant = 'error'

export interface ToastContextValue {
  showToast: (message: string, variant?: ToastVariant) => void
}

export const ToastContext = createContext<ToastContextValue | null>(null)
