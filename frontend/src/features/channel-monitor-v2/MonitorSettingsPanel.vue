<template>
  <section class="features-channel-monitor-v2-monitor-settings-panel__section">
    <header
      class="features-channel-monitor-v2-monitor-settings-panel__header page-header"
    >
      <div class="features-channel-monitor-v2-monitor-settings-panel__panel">
        <h2 class="features-channel-monitor-v2-monitor-settings-panel__heading page-title">
          <span class="features-channel-monitor-v2-monitor-settings-panel__text">
            <Icon name="chart" size="sm" />
          </span>
          {{ t('channelMonitorV2.settings.title') }}
        </h2>
        <p class="features-channel-monitor-v2-monitor-settings-panel__description page-description">
          {{ t('channelMonitorV2.settings.description') }}
        </p>
      </div>
      <button
        type="button"
        class="btn btn-primary"
        :disabled="saving || !dirty"
        @click="save"
      >
        <Icon name="check" size="sm" />
        {{ t('channelMonitorV2.settings.save') }}
      </button>
    </header>

    <div
      v-if="loading"
      class="features-channel-monitor-v2-monitor-settings-panel__panel-3 card"
    >
      <span class="features-channel-monitor-v2-monitor-settings-panel__text-2">{{ t('channelMonitorV2.settings.loading') }}</span>
    </div>

    <template v-else-if="draft">
      <div class="features-channel-monitor-v2-monitor-settings-panel__panel-4 card">
        <div class="features-channel-monitor-v2-monitor-settings-panel__panel-5">
          <div>
            <strong class="features-channel-monitor-v2-monitor-settings-panel__strong">{{ t('channelMonitorV2.settings.refreshTitle') }}</strong>
            <p class="features-channel-monitor-v2-monitor-settings-panel__description-2">{{ t('channelMonitorV2.settings.refreshHint') }}</p>
          </div>
          <div class="features-channel-monitor-v2-monitor-settings-panel__panel-6 tabs" role="group" :aria-label="t('channelMonitorV2.settings.refreshAria')">
            <button
              type="button"
              class="features-channel-monitor-v2-monitor-settings-panel__action"
              :class="draft.refresh_interval_seconds === 60 ? 'tab-active' : ''"
              @click="draft.refresh_interval_seconds = 60"
            >
              1 min
            </button>
            <button
              type="button"
              class="features-channel-monitor-v2-monitor-settings-panel__action"
              :class="draft.refresh_interval_seconds === 300 ? 'tab-active' : ''"
              @click="draft.refresh_interval_seconds = 300"
            >
              5 min
            </button>
          </div>
        </div>
      </div>

      <div class="features-channel-monitor-v2-monitor-settings-panel__panel-7 card">
        <div class="features-channel-monitor-v2-monitor-settings-panel__panel-8 card-header">
          <h3 class="features-channel-monitor-v2-monitor-settings-panel__strong">{{ t('channelMonitorV2.settings.platformsTitle') }}</h3>
          <p class="features-channel-monitor-v2-monitor-settings-panel__description-2">
            {{ t('channelMonitorV2.settings.platformsHint') }}
          </p>
        </div>
        <div class="features-channel-monitor-v2-monitor-settings-panel__panel-9">
          <div
            v-for="platform in draft.platforms"
            :key="platform.platform"
            class="features-channel-monitor-v2-monitor-settings-panel__panel-10"
          >
            <Toggle v-model="platform.enabled" />
            <strong class="features-channel-monitor-v2-monitor-settings-panel__strong-2">{{ platformLabel(platform.platform) }}</strong>
            <input
              class="input"
              :value="platform.models.join(', ')"
              type="text"
              :placeholder="t('channelMonitorV2.settings.modelsPlaceholder')"
              @change="setModels(platform, $event)"
            />
            <span
              class="features-channel-monitor-v2-monitor-settings-panel__text-3 badge"
              :class="platform.models.length ? 'badge-gray' : 'badge badge-primary'"
            >
              {{ platform.models.length ? t('channelMonitorV2.settings.badgeOther') : t('channelMonitorV2.settings.badgeAllModels') }}
            </span>
          </div>
        </div>
      </div>

      <div class="features-channel-monitor-v2-monitor-settings-panel__panel-7 card">
        <div class="features-channel-monitor-v2-monitor-settings-panel__panel-11 card-header">
          <div>
            <h3 class="features-channel-monitor-v2-monitor-settings-panel__strong">{{ t('channelMonitorV2.settings.groupsTitle') }}</h3>
            <p class="features-channel-monitor-v2-monitor-settings-panel__description-2">
              {{
                draft.group_ids.length
                  ? t('channelMonitorV2.settings.groupsSelected', { count: draft.group_ids.length })
                  : t('channelMonitorV2.settings.groupsAll')
              }}
            </p>
          </div>
          <button
            v-if="draft.group_ids.length"
            type="button"
            class="btn btn-ghost btn-sm"
            @click="draft.group_ids = []"
          >
            {{ t('channelMonitorV2.settings.groupsAll') }}
          </button>
        </div>
        <div class="features-channel-monitor-v2-monitor-settings-panel__panel-12">
          <div class="features-channel-monitor-v2-monitor-settings-panel__panel-13">
            <label
              v-for="group in groups"
              :key="group.id"
              class="features-channel-monitor-v2-monitor-settings-panel__label"
            >
              <input
                type="checkbox"
                class="features-channel-monitor-v2-monitor-settings-panel__field"
                :checked="draft.group_ids.includes(group.id)"
                @change="toggleGroup(group.id)"
              />
              <span class="features-channel-monitor-v2-monitor-settings-panel__text-4">{{ group.name }}</span>
              <small class="features-channel-monitor-v2-monitor-settings-panel__small">{{ platformLabel(group.platform) }} · #{{ group.id }}</small>
            </label>
          </div>
          <p v-if="groups.length === 0" class="features-channel-monitor-v2-monitor-settings-panel__description-3 empty-state">{{ t('channelMonitorV2.settings.groupsEmpty') }}</p>
        </div>
      </div>

      <div class="features-channel-monitor-v2-monitor-settings-panel__panel-7 card">
        <div class="features-channel-monitor-v2-monitor-settings-panel__panel-8 card-header">
          <h3 class="features-channel-monitor-v2-monitor-settings-panel__strong">{{ t('channelMonitorV2.settings.errorsTitle') }}</h3>
          <p class="features-channel-monitor-v2-monitor-settings-panel__description-2">
            {{ t('channelMonitorV2.settings.errorsHint') }}
          </p>
        </div>
        <div class="features-channel-monitor-v2-monitor-settings-panel__panel-14">
          <div class="features-channel-monitor-v2-monitor-settings-panel__panel-13">
            <label
              v-for="category in errorCategories"
              :key="category"
              class="features-channel-monitor-v2-monitor-settings-panel__label"
            >
              <input
                type="checkbox"
                class="features-channel-monitor-v2-monitor-settings-panel__field"
                :checked="isCategoryIgnored(category)"
                @change="toggleIgnoredCategory(category)"
              />
              <span class="features-channel-monitor-v2-monitor-settings-panel__text-4">
                {{ categoryLabel(category) }}
              </span>
              <small class="features-channel-monitor-v2-monitor-settings-panel__small-2">{{ category }}</small>
            </label>
          </div>
        </div>
        <div class="features-channel-monitor-v2-monitor-settings-panel__panel-15">
          {{
            t('channelMonitorV2.settings.ignoredSummary', {
              ignored: draft.ignored_error_categories?.length || 0,
              counted: countedErrorCategoryCount,
            })
          }}
        </div>
      </div>

      <div class="features-channel-monitor-v2-monitor-settings-panel__panel-7 card">
        <div class="features-channel-monitor-v2-monitor-settings-panel__panel-8 card-header">
          <h3 class="features-channel-monitor-v2-monitor-settings-panel__strong">{{ t('channelMonitorV2.settings.healthTitle') }}</h3>
          <p class="features-channel-monitor-v2-monitor-settings-panel__description-2">
            {{ t('channelMonitorV2.settings.healthHint') }}
          </p>
        </div>
        <div class="features-channel-monitor-v2-monitor-settings-panel__panel-16">
          <label class="features-channel-monitor-v2-monitor-settings-panel__label-2">
            <span class="input-label">{{ t('channelMonitorV2.settings.fields.minimumSample') }}</span>
            <input v-model.number="draft.health_thresholds.minimum_sample" class="input" type="number" min="1" max="10000" />
          </label>
          <label class="features-channel-monitor-v2-monitor-settings-panel__label-2">
            <span class="input-label">{{ t('channelMonitorV2.settings.fields.warningError') }}</span>
            <input v-model.number="warningErrorPercent" class="input" type="number" min="0" max="100" step="0.1" />
          </label>
          <label class="features-channel-monitor-v2-monitor-settings-panel__label-2">
            <span class="input-label">{{ t('channelMonitorV2.settings.fields.criticalError') }}</span>
            <input v-model.number="criticalErrorPercent" class="input" type="number" min="0" max="100" step="0.1" />
          </label>
          <label class="features-channel-monitor-v2-monitor-settings-panel__label-2">
            <span class="input-label">{{ t('channelMonitorV2.settings.fields.targetTtft') }}</span>
            <input v-model.number="draft.health_thresholds.target_ttft_ms" class="input" type="number" min="1" step="100" />
          </label>
          <label class="features-channel-monitor-v2-monitor-settings-panel__label-2">
            <span class="input-label">{{ t('channelMonitorV2.settings.fields.warningTtft') }}</span>
            <input v-model.number="draft.health_thresholds.warning_ttft_ms" class="input" type="number" min="1" step="100" />
          </label>
          <label class="features-channel-monitor-v2-monitor-settings-panel__label-2">
            <span class="input-label">{{ t('channelMonitorV2.settings.fields.criticalTtft') }}</span>
            <input v-model.number="draft.health_thresholds.critical_ttft_ms" class="input" type="number" min="1" step="100" />
          </label>
          <label class="features-channel-monitor-v2-monitor-settings-panel__label-2">
            <span class="input-label">{{ t('channelMonitorV2.settings.fields.warningCache') }}</span>
            <input v-model.number="warningCachePercent" class="input" type="number" min="0" max="100" step="0.1" />
          </label>
          <label class="features-channel-monitor-v2-monitor-settings-panel__label-2">
            <span class="input-label">{{ t('channelMonitorV2.settings.fields.criticalCache') }}</span>
            <input v-model.number="criticalCachePercent" class="input" type="number" min="0" max="100" step="0.1" />
          </label>
        </div>
      </div>

      <div class="features-channel-monitor-v2-monitor-settings-panel__panel-17">
        <div class="features-channel-monitor-v2-monitor-settings-panel__panel-18">
          <template v-if="namedModelCount === 0">
            {{ t('channelMonitorV2.settings.namedModelsEmpty') }}
          </template>
          <template v-else>
            {{ t('channelMonitorV2.settings.namedModelsCount', { count: namedModelCount }) }}
          </template>
        </div>
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Toggle from '@/components/common/Toggle.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import {
  getConfig,
  updateConfig,
  MONITOR_ERROR_CATEGORIES,
  type MonitorConfig,
} from '@/api/channelMonitorV2'
import { adminAPI } from '@/api/admin'
import type { AdminGroup } from '@/types'

