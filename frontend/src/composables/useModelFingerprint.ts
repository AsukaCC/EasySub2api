import { computed, onUnmounted, ref, watch, type Ref } from 'vue'
import { useEventListener } from '@vueuse/core'
import { getModelFingerprint, type ModelFingerprintSnapshot } from '@/api/admin/modelFingerprint'

const retentionMs = 2 * 60 * 60 * 1000

function retainedUntil(value: ModelFingerprintSnapshot) {
  // expires_at is the worker lease, used only for interrupted/legacy results.
  const finished = Date.parse(value.finished_at || value.expires_at || value.started_at)
  return Number.isFinite(finished) ? finished + retentionMs : 0
}

export function useModelFingerprintExpiry(source: Ref<ModelFingerprintSnapshot | null | undefined>, onExpire: () => void) {
  const now = ref(Date.now())
  let timer: ReturnType<typeof setTimeout> | undefined
  const snapshot = computed(() => source.value && retainedUntil(source.value) > now.value ? source.value : null)
  function update() {
    clearTimeout(timer)
    now.value = Date.now()
    if (!source.value) return
    const remaining = retainedUntil(source.value) - now.value
    if (remaining <= 0) onExpire()
    else timer = setTimeout(update, Math.min(remaining, retentionMs))
  }
  watch(source, update, { immediate: true })
  useEventListener(document, 'visibilitychange', update)
  onUnmounted(() => clearTimeout(timer))
  return snapshot
}

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
    if (value && retainedUntil(value) <= Date.now()) value = null
    snapshot.value = value
    onUpdate?.(value)
  }

  const visibleSnapshot = useModelFingerprintExpiry(snapshot, () => setSnapshot(null))

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
  return { snapshot: visibleSnapshot, pollingError, refresh, track }
}
