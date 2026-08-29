import { nextTick, onBeforeUnmount, onMounted, ref, type Ref } from 'vue'

export function useChartSize(
  element: Ref<HTMLElement | null>,
  fallbackWidth = 640,
  fallbackHeight = 280,
) {
  const width = ref(fallbackWidth)
  const height = ref(fallbackHeight)
  let observer: ResizeObserver | null = null

  function measure() {
    const node = element.value
    if (!node) return
    const rect = node.getBoundingClientRect()
    width.value = Math.max(240, Math.round(rect.width || node.clientWidth || fallbackWidth))
    height.value = Math.max(140, Math.round(rect.height || node.clientHeight || fallbackHeight))
  }

  onMounted(async () => {
    await nextTick()
    measure()
    if (typeof ResizeObserver !== 'undefined' && element.value) {
      observer = new ResizeObserver(measure)
      observer.observe(element.value)
    }
  })

  onBeforeUnmount(() => observer?.disconnect())

  return { width, height, measure }
}
