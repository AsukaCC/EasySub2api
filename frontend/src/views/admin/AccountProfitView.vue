<template>
  <AppLayout>
    <div class="account-profit-view">
      <TablePageLayout>
        <template #filters>
          <div class="account-profit-view__toolbar">
            <div class="account-profit-view__title">
              <span class="account-profit-view__title-icon" aria-hidden="true">
                <Icon name="chart" size="md" />
              </span>
              <h1>{{ t('admin.accounts.profit.pageTitle') }}</h1>
            </div>
            <div class="account-profit-view__toolbar-actions">
              <div class="account-profit-view__columns-menu-wrap">
                <button
                  type="button"
                  class="account-profit-view__icon-button"
                  :title="t('admin.accounts.profit.viewColumns')"
                  :aria-expanded="showColumnsMenu"
                  @click="showColumnsMenu = !showColumnsMenu"
                >
                  <Icon name="grid" size="sm" />
                </button>
                <div v-if="showColumnsMenu" class="account-profit-view__columns-menu">
                  <div class="account-profit-view__column-group">
                    <span class="account-profit-view__column-group-title">{{ t('admin.accounts.profit.quota7d') }}</span>
                    <label class="account-profit-view__column-option">
                      <input v-model="columnVisibility.quota_7d" type="checkbox" />
                      <span>{{ t('admin.accounts.profit.quota7d') }}</span>
                    </label>
                  </div>
                  <div v-for="group in metricGroups" :key="group" class="account-profit-view__column-group">
                    <span class="account-profit-view__column-group-title">{{ metricGroupLabel(group) }}</span>
                    <label v-for="column in metricColumns.filter(item => item.group === group)" :key="column.key" class="account-profit-view__column-option">
                      <input v-model="columnVisibility[column.key]" type="checkbox" />
                      <span>{{ column.label }}</span>
                    </label>
                  </div>
                </div>
              </div>
              <button type="button" class="account-profit-view__refresh" :disabled="loading" @click="reload">
                <Icon name="refresh" size="sm" />
                {{ t('common.refresh') }}
              </button>
            </div>
          </div>
          <div class="account-profit-view__filters">
            <SearchInput
              :model-value="filters.search"
              :placeholder="t('admin.accounts.searchAccounts')"
              class="account-profit-view__search"
              @update:model-value="updateFilter('search', $event)"
              @search="reload"
            />
            <Select :model-value="filters.platform" :options="platformOptions" @update:model-value="updatePlatform" @change="reload" />
            <Select :model-value="filters.subscription_tier" :options="tierOptions" :disabled="loadingTiers" @update:model-value="updateFilter('subscription_tier', $event)" @change="reload" />
            <Select :model-value="filters.status" :options="statusOptions" @update:model-value="updateFilter('status', $event)" @change="reload" />
            <Select :model-value="filters.expiry_status" :options="expiryOptions" @update:model-value="updateFilter('expiry_status', $event)" @change="reload" />
          </div>
        </template>

        <template #table>
          <div class="account-profit-view__table-wrap">
            <table class="account-profit-view__table">
              <thead>
                <tr class="account-profit-view__group-row">
                  <th rowspan="2" class="is-sticky is-account-column">{{ t('admin.accounts.columns.name') }}</th>
                  <th rowspan="2" class="is-sticky is-platform-column">{{ t('admin.accounts.columns.platform') }}</th>
                  <th rowspan="2" class="is-sticky is-tier-column">{{ t('admin.accounts.columns.subscriptionTier') }}</th>
                  <th rowspan="2" class="is-expiry-column account-profit-view__group account-profit-view__group--identity">
                    <button type="button" class="account-profit-view__sort-button" @click="toggleSort('expires_at')">
                      {{ t('admin.accounts.columns.expiresAt') }}
                      <Icon v-if="sortBy === 'expires_at'" :name="sortOrder === 'asc' ? 'chevronUp' : 'chevronDown'" size="xs" />
                    </button>
                  </th>
                  <th v-if="hasVisible('quota_7d')" rowspan="2" class="is-quota-column account-profit-view__group account-profit-view__group--quota">
                    <button type="button" class="account-profit-view__sort-button" @click="toggleSort('quota_7d_utilization')">
                      {{ t('admin.accounts.profit.quota7d') }}
                      <Icon v-if="sortBy === 'quota_7d_utilization'" :name="sortOrder === 'asc' ? 'chevronUp' : 'chevronDown'" size="xs" />
                    </button>
                  </th>
                  <th v-if="visibleGroupCount('period')" :colspan="visibleGroupCount('period')" class="account-profit-view__group account-profit-view__group--period">{{ t('admin.accounts.profit.period7d') }}</th>
                  <th v-if="visibleGroupCount('expiry')" :colspan="visibleGroupCount('expiry')" class="account-profit-view__group account-profit-view__group--expiry">{{ t('admin.accounts.profit.expiry30d') }}</th>
                  <th v-if="visibleGroupCount('lifetime')" :colspan="visibleGroupCount('lifetime')" class="account-profit-view__group account-profit-view__group--lifetime">{{ t('admin.accounts.profit.lifetime') }}</th>
                  <th rowspan="2">{{ t('admin.accounts.columns.actions') }}</th>
                </tr>
                <tr class="account-profit-view__metric-row">
                  <th v-for="column in visibleMetricColumns" :key="column.key" :class="['account-profit-view__metric-header', metricClass(column), `is-group-${column.group}`]">
                    <button type="button" class="account-profit-view__sort-button" @click="toggleSort(column.sortBy)">
                      {{ column.label }}
                      <Icon v-if="sortBy === column.sortBy" :name="sortOrder === 'asc' ? 'chevronUp' : 'chevronDown'" size="xs" />
                    </button>
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="loading">
                  <td :colspan="tableColumnCount" class="account-profit-view__state">{{ t('common.loading') }}</td>
                </tr>
                <tr v-else-if="error">
                  <td :colspan="tableColumnCount" class="account-profit-view__state is-error">{{ error }}</td>
                </tr>
                <tr v-else-if="!accounts.length">
                  <td :colspan="tableColumnCount" class="account-profit-view__state">{{ t('admin.accounts.profit.listEmpty') }}</td>
                </tr>
                <template v-else>
                  <tr v-for="account in accounts" :key="account.id">
                  <td class="is-sticky is-account-column">
                    <div class="account-profit-view__account-name">{{ account.name }}</div>
                    <small class="account-profit-view__account-id">{{ account.id }}</small>
                  </td>
                  <td class="is-sticky is-platform-column"><span class="account-profit-view__platform">{{ platformLabel(account.platform) }}</span></td>
                  <td class="is-sticky is-tier-column"><span class="account-profit-view__tier">{{ account.subscription_tier || '-' }}</span></td>
                  <td class="is-expiry-column"><span class="account-profit-view__expiry">{{ formatExpiry(account.expires_at) }}</span></td>
                  <td v-if="hasVisible('quota_7d')" class="is-quota-column">
                    <div v-if="account.quota_7d.known" class="account-profit-view__quota" :title="quotaTitle(account)">
                      <div class="account-profit-view__quota-values">
                        <strong>{{ account.quota_7d.used_percent.toFixed(1) }}%</strong>
                        <small>{{ account.quota_7d.remaining_percent.toFixed(1) }}% {{ t('admin.accounts.profit.quota7d') }}</small>
                      </div>
                      <span class="account-profit-view__quota-track" aria-hidden="true">
                        <span :style="{ width: `${Math.min(100, Math.max(0, account.quota_7d.used_percent))}%` }" />
                      </span>
                    </div>
                    <span v-else class="is-muted">{{ t('admin.accounts.profit.quotaUnavailable') }}</span>
                  </td>
                  <td v-for="column in visibleMetricColumns" :key="column.key" :class="['account-profit-view__metric-cell', metricClass(column), `is-group-${column.group}`]">
                    {{ formatMetric(account, column) }}
                  </td>
                  <td>
                    <button type="button" class="account-profit-view__detail" @click="openProfit(account)">
                      <Icon name="chart" size="xs" />
                      {{ t('admin.accounts.viewProfit') }}
                    </button>
                  </td>
                  </tr>
                </template>
              </tbody>
            </table>
          </div>
        </template>

        <template #pagination>
          <Pagination v-if="total > 0" :page="page" :total="total" :page-size="pageSize" @update:page="page = $event" @update:pageSize="pageSize = $event" />
        </template>
      </TablePageLayout>
    </div>
    <AccountProfitModal :show="showProfit" :account="selectedAccount" @close="showProfit = false; selectedAccount = null" />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'
