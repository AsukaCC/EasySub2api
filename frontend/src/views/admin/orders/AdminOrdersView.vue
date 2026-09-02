<template>
  <AppLayout>
    <div class="views-admin-orders-admin-orders-view__panel">
      <!-- Filters -->
      <div class="views-admin-orders-admin-orders-view__panel-2 card">
        <div class="views-admin-orders-admin-orders-view__panel-3">
          <div class="views-admin-orders-admin-orders-view__panel-4">
            <input v-model="orderSearch" type="text" :placeholder="t('payment.admin.searchOrders')" class="input" @input="debounceLoadOrders" />
          </div>
          <Select v-model="orderFilters.status" :options="statusFilterOptions" class="views-admin-orders-admin-orders-view__field" @change="loadOrders" />
          <Select v-model="orderFilters.payment_type" :options="paymentTypeFilterOptions" class="views-admin-orders-admin-orders-view__field-2" @change="loadOrders" />
          <Select v-model="orderFilters.order_type" :options="orderTypeFilterOptions" class="views-admin-orders-admin-orders-view__field" @change="loadOrders" />
          <div class="views-admin-orders-admin-orders-view__panel-5">
            <button @click="loadOrders" :disabled="ordersLoading" class="btn btn-secondary" :title="t('common.refresh')">
              <Icon name="refresh" size="md" :class="ordersLoading ? 'views-admin-orders-admin-orders-view__icon' : ''" />
            </button>
          </div>
        </div>
      </div>

      <!-- Table -->
      <OrderTable :orders="orders" :loading="ordersLoading" show-user>
        <template #actions="{ row }">
          <div class="views-admin-orders-admin-orders-view__panel-6">
            <button @click="showOrderDetail(row)" class="views-admin-orders-admin-orders-view__action">
              <Icon name="eye" size="sm" />
              {{ t('common.view') }}
            </button>
            <button v-if="row.status === 'PENDING'" @click="handleCancelOrder(row)" class="views-admin-orders-admin-orders-view__action-2">
              <Icon name="x" size="sm" />
              {{ t('payment.orders.cancel') }}
            </button>
            <button v-if="row.status === 'FAILED'" @click="handleRetryOrder(row)" class="views-admin-orders-admin-orders-view__action-3">
              <Icon name="refresh" size="sm" />
              {{ t('payment.admin.retry') }}
            </button>
            <button v-if="row.status === 'CANCELLED'" class="btn btn-danger btn-xs" @click="deleteTarget = row">
              <Icon name="trash" size="sm" />
              {{ t('common.delete') }}
            </button>
            <template v-if="row.status === 'REFUND_REQUESTED' && canDirectRefund(row)">
              <span v-if="row.refund_amount" class="views-admin-orders-admin-orders-view__text">{{ formatCNY(row.refund_amount) }}</span>
              <button @click="openRefundDialog(row)" class="views-admin-orders-admin-orders-view__action-4">
                <Icon name="check" size="sm" />
                {{ t('payment.admin.approveRefund') }}
              </button>
            </template>
            <button v-else-if="row.status === 'REFUND_FAILED' && canDirectRefund(row)" @click="openRefundDialog(row)" class="views-admin-orders-admin-orders-view__action-4">
              <Icon name="refresh" size="sm" />
              {{ t('payment.admin.retryRefund') }}
            </button>
            <button v-else-if="row.status === 'REFUND_PENDING'" :disabled="refundQueryingIds.has(row.id)" @click="handleQueryRefund(row)" class="views-admin-orders-admin-orders-view__action-5">
              <Icon name="refresh" size="sm" :class="refundQueryingIds.has(row.id) ? 'views-admin-orders-admin-orders-view__icon' : ''" />
              {{ t('payment.admin.queryRefundStatus') }}
            </button>
            <button v-else-if="canDirectRefund(row)" @click="openRefundDialog(row)" class="views-admin-orders-admin-orders-view__action-6">
              <Icon name="dollar" size="sm" />
              {{ t('payment.admin.refund') }}
            </button>
          </div>
        </template>
      </OrderTable>
      <Pagination v-if="orderPagination.total > 0" :page="orderPagination.page" :total="orderPagination.total" :page-size="orderPagination.page_size" @update:page="handleOrderPageChange" @update:pageSize="handleOrderPageSizeChange" />
    </div>

    <!-- Order Detail Dialog -->
    <BaseDialog :show="showDetailDialog" :title="t('payment.admin.orderDetail')" width="wide" @close="showDetailDialog = false">
      <div v-if="selectedOrder" class="views-admin-orders-admin-orders-view__panel">
        <div class="views-admin-orders-admin-orders-view__panel-7">
          <div><p class="views-admin-orders-admin-orders-view__description">{{ t('payment.orders.orderId') }}</p><p class="views-admin-orders-admin-orders-view__description-2">#{{ selectedOrder.id }}</p></div>
          <div><p class="views-admin-orders-admin-orders-view__description">{{ t('payment.orders.orderNo') }}</p><p class="views-admin-orders-admin-orders-view__description-3">{{ selectedOrder.out_trade_no }}</p></div>
          <div><p class="views-admin-orders-admin-orders-view__description">{{ t('payment.orders.status') }}</p><OrderStatusBadge :status="selectedOrder.status" /></div>
          <div><p class="views-admin-orders-admin-orders-view__description">{{ t(selectedOrder.order_type === 'balance' ? 'payment.orders.creditedPoints' : 'payment.orders.pointsPaid') }}</p><p class="views-admin-orders-admin-orders-view__description-3">{{ formatPoints(orderPoints(selectedOrder)) }}</p></div>
          <div v-if="selectedOrder.order_type === 'balance'"><p class="views-admin-orders-admin-orders-view__description">{{ t('payment.orders.payAmount') }}</p><p class="views-admin-orders-admin-orders-view__description-3">{{ formatCNY(selectedOrder.pay_amount) }}</p></div>
          <div><p class="views-admin-orders-admin-orders-view__description">{{ t('payment.orders.paymentMethod') }}</p><p class="views-admin-orders-admin-orders-view__description-4">{{ t('payment.methods.' + selectedOrder.payment_type, selectedOrder.payment_type) }}</p></div>
          <div><p class="views-admin-orders-admin-orders-view__description">{{ t('payment.admin.feeRate') }}</p><p class="views-admin-orders-admin-orders-view__description-4">{{ selectedOrder.fee_rate }}%</p></div>
          <div><p class="views-admin-orders-admin-orders-view__description">{{ t('payment.orders.createdAt') }}</p><p class="views-admin-orders-admin-orders-view__description-4">{{ formatDateTime(selectedOrder.created_at) }}</p></div>
          <div><p class="views-admin-orders-admin-orders-view__description">{{ t('payment.admin.expiresAt') }}</p><p class="views-admin-orders-admin-orders-view__description-4">{{ formatDateTime(selectedOrder.expires_at) }}</p></div>
          <div v-if="selectedOrder.paid_at"><p class="views-admin-orders-admin-orders-view__description">{{ t('payment.admin.paidAt') }}</p><p class="views-admin-orders-admin-orders-view__description-4">{{ formatDateTime(selectedOrder.paid_at) }}</p></div>
          <div v-if="selectedOrder.refund_amount"><p class="views-admin-orders-admin-orders-view__description">{{ t('payment.admin.refundAmount') }}</p><p class="views-admin-orders-admin-orders-view__description-5">{{ formatCNY(selectedOrder.refund_amount) }}</p></div>
          <div v-if="selectedOrder.refund_reason" class="views-admin-orders-admin-orders-view__panel-8"><p class="views-admin-orders-admin-orders-view__description">{{ t('payment.admin.refundReason') }}</p><p class="views-admin-orders-admin-orders-view__description-4">{{ selectedOrder.refund_reason }}</p></div>
          <!-- Refund request info -->
          <div v-if="selectedOrder.refund_requested_at" class="views-admin-orders-admin-orders-view__panel-9">
            <p class="views-admin-orders-admin-orders-view__description-6">{{ t('payment.admin.refundRequestInfo') }}</p>
            <div class="views-admin-orders-admin-orders-view__panel-7">
              <div>
                <p class="views-admin-orders-admin-orders-view__description">{{ t('payment.admin.refundRequestedAt') }}</p>
                <p class="views-admin-orders-admin-orders-view__description-4">{{ formatDateTime(selectedOrder.refund_requested_at) }}</p>
              </div>
              <div>
                <p class="views-admin-orders-admin-orders-view__description">{{ t('payment.admin.refundRequestedBy') }}</p>
                <p class="views-admin-orders-admin-orders-view__description-4">#{{ selectedOrder.refund_requested_by }}</p>
              </div>
              <div class="views-admin-orders-admin-orders-view__panel-8">
                <p class="views-admin-orders-admin-orders-view__description">{{ t('payment.admin.refundRequestReason') }}</p>
                <p class="views-admin-orders-admin-orders-view__description-4">{{ selectedOrder.refund_request_reason }}</p>
              </div>
            </div>
          </div>
        </div>
        <!-- Audit Logs -->
        <div v-if="orderAuditLogs.length > 0" class="views-admin-orders-admin-orders-view__panel-10">
          <p class="views-admin-orders-admin-orders-view__description-7">{{ t('payment.admin.auditLogs') }}</p>
          <div class="views-admin-orders-admin-orders-view__panel-11">
            <div v-for="log in orderAuditLogs" :key="log.id" class="views-admin-orders-admin-orders-view__panel-12">
              <div class="views-admin-orders-admin-orders-view__panel-13">
                <span class="views-admin-orders-admin-orders-view__text-2">{{ log.action }}</span>
                <span class="views-admin-orders-admin-orders-view__text-3">{{ formatDateTime(log.created_at) }}</span>
              </div>
              <div v-if="log.detail" class="views-admin-orders-admin-orders-view__panel-14">{{ log.detail }}</div>
              <div v-if="log.operator" class="views-admin-orders-admin-orders-view__panel-15">{{ t('payment.admin.operator') }}: {{ log.operator }}</div>
            </div>
          </div>
        </div>
      </div>
    </BaseDialog>

    <AdminRefundDialog :show="showRefundDialog" :order="selectedOrder" :submitting="refundSubmitting" @confirm="handleRefund" @cancel="closeRefundDialog" />

    <BaseDialog :show="!!deleteTarget" :title="t('payment.deleteOrder')" width="narrow" @close="deleteTarget = null">
      <p>{{ t('payment.confirmDeleteOrder') }}</p>
      <template #footer>
        <div class="order-delete-dialog__actions">
          <button class="btn btn-secondary" :disabled="deleteSubmitting" @click="deleteTarget = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-danger" :disabled="deleteSubmitting" @click="confirmDeleteOrder">
            {{ deleteSubmitting ? t('common.processing') : t('common.delete') }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { canDirectRefund, formatOrderDateTime } from '@/components/payment/orderUtils'
import type { PaymentOrder } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import AdminRefundDialog from '@/components/admin/payment/AdminRefundDialog.vue'
import { adminSupportTicketsAPI } from '@/api/supportTickets'
import OrderStatusBadge from '@/components/payment/OrderStatusBadge.vue'
import OrderTable from '@/components/payment/OrderTable.vue'
import { formatCNY, formatPoints } from '@/utils/format'

interface AuditLog {
  id: string
  action: string
  detail: string | null
  operator: string | null
  created_at: string
}

const { t } = useI18n()
const router = useRouter()
const appStore = useAppStore()

const ordersLoading = ref(true)
const orders = ref<PaymentOrder[]>([])
const orderSearch = ref('')
const orderFilters = reactive({ status: '', payment_type: '', order_type: '' })
const orderPagination = reactive({ page: 1, page_size: 20, total: 0 })
const selectedOrder = ref<PaymentOrder | null>(null)
const showDetailDialog = ref(false)
const showRefundDialog = ref(false)
const refundSubmitting = ref(false)
const refundQueryingIds = ref(new Set<string>())
const deleteTarget = ref<PaymentOrder | null>(null)
const deleteSubmitting = ref(false)
const orderAuditLogs = ref<AuditLog[]>([])
function orderPoints(order: PaymentOrder): number {
  if (order.order_type === 'balance') return order.credited_points ?? order.amount
  return order.wallet_amount > 0 ? order.wallet_amount : order.amount
}

let debounceTimer: ReturnType<typeof setTimeout> | null = null
function debounceLoadOrders() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => loadOrders(), 300)
}

