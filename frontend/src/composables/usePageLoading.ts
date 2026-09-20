import { onScopeDispose, watch, type WatchSource } from 'vue'
import { useNavigationLoadingState } from './useNavigationLoading'

// Each visible loading region owns its progress contribution until it finishes or unmounts.
export function usePageLoading(source: WatchSource<boolean>) {
  const state = useNavigationLoadingState()
  let finish: (() => void) | undefined
  watch(source, (loading) => {
    if (loading && !finish) finish = state.beginPageLoad()
    if (!loading) {
      finish?.()
      finish = undefined
    }
  }, { immediate: true, flush: 'sync' })
  onScopeDispose(() => finish?.())
}
