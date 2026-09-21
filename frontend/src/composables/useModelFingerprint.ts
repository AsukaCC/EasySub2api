import { onUnmounted, ref, watch, type Ref } from 'vue'
import { getModelFingerprint, type ModelFingerprintSnapshot } from '@/api/admin/modelFingerprint'

export function useModelFingerprint(accountId: Ref<string | null>, onUpdate?: (snapshot: ModelFingerprintSnapshot | null) => void) {
  const snapshot = ref<ModelFingerprintSnapshot | null>(null)
  const pollingError = ref(false)
  let timer: ReturnType<typeof setTimeout> | undefined
  let controller: AbortController | undefined
  let generation = 0

  function stop() {
    generation++
    clearTimeout(timer)
    controller?.abort()
  }

  function setSnapshot(value: ModelFingerprintSnapshot | null) {
    snapshot.value = value
    onUpdate?.(value)
  }

  async function refresh() {
    const id = accountId.value
    if (!id) return
    const current = ++generation
    clearTimeout(timer)
    controller?.abort()
    controller = new AbortController()
    try {
      const value = await getModelFingerprint(id, controller.signal)
      if (current !== generation) return
      pollingError.value = false
      setSnapshot(value)
    } catch {
      if (current !== generation) return
      pollingError.value = true
    }
    if (current === generation && (pollingError.value || snapshot.value?.status === 'running')) {
      timer = setTimeout(() => { void refresh() }, 3000)
    }
  }

  function track(value: ModelFingerprintSnapshot) {
    stop()
    setSnapshot(value)
    if (value.status === 'running') timer = setTimeout(() => { void refresh() }, 1500)
  }

  watch(accountId, () => {
    stop()
    snapshot.value = null
    pollingError.value = false
    if (accountId.value) void refresh()
  }, { immediate: true })
  onUnmounted(stop)
  return { snapshot, pollingError, refresh, track }
}
