<template>
  <div ref="dropdownRef" class="locale-switcher" @keydown.esc.prevent.stop="closeDropdown(true)">
    <button
      ref="triggerRef"
      type="button"
      @click="toggleDropdown"
      :disabled="switching"
      class="locale-trigger"
      :class="{ 'locale-trigger-open': isOpen }"
      :title="`Language: ${currentLocale?.name ?? currentLocaleCode}`"
      :aria-label="`Language: ${currentLocale?.name ?? currentLocaleCode}`"
      aria-haspopup="menu"
      :aria-expanded="isOpen"
    >
      <Icon name="globe" size="sm" class="components-common-locale-switcher__icon" />
      <span class="components-common-locale-switcher__text locale-current locale-current-full">
        {{ currentLocale?.name }}
      </span>
      <span class="locale-current locale-current-compact" aria-hidden="true">
        {{ currentLocaleCode.toUpperCase() }}
      </span>
      <Icon
        name="chevronDown"
        size="xs"
        class="locale-chevron"
        :class="{ 'components-common-locale-switcher__icon-3': isOpen }"
      />
    </button>

    <Teleport to="body">
      <transition name="dropdown">
        <div
          v-if="isOpen"
          ref="menuRef"
          class="locale-menu"
          :class="{ 'locale-menu-top': menuPlacement === 'top' }"
          :style="menuStyle"
          role="menu"
          aria-label="Language"
          @click.stop
          @keydown.esc.prevent.stop="closeDropdown(true)"
        >
          <button
            v-for="locale in availableLocales"
            :key="locale.code"
            type="button"
            :disabled="switching"
            @click="selectLocale(locale.code)"
            class="locale-option"
            :class="{
              'locale-option-active':
                locale.code === currentLocaleCode
            }"
            role="menuitemradio"
            :aria-checked="locale.code === currentLocaleCode"
          >
            <span
              class="locale-option-code"
              :class="{ 'locale-option-code-active': locale.code === currentLocaleCode }"
            >
              {{ locale.code.toUpperCase() }}
            </span>
            <span class="components-common-locale-switcher__text-2">{{ locale.name }}</span>
            <span class="components-common-locale-switcher__text-3">
              <Icon
                v-if="locale.code === currentLocaleCode"
                name="check"
                size="sm"
                class="components-common-locale-switcher__icon-2"
              />
            </span>
          </button>
        </div>
      </transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { setLocale, availableLocales } from '@/i18n'

const { locale } = useI18n()

const isOpen = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const menuRef = ref<HTMLElement | null>(null)
const menuPlacement = ref<'bottom' | 'top'>('bottom')
const menuStyle = ref<Record<string, string>>({})
const switching = ref(false)

const menuWidth = 176
const menuEstimatedHeight = 92
const menuGap = 8
const viewportPadding = 8

const currentLocaleCode = computed(() => locale.value)
const currentLocale = computed(() => availableLocales.find((l) => l.code === locale.value))

function toggleDropdown() {
  isOpen.value = !isOpen.value
}

function updateMenuPosition() {
  const trigger = triggerRef.value
  if (!trigger) return

  const rect = trigger.getBoundingClientRect()
  const width = Math.min(menuWidth, Math.max(0, window.innerWidth - viewportPadding * 2))
  const height = menuRef.value?.offsetHeight || menuEstimatedHeight
  const spaceBelow = window.innerHeight - rect.bottom - menuGap - viewportPadding
  const spaceAbove = rect.top - menuGap - viewportPadding
  const opensUp = spaceBelow < height && spaceAbove > spaceBelow
  const availableHeight = Math.max(0, opensUp ? spaceAbove : spaceBelow)
  const visibleHeight = Math.min(height, availableHeight)
  const left = Math.min(
    Math.max(viewportPadding, rect.right - width),
    Math.max(viewportPadding, window.innerWidth - viewportPadding - width)
  )
  const top = opensUp
    ? Math.max(viewportPadding, rect.top - menuGap - visibleHeight)
    : Math.min(rect.bottom + menuGap, window.innerHeight - viewportPadding - visibleHeight)

  menuPlacement.value = opensUp ? 'top' : 'bottom'
  menuStyle.value = {
    left: `${left}px`,
    top: `${top}px`,
    width: `${width}px`,
    maxHeight: `${availableHeight}px`
  }
}

function closeDropdown(restoreFocus = false) {
  isOpen.value = false
  if (restoreFocus) {
    void nextTick(() => triggerRef.value?.focus())
  }
}

async function selectLocale(code: string) {
  if (switching.value || code === currentLocaleCode.value) {
    closeDropdown(true)
    return
  }
  switching.value = true
  try {
    await setLocale(code)
  } finally {
    switching.value = false
    closeDropdown(true)
  }
}