import Select from '@/components/common/Select.vue'
import SearchInput from '@/components/common/SearchInput.vue'
import AccountProfitModal from '@/components/admin/account/AccountProfitModal.vue'
import { adminAPI } from '@/api/admin'
import type { AccountSubscriptionTierOption } from '@/api/admin/accounts'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatDateTime, formatNumber, formatPoints, formatUSD } from '@/utils/format'
import { accountPlatformOptions } from '@/utils/accountPlatforms'
import type { AccountPlatform, AccountProfitListItem, SelectOption } from '@/types'

type MetricGroup = 'period' | 'expiry' | 'lifetime'
type MetricKey =
  | 'period_7d_revenue'
  | 'period_7d_cost'
  | 'period_7d_profit'
  | 'period_7d_tokens'
  | 'expiry_30d_revenue'
  | 'expiry_30d_cost'
  | 'expiry_30d_profit'
  | 'expiry_30d_tokens'
  | 'lifetime_revenue'
  | 'lifetime_cost'
  | 'lifetime_profit'
  | 'lifetime_tokens'
  | 'quota_7d'

interface MetricColumn {
  key: MetricKey
  group: MetricGroup
  label: string
  sortBy: string
  value: 'revenue_points' | 'cost_usd' | 'profit_points' | 'tokens'
}

