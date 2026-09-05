<template>
  <AppLayout>
    <div class="views-user-channel-status-v2-view__panel">
      <!-- Ops-style elevated shell: title toolbar + filters (mirrors OpsDashboardHeader) -->
      <section
        class="views-user-channel-status-v2-view__section channel-monitor-shell card"
      >
        <header class="views-user-channel-status-v2-view__header page-header">
          <div class="views-user-channel-status-v2-view__panel-2">
            <h1 class="views-user-channel-status-v2-view__heading page-title">
              <span class="views-user-channel-status-v2-view__text">
                <Icon name="chart" size="sm" />
              </span>
              {{ t('channelMonitorV2.title') }}
            </h1>
            <div class="views-user-channel-status-v2-view__panel-3 page-description">
              <span class="views-user-channel-status-v2-view__text-2">
                <span
                  class="views-user-channel-status-v2-view__text-3"
                  :class="loading || refreshing ? 'views-user-channel-status-v2-view__text-21' : 'views-user-channel-status-v2-view__text-22'"
                ></span>
              </span>
              <span v-if="refreshing" class="views-user-channel-status-v2-view__text-4">
                <LoadingSpinner size="sm" />
                {{ t('channelMonitorV2.updating') }}
              </span>
              <span v-else-if="snapshot?.coverage.data_through">
                {{ t('channelMonitorV2.updatedTo', { time: formatTime(snapshot.coverage.data_through) }) }}
              </span>
              <span v-else class="views-user-channel-status-v2-view__text-5">{{ t('common.loading') }}</span>
              <span
                v-if="snapshot && !snapshot.coverage.coverage_complete && !bootstrapActive"
                class="badge badge-warning"
              >
                {{ t('channelMonitorV2.partialCoverage') }}
              </span>
              <span
                v-if="bootstrapActive"
                class="views-user-channel-status-v2-view__text-6 badge badge-primary"
              >
                <LoadingSpinner size="sm" />
                {{ t('channelMonitorV2.bootstrap.progress', { percent: bootstrapPercent }) }}
              </span>
            </div>
          </div>
          <button
            class="views-user-channel-status-v2-view__action btn btn-secondary btn-icon"
            type="button"
            :title="t('common.refresh')"
            :disabled="loading"
            @click="reload(false)"
          >
            <Icon name="refresh" size="sm" :class="loading ? 'views-user-channel-status-v2-view__icon' : ''" />
          </button>
        </header>

        <!-- First-upgrade silent backfill: show until 30d product window is covered -->
        <div
          v-if="bootstrapActive"
          class="views-user-channel-status-v2-view__panel-4"
          role="status"
          aria-live="polite"
        >
          <div class="views-user-channel-status-v2-view__panel-5">
            <div class="views-user-channel-status-v2-view__panel-6">
              <p class="views-user-channel-status-v2-view__description">
                {{ t('channelMonitorV2.bootstrap.title') }}
              </p>
              <p class="views-user-channel-status-v2-view__description-2">
                {{ t('channelMonitorV2.bootstrap.description') }}
              </p>
            </div>
            <span class="views-user-channel-status-v2-view__text-7">
              {{ t('channelMonitorV2.bootstrap.progress', { percent: bootstrapPercent }) }}
            </span>
          </div>
          <div
            class="views-user-channel-status-v2-view__panel-7"
            role="progressbar"
            :aria-valuenow="bootstrapPercent"
            aria-valuemin="0"
            aria-valuemax="100"
            :aria-label="t('channelMonitorV2.bootstrap.working')"
          >
            <div
              class="views-user-channel-status-v2-view__panel-8"
              :style="{ width: `${bootstrapPercent}%` }"
            />
          </div>
        </div>

        <!-- Single compact toolbar row: range · filters · view controls -->
        <div class="views-user-channel-status-v2-view__panel-9 monitor-toolbar">
          <div
            class="views-user-channel-status-v2-view__panel-10 tabs monitor-tabs monitor-range-tabs"
            role="group"
            :aria-label="t('channelMonitorV2.timeRange')"
          >
            <button
              v-for="option in ranges"
              :key="option.value"
              type="button"
              class="views-user-channel-status-v2-view__action-2 tab monitor-tabs__tab"
              :class="filter.range === option.value ? 'tab-active' : ''"
              @click="setRange(option.value)"
            >
              {{ option.label }}
            </button>
          </div>

          <span class="views-user-channel-status-v2-view__text-8" aria-hidden="true"></span>

          <FilterMultiSelect
            v-model="filter.platforms"
            compact
            :label="t('channelMonitorV2.filters.platform')"
            :all-label="t('channelMonitorV2.filters.allPlatforms')"
            :options="platformOptions"
          />
          <FilterMultiSelect
            v-model="selectedGroupIds"
            compact
            :label="t('channelMonitorV2.filters.group')"
            :all-label="t('channelMonitorV2.filters.allGroups')"
            :options="groupOptions"
          />
          <FilterMultiSelect
            v-model="filter.models"
            compact
            :label="t('channelMonitorV2.filters.model')"
            :all-label="t('channelMonitorV2.filters.allModels')"
            :options="modelOptions"
          />
          <button
            type="button"
            class="views-user-channel-status-v2-view__action-3 btn btn-ghost btn-sm"
            :disabled="!hasDimensionFilter"
            :class="!hasDimensionFilter ? 'views-user-channel-status-v2-view__action-7' : ''"
            @click="clearDimensions"
          >
            {{ t('channelMonitorV2.clearFilters') }}
          </button>
        </div>
      </section>

      <!-- Overview KPI: success · TTFT · tokens/s(optional) · cache · (+ RPM when throughput visible) -->
      <section
        v-if="snapshot"
        class="views-user-channel-status-v2-view__section-2"
        :class="showThroughput ? 'views-user-channel-status-v2-view__section-4' : 'views-user-channel-status-v2-view__section-5'"
        :aria-label="t('channelMonitorV2.summaryAria')"
      >
        <MetricCell
          :label="t('channelMonitorV2.metrics.successRate')"
          :value="formatPercent(1 - snapshot.metrics.error_rate)"
          :detail="t('channelMonitorV2.metrics.errorRateValue', { value: formatPercent(snapshot.metrics.error_rate) })"
          :state="snapshot.health.error_rate"
        />
        <MetricCell
          :label="t('channelMonitorV2.metrics.ttftP50')"
          :value="formatMs(snapshot.metrics.ttft.p50_ms)"
          :detail="latencyKpiSecondary(snapshot.metrics.ttft)"
          :title="latencyDetail(snapshot.metrics.ttft)"
          :state="snapshot.health.ttft"
        />
        <MetricCell
          v-if="showThroughput"
          :label="t('channelMonitorV2.metrics.tps')"
          :value="formatTps(snapshot.metrics.tpm)"
          :detail="t('channelMonitorV2.metrics.tpsDetail')"
          :title="exactTps(snapshot.metrics.tpm)"
        />
        <MetricCell
          :label="t('channelMonitorV2.metrics.cacheRate')"
          :value="formatPercent(snapshot.metrics.cache_rate)"
          :detail="t('channelMonitorV2.metrics.cacheDetail')"
          :state="snapshot.health.cache || snapshot.health.overall"
        />
        <MetricCell
          v-if="showThroughput"
          :label="t('channelMonitorV2.metrics.rpm')"
          :value="formatRate(snapshot.metrics.rpm)"
          :detail="t('channelMonitorV2.metrics.rpmDetail')"
          :title="exactRate(snapshot.metrics.rpm)"
        />
      </section>
      <section
        v-else-if="loading"
        class="views-user-channel-status-v2-view__section-2"
        :class="showThroughput ? 'views-user-channel-status-v2-view__section-4' : 'views-user-channel-status-v2-view__section-5'"
        aria-hidden="true"
      >
        <div
          v-for="i in (showThroughput ? 5 : 4)"
          :key="i"
          class="views-user-channel-status-v2-view__panel-11"
        />
      </section>

      <div class="views-user-channel-status-v2-view__panel-12">
        <MonitorTrendChart
          :trend="snapshot?.trend || []"
          :coverage="snapshot?.coverage || null"
          :loading="loading && !snapshot"
        />
      </div>

      <section class="views-user-channel-status-v2-view__section-3 card">
        <div class="views-user-channel-status-v2-view__panel-14 monitor-detail-header">
          <nav class="views-user-channel-status-v2-view__navigation tabs monitor-tabs monitor-detail-tabs" role="tablist" :aria-label="t('channelMonitorV2.tabs.aria')">
            <button
              v-for="item in tabs"
              :key="item.value"
              type="button"
              role="tab"
              class="views-user-channel-status-v2-view__action-5 tab monitor-tabs__tab"
              :aria-selected="activeTab === item.value"
              :class="activeTab === item.value ? 'tab-active' : ''"
              @click="activeTab = item.value"
            >
              {{ item.label }}
            </button>
          </nav>
        </div>
        <div class="views-user-channel-status-v2-view__panel-15">
          <div v-if="activeTab === 'models'" class="monitor-table-container table-container">
            <table class="table monitor-table monitor-table--models">
              <thead>
                <tr>
                  <th>{{ t('channelMonitorV2.table.platformModel') }}</th>
                  <th>{{ t('channelMonitorV2.metrics.successRate') }}</th>
                  <th>{{ t('channelMonitorV2.metrics.ttftP50') }}</th>
                  <th v-if="showThroughput">{{ t('channelMonitorV2.metrics.tps') }}</th>
                  <th>{{ t('channelMonitorV2.metrics.cacheRate') }}</th>
                  <th v-if="showThroughput">{{ t('channelMonitorV2.metrics.rpm') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="row in modelRows"
                  :key="`${row.platform}:${row.model}`"
                  class="views-user-channel-status-v2-view__row monitor-table__row--interactive"
                  @click="drillModel(row)"
                >
                  <td>
                    <div class="views-user-channel-status-v2-view__panel-17">
                      <span :class="statusDot(row.health)" aria-hidden="true"></span>
                      <div>
                        <span class="views-user-channel-status-v2-view__text-11">{{ row.platform }}</span>
                        <strong class="views-user-channel-status-v2-view__strong">
                          {{ row.model === '__other__' ? t('channelMonitorV2.otherModels') : row.model }}
                        </strong>
                      </div>
                    </div>
                  </td>
                  <td>
                    <span class="views-user-channel-status-v2-view__text-12">{{ formatPercent(1 - row.metrics.error_rate) }}</span>
                    <small class="views-user-channel-status-v2-view__small">{{ t('channelMonitorV2.metrics.errorRateValue', { value: formatPercent(row.metrics.error_rate) }) }}</small>
                  </td>
                  <td>
                    <span class="views-user-channel-status-v2-view__text-12">{{ formatMs(row.metrics.ttft.p50_ms) }}</span>
                    <small class="views-user-channel-status-v2-view__small">{{ latencyDetail(row.metrics.ttft) }}</small>
                  </td>
                  <td v-if="showThroughput" :title="exactTps(row.metrics.tpm)">{{ formatTps(row.metrics.tpm) }}</td>
                  <td>{{ formatPercent(row.metrics.cache_rate) }}</td>
                  <td v-if="showThroughput">{{ formatRate(row.metrics.rpm) }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-else-if="activeTab === 'errors'" class="views-user-channel-status-v2-view__panel-18">
            <div
              v-for="row in errorRows"
              :key="row.category"
              class="views-user-channel-status-v2-view__panel-19"
              :class="row.ignored ? 'views-user-channel-status-v2-view__panel-25' : ''"
            >
              <button
                type="button"
                class="views-user-channel-status-v2-view__action-6"
                @click="toggleError(row.category)"
              >
                <span class="views-user-channel-status-v2-view__text-13">
                  <span class="views-user-channel-status-v2-view__text-14">{{ errorLabel(row.category) }}</span>
                  <span v-if="row.ignored" class="views-user-channel-status-v2-view__text-15 badge badge-gray">{{ t('channelMonitorV2.ignored') }}</span>
                </span>
                <span class="views-user-channel-status-v2-view__text-16">
                  <i
                    class="views-user-channel-status-v2-view__i"
                    :class="row.ignored ? 'views-user-channel-status-v2-view__i-2' : 'views-user-channel-status-v2-view__i-3'"
                    :style="{ width: `${Math.max(2, row.rate * 100)}%` }"
                  ></i>
                </span>
                <small
                  class="views-user-channel-status-v2-view__small-2"
                  :class="row.ignored ? 'views-user-channel-status-v2-view__text-5' : 'views-user-channel-status-v2-view__small-3'"
                >{{ formatPercent(row.rate) }}</small>
                <Icon name="chevronDown" size="sm" :class="['views-user-channel-status-v2-view__icon-2', expandedErrors.has(row.category) ? 'views-user-channel-status-v2-view__icon-3' : '']" />
              </button>
              <div v-if="expandedErrors.has(row.category)" class="views-user-channel-status-v2-view__panel-20">
                <template v-if="isAdmin && (row.details || []).length">
                  <div
                    v-for="(detail, index) in row.details || []"
                    :key="`${row.category}:${index}:${detail.message}`"
                    class="views-user-channel-status-v2-view__panel-21"
                  >
                    <div class="views-user-channel-status-v2-view__panel-22">
                      <span class="views-user-channel-status-v2-view__text-17 badge badge-gray">{{ detail.platform || '-' }}</span>
                      <span class="views-user-channel-status-v2-view__text-18">{{ detail.model || '-' }}</span>
                      <span v-if="detail.status_code" class="views-user-channel-status-v2-view__text-5">{{ t('channelMonitorV2.errorDetail.http', { code: detail.status_code }) }}</span>
                      <span v-if="detail.upstream_status_code" class="views-user-channel-status-v2-view__text-5">{{ t('channelMonitorV2.errorDetail.upstream', { code: detail.upstream_status_code }) }}</span>
                      <span class="views-user-channel-status-v2-view__text-19">×{{ detail.count }}</span>
                    </div>
                    <p class="views-user-channel-status-v2-view__description-3">{{ detail.message || detail.error_type || t('channelMonitorV2.errorDetail.noMessage') }}</p>
                  </div>
                </template>
                <p v-else class="views-user-channel-status-v2-view__small">{{ t('channelMonitorV2.errorDetail.empty') }}</p>
              </div>
            </div>
          </div>

          <div v-else class="monitor-table-container table-container">
            <table class="table monitor-table monitor-table--users">
              <thead>
                <tr>
                  <th class="views-user-channel-status-v2-view__heading-2">{{ t('channelMonitorV2.table.rank') }}</th>
                  <th>{{ t('channelMonitorV2.table.user') }}</th>
                  <th>{{ t('channelMonitorV2.metrics.successRate') }}</th>
                  <th>{{ t('channelMonitorV2.metrics.ttftP50') }}</th>
                  <th v-if="showThroughput">{{ t('channelMonitorV2.metrics.tps') }}</th>
                  <th>{{ t('channelMonitorV2.metrics.cacheRate') }}</th>
                  <th v-if="showThroughput">{{ t('channelMonitorV2.metrics.rpm') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="row in userRows"
                  :key="row.user_id || row.display_label"
                  :class="row.is_self
                    ? 'views-user-channel-status-v2-view__row-2'
                    : ''"
                >
                  <td><MonitorRankBadge :rank="row.rank" /></td>
                  <td>
                    <strong
                      class="views-user-channel-status-v2-view__strong-2"
                      :class="row.is_self ? 'views-user-channel-status-v2-view__strong-3' : 'views-user-channel-status-v2-view__strong-4'"
                    >
                      {{ row.display_label }}
                      <span
                        v-if="row.is_self"
                        class="views-user-channel-status-v2-view__text-20 badge badge-primary"
                      >{{ t('channelMonitorV2.currentUser') }}</span>
                    </strong>
                  </td>
                  <td>
                    <span class="views-user-channel-status-v2-view__text-12">{{ formatPercent(1 - row.metrics.error_rate) }}</span>
                    <small class="views-user-channel-status-v2-view__small">{{ t('channelMonitorV2.metrics.errorRateValue', { value: formatPercent(row.metrics.error_rate) }) }}</small>
                  </td>
                  <td>
                    <span class="views-user-channel-status-v2-view__text-12">{{ formatMs(row.metrics.ttft.p50_ms) }}</span>
                    <small class="views-user-channel-status-v2-view__small">{{ latencyDetail(row.metrics.ttft) }}</small>
                  </td>
                  <td v-if="showThroughput" :title="exactTps(row.metrics.tpm)">{{ formatTps(row.metrics.tpm) }}</td>
                  <td>{{ formatPercent(row.metrics.cache_rate) }}</td>
                  <td v-if="showThroughput">{{ formatRate(row.metrics.rpm) }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-if="tabLoading" class="views-user-channel-status-v2-view__panel-23 empty-state">{{ t('common.loading') }}</div>
          <div v-else-if="activeRowsEmpty" class="views-user-channel-status-v2-view__panel-24 empty-state">
            <p class="views-user-channel-status-v2-view__description-4 empty-state-title">
              {{
                bootstrapActive
                  ? t('channelMonitorV2.bootstrap.title')
                  : t('channelMonitorV2.empty.title')
              }}
            </p>
            <p class="empty-state-description">
              {{
                bootstrapActive
                  ? t('channelMonitorV2.bootstrap.description')
                  : t('channelMonitorV2.empty.description')
              }}
            </p>
          </div>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import FilterMultiSelect from '@/features/channel-monitor-v2/FilterMultiSelect.vue'
import MetricCell from '@/features/channel-monitor-v2/MetricCell.vue'
import MonitorRankBadge from '@/features/channel-monitor-v2/MonitorRankBadge.vue'
import MonitorTrendChart from '@/features/channel-monitor-v2/MonitorTrendChart.vue'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'
import { extractApiErrorMessage } from '@/utils/apiError'
import { isChannelMonitorThroughputHidden } from '@/utils/featureFlags'
import * as api from '@/api/channelMonitorV2'
import type {
  HealthState,
  MonitorDimensions,
  MonitorErrorRow,
  MonitorFilter,
  MonitorHealth,
  MonitorModelRow,
  MonitorRange,
  MonitorSnapshot,
  MonitorUserRow,
} from '@/api/channelMonitorV2'
import {
  formatLatencyKpiSecondary,
  formatLatencyPrivacy,
  formatMonitorMs,
  formatMonitorPercent,
  formatMonitorThroughput,
  formatMonitorTokensPerSecond,
  tokensPerSecondFromTpm,
  healthScoreClass,
  monitorErrorCategoryLabel,
} from '@/features/channel-monitor-v2/monitorFormat'

type Tab = 'models' | 'errors' | 'users'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()
const { t, te, locale } = useI18n()
const isAdmin = computed(() => authStore.isAdmin)
/** Admins always see RPM/TPM; users honor the hide-throughput system setting. */
const showThroughput = computed(() => isAdmin.value || !isChannelMonitorThroughputHidden())

const ranges = computed(() => [
  { value: '90m' as MonitorRange, label: t('channelMonitorV2.ranges.90m') },
  { value: '24h' as MonitorRange, label: t('channelMonitorV2.ranges.24h') },
  { value: '7d' as MonitorRange, label: t('channelMonitorV2.ranges.7d') },
  { value: '14d' as MonitorRange, label: t('channelMonitorV2.ranges.14d') },
  { value: '30d' as MonitorRange, label: t('channelMonitorV2.ranges.30d') },
])
const tabs = computed(() => [
  { value: 'models' as Tab, label: t('channelMonitorV2.tabs.models') },
  { value: 'errors' as Tab, label: t('channelMonitorV2.tabs.errors') },
  { value: 'users' as Tab, label: t('channelMonitorV2.tabs.users') },
])
const filter = ref<MonitorFilter>({
  range: parseRange(route.query.range),
  platforms: csv(route.query.platform),
  groupIds: csv(route.query.group),
  models: csv(route.query.model),
})
const activeTab = ref<Tab>(
  (['models', 'errors', 'users'].includes(String(route.query.tab)) ? route.query.tab : 'models') as Tab
)
const dimensions = ref<MonitorDimensions>({ platforms: [], groups: [], models: [] })
const snapshot = ref<MonitorSnapshot | null>(null)
const modelRows = ref<MonitorModelRow[]>([])
const errorRows = ref<MonitorErrorRow[]>([])
const userRows = ref<MonitorUserRow[]>([])
const loading = ref(false)
const tabLoading = ref(false)
const refreshing = ref(false)
const expandedErrors = ref(new Set<string>())
let controller: AbortController | null = null
let sequence = 0
let autoRefreshTimer: number | null = null

const hasDimensionFilter = computed(
  () => filter.value.platforms.length + filter.value.groupIds.length + filter.value.models.length > 0
)
// Full platform catalog (never pruned). Groups/models cascade by selected platforms
// so choosing a platform narrows the other pickers without collapsing platforms.
const platformOptions = computed(() =>
  (dimensions.value.platforms || []).map((item) => ({
    value: item.value,
    label: item.label,
  }))
)
const selectedPlatforms = computed(() => new Set(filter.value.platforms))
const groupOptions = computed(() =>
  (dimensions.value.groups || [])
    .filter(
      (item) =>
        selectedPlatforms.value.size === 0 ||
        !item.platform ||
        selectedPlatforms.value.has(item.platform),
    )
    .map((item) => ({
      value: String(item.id),
      label: item.platform ? `${item.platform} / ${item.name || `#${item.id}`}` : item.name || `#${item.id}`,
    }))
)
const modelOptions = computed(() =>
  (dimensions.value.models || [])
    .filter(
      (item) =>
        selectedPlatforms.value.size === 0 ||
        !item.platform ||
        selectedPlatforms.value.has(item.platform),
    )
    .map((item) => ({
      value: item.value,
      label:
        item.platform && !item.label.includes(item.platform)
          ? `${item.platform} / ${item.label}`
          : item.label,
    }))
)
const selectedGroupIds = computed({
  get: () => filter.value.groupIds,
  set: (value: string[]) => {
    filter.value.groupIds = value.filter(Boolean)
  },
})
// Soft-prune group/model selections that fall outside the platform cascade.
// Do NOT wipe when options are temporarily empty (loading); only drop invalid ids.
watch(
  [groupOptions, modelOptions],
  () => {
    if (groupOptions.value.length > 0) {
      const allowed = new Set(groupOptions.value.map((item) => item.value))
      const next = filter.value.groupIds.filter((id) => allowed.has(id))
      if (next.length !== filter.value.groupIds.length) {
        filter.value.groupIds = next
      }
    }
    if (modelOptions.value.length > 0) {
      const allowed = new Set(modelOptions.value.map((item) => item.value))
      const next = filter.value.models.filter((model) => allowed.has(model))
      if (next.length !== filter.value.models.length) {
        filter.value.models = next
      }
    }
  },
  { flush: 'post' },
)
const activeRowsEmpty = computed(() =>
  activeTab.value === 'models'
    ? modelRows.value.length === 0
    : activeTab.value === 'errors'
      ? errorRows.value.length === 0
      : userRows.value.length === 0
)
/** First-upgrade backfill toward 90m/24h/7d/14d/30d; banner hides when backend omits bootstrap. */
const bootstrapActive = computed(() => Boolean(snapshot.value?.coverage?.bootstrap?.active))
const bootstrapPercent = computed(() => {
  const raw = snapshot.value?.coverage?.bootstrap?.progress_percent
  if (typeof raw !== 'number' || Number.isNaN(raw)) return 0
  return Math.min(100, Math.max(0, Math.round(raw)))
})
function csv(value: unknown) {
  return typeof value === 'string' ? value.split(',').filter(Boolean) : []
}
function parseRange(value: unknown): MonitorRange {
  return ['90m', '24h', '7d', '14d', '30d'].includes(String(value)) ? (value as MonitorRange) : '90m'
}
function syncQuery() {
  void router.replace({
    query: {
      range: filter.value.range,
      platform: filter.value.platforms.join(',') || undefined,
      group: filter.value.groupIds.join(',') || undefined,
      model: filter.value.models.join(',') || undefined,
      tab: activeTab.value,
    },
  })
}
/** Dimensions catalog: range only — never re-filtered by platform/group/model selection. */
async function loadDimensions(signal?: AbortSignal, id = sequence) {
  const rangeOnly: MonitorFilter = {
    range: filter.value.range,
    platforms: [],
    groupIds: [],
    models: [],
  }
  const next = await api.getDimensions(rangeOnly, isAdmin.value, signal)
  if (id !== sequence) return
  dimensions.value = next
}

async function loadMetrics(signal?: AbortSignal, id = sequence) {
  const nextSnapshot = await api.getSnapshot(filter.value, isAdmin.value, signal)
  if (id !== sequence) return
  snapshot.value = nextSnapshot
  scheduleAutoRefresh()
  await loadTab(signal, id)
}

async function reload(silent = true) {
  controller?.abort()
  const request = new AbortController()
  controller = request
  const id = ++sequence
  refreshing.value = true
  if (!silent) loading.value = true
  try {
    // Catalog + metrics in parallel; catalog ignores dimension filters so options never shrink.
    await Promise.all([
      loadDimensions(request.signal, id),
      loadMetrics(request.signal, id),
    ])
  } catch (error) {
    if ((error as { name?: string }).name !== 'CanceledError') {
      appStore.showError(extractApiErrorMessage(error, t('channelMonitorV2.loadFailed')))
    }
  } finally {
    if (id === sequence) {
      loading.value = false
      tabLoading.value = false
      refreshing.value = false
    }
  }
}

/** When only range changes, still refresh dimensions; dimension filters only re-load metrics. */
async function reloadMetricsOnly(silent = true) {
  controller?.abort()
  const request = new AbortController()
  controller = request
  const id = ++sequence
  refreshing.value = true
  if (!silent) loading.value = true
  try {
    await loadMetrics(request.signal, id)
  } catch (error) {
    if ((error as { name?: string }).name !== 'CanceledError') {
      appStore.showError(extractApiErrorMessage(error, t('channelMonitorV2.loadFailed')))
    }
  } finally {
    if (id === sequence) {
      loading.value = false
      tabLoading.value = false
      refreshing.value = false
    }
  }
}
async function loadTab(signal?: AbortSignal, id = sequence) {
  tabLoading.value = true
  try {
    if (activeTab.value === 'models') {
      modelRows.value = (await api.getModels(filter.value, isAdmin.value, signal)).items || []
    } else if (activeTab.value === 'errors') {
      errorRows.value = (await api.getErrors(filter.value, isAdmin.value, signal)).items || []
    } else {
      userRows.value = (await api.getUsers(filter.value, isAdmin.value, signal)).items || []
    }
  } catch (error) {
    const e = error as { name?: string; code?: string }
    if (e?.name === 'AbortError' || e?.name === 'CanceledError' || e?.code === 'ERR_CANCELED') return
    appStore.showError(extractApiErrorMessage(error, t('channelMonitorV2.detailLoadFailed')))
  } finally {
    if (id === sequence) tabLoading.value = false
  }
}
function setRange(value: MonitorRange) {
  filter.value.range = value
}
function clearDimensions() {
  // Replace arrays so deep watch always fires and metrics reload full window.
  filter.value = {
    ...filter.value,
    platforms: [],
    groupIds: [],
    models: [],
  }
}
function scheduleAutoRefresh() {
  if (autoRefreshTimer) {
    window.clearInterval(autoRefreshTimer)
    autoRefreshTimer = null
  }
  // Poll faster while first-upgrade bootstrap is filling 90m→30d so the progress bar moves.
  const seconds = bootstrapActive.value
    ? 10
    : snapshot.value?.config?.refresh_interval_seconds || 300
  autoRefreshTimer = window.setInterval(() => {
    if (!loading.value && !refreshing.value) {
      void reload(true)
    }
  }, Math.max(bootstrapActive.value ? 10 : 60, seconds) * 1000)
}
function drillModel(row: MonitorModelRow) {
  filter.value.platforms = [row.platform]
  filter.value.models = [row.model]
}
function formatRate(value: number) {
  return formatMonitorThroughput(value)
}
function exactRate(value: number) {
  return Intl.NumberFormat(locale.value || undefined, { maximumFractionDigits: 2 }).format(value || 0)
}
function formatTps(tpm: number | null | undefined) {
  return formatMonitorTokensPerSecond(tpm)
}
function exactTps(tpm: number | null | undefined) {
  return Intl.NumberFormat(locale.value || undefined, { maximumFractionDigits: 3 }).format(
    tokensPerSecondFromTpm(tpm),
  )
}
function formatPercent(value: number) {
  return formatMonitorPercent(value)
}
function formatMs(value: number | null) {
  return formatMonitorMs(value)
}
function latencyDetail(metric: {
  p50_ms: number | null
  p90_ms?: number | null
  p95_ms: number | null
  avg_ms?: number | null
}) {
  return formatLatencyPrivacy(metric.p50_ms, metric.p90_ms, metric.avg_ms, metric.p95_ms)
}
/** KPI secondary: AVG · P90 under the P50 primary value. */
function latencyKpiSecondary(metric: {
  p90_ms?: number | null
  p95_ms: number | null
  avg_ms?: number | null
}) {
  return formatLatencyKpiSecondary(metric.avg_ms, metric.p90_ms, metric.p95_ms)
}
function formatTime(value: string) {
  return new Intl.DateTimeFormat(locale.value || undefined, {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  }).format(new Date(value))
}
function statusDot(health?: MonitorHealth | HealthState) {
  if (!health || typeof health === 'string') {
    return `status-dot health-${health || 'unknown'}`
  }
  // Prefer multi-band score when available; otherwise fall back to the coarse
  // overall state for mixed-version/older payloads.
  const klass =
    health.score != null
      ? healthScoreClass(health, 'overall', 0)
      : `health-${health.overall || 'unknown'}`
  return `status-dot ${klass}`
}
function errorLabel(value: string) {
  const key = `channelMonitorV2.errorCategories.${value}`
  return te(key) ? t(key) : monitorErrorCategoryLabel(value)
}
function toggleError(category: string) {
  const next = new Set(expandedErrors.value)
  if (next.has(category)) next.delete(category)
  else next.add(category)
  expandedErrors.value = next
}

let lastRange: MonitorRange = filter.value.range
watch(
  filter,
  () => {
    syncQuery()
    const rangeChanged = filter.value.range !== lastRange
    lastRange = filter.value.range
    if (rangeChanged) void reload(true)
    else void reloadMetricsOnly(true)
  },
  { deep: true }
)
watch(activeTab, () => {
  syncQuery()
  void loadTab()
})
onMounted(() => void reload(false))
onBeforeUnmount(() => {
  controller?.abort()
  if (autoRefreshTimer) window.clearInterval(autoRefreshTimer)
})
</script>

<style scoped>
.monitor-toolbar {
  flex-wrap: wrap;
  overflow: visible;
}

.monitor-toolbar :deep(.filter-menu) {
  flex: 1 1 8.5rem;
}

.monitor-tabs {
  width: fit-content;
  max-width: 100%;
  flex-wrap: nowrap;
}

.monitor-range-tabs {
  flex: 0 0 auto;
}

.monitor-tabs__tab {
  flex: 0 0 auto;
  min-height: 2rem;
  white-space: nowrap;
}

.monitor-detail-header {
  overflow-x: auto;
  scrollbar-width: none;
}

.monitor-detail-header::-webkit-scrollbar {
  display: none;
}

.monitor-detail-tabs {
  width: max-content;
}

.monitor-table-container {
  isolation: isolate;
  box-shadow:
    var(--glass-shadow),
    0 1px 0 var(--glass-highlight) inset;
}

.monitor-table {
  border-spacing: 0;
  border-collapse: separate;
  font-variant-numeric: tabular-nums;
}

.monitor-table td {
  transition:
    background-color 160ms ease,
    backdrop-filter 160ms ease;
}

.monitor-table tbody tr:hover td {
  background: var(--glass-bg-interactive);
  -webkit-backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-blur-xs-hover)) saturate(var(--glass-saturate));
}

.monitor-table--models {
  min-width: 45rem;
}

.monitor-table--users {
  min-width: 40rem;
}

.monitor-table__row--interactive {
  cursor: pointer;
}

@media (max-width: 639px) {
  .monitor-range-tabs {
    display: grid;
    width: 100%;
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }

  .monitor-range-tabs .monitor-tabs__tab {
    min-width: 0;
    text-align: center;
  }

  .monitor-toolbar :deep(.filter-menu) {
    flex-basis: calc(50% - 0.25rem);
  }
}

@media (max-width: 420px) {
  .monitor-toolbar :deep(.filter-menu) {
    flex-basis: 100%;
  }
}

.status-dot {
  display: inline-block;
  height: 0.5rem;
  width: 0.5rem;
  flex: none;
  border-radius: 9999px;
}
/* Multi-stop green → yellow → red score bands */
.health-score10 { background: #16a34a; }
.health-score9  { background: #22c55e; }
.health-score8  { background: #4ade80; }
.health-score7  { background: #a3e635; }
.health-score6  { background: #facc15; }
.health-score5  { background: #fbbf24; }
.health-score4  { background: #f59e0b; }
.health-score3  { background: #f97316; }
.health-score2  { background: #fb7185; }
.health-score1  { background: #f87171; }
.health-score0  { background: rgb(239, 67, 67); }
.health-healthy  { background: #22c55e; }
.health-warning  { background: #f59e0b; }
.health-critical { background: #ef4444; }
.health-unknown  { background: #9ca3af; }
.matrix-select {
  min-width: 10rem;
}
details > summary::-webkit-details-marker {
  display: none;
}
</style>
