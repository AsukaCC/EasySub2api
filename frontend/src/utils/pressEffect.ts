import gsap from 'gsap'

// 按压微动效:document 级事件委托,统一给按钮类控件加
// pointerdown 缩小 + 释放弹性回弹。集中在一处注册,组件无需各自引入。
//
// 注意:.btn 的 CSS transition 已不含 transform(见 _components.scss),
// 避免与 GSAP 的行内 transform 互相竞争。

const PRESSABLE_SELECTOR = [
  '.btn',
  '.button',
  '.quick-actions__item',
  '.dashboard-key-usage__segment',
].join(', ')

const PRESS_SCALE = 0.97

function findPressable(target: EventTarget | null): HTMLElement | null {
  if (!(target instanceof Element)) return null
  const el = target.closest<HTMLElement>(PRESSABLE_SELECTOR)
  if (!el || el.hasAttribute('disabled') || el.getAttribute('aria-disabled') === 'true') {
    return null
  }
  return el
}

function pressDown(el: HTMLElement) {
  gsap.to(el, {
    scale: PRESS_SCALE,
    duration: 0.12,
    ease: 'power2.out',
    overwrite: 'auto',
  })
}

function pressRelease(el: HTMLElement) {
  gsap.to(el, {
    scale: 1,
    duration: 0.45,
    ease: 'elastic.out(1.1, 0.55)',
    overwrite: 'auto',
    // 回弹结束后清掉行内 transform,不残留样式
    onComplete: () => gsap.set(el, { clearProps: 'scale' }),
  })
}

export function installPressEffect(root: Document = document) {
  // 尊重系统减弱动态偏好:直接不注册,保留纯 CSS 反馈
  if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return

  let pressed: HTMLElement | null = null

  root.addEventListener(
    'pointerdown',
    (event) => {
      // 仅响应主键/触摸
      if (event instanceof PointerEvent && event.button !== 0) return
      const el = findPressable(event.target)
      if (!el) return
      pressed = el
      pressDown(el)
    },
    { passive: true, capture: true }
  )

  const release = () => {
    if (!pressed) return
    pressRelease(pressed)
    pressed = null
  }

  root.addEventListener('pointerup', release, { passive: true, capture: true })
  root.addEventListener('pointercancel', release, { passive: true, capture: true })
}