const metricGroups: MetricGroup[] = ['period', 'expiry', 'lifetime']

const { t } = useI18n()
const accounts = ref<AccountProfitListItem[]>([])
const tiers = ref<AccountSubscriptionTierOption[]>([])
const loadingTiers = ref(false)
const loading = ref(false)
const error = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)
const sortBy = ref('created_at')
const sortOrder = ref<'asc' | 'desc'>('desc')
const showColumnsMenu = ref(false)
const selectedAccount = ref<AccountProfitListItem | null>(null)
const showProfit = ref(false)

const filters = reactive<Record<string, string>>({
  search: '',
  platform: '',
  status: '',
  expiry_status: '',
  subscription_tier: ''
})

const platformOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('admin.accounts.allPlatforms') },
  ...accountPlatformOptions(t)
])
const tierOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('admin.accounts.allSubscriptionTiers') },
  ...tiers.value.map(tier => ({
    value: tier.value,
    label: `${tier.value === '__unrecognized__' ? t('admin.accounts.subscriptionTierUnrecognized') : tier.label} (${tier.account_count})`
  }))
])
const statusOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('admin.accounts.allStatus') },
  { value: 'active', label: t('admin.accounts.status.active') },
  { value: 'inactive', label: t('admin.accounts.status.inactive') },
  { value: 'error', label: t('admin.accounts.status.error') },
  { value: 'rate_limited', label: t('admin.accounts.status.rateLimited') },
  { value: 'temp_unschedulable', label: t('admin.accounts.status.tempUnschedulable') },
  { value: 'unschedulable', label: t('admin.accounts.status.unschedulable') }
])
const expiryOptions = computed<SelectOption[]>(() => [
  { value: '', label: t('admin.accounts.allExpiryStatuses') },
  { value: 'expiring', label: t('admin.accounts.expiringWithin7Days') },
  { value: 'expired', label: t('admin.accounts.expiredAccounts') }
])

