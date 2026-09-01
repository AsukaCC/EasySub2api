<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="views-admin-affiliates-admin-affiliate-records-table__panel">
          <div class="views-admin-affiliates-admin-affiliate-records-table__panel-2">
            <Icon name="search" size="md" class="views-admin-affiliates-admin-affiliate-records-table__icon" />
            <input v-model="filters.search" type="text" class="views-admin-affiliates-admin-affiliate-records-table__field input" :placeholder="t('admin.affiliates.records.searchPlaceholder')" @input="debounceLoad" />
          </div>
          <input v-model="filters.start_at" type="date" class="views-admin-affiliates-admin-affiliate-records-table__field-2 input" :title="t('admin.affiliates.records.startAt')" @change="reloadFromFirstPage" />
          <input v-model="filters.end_at" type="date" class="views-admin-affiliates-admin-affiliate-records-table__field-2 input" :title="t('admin.affiliates.records.endAt')" @change="reloadFromFirstPage" />
          <button class="views-admin-affiliates-admin-affiliate-records-table__action btn btn-secondary" :disabled="loading" :title="t('common.refresh')" @click="loadRecords">
            <Icon name="refresh" size="md" :class="loading ? 'views-admin-affiliates-admin-affiliate-records-table__icon-2' : ''" />
          </button>
        </div>
      </template>

      <template #table>
        <DataTable
          :columns="columns"
          :data="records"
          :loading="loading"
          :server-side-sort="true"
          default-sort-key="created_at"
          default-sort-order="desc"
          :sort-storage-key="sortStorageKey"
          @sort="handleSort"
        >
          <template #cell-inviter="{ row }">
            <UserCell
              :id="row.inviter_id"
              :email="row.inviter_email"
              :username="row.inviter_username"
              :clickable="props.type !== 'transfers'"
              @open="openUserOverview"
            />
          </template>
          <template #cell-invitee="{ row }">
            <UserCell
              :id="row.invitee_id"
              :email="row.invitee_email"
              :username="row.invitee_username"
              :clickable="props.type !== 'transfers'"
              @open="openUserOverview"
            />
          </template>
          <template #cell-recipient="{ row }">
            <UserCell
              :id="row.recipient_id"
              :email="row.recipient_email"
              :username="row.recipient_username"
              :clickable="true"
              @open="openUserOverview"
            />
          </template>
          <template #cell-user="{ row }">
            <UserCell
              :id="row.user_id"
              :email="row.user_email"
              :username="row.username"
              :clickable="true"
              @open="openUserOverview"
            />
          </template>
          <template #cell-aff_code="{ row }">
            <span class="views-admin-affiliates-admin-affiliate-records-table__text">{{ row.aff_code || '-' }}</span>
          </template>
          <template #cell-order="{ row }">
            <div class="views-admin-affiliates-admin-affiliate-records-table__panel-3">
              <div class="views-admin-affiliates-admin-affiliate-records-table__panel-4">#{{ row.order_id }}</div>
              <div class="views-admin-affiliates-admin-affiliate-records-table__panel-5">{{ row.out_trade_no }}</div>
            </div>
          </template>
          <template #cell-payment_type="{ row }">
            {{ t('payment.methods.' + row.payment_type, row.payment_type || '-') }}
          </template>
          <template #cell-order_status="{ row }">
            <OrderStatusBadge :status="row.order_status" />
          </template>
          <template #cell-total_rebate="{ row }">
            <AmountText :value="row.total_rebate" />
          </template>
          <template #cell-order_amount="{ row }">
            <AmountText :value="row.order_amount" />
          </template>
          <template #cell-pay_amount="{ row }">
            <span class="views-admin-affiliates-admin-affiliate-records-table__text-2">{{ formatCNY(row.pay_amount) }}</span>
          </template>
          <template #cell-rebate_amount="{ row }">
            <AmountText :value="row.rebate_amount" strong />
          </template>
          <template #cell-reversed_points="{ row }">
            <div class="views-admin-affiliates-admin-affiliate-records-table__panel-3">
              <AmountText :value="row.reversed_points" />
              <span
                v-if="row.reserved_reversal_points > 0"
                class="views-admin-affiliates-admin-affiliate-records-table__text-3"
              >
                {{ t('admin.affiliates.records.reservedReversalPoints', { amount: formatPoints(row.reserved_reversal_points) }) }}
              </span>
            </div>
          </template>
          <template #cell-net_rebate_points="{ row }">
            <AmountText :value="row.net_rebate_points" strong />
          </template>
          <template #cell-amount="{ row }">
            <AmountText :value="row.amount" strong />
          </template>
          <template #cell-balance_after="{ row }">
            <NullableAmountText :value="row.balance_after" />
          </template>
          <template #cell-available_quota_after="{ row }">
            <NullableAmountText :value="row.available_quota_after" />
          </template>
          <template #cell-frozen_quota_after="{ row }">
            <NullableAmountText :value="row.frozen_quota_after" />
          </template>
          <template #cell-history_quota_after="{ row }">
            <NullableAmountText :value="row.history_quota_after" />
          </template>
          <template #cell-created_at="{ row }">
            <span class="views-admin-affiliates-admin-affiliate-records-table__text-3">{{ formatDateTime(row.created_at) }}</span>
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <BaseDialog
      :show="overviewDialog"
      :title="t('admin.affiliates.overview.title')"
      width="normal"
      @close="overviewDialog = false"
    >
      <div v-if="overviewLoading" class="views-admin-affiliates-admin-affiliate-records-table__panel-6">
        <div class="views-admin-affiliates-admin-affiliate-records-table__panel-7"></div>
      </div>
      <div v-else-if="selectedOverview" class="views-admin-affiliates-admin-affiliate-records-table__panel-8">
        <div class="views-admin-affiliates-admin-affiliate-records-table__panel-9">
          <div class="views-admin-affiliates-admin-affiliate-records-table__panel-4">#{{ selectedOverview.user_id }}</div>
          <div class="views-admin-affiliates-admin-affiliate-records-table__panel-10">{{ selectedOverview.email || '-' }}</div>
          <div class="views-admin-affiliates-admin-affiliate-records-table__panel-11">{{ selectedOverview.username || '-' }}</div>
        </div>
        <div class="views-admin-affiliates-admin-affiliate-records-table__panel-12">
          <OverviewStat :label="t('admin.affiliates.overview.affCode')" :value="selectedOverview.aff_code || '-'" mono />
          <OverviewStat :label="t('admin.affiliates.overview.rebateRate')" :value="formatPercent(selectedOverview.rebate_rate_percent)" />
          <OverviewStat :label="t('admin.affiliates.overview.invitedCount')" :value="String(selectedOverview.invited_count)" />
          <OverviewStat :label="t('admin.affiliates.overview.rebatedInviteeCount')" :value="String(selectedOverview.rebated_invitee_count)" />
          <OverviewStat :label="t('admin.affiliates.overview.availableQuota')" :value="formatPoints(selectedOverview.available_quota)" />
          <OverviewStat :label="t('admin.affiliates.overview.historyQuota')" :value="formatPoints(selectedOverview.history_quota)" />
        </div>
      </div>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref, type PropType } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import OrderStatusBadge from '@/components/payment/OrderStatusBadge.vue'
