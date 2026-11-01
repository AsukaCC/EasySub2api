import { mount } from '@vue/test-utils'
import { defineComponent, nextTick, ref } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { TOTP } from 'otpauth'
import { useAccountTotp } from '../useAccountTotp'

describe('account TOTP lifecycle', () => {
  afterEach(() => { vi.useRealTimers(); vi.restoreAllMocks() })
  it('computes once per period and stops on hidden, close, and unmount', async () => {
    vi.useFakeTimers()
    vi.setSystemTime(59_000)
    const hidden = vi.spyOn(document, 'hidden', 'get').mockReturnValue(false)
    const generate = vi.spyOn(TOTP.prototype, 'generate')
    const active = ref(true)
    const secret = ref('GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ')
    let state!: ReturnType<typeof useAccountTotp>
    const wrapper = mount(defineComponent({ setup() { state = useAccountTotp(secret, active); return () => null } }))
    expect(state.code.value).toBe('287082')
    expect(generate).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(1000)
    expect(generate).toHaveBeenCalledTimes(2)
    await vi.advanceTimersByTimeAsync(20_000)
    expect(generate).toHaveBeenCalledTimes(2)
    hidden.mockReturnValue(true)
    document.dispatchEvent(new Event('visibilitychange'))
    await nextTick()
    expect(vi.getTimerCount()).toBe(0)
    expect(state.code.value).toBe('')
    await vi.advanceTimersByTimeAsync(60_000)
    hidden.mockReturnValue(false)
    document.dispatchEvent(new Event('visibilitychange'))
    await nextTick()
    expect(generate).toHaveBeenCalledTimes(3)
    expect(state.code.value).not.toBe('287082')
    active.value = false
    await nextTick()
    expect(vi.getTimerCount()).toBe(0)
    active.value = true
    await nextTick()
    expect(vi.getTimerCount()).toBe(1)
    wrapper.unmount()
    expect(vi.getTimerCount()).toBe(0)
  })
})
