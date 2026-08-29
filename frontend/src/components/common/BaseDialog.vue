<template>
  <Teleport to="body">
    <Transition name="modal" v-bind="transitionProps">
      <div
        v-if="show"
        class="modal-overlay"
        :style="zIndexStyle"
        :aria-labelledby="dialogId"
        role="dialog"
        aria-modal="true"
        @click.self="handleClose"
      >
        <!-- Modal panel -->
        <div ref="dialogRef" :class="['modal-content', widthClasses]" @click.stop>
          <!-- Header -->
          <div class="modal-header">
            <h3 :id="dialogId" class="modal-title">
              {{ title }}
            </h3>
            <button
              v-if="showCloseButton"
              @click="emit('close')"
              class="base-dialog__close"
              aria-label="Close modal"
            >
              <Icon name="x" size="md" />
            </button>
          </div>

          <!-- Body -->
          <div ref="modalBodyRef" class="modal-body">
            <slot></slot>
          </div>

          <!-- Footer -->
          <div v-if="$slots.footer" class="modal-footer">
            <slot name="footer"></slot>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, watch, onMounted, onUnmounted, ref, nextTick } from 'vue'
import gsap from 'gsap'
import Icon from '@/components/icons/Icon.vue'

// GSAP 进出场:overlay 淡入 + 内容轻微上浮缩放。
// 减弱动态偏好时不接管,回退到 _feedback.scss 的 .modal CSS Transition。
const prefersReducedMotion =
  typeof window !== 'undefined' &&
  window.matchMedia('(prefers-reduced-motion: reduce)').matches

const dialogEnter = (el: Element, done: () => void) => {
  const overlay = el as HTMLElement
  const content = overlay.querySelector<HTMLElement>('.modal-content')
  gsap.fromTo(overlay, { opacity: 0 }, { opacity: 1, duration: 0.22, ease: 'power2.out' })
  if (!content) {
    done()
    return
  }
  gsap.fromTo(
    content,
    { opacity: 0, scale: 0.96, y: 8 },
    {
      opacity: 1,
      scale: 1,
      y: 0,
      duration: 0.3,
      ease: 'power3.out',
      onComplete: () => {
        gsap.set(content, { clearProps: 'opacity,transform' })
        done()
      },
    }
  )
}

const dialogLeave = (el: Element, done: () => void) => {
  const overlay = el as HTMLElement
  const content = overlay.querySelector<HTMLElement>('.modal-content')
  gsap.to(overlay, { opacity: 0, duration: 0.18, ease: 'power2.in', delay: 0.04 })
  if (!content) {
    gsap.delayedCall(0.22, done)
    return
  }
  gsap.to(content, {
    opacity: 0,
    scale: 0.97,
    y: 6,
    duration: 0.18,
    ease: 'power2.in',
    onComplete: done,
  })
}

// 减弱动态时传空对象,保留 name="modal" 的 CSS 过渡;
// 否则 css:false 完全交给 GSAP 钩子控制
const transitionProps = prefersReducedMotion
  ? {}
  : { css: false, onEnter: dialogEnter, onLeave: dialogLeave }

// 生成唯一ID以避免多个对话框时ID冲突
let dialogIdCounter = 0
const dialogId = `modal-title-${++dialogIdCounter}`

// 焦点管理
const dialogRef = ref<HTMLElement | null>(null)
const modalBodyRef = ref<HTMLElement | null>(null)
let previousActiveElement: HTMLElement | null = null

type DialogWidth = 'narrow' | 'normal' | 'wide' | 'extra-wide' | 'full'

interface Props {
  show: boolean
  title: string
  width?: DialogWidth
  closeOnEscape?: boolean
  closeOnClickOutside?: boolean
  showCloseButton?: boolean
  zIndex?: number
}

interface Emits {
  (e: 'close'): void
}

const props = withDefaults(defineProps<Props>(), {
  width: 'normal',
  closeOnEscape: true,
  closeOnClickOutside: false,
  showCloseButton: true,
  zIndex: 50
})

const emit = defineEmits<Emits>()

// Keep the existing 40/50/60/80 API as relative modal levels while moving
// the whole modal stack above sticky page sections and ordinary popovers.
const zIndexStyle = computed(() => {
  return { '--modal-layer-offset': String(props.zIndex - 50) }
})

const widthClasses = computed(() => {
  // Width guidance: narrow=confirm/short prompts, normal=standard forms,
  // wide=multi-section forms or rich content, extra-wide=analytics/tables,
  // full=full-screen or very dense layouts.
  const widths: Record<DialogWidth, string> = {
    narrow: 'modal-content--narrow',
    normal: 'modal-content--normal',
    wide: 'modal-content--wide',
    'extra-wide': 'modal-content--extra-wide',
    full: 'modal-content--full'
  }
  return widths[props.width]
})

const handleClose = () => {
  if (props.closeOnClickOutside) {
    emit('close')
  }
}

const handleEscape = (event: KeyboardEvent) => {
  if (props.show && props.closeOnEscape && event.key === 'Escape') {
    emit('close')
  }
}

// Prevent body scroll when modal is open and manage focus
watch(
  () => props.show,
  async (isOpen) => {
    if (isOpen) {
      // 保存当前焦点元素
      previousActiveElement = document.activeElement as HTMLElement
      // 使用CSS类而不是直接操作style,更易于管理多个对话框
      document.body.classList.add('modal-open')

      // 等待DOM更新后设置焦点到对话框
      await nextTick()
      if (modalBodyRef.value) {
        modalBodyRef.value.scrollTop = 0
      }
      if (dialogRef.value) {
        const firstFocusable = dialogRef.value.querySelector<HTMLElement>(
          'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
        )
        firstFocusable?.focus()
      }
    } else {
      document.body.classList.remove('modal-open')
      // 恢复之前的焦点
      if (previousActiveElement && typeof previousActiveElement.focus === 'function') {
        previousActiveElement.focus()
      }
      previousActiveElement = null
    }
  },
  { immediate: true }
)

onMounted(() => {
  document.addEventListener('keydown', handleEscape)
})

onUnmounted(() => {
  document.removeEventListener('keydown', handleEscape)
  // 确保组件卸载时移除滚动锁定
  document.body.classList.remove('modal-open')
})
</script>

<style>
.base-dialog__close {
  margin-right: -0.5rem;
  padding: 0.5rem;
  border: 1px solid transparent;
  border-radius: var(--radius-lg);
  background-color: transparent;
  color: var(--color-text-tertiary);
  transition: color 150ms ease, border-color 150ms ease, box-shadow 150ms ease;
}

.base-dialog__close:hover {
  border-color: var(--glass-border);
  background-color: var(--glass-bg-interactive-hover);
  color: var(--color-text-secondary);
  box-shadow: 0 1px 0 var(--glass-highlight-hover) inset;
  -webkit-backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
  backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
}

.base-dialog__close:focus-visible {
  border-color: var(--glass-border-active);
  outline: none;
  box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.22);
}
</style>