const { t, te } = useI18n()
const appStore = useAppStore()
const loading = ref(true)
const saving = ref(false)
const draft = ref<MonitorConfig | null>(null)
const original = ref('')
const groups = ref<AdminGroup[]>([])

const dirty = computed(() => (draft.value ? JSON.stringify(draft.value) !== original.value : false))
const namedModelCount = computed(
  () => draft.value?.platforms.filter((p) => p.enabled).reduce((sum, p) => sum + p.models.length, 0) || 0
)
const errorCategories = MONITOR_ERROR_CATEGORIES
const countedErrorCategoryCount = computed(
  () => errorCategories.length - (draft.value?.ignored_error_categories?.length || 0)
)
const defaultThresholds = {
  minimum_sample: 50,
  warning_error_rate: 0.05,
  critical_error_rate: 0.20,
  target_ttft_ms: 3000,
  warning_ttft_ms: 3000,
  critical_ttft_ms: 10000,
  // Higher is better: below 85% watch, below 60% critical.
  warning_cache_rate: 0.85,
  critical_cache_rate: 0.60,
  error_weight: 0.60,
  ttft_weight: 0.20,
  cache_weight: 0.20,
}

/** Factory ignored categories (matches backend DefaultChannelMonitorV2IgnoredErrorCategories). */
const defaultIgnoredErrorCategories = [
  'authentication',
  'client_cancelled',
  'content_policy',
  'context_limit',
  'group_access',
  'model_unsupported',
  'not_found',
  'quota_or_balance',
] as const
function percentModel(key: 'warning_error_rate' | 'critical_error_rate' | 'warning_cache_rate' | 'critical_cache_rate') {
  return computed({
    get: () => ((draft.value?.health_thresholds?.[key] ?? defaultThresholds[key]) * 100),
    set: (value: number) => {
      if (!draft.value) return
      draft.value.health_thresholds[key] = Math.max(0, Math.min(100, Number(value) || 0)) / 100
    },
  })
}
const warningErrorPercent = percentModel('warning_error_rate')
const criticalErrorPercent = percentModel('critical_error_rate')
const warningCachePercent = percentModel('warning_cache_rate')
const criticalCachePercent = percentModel('critical_cache_rate')