import type { Column } from '@/components/common/types'
import { useAppStore } from '@/stores/app'
import { affiliatesAPI, type AffiliateInviteRecord, type AffiliateRebateRecord, type AffiliateTransferRecord, type AffiliateUserOverview, type ListAffiliateRecordsParams } from '@/api/admin/affiliates'
import type { PaginatedResponse } from '@/types'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatCNY, formatDateTime as formatDisplayDateTime, formatPoints } from '@/utils/format'

type RecordType = 'invites' | 'rebates' | 'transfers'
type AffiliateRecord = AffiliateInviteRecord | AffiliateRebateRecord | AffiliateTransferRecord

const props = defineProps<{
  type: RecordType
}>()

const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(true)
const records = ref<AffiliateRecord[]>([])
const filters = reactive({ search: '', start_at: '', end_at: '' })
const pagination = reactive({ page: 1, page_size: 20, total: 0 })
const overviewDialog = ref(false)
const overviewLoading = ref(false)
const selectedOverview = ref<AffiliateUserOverview | null>(null)
let debounceTimer: ReturnType<typeof setTimeout> | null = null

const columns = computed<Column[]>(() => {
  if (props.type === 'invites') {
    return [
      { key: 'inviter', label: t('admin.affiliates.records.inviter'), sortable: true },
      { key: 'invitee', label: t('admin.affiliates.records.invitee'), sortable: true },
      { key: 'aff_code', label: t('admin.affiliates.records.affCode'), sortable: true },
      { key: 'total_rebate', label: t('admin.affiliates.records.totalRebate'), sortable: true },
      { key: 'created_at', label: t('admin.affiliates.records.invitedAt'), sortable: true },
    ]
  }
  if (props.type === 'rebates') {
    return [
      { key: 'order', label: t('admin.affiliates.records.order'), sortable: true },
      { key: 'inviter', label: t('admin.affiliates.records.inviter'), sortable: true },
      { key: 'invitee', label: t('admin.affiliates.records.invitee'), sortable: true },
      { key: 'recipient', label: t('admin.affiliates.records.recipient'), sortable: true },
      { key: 'order_amount', label: t('admin.affiliates.records.orderAmount'), sortable: true },
      { key: 'pay_amount', label: t('admin.affiliates.records.payAmount'), sortable: true },
      { key: 'rebate_amount', label: t('admin.affiliates.records.rebateAmount') },
      { key: 'reversed_points', label: t('admin.affiliates.records.reversedPoints') },
      { key: 'net_rebate_points', label: t('admin.affiliates.records.netRebatePoints') },
      { key: 'payment_type', label: t('admin.affiliates.records.paymentType'), sortable: true },
      { key: 'order_status', label: t('admin.affiliates.records.orderStatus'), sortable: true },
      { key: 'created_at', label: t('admin.affiliates.records.rebatedAt'), sortable: true },
    ]
  }
  return [
    { key: 'user', label: t('admin.affiliates.records.user'), sortable: true },
    { key: 'amount', label: t('admin.affiliates.records.transferAmount'), sortable: true },
    { key: 'balance_after', label: t('admin.affiliates.records.balanceAfter'), sortable: true },
    { key: 'available_quota_after', label: t('admin.affiliates.records.availableQuotaAfter'), sortable: true },
    { key: 'frozen_quota_after', label: t('admin.affiliates.records.frozenQuotaAfter'), sortable: true },
    { key: 'history_quota_after', label: t('admin.affiliates.records.historyQuotaAfter'), sortable: true },
    { key: 'created_at', label: t('admin.affiliates.records.transferredAt'), sortable: true },
  ]
})