const metricColumns = computed<MetricColumn[]>(() => [
  { key: 'period_7d_revenue', group: 'period', label: t('admin.accounts.profit.revenuePoints'), sortBy: 'period_7d_revenue', value: 'revenue_points' },
  { key: 'period_7d_cost', group: 'period', label: t('admin.accounts.profit.upstreamCost'), sortBy: 'period_7d_cost', value: 'cost_usd' },
  { key: 'period_7d_profit', group: 'period', label: t('admin.accounts.profit.profit'), sortBy: 'period_7d_profit', value: 'profit_points' },
  { key: 'period_7d_tokens', group: 'period', label: t('admin.accounts.profit.tokens'), sortBy: 'period_7d_tokens', value: 'tokens' },
  { key: 'expiry_30d_revenue', group: 'expiry', label: t('admin.accounts.profit.revenuePoints'), sortBy: 'expiry_30d_revenue', value: 'revenue_points' },
  { key: 'expiry_30d_cost', group: 'expiry', label: t('admin.accounts.profit.upstreamCost'), sortBy: 'expiry_30d_cost', value: 'cost_usd' },
  { key: 'expiry_30d_profit', group: 'expiry', label: t('admin.accounts.profit.profit'), sortBy: 'expiry_30d_profit', value: 'profit_points' },
  { key: 'expiry_30d_tokens', group: 'expiry', label: t('admin.accounts.profit.tokens'), sortBy: 'expiry_30d_tokens', value: 'tokens' },
  { key: 'lifetime_revenue', group: 'lifetime', label: t('admin.accounts.profit.revenuePoints'), sortBy: 'lifetime_revenue', value: 'revenue_points' },
  { key: 'lifetime_cost', group: 'lifetime', label: t('admin.accounts.profit.upstreamCost'), sortBy: 'lifetime_cost', value: 'cost_usd' },
  { key: 'lifetime_profit', group: 'lifetime', label: t('admin.accounts.profit.profit'), sortBy: 'lifetime_profit', value: 'profit_points' },
  { key: 'lifetime_tokens', group: 'lifetime', label: t('admin.accounts.profit.tokens'), sortBy: 'lifetime_tokens', value: 'tokens' }
])

const columnVisibility = reactive<Record<MetricKey, boolean>>({
  quota_7d: true,
  period_7d_revenue: true,
  period_7d_cost: true,
  period_7d_profit: false,
  period_7d_tokens: true,
  expiry_30d_revenue: true,
  expiry_30d_cost: true,
  expiry_30d_profit: false,
  expiry_30d_tokens: true,
  lifetime_revenue: true,
  lifetime_cost: true,
  lifetime_profit: false,
  lifetime_tokens: true
})

const visibleMetricColumns = computed(() => metricColumns.value.filter(column => columnVisibility[column.key]))
const tableColumnCount = computed(() => 4 + (columnVisibility.quota_7d ? 1 : 0) + visibleMetricColumns.value.length + 1)

function hasVisible(key: MetricKey): boolean {
  return columnVisibility[key]
}

function visibleGroupCount(group: MetricGroup): number {
  return visibleMetricColumns.value.filter(column => column.group === group).length
}

function updateFilter(key: string, value: string | number | boolean | null) {
  filters[key] = value === null ? '' : String(value)
}

