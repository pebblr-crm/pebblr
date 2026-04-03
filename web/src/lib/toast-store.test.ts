import { onToast, emitToast } from './toast-store'

describe('toast-store', () => {
  it('delivers a toast to the listener', () => {
    const listener = vi.fn()
    const unsubscribe = onToast(listener)

    emitToast('Something failed', 'error')
    expect(listener).toHaveBeenCalledWith('Something failed', 'error')

    unsubscribe()
  })

  it('defaults variant to error', () => {
    const listener = vi.fn()
    const unsubscribe = onToast(listener)

    emitToast('Oops')
    expect(listener).toHaveBeenCalledWith('Oops', 'error')

    unsubscribe()
  })

  it('does not deliver after unsubscribe', () => {
    const listener = vi.fn()
    const unsubscribe = onToast(listener)
    unsubscribe()

    emitToast('Should not arrive')
    expect(listener).not.toHaveBeenCalled()
  })

  it('does nothing when no listener is registered', () => {
    // Should not throw
    expect(() => emitToast('No listener')).not.toThrow()
  })

  it('replaces previous listener on new subscription', () => {
    const first = vi.fn()
    const second = vi.fn()
    const unsub1 = onToast(first)
    const unsub2 = onToast(second)

    emitToast('Hello', 'info')
    expect(first).not.toHaveBeenCalled()
    expect(second).toHaveBeenCalledWith('Hello', 'info')

    unsub1()
    unsub2()
  })
})
