import { useEffect } from 'react'
import { useToast } from './Toast'
import { onToast } from '@/lib/toast-store'

/**
 * Renders a global toast container that listens to the toast-store.
 * Mount once near the app root so mutations and other non-component code
 * can surface errors via emitToast().
 */
export function GlobalToast() {
  const { showToast, ToastContainer } = useToast()

  useEffect(() => {
    const unsubscribe = onToast((message, variant) => {
      showToast(message, variant)
    })
    return unsubscribe
  }, [showToast])

  return <ToastContainer />
}