function updatePlatform(value: string | number | boolean | null) {
  filters.platform = value === null ? '' : String(value)
  filters.subscription_tier = ''
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const result = await adminAPI.accounts.listProfit(page.value, pageSize.value, {
      search: filters.search || undefined,
      platform: filters.platform || undefined,
      status: filters.status || undefined,
      expiry_status: filters.expiry_status || undefined,
      subscription_tier: filters.subscription_tier || undefined,
      sort_by: sortBy.value,
      sort_order: sortOrder.value
    })
    accounts.value = result.items
    total.value = result.total
  } catch (err) {
    error.value = extractApiErrorMessage(err, t('admin.accounts.profit.loadFailed'))
    accounts.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function reload() {
  page.value = 1
  void load()
}

function toggleSort(field: string) {
  if (sortBy.value === field) {
    sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortBy.value = field
    sortOrder.value = 'desc'
  }
  page.value = 1
}

function openProfit(account: AccountProfitListItem) {
  selectedAccount.value = account
  showProfit.value = true
}

function platformLabel(value: AccountPlatform) {
  return t(`admin.accounts.platforms.${value}`)
}

function formatExpiry(timestamp?: number | null) {
  return timestamp ? formatDateTime(new Date(timestamp * 1000)) : '-'
}

function quotaTitle(account: AccountProfitListItem) {
  const quota = account.quota_7d
  const reset = quota.reset_at ? formatDateTime(quota.reset_at) : '-'
  const observed = quota.observed_at ? formatDateTime(quota.observed_at) : '-'
  return `${t('admin.accounts.profit.quotaResetAt')}: ${reset}\n${t('admin.accounts.profit.quotaObservedAt')}: ${observed}`
}

function metricPeriod(account: AccountProfitListItem, column: MetricColumn) {
  if (column.group === 'period') return account.period_7d
  if (column.group === 'expiry') return account.expiry_30d
  return account.lifetime
}

function formatMetric(account: AccountProfitListItem, column: MetricColumn) {
  const period = metricPeriod(account, column)
  if (!period) return '-'
  const value = period[column.value]
  if (column.value === 'cost_usd') return formatUSD(value)
  if (column.value === 'tokens') return formatNumber(value)
  return formatPoints(value)
}

function metricClass(column: MetricColumn) {
  if (column.value === 'revenue_points') return 'is-revenue'
  if (column.value === 'cost_usd') return 'is-cost'
  if (column.value === 'profit_points') return 'is-profit'
  return 'is-tokens'
}

function metricGroupLabel(group: MetricGroup) {
  if (group === 'period') return t('admin.accounts.profit.period7d')
  if (group === 'expiry') return t('admin.accounts.profit.expiry30d')
  return t('admin.accounts.profit.lifetime')
}

watch([page, pageSize, sortBy, sortOrder], () => void load())

watch(() => filters.platform, async (platform) => {
  loadingTiers.value = true
  try {
    tiers.value = await adminAPI.accounts.listSubscriptionTiers(platform)
  } catch {
    tiers.value = []
  } finally {
    loadingTiers.value = false
  }
}, { immediate: true })

onMounted(() => void load())
</script>

<style scoped>
.account-profit-view {
  --account-col-width: 13.25rem;
  --platform-col-width: 6.5rem;
  --tier-col-width: 7.75rem;
  --table-head-row-height: 2.25rem;
  min-height: 100%;
}

.account-profit-view :deep(.table-page-layout) {
  min-height: 100%;
  gap: 0.75rem;
}

.account-profit-view__toolbar,
.account-profit-view__filters {
  border: 1px solid var(--glass-border);
  background: var(--glass-layer-inset-bg);
  box-shadow: 0 1px 0 var(--glass-highlight) inset;
  -webkit-backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
}

.account-profit-view__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  min-height: 3.25rem;
  padding: 0.5rem 0.75rem;
  border-radius: var(--radius-lg);
}

.account-profit-view__title {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 0.625rem;
}

.account-profit-view__title-icon {
  display: grid;
  width: 2.125rem;
  height: 2.125rem;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid color-mix(in srgb, var(--theme-accent) 28%, transparent);
  border-radius: var(--radius-md);
  background: var(--glass-tint-brand);
  color: var(--color-text-brand);
}