async function loadOrders() {
  ordersLoading.value = true
  try {
    const res = await adminPaymentAPI.getOrders({
      page: orderPagination.page, page_size: orderPagination.page_size,
      keyword: orderSearch.value || undefined, status: orderFilters.status || undefined,
      payment_type: orderFilters.payment_type || undefined, order_type: orderFilters.order_type || undefined,
    })
    orders.value = res.data.items || []
    orderPagination.total = res.data.total || 0
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally { ordersLoading.value = false }
}

function handleOrderPageChange(page: number) { orderPagination.page = page; loadOrders() }
function handleOrderPageSizeChange(size: number) { orderPagination.page_size = size; orderPagination.page = 1; loadOrders() }

const statusFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allStatuses') },
  { value: 'PENDING', label: t('payment.status.pending') },
  { value: 'PAID', label: t('payment.status.paid') },
  { value: 'COMPLETED', label: t('payment.status.completed') },
  { value: 'EXPIRED', label: t('payment.status.expired') },
  { value: 'CANCELLED', label: t('payment.status.cancelled') },
  { value: 'FAILED', label: t('payment.status.failed') },
  { value: 'REFUNDED', label: t('payment.status.refunded') },
  { value: 'REFUND_REQUESTED', label: t('payment.status.refund_requested') },
  { value: 'REFUND_PENDING', label: t('payment.status.refund_pending') },
  { value: 'REFUND_FAILED', label: t('payment.status.refund_failed') },
])

const paymentTypeFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allPaymentTypes') },
  { value: 'alipay', label: t('payment.methods.alipay') },
  { value: 'wxpay', label: t('payment.methods.wxpay') },
  { value: 'stripe', label: t('payment.methods.stripe') },
  { value: 'airwallex', label: t('payment.methods.airwallex') },
])

const orderTypeFilterOptions = computed(() => [
  { value: '', label: t('payment.admin.allOrderTypes') },
  { value: 'balance', label: t('payment.admin.balanceOrder') },
  { value: 'subscription', label: t('payment.admin.subscriptionOrder') },
])

async function showOrderDetail(order: PaymentOrder) {
  selectedOrder.value = order
  orderAuditLogs.value = []
  showDetailDialog.value = true
  try {
    const res = await adminPaymentAPI.getOrder(order.id)
    const data = res.data as unknown as Record<string, unknown>
    if (data.order) selectedOrder.value = data.order as PaymentOrder
    orderAuditLogs.value = ((data.auditLogs || data.audit_logs || []) as unknown) as AuditLog[]
  } catch (_err: unknown) { /* keep cached order data */ }
}

async function handleCancelOrder(order: PaymentOrder) {
  try { await adminPaymentAPI.cancelOrder(order.id); appStore.showSuccess(t('payment.admin.orderCancelled')); loadOrders() }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
}

