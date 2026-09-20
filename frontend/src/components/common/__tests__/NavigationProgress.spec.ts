/**
 * NavigationProgress 组件单元测试
 */
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { enableAutoUnmount, mount } from '@vue/test-utils'
import { nextTick, ref } from 'vue'
import NavigationProgress from '../../common/NavigationProgress.vue'

// Mock useNavigationLoadingState
const mockIsLoading = ref(false)
enableAutoUnmount(afterEach)
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: () => 'Loading' }) }))

vi.mock('@/composables/useNavigationLoading', () => ({
  useNavigationLoadingState: () => ({
    isLoading: mockIsLoading
  })
}))

describe('NavigationProgress', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    mockIsLoading.value = false
  })
  afterEach(() => { vi.useRealTimers() })

  it('isLoading=false 时进度条应该隐藏', () => {
    mockIsLoading.value = false
    const wrapper = mount(NavigationProgress)

    const progressBar = wrapper.find('.navigation-progress')
    // v-show 会设置 display: none
    expect(progressBar.isVisible()).toBe(false)
  })

  it('isLoading=true 时进度条应该可见', async () => {
    mockIsLoading.value = true
    const wrapper = mount(NavigationProgress)

    await wrapper.vm.$nextTick()

    const progressBar = wrapper.find('.navigation-progress')
    expect(progressBar.exists()).toBe(true)
    expect(progressBar.isVisible()).toBe(true)
  })

  it('应该有正确的 ARIA 属性', () => {
    mockIsLoading.value = true
    const wrapper = mount(NavigationProgress)

    const progressBar = wrapper.find('.navigation-progress')
    expect(progressBar.attributes('role')).toBe('progressbar')
    expect(progressBar.attributes('aria-label')).toBe('Loading')
    expect(progressBar.attributes('aria-valuenow')).toBeUndefined()
  })

  it('进度条应该有动画 class', () => {
    mockIsLoading.value = true
    const wrapper = mount(NavigationProgress)

    const bar = wrapper.find('.navigation-progress-bar')
    expect(bar.exists()).toBe(true)
  })

  it('应该正确响应 isLoading 状态变化', async () => {
    // 测试初始状态为 false
    mockIsLoading.value = false
    const wrapper = mount(NavigationProgress)
    await wrapper.vm.$nextTick()

    // 初始状态隐藏
    expect(wrapper.find('.navigation-progress').isVisible()).toBe(false)

    // 卸载后重新挂载以测试 true 状态
    wrapper.unmount()

    // 改变为 true 后重新挂载
    mockIsLoading.value = true
    const wrapper2 = mount(NavigationProgress)
    await wrapper2.vm.$nextTick()
    expect(wrapper2.find('.navigation-progress').isVisible()).toBe(true)

    // 清理
    wrapper2.unmount()
  })

  it('finishes at the end before hiding, and cancels completion when another load begins', async () => {
    mockIsLoading.value = true
    const wrapper = mount(NavigationProgress)
    await vi.advanceTimersByTimeAsync(1000)
    const before = wrapper.get('.navigation-progress-bar').attributes('style')
    expect(before).not.toContain('scaleX(1)')
    mockIsLoading.value = false
    await nextTick()
    await vi.advanceTimersByTimeAsync(100)
    mockIsLoading.value = true
    await nextTick()
    await vi.advanceTimersByTimeAsync(400)
    expect(wrapper.get('.navigation-progress').isVisible()).toBe(true)
    expect(wrapper.get('.navigation-progress-bar').attributes('style')).not.toContain('scaleX(1)')
    mockIsLoading.value = false
    await nextTick()
    await vi.advanceTimersByTimeAsync(150)
    expect(wrapper.get('.navigation-progress-bar').attributes('style')).toContain('scaleX(1)')
    await vi.advanceTimersByTimeAsync(220)
    await vi.advanceTimersByTimeAsync(200)
    expect(wrapper.get('.navigation-progress').attributes('style')).toContain('display: none')
  })

  it('clears its timers when unmounted during loading', () => {
    mockIsLoading.value = true
    const wrapper = mount(NavigationProgress)
    wrapper.unmount()
    expect(vi.getTimerCount()).toBe(0)
  })
})
