<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useClipboard } from '@/composables/useClipboard'
import type { CustomEndpoint } from '@/types'

const props = defineProps<{
  apiBaseUrl: string
  customEndpoints: CustomEndpoint[]
}>()

const { t } = useI18n()
const { copyToClipboard } = useClipboard()
const copiedEndpoint = ref<string | null>(null)

let copiedResetTimer: number | undefined

const allEndpoints = computed(() => {
  const items: Array<{ name: string; endpoint: string; description: string; isDefault: boolean }> = []
  if (props.apiBaseUrl) {
    items.push({
      name: t('keys.endpoints.title'),
      endpoint: props.apiBaseUrl,
      description: '',
      isDefault: true,
    })
  }
  for (const ep of props.customEndpoints) {
    items.push({ ...ep, isDefault: false })
  }
  return items
})

async function copy(url: string) {
  const success = await copyToClipboard(url, t('keys.endpoints.copied'))
  if (!success) return

  copiedEndpoint.value = url
  if (copiedResetTimer !== undefined) {
    window.clearTimeout(copiedResetTimer)
  }
  copiedResetTimer = window.setTimeout(() => {
    if (copiedEndpoint.value === url) {
      copiedEndpoint.value = null
    }
  }, 1800)
}

function tooltipHint(endpoint: string): string {
  return copiedEndpoint.value === endpoint
    ? t('keys.endpoints.copiedHint')
    : t('keys.endpoints.clickToCopy')
}

function speedTestUrl(endpoint: string): string {
  return `https://www.tcptest.cn/http/${encodeURIComponent(endpoint)}`
}

onBeforeUnmount(() => {
  if (copiedResetTimer !== undefined) {
    window.clearTimeout(copiedResetTimer)
  }
})
</script>

<template>
  <div v-if="allEndpoints.length > 0" class="components-keys-endpoint-popover__panel">
    <div
      v-for="(item, index) in allEndpoints"
      :key="index"
      class="components-keys-endpoint-popover__panel-2"
    >
      <span class="components-keys-endpoint-popover__text">{{ item.name }}</span>
      <span
        v-if="item.isDefault"
        class="components-keys-endpoint-popover__text-2"
      >{{ t('keys.endpoints.default') }}</span>

      <span class="components-keys-endpoint-popover__text-3">|</span>

      <div class="components-keys-endpoint-popover__panel-3">
        <div
          class="components-keys-endpoint-popover__panel-4 glass-popover"
        >
          <p
            v-if="item.description"
            class="components-keys-endpoint-popover__description"
          >
            {{ item.description }}
          </p>
          <p
            class="components-keys-endpoint-popover__description-2"
            :class="item.description ? 'components-keys-endpoint-popover__description-3' : ''"
          >
            <span class="components-keys-endpoint-popover__text-4"></span>
            {{ tooltipHint(item.endpoint) }}
          </p>
          <div class="components-keys-endpoint-popover__panel-5"></div>
        </div>

        <code
          class="components-keys-endpoint-popover__code"
          role="button"
          tabindex="0"
          @click="copy(item.endpoint)"
          @keydown.enter.prevent="copy(item.endpoint)"
          @keydown.space.prevent="copy(item.endpoint)"
        >{{ item.endpoint }}</code>

        <button
          type="button"
          class="components-keys-endpoint-popover__action"
          :class="copiedEndpoint === item.endpoint
            ? 'components-keys-endpoint-popover__action-2'
            : 'components-keys-endpoint-popover__action-3'"
          :aria-label="tooltipHint(item.endpoint)"
          @click="copy(item.endpoint)"
        >
          <svg v-if="copiedEndpoint === item.endpoint" class="components-keys-endpoint-popover__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
          </svg>
          <svg v-else class="components-keys-endpoint-popover__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
        </button>

        <a
          :href="speedTestUrl(item.endpoint)"
          target="_blank"
          rel="noopener noreferrer"
          class="components-keys-endpoint-popover__link"
          :title="t('keys.endpoints.speedTest')"
        >
          <svg class="components-keys-endpoint-popover__icon" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
        </a>
      </div>
    </div>
  </div>
</template>
