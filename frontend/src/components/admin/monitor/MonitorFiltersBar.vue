<template>
  <div class="components-admin-monitor-monitor-filters-bar__panel">
    <!-- Left: Search + Filters -->
    <div class="components-admin-monitor-monitor-filters-bar__panel-2">
      <div class="components-admin-monitor-monitor-filters-bar__panel-3">
        <Icon
          name="search"
          size="md"
          class="components-admin-monitor-monitor-filters-bar__icon"
        />
        <input
          v-model="search"
          type="text"
          :placeholder="t('admin.channelMonitor.searchPlaceholder')"
          class="components-admin-monitor-monitor-filters-bar__field input"
          @input="$emit('search-input')"
        />
      </div>

      <Select
        v-model="provider"
        :options="providerFilterOptions"
        :placeholder="t('admin.channelMonitor.allProviders')"
        class="components-admin-monitor-monitor-filters-bar__field-2"
        @change="$emit('reload')"
      />

      <Select
        v-model="enabled"
        :options="enabledFilterOptions"
        :placeholder="t('admin.channelMonitor.enabledFilter')"
        class="components-admin-monitor-monitor-filters-bar__field-3"
        @change="$emit('reload')"
      />
    </div>

    <!-- Right: Actions -->
    <div class="components-admin-monitor-monitor-filters-bar__panel-4">
      <button
        @click="$emit('reload')"
        :disabled="loading"
        class="btn btn-secondary"
        :title="t('common.refresh')"
      >
        <Icon name="refresh" size="md" :class="loading ? 'components-admin-monitor-monitor-filters-bar__icon-3' : ''" />
      </button>
      <button
        @click="$emit('manage-templates')"
        class="btn btn-secondary"
        :title="t('admin.channelMonitor.template.manageButton')"
      >
        <Icon name="cog" size="md" class="components-admin-monitor-monitor-filters-bar__icon-2" />
        {{ t('admin.channelMonitor.template.manageButton') }}
      </button>
      <button @click="$emit('create')" class="btn btn-primary">
        <Icon name="plus" size="md" class="components-admin-monitor-monitor-filters-bar__icon-2" />
        {{ t('admin.channelMonitor.createButton') }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { Provider } from '@/api/admin/channelMonitor'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import {
  PROVIDER_OPENAI,
  PROVIDER_ANTHROPIC,
  PROVIDER_GEMINI,
  PROVIDER_GROK,
  PROVIDER_ANTIGRAVITY,
  PROVIDER_KIMI,
  PROVIDER_ZHIPU,
  PROVIDER_DEEPSEEK,
} from '@/constants/channelMonitor'

defineProps<{
  loading: boolean
}>()

defineEmits<{
  (e: 'reload'): void
  (e: 'create'): void
  (e: 'manage-templates'): void
  (e: 'search-input'): void
}>()

const search = defineModel<string>('search', { required: true })
const provider = defineModel<Provider | ''>('provider', { required: true })
const enabled = defineModel<'' | 'true' | 'false'>('enabled', { required: true })

const { t } = useI18n()

const providerFilterOptions = computed(() => [
  { value: '', label: t('admin.channelMonitor.allProviders') },
  { value: PROVIDER_OPENAI, label: t('monitorCommon.providers.openai') },
  { value: PROVIDER_ANTHROPIC, label: t('monitorCommon.providers.anthropic') },
  { value: PROVIDER_GEMINI, label: t('monitorCommon.providers.gemini') },
  { value: PROVIDER_GROK, label: t('monitorCommon.providers.grok') },
  { value: PROVIDER_ANTIGRAVITY, label: t('monitorCommon.providers.antigravity') },
  { value: PROVIDER_KIMI, label: t('monitorCommon.providers.kimi') },
  { value: PROVIDER_ZHIPU, label: t('monitorCommon.providers.zhipu') },
  { value: PROVIDER_DEEPSEEK, label: t('monitorCommon.providers.deepseek') },
])

const enabledFilterOptions = computed(() => [
  { value: '', label: t('admin.channelMonitor.allStatus') },
  { value: 'true', label: t('admin.channelMonitor.onlyEnabled') },
  { value: 'false', label: t('admin.channelMonitor.onlyDisabled') },
])
</script>