function setModels(platform: MonitorConfig['platforms'][number], event: Event) {
  platform.models = [
    ...new Set(
      (event.target as HTMLInputElement).value
        .split(',')
        .map((v) => v.trim())
        .filter(Boolean)
    ),
  ].sort()
}

function toggleGroup(id: string) {
  if (!draft.value) return
  draft.value.group_ids = draft.value.group_ids.includes(id)
    ? draft.value.group_ids.filter((value) => value !== id)
    : [...draft.value.group_ids, id].sort()
}

function isCategoryIgnored(category: string): boolean {
  return Boolean(draft.value?.ignored_error_categories?.includes(category))
}

function toggleIgnoredCategory(category: string) {
  if (!draft.value) return
  const current = new Set(draft.value.ignored_error_categories || [])
  if (current.has(category)) current.delete(category)
  else current.add(category)
  draft.value.ignored_error_categories = [...current].sort()
}

function categoryLabel(category: string) {
  const key = `channelMonitorV2.errorCategories.${category}`
  return te(key) ? t(key) : category
}

function platformLabel(value: string) {
  return (
    {
      anthropic: 'Claude',
      openai: 'OpenAI',
      grok: 'Grok',
      kimi: 'Kimi',
      zhipu: 'Zhipu',
      deepseek: 'DeepSeek',
    } as Record<string, string>
  )[value] || value
}