const sortStorageKey = computed(() => `admin-affiliate-${props.type}-table-sort`)

function loadInitialSortState(): { sort_by: string; sort_order: 'asc' | 'desc' } {
  const fallback = { sort_by: 'created_at', sort_order: 'desc' as 'asc' | 'desc' }
  try {
    const raw = localStorage.getItem(sortStorageKey.value)
    if (!raw) return fallback
    const parsed = JSON.parse(raw) as { key?: string; order?: string }
    const key = typeof parsed.key === 'string' ? parsed.key : ''
    if (!columns.value.some((column) => column.key === key && column.sortable)) return fallback
    return {
      sort_by: key,
      sort_order: parsed.order === 'asc' ? 'asc' : 'desc',
    }
  } catch {
    return fallback
  }
}

const sortState = reactive(loadInitialSortState())

function userTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone
  } catch {
    return 'UTC'
  }
}

function buildParams(): ListAffiliateRecordsParams {
  return {
    page: pagination.page,
    page_size: pagination.page_size,
    search: filters.search.trim() || undefined,
    start_at: filters.start_at || undefined,
    end_at: filters.end_at || undefined,
    sort_by: sortState.sort_by,
    sort_order: sortState.sort_order,
    timezone: userTimezone(),
  }
}

async function fetchRecords(params: ListAffiliateRecordsParams): Promise<PaginatedResponse<AffiliateRecord>> {
  if (props.type === 'invites') {
    return affiliatesAPI.listInviteRecords(params)
  }
  if (props.type === 'rebates') {
    return affiliatesAPI.listRebateRecords(params)
  }
  return affiliatesAPI.listTransferRecords(params)
}

