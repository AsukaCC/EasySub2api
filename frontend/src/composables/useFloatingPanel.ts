import { computed, nextTick, onBeforeUnmount, ref, watch, type Ref } from 'vue'
import {
  getFloatingPanelPosition,
  type FloatingPanelOptions,
  type FloatingPanelPosition
} from '@/utils/floatingPanel'

export interface UseFloatingPanelOptions extends FloatingPanelOptions {
  zIndex?: number
}

const emptyPosition: FloatingPanelPosition = {
  top: null,
  bottom: null,
  left: 0,
  width: 0,
  maxHeight: 0
}

/**
 * Positions a teleported panel relative to its trigger and keeps it aligned
 * while nested scroll containers or the viewport move.
 */
export function useFloatingPanel(
  triggerRef: Ref<HTMLElement | null>,
  open: Ref<boolean>,
  options: UseFloatingPanelOptions = {}
) {
  const panelRef = ref<HTMLElement | null>(null)
  const position = ref<FloatingPanelPosition>({ ...emptyPosition })

  const style = computed<Record<string, string>>(() => ({
    position: 'fixed',
    zIndex: String(options.zIndex ?? 1000),
    top: position.value.top == null ? 'auto' : `${position.value.top}px`,
    bottom: position.value.bottom == null ? 'auto' : `${position.value.bottom}px`,
    left: `${position.value.left}px`,
    width: `${position.value.width}px`,
    maxHeight: `${position.value.maxHeight}px`
  }))

  const update = () => {
    const trigger = triggerRef.value
    if (!trigger) return

    position.value = getFloatingPanelPosition(
      trigger.getBoundingClientRect(),
      document.documentElement.clientWidth || window.innerWidth,
      window.innerHeight,
      options
    )
  }

  const handleViewportChange = () => update()

  const startTracking = () => {
    update()
    void nextTick(update)
    window.addEventListener('resize', handleViewportChange)
    window.addEventListener('scroll', handleViewportChange, true)
  }

  const stopTracking = () => {
    window.removeEventListener('resize', handleViewportChange)
    window.removeEventListener('scroll', handleViewportChange, true)
  }

  watch(open, (isOpen) => {
    if (isOpen) startTracking()
    else stopTracking()
  })

  onBeforeUnmount(stopTracking)

  return {
    panelRef,
    position,
    style,
    update
  }
}