.account-profit-view__toolbar h1 {
  margin: 0;
  color: var(--color-text-primary);
  font-size: var(--type-section-title-size);
  font-weight: 650;
  line-height: 1.2;
}

.account-profit-view__toolbar-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex: 0 0 auto;
}

.account-profit-view__filters {
  display: grid;
  grid-template-columns: minmax(13rem, 1.6fr) repeat(4, minmax(8.5rem, 1fr));
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 0.75rem;
  border-radius: var(--radius-lg);
}

.account-profit-view__search {
  min-width: 0;
}

.account-profit-view__search :deep(.search-input__field),
.account-profit-view__filters :deep(.select-trigger) {
  min-height: 2.25rem;
  font-size: var(--type-control-size);
}

.account-profit-view__search :deep(.search-input__field) {
  padding-top: 0.5rem;
  padding-bottom: 0.5rem;
}

.account-profit-view__columns-menu-wrap {
  position: relative;
}

.account-profit-view__icon-button,
.account-profit-view__refresh,
.account-profit-view__detail {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.375rem;
  min-height: 2.25rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-md);
  background: var(--glass-bg-interactive);
  color: var(--color-text-secondary);
  font-size: var(--type-control-size);
  font-weight: 600;
  cursor: pointer;
  transition: color 160ms ease, border-color 160ms ease, background-color 160ms ease, box-shadow 160ms ease;
}

.account-profit-view__icon-button {
  width: 2.25rem;
  padding: 0;
}

.account-profit-view__refresh,
.account-profit-view__detail {
  padding: 0 0.75rem;
}

.account-profit-view__icon-button:hover,
.account-profit-view__refresh:hover:not(:disabled),
.account-profit-view__detail:hover {
  border-color: var(--glass-border-hover);
  background: var(--glass-bg-interactive-hover);
  color: var(--color-text-primary);
  box-shadow: 0 1px 0 var(--glass-highlight-hover) inset;
}

.account-profit-view__icon-button:focus-visible,
.account-profit-view__refresh:focus-visible,
.account-profit-view__detail:focus-visible,
.account-profit-view__sort-button:focus-visible {
  outline: none;
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--theme-accent) 20%, transparent);
}

.account-profit-view__refresh:disabled {
  cursor: wait;
  opacity: 0.6;
}

.account-profit-view__columns-menu {
  position: absolute;
  z-index: 10;
  top: calc(100% + 0.375rem);
  right: 0;
  display: grid;
  min-width: 13rem;
  max-height: min(26rem, 70vh);
  gap: 0.125rem;
  overflow-y: auto;
  padding: 0.5rem;
  border: 1px solid var(--glass-border-hover);
  border-radius: var(--radius-lg);
  background: var(--glass-layer-floating-bg);
  box-shadow: var(--glass-shadow-hover), 0 1px 0 var(--glass-highlight) inset;
  -webkit-backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-layer-floating-blur)) saturate(var(--glass-saturate));
}

.account-profit-view__column-group + .account-profit-view__column-group {
  margin-top: 0.375rem;
  padding-top: 0.375rem;
  border-top: 1px solid var(--color-border-subtle);
}

.account-profit-view__column-group-title {
  display: block;
  padding: 0.25rem 0.5rem 0.125rem;
  color: var(--color-text-quaternary);
  font-size: 0.625rem;
  font-weight: 700;
  letter-spacing: 0.035em;
  text-transform: uppercase;
}

.account-profit-view__column-option {
  display: flex;
  min-height: 2rem;
  align-items: center;
  gap: 0.5rem;
  padding: 0.25rem 0.5rem;
  border-radius: var(--radius-sm);
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
  cursor: pointer;
}

.account-profit-view__column-option:hover {
  background: var(--glass-bg-interactive-hover);
  color: var(--color-text-primary);
}

.account-profit-view__column-option input {
  accent-color: var(--theme-accent);
}