function handleClickOutside(event: MouseEvent) {
  const target = event.target as Node
  if (dropdownRef.value?.contains(target) || menuRef.value?.contains(target)) return
  closeDropdown()
}

watch(isOpen, (open) => {
  if (open) {
    updateMenuPosition()
    void nextTick(updateMenuPosition)
    window.addEventListener('resize', updateMenuPosition)
    window.addEventListener('scroll', updateMenuPosition, true)
  } else {
    window.removeEventListener('resize', updateMenuPosition)
    window.removeEventListener('scroll', updateMenuPosition, true)
  }
})

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleClickOutside)
  window.removeEventListener('resize', updateMenuPosition)
  window.removeEventListener('scroll', updateMenuPosition, true)
})
</script>

<style scoped>
.locale-switcher {
  position: relative;
}

.locale-trigger {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  min-width: 5.5rem;
  min-height: 2.25rem;
  padding: 0.375rem 0.625rem;
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  font-weight: 500;
  cursor: pointer;
  transition: background-color 160ms ease, border-color 160ms ease;
  background-color: var(--glass-layer-inset-bg);
  border: 1px solid var(--glass-border);
  -webkit-backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
  box-shadow:
    0 4px 14px rgb(12 12 14 / 0.06),
    0 1px 0 var(--glass-highlight) inset;
}

.locale-trigger:hover:not(:disabled) {
  background-color: var(--glass-layer-inset-bg);
  border-color: var(--glass-border-hover);
  -webkit-backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
  backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
  box-shadow: 0 4px 16px rgba(15, 23, 42, 0.08), 0 1px 0 var(--glass-highlight-hover) inset;
}

.locale-trigger:focus-visible {
  border-color: var(--color-primary);
  outline: none;
  box-shadow:
    0 0 0 3px rgba(10, 132, 255, 0.22),
    0 1px 0 var(--glass-highlight) inset;
}

.locale-trigger-open {
  background-color: var(--glass-tint-brand);
  border-color: var(--color-primary-border);
  color: var(--color-text-brand);
  -webkit-backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
  backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
}

.locale-current {
  min-width: 0;
  overflow: hidden;
  max-width: 7rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.locale-current-compact {
  display: none;
  font-size: var(--font-size-xs);
  font-weight: var(--font-weight-semibold);
}

.locale-chevron {
  flex-shrink: 0;
  color: var(--color-text-tertiary);
}

.locale-menu {
  position: fixed;
  z-index: 100;
  overflow-y: auto;
  padding: 0.375rem;
  border-radius: var(--radius-lg);
  background-color: var(--glass-layer-floating-bg);
  border: 1px solid var(--glass-border);
  -webkit-backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  box-shadow:
    var(--glass-shadow-hover),
    0 1px 0 var(--glass-highlight) inset;
  transform-origin: top right;
}

.locale-menu-top {
  transform-origin: bottom right;
}

.locale-option {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
  padding: 0.5rem 0.75rem;
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-medium);
  text-align: left;
  cursor: pointer;
  transition: color 150ms ease, background-color 150ms ease, border-color 150ms ease;
}

.locale-option:hover:not(:disabled) {
  color: var(--color-text-primary);
  border-color: var(--glass-border);
  background-color: var(--glass-bg-interactive-hover);
  -webkit-backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
  backdrop-filter: blur(var(--glass-layer-inset-blur-hover)) saturate(var(--glass-saturate-hover));
  box-shadow: 0 1px 0 var(--glass-highlight) inset;
}

.locale-option-active {
  background-color: var(--glass-tint-brand);
  border-color: var(--color-primary-border);
  color: var(--color-text-brand);
  font-weight: var(--font-weight-semibold);
}

.locale-option-code {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  min-width: 2rem;
  padding: 0.125rem 0.25rem;
  border-radius: var(--radius-sm);
  background: var(--glass-layer-inset-bg);
  color: var(--color-text-tertiary);
  font-size: var(--font-size-2xs);
  font-weight: 600;
}

.locale-option-code-active {
  background: var(--glass-tint-brand-active);
  color: var(--color-primary);
}

@media (max-width: 639px) {
  .locale-trigger {
    min-width: 4rem;
    padding-inline: 0.5rem;
  }

  .locale-current-full {
    display: none;
  }

  .locale-current-compact {
    display: inline-block;
  }
}

.dropdown-enter-active,
.dropdown-leave-active {
  transition: opacity 0.16s ease, transform 0.16s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-4px);
}

.locale-menu-top.dropdown-enter-from,
.locale-menu-top.dropdown-leave-to {
  transform: scale(0.95) translateY(4px);
}
</style>
