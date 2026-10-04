<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <form class="point-changes__filters" @submit.prevent="search">
          <label>
            <span class="input-label">{{ t(`${key}.keyword`) }}</span>
            <input v-model.trim="filters.keyword" class="input" type="search" :placeholder="t(`${key}.keyword`)" />
          </label>
          <label>
            <span class="input-label">{{ t('admin.users.balanceType') }}</span>
            <Select v-model="filters.point_type" :options="pointTypes" :aria-label="t('admin.users.balanceType')" />
          </label>
          <label>
            <span class="input-label">{{ t(`${key}.direction`) }}</span>
            <Select v-model="filters.direction" :options="directions" :aria-label="t(`${key}.direction`)" />
          </label>
          <div>
            <span class="input-label">{{ t(`${key}.date`) }}</span>
            <DateRangePicker v-model:start-date="startDate" v-model:end-date="endDate" />
          </div>
          <div class="point-changes__actions">
            <button class="btn btn-primary" :disabled="loading" type="submit"><Icon name="search" size="sm" />{{ t('common.search') }}</button>
            <button class="btn btn-secondary" type="button" :disabled="loading" @click="reset">{{ t('common.reset') }}</button>
          </div>
        </form>
        <div v-if="fixedUser" class="point-changes__user">
          <span>{{ t(`${key}.selectedUser`) }}: {{ userLabel }}</span>
          <button class="btn btn-ghost btn-sm" type="button" @click="router.replace({ path: route.path })">{{ t(`${key}.allUsers`) }}</button>
        </div>
        <p v-if="error" class="point-changes__error" role="alert">{{ error }}</p>
      </template>
      <template #table>
        <DataTable :columns="columns" :data="items" :loading="loading" row-key="id">
          <template #cell-created_at="{ value }">{{ formatDateTime(value) }}</template>
          <template #cell-user="{ row }"><div class="point-changes__identity"><span>{{ row.email }}</span><small>{{ row.username || '-' }}</small></div></template>
          <template #cell-point_type="{ value }">{{ t(value === 'bonus' ? 'admin.users.bonusBalance' : 'admin.users.rechargeBalance') }}</template>
          <template #cell-amount="{ row }"><span class="point-changes__number" :class="{ 'is-negative': row.amount < 0, 'is-positive': row.amount > 0 }">{{ signed(row.amount) }}</span></template>
          <template #cell-balance="{ row }"><span v-if="row.balance_before !== null && row.balance_after !== null" class="point-changes__number">{{ points(row.balance_before) }} → {{ points(row.balance_after) }}</span><span v-else :title="t(`${key}.snapshotUnavailable`)">-</span></template>
          <template #cell-reason="{ row }"><div class="point-changes__identity"><span>{{ reason(row) }}</span><small v-if="row.frozen_amount">{{ t(`${key}.frozen`) }} {{ signed(row.frozen_amount) }}</small></div></template>
          <template #cell-source="{ row }"><div class="point-changes__source"><span :title="row.source_id">{{ row.source_id || '-' }}</span><small>{{ row.notes }}</small></div></template>
        </DataTable>
      </template>
      <template #pagination>
        <Pagination :total="total" :page="page" :page-size="pageSize" :page-size-options="[20, 50, 100]" @update:page="changePage" @update:page-size="changeSize" />
      </template>
    </TablePageLayout>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Icon from '@/components/icons/Icon.vue'
import { listPointChanges, type PointChange, type PointChangeQuery } from '@/api/admin/pointChanges'
import { formatDateTime } from '@/utils/format'
import { parsePickerValue } from '@/utils/datetime'

