<template>
  <div class="components-common-auto-refresh-button__panel" ref="dropdownRef" @keydown.esc.prevent.stop="showDropdown = false">
    <button
      ref="triggerRef"
      type="button"
      @click="showDropdown = !showDropdown"
      class="components-common-auto-refresh-button__action"
      :title="t('common.autoRefresh.title')"
    >
      <svg
        class="components-common-auto-refresh-button__icon"
        :class="enabled ? 'components-common-auto-refresh-button__icon-3' : ''"
        xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor"
      >
        <path fill-rule="evenodd" d="M15.312 11.424a5.5 5.5 0 01-9.201 2.466l-.312-.311h2.433a.75.75 0 000-1.5H4.598a.75.75 0 00-.75.75v3.634a.75.75 0 001.5 0v-2.033l.312.312a7 7 0 0011.712-3.138.75.75 0 00-1.449-.39zm-10.624-2.848a5.5 5.5 0 019.201-2.466l.312.311H11.768a.75.75 0 000 1.5h3.634a.75.75 0 00.75-.75V3.537a.75.75 0 00-1.5 0v2.034l-.312-.312A7 7 0 002.628 8.397a.75.75 0 001.449.39z" clip-rule="evenodd" />
      </svg>
      <span>
        {{ enabled
          ? t('common.autoRefresh.countdown', { seconds: countdown })
          : t('common.autoRefresh.title')
        }}
      </span>
    </button>

    <Teleport to="body">
      <div
        v-if="showDropdown"
        ref="panelRef"
        :style="panelStyle"
        class="components-common-auto-refresh-button__panel-2"
        @click.stop
      >
      <div class="components-common-auto-refresh-button__panel-3">
        <button
          @click="$emit('update:enabled', !enabled)"
          class="components-common-auto-refresh-button__action-2"
        >
          <span>{{ t('common.autoRefresh.enable') }}</span>
          <svg v-if="enabled" class="components-common-auto-refresh-button__icon-2" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M16.704 4.153a.75.75 0 01.143 1.052l-8 10.5a.75.75 0 01-1.127.075l-4.5-4.5a.75.75 0 011.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 011.05-.143z" clip-rule="evenodd" />
          </svg>
        </button>
        <div class="components-common-auto-refresh-button__panel-4"></div>
        <button
          v-for="sec in intervals"
          :key="sec"
          @click="$emit('update:interval', sec)"
          class="components-common-auto-refresh-button__action-2"
        >
          <span>{{ t('common.autoRefresh.seconds', { n: sec }) }}</span>
          <svg v-if="intervalSeconds === sec" class="components-common-auto-refresh-button__icon-2" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M16.704 4.153a.75.75 0 01.143 1.052l-8 10.5a.75.75 0 01-1.127.075l-4.5-4.5a.75.75 0 011.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 011.05-.143z" clip-rule="evenodd" />
          </svg>
        </button>
      </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useFloatingPanel } from '@/composables/useFloatingPanel'

defineProps<{
  enabled: boolean
  intervalSeconds: number
  countdown: number
  intervals: readonly number[]
}>()

defineEmits<{
  (e: 'update:enabled', value: boolean): void
  (e: 'update:interval', value: number): void
}>()

const { t } = useI18n()
const showDropdown = ref(false)
const dropdownRef = ref<HTMLElement | null>(null)
const triggerRef = ref<HTMLButtonElement | null>(null)
const { panelRef, style: panelStyle } = useFloatingPanel(triggerRef, showDropdown, {
  maxWidth: 176,
  minComfortableHeight: 120
})

function handleClickOutside(event: MouseEvent) {
  const target = event.target as Node
  if (dropdownRef.value?.contains(target) || panelRef.value?.contains(target)) return
  showDropdown.value = false
}

onMounted(() => document.addEventListener('click', handleClickOutside))
onBeforeUnmount(() => document.removeEventListener('click', handleClickOutside))
</script>