async function loadRecords() {
  loading.value = true
  try {
    const res = await fetchRecords(buildParams())
    records.value = res.items || []
    pagination.total = res.total || 0
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.affiliates.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

function debounceLoad() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => reloadFromFirstPage(), 300)
}

function reloadFromFirstPage() {
  pagination.page = 1
  void loadRecords()
}

function handlePageChange(page: number) {
  pagination.page = page
  void loadRecords()
}

function handlePageSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  void loadRecords()
}

function handleSort(key: string, order: 'asc' | 'desc') {
  sortState.sort_by = key
  sortState.sort_order = order
  pagination.page = 1
  void loadRecords()
}

function formatPercent(value: number | null | undefined): string {
  const rounded = Math.round(Number(value || 0) * 100) / 100
  return `${Number.isInteger(rounded) ? rounded.toString() : rounded.toString()}%`
}

function formatDateTime(value: string | null | undefined): string {
  return value ? formatDisplayDateTime(value) : '-'
}

async function openUserOverview(userId: string) {
  if (!userId) return
  overviewDialog.value = true
  overviewLoading.value = true
  selectedOverview.value = null
  try {
    selectedOverview.value = await affiliatesAPI.getUserOverview(userId)
  } catch (error) {
    overviewDialog.value = false
    appStore.showError(extractI18nErrorMessage(error, t, 'admin.affiliates.errors', t('common.error')))
  } finally {
    overviewLoading.value = false
  }
}

const UserCell = defineComponent({
  props: {
    id: { type: String, required: true },
    email: { type: String, default: '' },
    username: { type: String, default: '' },
    clickable: { type: Boolean, default: false },
  },
  emits: ['open'],
  setup(cellProps, { emit }) {
    return () => h('div', { class: 'views-admin-affiliates-admin-affiliate-records-table__panel-3' }, [
      h('div', { class: 'views-admin-affiliates-admin-affiliate-records-table__panel-4' }, `#${cellProps.id}`),
      h(cellProps.clickable ? 'button' : 'div', {
        class: cellProps.clickable
          ? 'views-admin-affiliates-admin-affiliate-records-table__state'
          : 'views-admin-affiliates-admin-affiliate-records-table__state-2',
        type: cellProps.clickable ? 'button' : undefined,
        onClick: cellProps.clickable ? () => emit('open', cellProps.id) : undefined,
      }, cellProps.email || '-'),
      h('div', { class: 'views-admin-affiliates-admin-affiliate-records-table__panel-5' }, cellProps.username || '-'),
    ])
  },
})

const AmountText = defineComponent({
  props: {
    value: { type: Number, default: 0 },
    strong: { type: Boolean, default: false },
  },
  setup(amountProps) {
    return () => h('span', {
      class: amountProps.strong
        ? 'views-admin-affiliates-admin-affiliate-records-table__state-3'
        : 'views-admin-affiliates-admin-affiliate-records-table__text-2',
    }, formatPoints(amountProps.value))
  },
})

const NullableAmountText = defineComponent({
  props: {
    value: { type: Number as PropType<number | null | undefined>, default: null },
  },
  setup(amountProps) {
    return () => {
      const value = amountProps.value
      if (value === null || value === undefined) {
        return h('span', { class: 'views-admin-affiliates-admin-affiliate-records-table__render' }, '-')
      }
      return h(AmountText, { value })
    }
  },
})

const OverviewStat = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: String, required: true },
    mono: { type: Boolean, default: false },
  },
  setup(statProps) {
    return () => h('div', { class: 'views-admin-affiliates-admin-affiliate-records-table__render-2' }, [
      h('div', { class: 'views-admin-affiliates-admin-affiliate-records-table__render-3' }, statProps.label),
      h('div', {
        class: statProps.mono
          ? 'views-admin-affiliates-admin-affiliate-records-table__state-4'
          : 'views-admin-affiliates-admin-affiliate-records-table__state-5',
      }, statProps.value),
    ])
  },
})

onMounted(() => {
  void loadRecords()
})
</script>
