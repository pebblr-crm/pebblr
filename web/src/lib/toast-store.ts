type ToastVariant = 'info' | 'warning' | 'error'
type ToastListener = (message: string, variant: ToastVariant) => void

let listener: ToastListener | null = null

/**
 * Subscribe to toast events. Only one listener at a time (the ToastContainer).
 * Returns an unsubscribe function.
 */
export function onToast(fn: ToastListener): () => void {
  listener = fn
  return () => {
    if (listener === fn) listener = null
  }
}

/**
 * Emit a toast from anywhere (including outside React components).
 * Used by the global MutationCache onError handler.
 */
export function emitToast(message: string, variant: ToastVariant = 'error'): void {
  listener?.(message, variant)
}