async function handleRetryOrder(order: PaymentOrder) {
  try { await adminPaymentAPI.retryRecharge(order.id); appStore.showSuccess(t('payment.admin.retrySuccess')); loadOrders() }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
}

async function confirmDeleteOrder() {
  if (!deleteTarget.value) return
  deleteSubmitting.value = true
  try {
    await adminPaymentAPI.deleteOrder(deleteTarget.value.id)
    appStore.showSuccess(t('payment.orderDeleted'))
    deleteTarget.value = null
    if (orders.value.length === 1 && orderPagination.page > 1) orderPagination.page -= 1
    await loadOrders()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    deleteSubmitting.value = false
  }
}

function openRefundDialog(order: PaymentOrder) {
  selectedOrder.value = order
  showRefundDialog.value = true
}

function closeRefundDialog() {
  showRefundDialog.value = false
}

async function handleRefund(data: { principal_amount: number; reason: string; idempotency_key: string }) {
  if (!selectedOrder.value) return
  refundSubmitting.value = true
  try {
    const res = await adminSupportTicketsAPI.createRefund({
      order_id: selectedOrder.value.id,
      approved_principal_amount: data.principal_amount,
      message: data.reason,
    })
    const result = res.data as { ticket?: { id?: string } }
    appStore.showSuccess(t('payment.admin.refundPending'))
    closeRefundDialog()
    if (result.ticket?.id) {
      await router.push({ path: '/admin/tickets', query: { ticket: result.ticket.id } })
    } else {
      await loadOrders()
    }
  } catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
  finally { refundSubmitting.value = false }
}

async function handleQueryRefund(order: PaymentOrder) {
  refundQueryingIds.value = new Set(refundQueryingIds.value).add(order.id)
  try {
    const res = await adminPaymentAPI.queryRefund(order.id)
    if (res.data.status === 'SUCCEEDED') {
      appStore.showSuccess(t('payment.admin.refundSuccess'))
    } else if (res.data.status !== 'FAILED') {
      appStore.showSuccess(t('payment.admin.refundPending'))
    } else {
      appStore.showError(res.data.error_message || t('common.error'))
    }
    loadOrders()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    const next = new Set(refundQueryingIds.value)
    next.delete(order.id)
    refundQueryingIds.value = next
  }
}

function formatDateTime(dateStr: string): string { return formatOrderDateTime(dateStr) }

onMounted(() => loadOrders())
</script>

<style scoped>
.order-delete-dialog__actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

.admin-orders-workspace-tabs {
  display: inline-grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.25rem;
  margin-bottom: 1rem;
  padding: 0.25rem;
  border: 1px solid var(--color-border);
  border-radius: 0.5rem;
  background: var(--color-surface);
}

.admin-orders-workspace-tabs button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.375rem;
  min-height: 2.25rem;
  padding: 0.375rem 0.75rem;
  border-radius: 0.375rem;
  color: var(--color-text-secondary);
  font-size: var(--font-size-sm);
  font-weight: 500;
}

.admin-orders-workspace-tabs button.active {
  background: var(--color-primary, #2563eb);
  color: #fff;
}
</style>