function normalizeConfig(value: MonitorConfig): MonitorConfig {
  const ignored = value.ignored_error_categories
  return {
    ...value,
    health_thresholds: { ...defaultThresholds, ...(value.health_thresholds || {}) },
    // Preserve explicit empty arrays from the server (operator cleared all).
    ignored_error_categories: [
      ...(ignored == null ? [...defaultIgnoredErrorCategories] : ignored),
    ].sort(),
  }
}

async function load() {
  loading.value = true
  try {
    const [value, groupRows] = await Promise.all([getConfig(), adminAPI.groups.getAllIncludingInactive()])
    const normalized = normalizeConfig(value)
    draft.value = structuredClone(normalized)
    groups.value = groupRows
    original.value = JSON.stringify(normalized)
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('channelMonitorV2.settings.loadFailed')))
  } finally {
    loading.value = false
  }
}

async function save() {
  if (!draft.value) return
  saving.value = true
  try {
    const payload = normalizeConfig(draft.value)
    const value = await updateConfig(payload)
    const normalized = normalizeConfig(value)
    draft.value = structuredClone(normalized)
    original.value = JSON.stringify(normalized)
    appStore.showSuccess(t('channelMonitorV2.settings.saveSuccess'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('channelMonitorV2.settings.saveFailed')))
    await load()
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>