.account-profit-view__table-wrap {
  max-width: 100%;
  overflow: auto;
  overscroll-behavior: contain;
  scrollbar-color: color-mix(in srgb, var(--color-text-tertiary) 38%, transparent) transparent;
  scrollbar-width: thin;
}

.account-profit-view__table {
  width: max-content;
  min-width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  color: var(--color-text-secondary);
  font-size: var(--font-size-xs);
  font-variant-numeric: tabular-nums;
}

.account-profit-view__table th,
.account-profit-view__table td {
  padding: 0.625rem 0.75rem;
  border-bottom: 1px solid var(--color-border-subtle);
  text-align: right;
  white-space: nowrap;
}

.account-profit-view__table thead th {
  position: sticky;
  top: 0;
  z-index: 5;
  height: var(--table-head-row-height);
  background: var(--glass-layer-floating-bg);
  color: var(--color-text-tertiary);
  font-size: var(--font-size-2xs);
  font-weight: 650;
  line-height: 1.15;
  -webkit-backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
}

.account-profit-view__table thead .account-profit-view__metric-row th {
  top: var(--table-head-row-height);
  border-bottom-color: var(--color-border);
  background: var(--color-surface-elevated);
}

.account-profit-view__group-row th {
  border-bottom-color: var(--color-border);
  text-align: center;
  text-transform: uppercase;
  letter-spacing: 0.035em;
}

.account-profit-view__group {
  box-shadow: inset 0 2px 0 var(--group-accent, var(--color-border-strong));
}

.account-profit-view__group--identity {
  --group-accent: var(--color-text-tertiary);
}

.account-profit-view__group--quota {
  --group-accent: var(--color-info);
  background: color-mix(in srgb, var(--color-info) 8%, var(--glass-layer-floating-bg)) !important;
}

.account-profit-view__group--period {
  --group-accent: var(--color-success);
  background: color-mix(in srgb, var(--color-success) 8%, var(--glass-layer-floating-bg)) !important;
}

.account-profit-view__group--expiry {
  --group-accent: var(--color-warning);
  background: color-mix(in srgb, var(--color-warning) 9%, var(--glass-layer-floating-bg)) !important;
}

.account-profit-view__group--lifetime {
  --group-accent: var(--theme-accent);
  background: color-mix(in srgb, var(--theme-accent) 8%, var(--glass-layer-floating-bg)) !important;
}

.account-profit-view__sort-button {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  gap: 0.25rem;
  min-height: 1.5rem;
  padding: 0;
  border: 0;
  border-radius: var(--radius-xs);
  background: none;
  color: inherit;
  font: inherit;
  cursor: pointer;
}

.account-profit-view__sort-button:hover {
  color: var(--color-text-primary);
}

.account-profit-view__table td:first-child,
.account-profit-view__table th:first-child {
  text-align: left;
}

.account-profit-view__table tbody tr {
  transition: background-color 140ms ease;
}

.account-profit-view__table tbody tr:hover td {
  background: color-mix(in srgb, var(--theme-accent) 5%, transparent);
}

.account-profit-view__table .is-sticky {
  position: sticky;
  z-index: 2;
  background: var(--color-surface);
}

.account-profit-view__table thead th.is-sticky {
  z-index: 8;
  background: var(--color-surface-elevated);
}

.account-profit-view__table tbody tr:hover .is-sticky {
  background: color-mix(in srgb, var(--theme-accent) 5%, var(--color-surface));
}

.account-profit-view__table .is-account-column {
  left: 0;
  width: var(--account-col-width);
  min-width: var(--account-col-width);
  text-align: left;
}

.account-profit-view__table .is-platform-column {
  left: var(--account-col-width);
  width: var(--platform-col-width);
  min-width: var(--platform-col-width);
  text-align: left;
}

.account-profit-view__table .is-tier-column {
  left: calc(var(--account-col-width) + var(--platform-col-width));
  width: var(--tier-col-width);
  min-width: var(--tier-col-width);
  text-align: left;
  box-shadow: 8px 0 14px -14px var(--color-text-primary);
}