const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const key = 'admin.users.pointChanges'
const fixedUser = computed(() => typeof route.query.user_id === 'string' ? route.query.user_id : '')
const userLabel = computed(() => items.value[0]?.email || fixedUser.value)
const filters = reactive({ keyword: '', point_type: '', direction: '' })
const startDate = ref('')
const endDate = ref('')
const items = ref<PointChange[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const error = ref('')
let requestVersion = 0
let applied: Omit<PointChangeQuery, 'page' | 'page_size'> = {}
const pointTypes = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'recharge', label: t('admin.users.rechargeBalance') },
  { value: 'bonus', label: t('admin.users.bonusBalance') }
])
const directions = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'increase', label: t(`${key}.increase`) },
  { value: 'decrease', label: t(`${key}.decrease`) }
])
const columns = computed(() => [
  { key: 'created_at', label: t(`${key}.date`) },
  { key: 'user', label: t(`${key}.user`) },
  { key: 'point_type', label: t('admin.users.balanceType') },
  { key: 'amount', label: t(`${key}.amount`) },
  { key: 'balance', label: t(`${key}.balance`) },
  { key: 'reason', label: t(`${key}.reason`) },
  { key: 'source', label: t(`${key}.source`) }
])
const points = (value: number) => new Intl.NumberFormat(locale.value, { maximumFractionDigits: 8 }).format(value)
const signed = (value: number) => `${value > 0 ? '+' : ''}${points(value)}`
function reason(row: PointChange) {
  if (row.action === 'capture') return t('admin.users.walletActions.capture')
  if (row.source_type === 'payment_refund') return t(`${key}.${row.action === 'release' ? 'refundRestore' : 'refundRecovery'}`)
  if (row.source_type === 'affiliate_refund') return t(`${key}.${row.action === 'release' ? 'refundRestore' : 'affiliateRecovery'}`)
  if (row.source_type === 'admin_adjustment') return t('admin.users.walletActions.adjust')
  const sources: Record<string, string> = { payment_order: 'recharge', redeem_code: 'redeem', promo_code: 'promotion', affiliate_transfer: 'affiliate' }
  if (row.source_type === 'payment_order' && row.point_type === 'bonus' && row.action === 'credit') return t('admin.users.rechargeTierBonus')
  if ((row.action === 'credit' || row.action === 'bonus') && sources[row.source_type]) return t(`${key}.${sources[row.source_type]}`)
  return t(`admin.users.walletActions.${row.action}`)
}
async function load() {
  const version = ++requestVersion
  loading.value = true
  error.value = ''
  try {
    const data = await listPointChanges({ ...applied, user_id: fixedUser.value || undefined, page: page.value, page_size: pageSize.value })
    if (version !== requestVersion) return
    items.value = data.items
    total.value = data.total
  } catch {
    if (version !== requestVersion) return
    items.value = []
    total.value = 0
    error.value = t(`${key}.loadFailed`)
  } finally {
    if (version === requestVersion) loading.value = false
  }
}
function search() {
  const start = startDate.value ? parsePickerValue(startDate.value, 'date') : null
  const end = endDate.value ? parsePickerValue(endDate.value, 'date') : null
  if ((startDate.value && !start) || (endDate.value && !end) || (start && end && end.isBefore(start))) {
    error.value = t(`${key}.invalidDate`)
    return
  }
  applied = { ...filters, start_time: start?.startOf('day').toISOString(), end_time: end?.add(1, 'day').startOf('day').toISOString() }
  page.value = 1
  void load()
}
function reset() { Object.assign(filters, { keyword: '', point_type: '', direction: '' }); startDate.value = ''; endDate.value = ''; search() }
function changePage(value: number) { page.value = value; void load() }
function changeSize(value: number) { pageSize.value = Math.min(value, 100); page.value = 1; void load() }
watch(fixedUser, () => reset(), { immediate: true })
onBeforeUnmount(() => { requestVersion++ })
</script>

<style scoped>
.point-changes__filters { display: flex; flex-wrap: wrap; align-items: end; gap: 1rem; padding-bottom: 1rem; }
.point-changes__filters > label { flex: 1 1 180px; min-width: 0; }
.point-changes__actions, .point-changes__user { display: flex; align-items: center; gap: .5rem; }
.point-changes__user { flex-wrap: wrap; overflow-wrap: anywhere; padding-bottom: .75rem; }
.point-changes__identity, .point-changes__source { display: grid; gap: .25rem; }
.point-changes__identity { max-width: 280px; overflow-wrap: anywhere; }
.point-changes__identity small, .point-changes__source small { color: var(--color-text-secondary); }
.point-changes__source { max-width: 280px; overflow-wrap: anywhere; white-space: normal; }
.point-changes__number { font-variant-numeric: tabular-nums; white-space: nowrap; }
.point-changes__error, .is-negative { color: var(--color-danger); }
.is-positive { color: var(--color-success); }
@media (max-width: 640px) { .point-changes__filters > * { flex: 1 1 100%; min-width: 0; } }
</style>
