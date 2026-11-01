import { computed, onUnmounted, ref, watch, type Ref } from 'vue'
import { useEventListener } from '@vueuse/core'
import { parseAccountTotp } from '@/utils/accountTotp'

export function useAccountTotp(secret: Ref<string>, active: Ref<boolean>) {
  const visible = ref(!document.hidden)
  const code = ref('')
  const remaining = ref(0)
  const parsed = computed(() => {
    if (!secret.value.trim()) return { generator: null, invalid: false }
    try { return { generator: parseAccountTotp(secret.value), invalid: false } }
    catch { return { generator: null, invalid: true } }
  })
  let timer: ReturnType<typeof setInterval> | undefined
  let lastCounter = -1
  function stop() {
    clearInterval(timer)
    timer = undefined
    code.value = ''
    remaining.value = 0
    lastCounter = -1
  }
  function refresh() {
    const generator = parsed.value.generator
    if (!active.value || !visible.value || !generator) return
    const timestamp = Date.now()
    const counter = Math.floor(timestamp / (generator.period * 1000))
    // The one-second tick updates only the countdown; HMAC runs once per period.
    if (counter !== lastCounter) {
      code.value = generator.generate({ timestamp })
      lastCounter = counter
    }
    remaining.value = generator.period - Math.floor(timestamp / 1000) % generator.period
  }
  useEventListener(document, 'visibilitychange', () => { visible.value = !document.hidden })
  watch([parsed, active, visible], () => {
    stop()
    if (active.value && visible.value && parsed.value.generator) {
      refresh()
      timer = setInterval(refresh, 1000)
    }
  }, { immediate: true })
  onUnmounted(stop)
  return { code, remaining, invalid: computed(() => parsed.value.invalid), refresh }
}
