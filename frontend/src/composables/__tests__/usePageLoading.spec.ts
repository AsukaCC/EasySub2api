import { effectScope, ref } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { usePageLoading } from '../usePageLoading'
import { _resetNavigationLoadingInstance, useNavigationLoadingState } from '../useNavigationLoading'

beforeEach(() => { vi.useFakeTimers(); _resetNavigationLoadingInstance() })
afterEach(() => { _resetNavigationLoadingInstance(); vi.useRealTimers() })

describe('page loading lifecycle', () => {
  it('releases pending loads on scope disposal and tracks subsequent reloads', () => {
    const scope = effectScope()
    const loading = ref(true)
    const state = useNavigationLoadingState()
    scope.run(() => usePageLoading(loading))
    vi.advanceTimersByTime(100)
    expect(state.isLoading.value).toBe(true)
    loading.value = false
    expect(state.isLoading.value).toBe(false)
    loading.value = true
    vi.advanceTimersByTime(100)
    expect(state.isLoading.value).toBe(true)
    scope.stop()
    expect(state.isLoading.value).toBe(false)
    expect(vi.getTimerCount()).toBe(0)
  })
})