.account-profit-view__table .is-expiry-column {
  min-width: 8.75rem;
}

.account-profit-view__table .is-quota-column {
  min-width: 9.5rem;
}

.account-profit-view__table .is-group-expiry,
.account-profit-view__table .is-group-lifetime {
  border-left: 1px solid var(--color-border);
}

.account-profit-view__account-name,
.account-profit-view__platform,
.account-profit-view__tier,
.account-profit-view__expiry {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
}

.account-profit-view__account-name {
  color: var(--color-text-primary);
  font-weight: 650;
}

.account-profit-view__account-id {
  display: block;
  margin-top: 0.125rem;
  overflow: hidden;
  color: var(--color-text-quaternary);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 0.625rem;
  text-overflow: ellipsis;
}

.account-profit-view__platform {
  color: var(--color-text-secondary);
  font-weight: 600;
}

.account-profit-view__tier {
  max-width: calc(var(--tier-col-width) - 1.5rem);
  color: var(--color-text-tertiary);
}

.account-profit-view__expiry {
  color: var(--color-text-tertiary);
}

.account-profit-view__quota {
  display: grid;
  min-width: 7.5rem;
  gap: 0.3rem;
  text-align: left;
}

.account-profit-view__quota-values {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.5rem;
}

.account-profit-view__quota-values strong {
  color: var(--color-text-info);
  font-size: var(--font-size-xs);
  font-weight: 700;
}

.account-profit-view__quota-values small {
  color: var(--color-text-quaternary);
  font-size: 0.625rem;
}

.account-profit-view__quota-track {
  display: block;
  height: 0.3rem;
  overflow: hidden;
  border-radius: var(--radius-full);
  background: color-mix(in srgb, var(--color-info) 12%, var(--color-border-subtle));
}

.account-profit-view__quota-track span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--color-success), var(--color-warning));
  transition: width 180ms ease;
}

.account-profit-view__state {
  padding: 3.5rem 1rem !important;
  text-align: center !important;
  color: var(--color-text-muted);
}

.account-profit-view__state.is-error {
  color: var(--color-text-danger);
}

.account-profit-view__detail {
  min-height: 1.875rem;
  border-color: color-mix(in srgb, var(--theme-accent) 28%, var(--glass-border));
  background: var(--glass-tint-brand);
  color: var(--color-text-brand);
  font-size: var(--font-size-2xs);
}

.is-revenue { color: var(--color-text-success); }
.is-cost { color: var(--color-text-warning); }
.is-profit { color: var(--color-text-brand); }
.is-tokens { color: #7c3aed; }
.is-muted { color: var(--color-text-muted); }

@media (max-width: 1200px) {
  .account-profit-view__filters {
    grid-template-columns: minmax(13rem, 1.4fr) repeat(2, minmax(9rem, 1fr));
  }

  .account-profit-view__search {
    grid-column: span 2;
  }
}

@media (max-width: 900px) {
  .account-profit-view {
    --account-col-width: 11rem;
    --platform-col-width: 6.25rem;
    --tier-col-width: 7.25rem;
  }

  .account-profit-view__toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .account-profit-view__toolbar-actions {
    justify-content: flex-end;
  }

  .account-profit-view__filters {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .account-profit-view__search {
    grid-column: 1 / -1;
  }
}

@media (max-width: 560px) {
  .account-profit-view__filters {
    grid-template-columns: 1fr;
  }

  .account-profit-view__search {
    grid-column: auto;
  }

  .account-profit-view__toolbar,
  .account-profit-view__filters {
    padding-inline: 0.625rem;
  }

  .account-profit-view__title-icon {
    width: 2rem;
    height: 2rem;
  }

  .account-profit-view__toolbar h1 {
    font-size: var(--font-size-lg);
  }
}
</style>
