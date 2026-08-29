<template>
  <AppLayout>
    <div class="views-user-user-orders-view__panel">
      <!-- Filters -->
      <div class="views-user-user-orders-view__panel-2 card">
        <div class="views-user-user-orders-view__panel-3">
          <Select v-model="currentFilter" :options="statusFilters" class="views-user-user-orders-view__field" @change="fetchOrders" />
          <div class="views-user-user-orders-view__panel-4">
            <button @click="fetchOrders" :disabled="loading" class="btn btn-secondary" :title="t('common.refresh')">
              <Icon name="refresh" size="md" :class="loading ? 'views-user-user-orders-view__icon' : ''" />
            </button>
            <button class="btn btn-primary" @click="router.push('/purchase')">{{ t('payment.result.backToRecharge') }}</button>
          </div>
        </div>
      </div>

      <!-- Table -->
      <OrderTable :orders="orders" :loading="loading">
        <template #actions="{ row }">
          <div class="views-user-user-orders-view__panel-5">
            <button v-if="row.status === 'PENDING'" @click="handleCancel(row.id)" class="views-user-user-orders-view__action">
              <Icon name="x" size="sm" />
              <span>{{ t('payment.orders.cancel') }}</span>
            </button>
            <button v-if="canRequestRefund(row)" @click="openRefundDialog(row)" class="views-user-user-orders-view__action-2">
              <Icon name="dollar" size="sm" />
              <span>{{ t('payment.orders.requestRefund') }}</span>
            </button>
          </div>
        </template>
      </OrderTable>

      <!-- Pagination -->
      <Pagination
        v-if="pagination.total > 0"
        :page="pagination.page"
        :total="pagination.total"
        :page-size="pagination.page_size"
        @update:page="handlePageChange"
        @update:pageSize="handlePageSizeChange"
      />

      <section class="refund-tickets card" aria-labelledby="refund-ticket-heading">
        <div class="refund-tickets__header">
          <div>
            <h2 id="refund-ticket-heading" class="refund-tickets__title">{{ t('payment.refundTickets.title') }}</h2>
            <p class="refund-tickets__description">{{ t('payment.refundTickets.description') }}</p>
          </div>
          <button class="btn btn-secondary" :disabled="ticketsLoading" :title="t('common.refresh')" @click="fetchRefundTickets">
            <Icon name="refresh" size="sm" :class="ticketsLoading ? 'views-user-user-orders-view__icon' : ''" />
          </button>
        </div>

        <div v-if="ticketsLoading && refundTickets.length === 0" class="refund-tickets__empty">
          {{ t('common.loading') }}
        </div>
        <div v-else-if="refundTickets.length === 0" class="refund-tickets__empty">
          {{ t('payment.refundTickets.empty') }}
        </div>
        <div v-else class="table-container">
          <table class="refund-tickets__table">
            <thead>
              <tr>
                <th>{{ t('payment.orders.orderId') }}</th>
                <th>{{ t('payment.refundTickets.comment') }}</th>
                <th>{{ t('payment.orders.status') }}</th>
                <th>{{ t('payment.refundTickets.approvedPrincipal') }}</th>
                <th>{{ t('payment.orders.createdAt') }}</th>
                <th class="refund-tickets__actions-heading">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="ticket in refundTickets" :key="ticket.id">
                <td class="refund-tickets__mono">#{{ ticket.order_id }}</td>
                <td>
                  <p>{{ ticket.comment || '-' }}</p>
                  <p v-if="ticket.review_note" class="refund-tickets__note">{{ t('payment.refundTickets.reviewNote') }}: {{ ticket.review_note }}</p>
                </td>
                <td><span :class="['badge', refundTicketBadgeClass(ticket.status)]">{{ refundTicketStatusLabel(ticket.status) }}</span></td>
                <td>{{ ticket.approved_principal_amount == null ? '-' : formatCNY(ticket.approved_principal_amount, localeCode) }}</td>
                <td>{{ formatOrderDateTime(ticket.created_at) }}</td>
                <td class="refund-tickets__actions">
                  <button
                    v-if="ticket.status === 'PENDING'"
                    class="btn btn-secondary"
                    :disabled="cancellingTicketIds.has(ticket.id)"
                    @click="cancelRefundTicket(ticket.id)"
                  >
                    <Icon name="x" size="sm" />
                    {{ t('payment.refundTickets.cancel') }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <Pagination
          v-if="ticketPagination.total > ticketPagination.page_size"
          :page="ticketPagination.page"
          :total="ticketPagination.total"
          :page-size="ticketPagination.page_size"
          @update:page="handleTicketPageChange"
          @update:pageSize="handleTicketPageSizeChange"
        />
      </section>
    </div>

    <!-- Cancel Confirm Dialog -->
    <BaseDialog :show="!!cancelTargetId" :title="t('payment.orders.cancel')" width="narrow" @close="cancelTargetId = null">
      <p class="views-user-user-orders-view__description">{{ t('payment.confirmCancel') }}</p>
      <template #footer>
        <div class="views-user-user-orders-view__panel-6">
          <button class="btn btn-secondary" @click="cancelTargetId = null">{{ t('common.cancel') }}</button>
          <button class="btn btn-danger" :disabled="actionLoading" @click="confirmCancel">{{ actionLoading ? t('common.processing') : t('payment.orders.cancel') }}</button>
        </div>
      </template>
    </BaseDialog>

    <!-- Refund Dialog -->
    <BaseDialog :show="!!refundTarget" :title="t('payment.orders.requestRefund')" width="wide" @close="closeRefundDialog">
      <div v-if="refundTarget" class="refund-dialog">
        <div class="views-user-user-orders-view__panel-7">
          <div class="views-user-user-orders-view__panel-8">
            <span class="views-user-user-orders-view__text">{{ t('payment.orders.orderId') }}</span>
            <span class="views-user-user-orders-view__text-2">#{{ refundTarget.id }}</span>
          </div>
          <div class="views-user-user-orders-view__panel-9">
            <span class="views-user-user-orders-view__text">{{ t('payment.orders.baseAmount') }}</span>
            <span class="views-user-user-orders-view__text-3">{{ formatCNY(rechargePrincipalAmount(refundTarget), localeCode) }}</span>
          </div>
          <div class="views-user-user-orders-view__panel-9">
            <span class="views-user-user-orders-view__text">{{ t('payment.orders.creditedPoints') }}</span>
            <span class="views-user-user-orders-view__text-3">{{ formatPoints(rechargeCreditedPoints(refundTarget), localeCode) }}</span>
          </div>
        </div>

        <div v-if="refundQuoteLoading && !refundQuote" class="refund-dialog__state">
          <Icon name="refresh" size="sm" class="views-user-user-orders-view__icon" />
          {{ t('payment.refundQuote.loading') }}
        </div>
        <div v-else-if="refundQuoteError" class="refund-dialog__alert refund-dialog__alert--danger">
          <Icon name="exclamationCircle" size="sm" />
          <span>{{ refundQuoteError }}</span>
          <button class="btn btn-secondary" @click="loadRefundQuote()">{{ t('payment.refundQuote.retry') }}</button>
        </div>
        <template v-else-if="refundQuote">
          <div v-if="refundQuote.requires_ticket" class="refund-dialog__alert refund-dialog__alert--warning">
            <Icon name="clock" size="sm" />
            <div>
              <strong>{{ t('payment.refundQuote.ticketRequired') }}</strong>
              <p>{{ refundQuote.blocked_reason || t('payment.refundQuote.ticketRequiredHint') }}</p>
            </div>
          </div>

          <div v-else>
            <label class="input-label" for="refund-principal">{{ t('payment.refundQuote.principalInput') }}</label>
            <div class="refund-dialog__amount-input">
              <span>RMB</span>
              <input
                id="refund-principal"
                v-model.number="refundPrincipal"
                class="input"
                type="number"
                min="0.01"
                step="0.01"
                :max="refundQuote.remaining_principal_amount"
                @input="scheduleRefundQuote"
              />
            </div>
            <p class="refund-dialog__hint">
              {{ t('payment.refundQuote.maximum', { amount: formatCNY(refundQuote.max_refundable_principal_amount, localeCode) }) }}
            </p>
            <p v-if="quoteWasClamped" class="refund-dialog__clamp">
              {{ t('payment.refundQuote.clamped', { amount: formatCNY(refundQuote.principal_amount, localeCode) }) }}
            </p>
          </div>

          <dl class="refund-dialog__breakdown">
            <div><dt>{{ t('payment.refundQuote.principal') }}</dt><dd>{{ formatCNY(refundQuote.principal_amount, localeCode) }}</dd></div>
            <div><dt>{{ t('payment.refundQuote.fee') }}</dt><dd>{{ formatCNY(refundQuote.fee_amount, localeCode) }}</dd></div>
            <div class="refund-dialog__breakdown-total"><dt>{{ t('payment.refundQuote.gatewayAmount') }}</dt><dd>{{ formatCNY(refundQuote.gateway_amount, localeCode) }}</dd></div>
            <div><dt>{{ t('payment.refundQuote.basePoints') }}</dt><dd>{{ formatPoints(refundQuote.base_points, localeCode) }}</dd></div>
            <div><dt>{{ t('payment.refundQuote.bonusPoints') }}</dt><dd>{{ formatPoints(refundQuote.bonus_points, localeCode) }}</dd></div>
            <div><dt>{{ t('payment.refundQuote.expiredBonusOffset') }}</dt><dd>{{ formatPoints(refundQuote.bonus_expired_offset, localeCode) }}</dd></div>
            <div class="refund-dialog__breakdown-total"><dt>{{ t('payment.refundQuote.pointsToHold') }}</dt><dd>{{ formatPoints(refundQuote.points_to_hold, localeCode) }}</dd></div>
            <div><dt>{{ t('payment.refundQuote.affiliatePoints') }}</dt><dd>{{ formatPoints(refundQuote.affiliate_rebate_points, localeCode) }}</dd></div>
          </dl>
          <p v-if="refundQuote.refund_deadline" class="refund-dialog__hint">
            {{ t('payment.refundQuote.deadline', { date: formatOrderDateTime(refundQuote.refund_deadline) }) }}
          </p>
        </template>

        <div>
          <label class="input-label">{{ t('payment.refundReason') }}</label>
          <textarea v-model="refundReason" rows="3" class="views-user-user-orders-view__field-2 input" :placeholder="refundQuote?.requires_ticket ? t('payment.refundTickets.commentPlaceholder') : t('payment.refundReasonPlaceholder')" />
        </div>
      </div>
      <template #footer>
        <div class="views-user-user-orders-view__panel-6">
          <button class="btn btn-secondary" @click="closeRefundDialog">{{ t('common.cancel') }}</button>
          <button
            class="btn btn-primary"
            :disabled="actionLoading || refundQuoteLoading || !refundQuote || !refundReason.trim() || (!refundQuote.requires_ticket && (!refundPrincipal || refundPrincipal <= 0))"
            @click="confirmRefund"
          >
            {{ actionLoading ? t('common.processing') : refundQuote?.requires_ticket ? t('payment.refundTickets.submit') : t('payment.refundQuote.confirm') }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { PaymentOrder, RefundQuote, RefundTicket } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import OrderTable from '@/components/payment/OrderTable.vue'
import { formatOrderDateTime, rechargeCreditedPoints, rechargePrincipalAmount } from '@/components/payment/orderUtils'
import { formatCNY, formatPoints } from '@/utils/format'

const { t, locale } = useI18n()
const localeCode = computed(() => String(locale.value || ''))
const router = useRouter()
const appStore = useAppStore()

const loading = ref(true)
const actionLoading = ref(false)
const orders = ref<PaymentOrder[]>([])
const refundCapableProviders = ref<Set<string>>(new Set())
const refundTickets = ref<RefundTicket[]>([])
const ticketsLoading = ref(false)
const cancellingTicketIds = ref(new Set<string>())
const currentFilter = ref('')
const cancelTargetId = ref<string | null>(null)
const refundTarget = ref<PaymentOrder | null>(null)
const refundReason = ref('')
const refundPrincipal = ref<number | null>(null)
const refundQuote = ref<RefundQuote | null>(null)
const refundQuoteLoading = ref(false)
const refundQuoteError = ref('')
const refundIdempotencyKey = ref('')
const pagination = reactive({ page: 1, page_size: 20, total: 0 })
const ticketPagination = reactive({ page: 1, page_size: 10, total: 0 })

let refundQuoteTimer: ReturnType<typeof setTimeout> | null = null
let refundQuoteRequestId = 0

const quoteWasClamped = computed(() => {
  if (!refundQuote.value || refundPrincipal.value == null) return false
  return refundPrincipal.value - refundQuote.value.principal_amount >= 0.005
})

const statusFilters = computed(() => [
  { value: '', label: t('common.all') },
  { value: 'PENDING', label: t('payment.status.pending') },
  { value: 'COMPLETED', label: t('payment.status.completed') },
  { value: 'FAILED', label: t('payment.status.failed') },
  { value: 'REFUNDED', label: t('payment.status.refunded') },
])

async function fetchOrders() {
  loading.value = true
  try {
    const res = await paymentAPI.getMyOrders({
      page: pagination.page,
      page_size: pagination.page_size,
      status: currentFilter.value || undefined,
    })
    orders.value = res.data.items || []
    pagination.total = res.data.total || 0
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) { pagination.page = page; fetchOrders() }
function handlePageSizeChange(size: number) { pagination.page_size = size; pagination.page = 1; fetchOrders() }

async function fetchRefundTickets() {
  ticketsLoading.value = true
  try {
    const res = await paymentAPI.getMyRefundTickets({
      page: ticketPagination.page,
      page_size: ticketPagination.page_size,
    })
    refundTickets.value = res.data.items || []
    ticketPagination.total = res.data.total || 0
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    ticketsLoading.value = false
  }
}

function handleTicketPageChange(page: number) { ticketPagination.page = page; fetchRefundTickets() }
function handleTicketPageSizeChange(size: number) { ticketPagination.page_size = size; ticketPagination.page = 1; fetchRefundTickets() }

function handleCancel(orderId: string) { cancelTargetId.value = orderId }

async function confirmCancel() {
  if (!cancelTargetId.value) return
  actionLoading.value = true
  try {
    await paymentAPI.cancelOrder(cancelTargetId.value)
    appStore.showSuccess(t('common.success'))
    cancelTargetId.value = null
    await fetchOrders()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

function createRefundIdempotencyKey(orderId: string): string {
  const randomPart = globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(16).slice(2)}`
  return `self-refund-${orderId}-${randomPart}`
}

function openRefundDialog(order: PaymentOrder) {
  refundTarget.value = order
  refundReason.value = ''
  refundPrincipal.value = null
  refundQuote.value = null
  refundQuoteError.value = ''
  refundIdempotencyKey.value = createRefundIdempotencyKey(order.id)
  void loadRefundQuote()
}

function closeRefundDialog() {
  if (refundQuoteTimer) clearTimeout(refundQuoteTimer)
  refundQuoteRequestId += 1
  refundTarget.value = null
  refundReason.value = ''
  refundPrincipal.value = null
  refundQuote.value = null
  refundQuoteError.value = ''
  refundIdempotencyKey.value = ''
}

async function loadRefundQuote(principalAmount: number | null = refundPrincipal.value) {
  if (!refundTarget.value) return
  const requestId = ++refundQuoteRequestId
  const orderId = refundTarget.value.id
  refundQuoteLoading.value = true
  refundQuoteError.value = ''
  try {
    const res = await paymentAPI.getRefundQuote(
      orderId,
      principalAmount != null && principalAmount > 0 ? principalAmount : undefined,
    )
    if (requestId !== refundQuoteRequestId || refundTarget.value?.id !== orderId) return
    refundQuote.value = res.data
    if (principalAmount == null) refundPrincipal.value = res.data.remaining_principal_amount
  } catch (err: unknown) {
    if (requestId !== refundQuoteRequestId) return
    refundQuote.value = null
    refundQuoteError.value = extractI18nErrorMessage(err, t, 'payment.errors', t('payment.refundQuote.loadFailed'))
  } finally {
    if (requestId === refundQuoteRequestId) refundQuoteLoading.value = false
  }
}

function scheduleRefundQuote() {
  if (refundQuoteTimer) clearTimeout(refundQuoteTimer)
  if (refundPrincipal.value == null || refundPrincipal.value <= 0) return
  refundQuoteTimer = setTimeout(() => void loadRefundQuote(refundPrincipal.value), 350)
}

async function confirmRefund() {
  if (!refundTarget.value || !refundQuote.value || !refundReason.value.trim()) return
  actionLoading.value = true
  try {
    if (refundQuote.value.requires_ticket) {
      await paymentAPI.createRefundTicket(refundTarget.value.id, { comment: refundReason.value.trim() })
      appStore.showSuccess(t('payment.refundTickets.submitted'))
    } else {
      if (refundPrincipal.value == null || refundPrincipal.value <= 0 || !refundIdempotencyKey.value) return
      await paymentAPI.createRefund(
        refundTarget.value.id,
        { principal_amount: refundPrincipal.value, reason: refundReason.value.trim() },
        refundIdempotencyKey.value,
      )
      appStore.showSuccess(t('payment.refundQuote.submitted'))
    }
    closeRefundDialog()
    await Promise.all([fetchOrders(), fetchRefundTickets()])
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    actionLoading.value = false
  }
}

async function cancelRefundTicket(ticketId: string) {
  cancellingTicketIds.value = new Set(cancellingTicketIds.value).add(ticketId)
  try {
    await paymentAPI.cancelRefundTicket(ticketId)
    appStore.showSuccess(t('payment.refundTickets.cancelled'))
    await fetchRefundTickets()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    const next = new Set(cancellingTicketIds.value)
    next.delete(ticketId)
    cancellingTicketIds.value = next
  }
}

function refundTicketStatusLabel(status: string): string {
  return t(`payment.refundTickets.status.${status.toLowerCase()}`, status)
}

function refundTicketBadgeClass(status: string): string {
  if (status === 'COMPLETED') return 'badge-success'
  if (status === 'REJECTED' || status === 'FAILED') return 'badge-danger'
  if (status === 'CANCELLED') return 'badge-gray'
  if (status === 'PENDING') return 'badge-warning'
  return 'badge-primary'
}

function canRequestRefund(order: PaymentOrder): boolean {
  if (order.order_type !== 'balance') return false
  if (!['COMPLETED', 'PARTIALLY_REFUNDED', 'REFUND_FAILED'].includes(order.status)) return false
  if (!order.provider_instance_id) return false
  return refundCapableProviders.value.has(order.provider_instance_id)
}

async function loadRefundEligibility() {
  try {
    const res = await paymentAPI.getRefundEligibleProviders()
    refundCapableProviders.value = new Set(
      res.data.refund_enabled_provider_instance_ids
        || res.data.provider_instance_ids
        || [],
    )
  } catch { /* ignore — default to hiding refund button */ }
}

onMounted(() => { fetchOrders(); fetchRefundTickets(); loadRefundEligibility() })
onUnmounted(() => { if (refundQuoteTimer) clearTimeout(refundQuoteTimer) })
</script>

<style scoped>
.refund-tickets {
  padding: 1rem 1.25rem;
}

.refund-tickets__header,
.refund-dialog__amount-input,
.refund-dialog__alert {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.refund-tickets__header {
  justify-content: space-between;
  margin-bottom: 1rem;
}

.refund-tickets__title {
  font-size: var(--font-size-base);
  font-weight: 600;
}

.refund-tickets__description,
.refund-tickets__note,
.refund-dialog__hint {
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
}

.refund-tickets__empty,
.refund-dialog__state {
  padding: 1.5rem;
  color: var(--text-secondary);
  text-align: center;
}

.refund-tickets__table {
  width: 100%;
  border-collapse: collapse;
}

.refund-tickets__table th,
.refund-tickets__table td {
  padding: 0.75rem;
  border-bottom: 1px solid var(--border-color);
  text-align: left;
  vertical-align: top;
}

.refund-tickets__mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: var(--font-size-sm);
}

.refund-tickets__actions-heading,
.refund-tickets__actions {
  text-align: right !important;
}

.refund-dialog {
  display: grid;
  gap: 1rem;
}

.refund-dialog__state {
  display: flex;
  justify-content: center;
  gap: 0.5rem;
}

.refund-dialog__alert {
  align-items: flex-start;
  justify-content: space-between;
  padding: 0.75rem;
  border: 1px solid var(--border-color);
  border-radius: 0.5rem;
}

.refund-dialog__alert--warning {
  border-color: rgb(217 119 6 / 0.35);
  background: rgb(245 158 11 / 0.08);
}

.refund-dialog__alert--danger {
  border-color: rgb(220 38 38 / 0.35);
  background: rgb(239 68 68 / 0.08);
}

.refund-dialog__amount-input > span {
  flex: 0 0 auto;
  font-size: var(--font-size-sm);
  font-weight: 600;
}

.refund-dialog__amount-input .input {
  min-width: 0;
}

.refund-dialog__clamp {
  margin-top: 0.375rem;
  color: rgb(180 83 9);
  font-size: var(--font-size-sm);
}

.refund-dialog__breakdown {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  border: 1px solid var(--border-color);
  border-radius: 0.5rem;
  overflow: hidden;
}

.refund-dialog__breakdown > div {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.625rem 0.75rem;
  border-bottom: 1px solid var(--border-color);
}

.refund-dialog__breakdown dt {
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
}

.refund-dialog__breakdown dd {
  font-weight: 600;
}

.refund-dialog__breakdown-total {
  background: rgb(127 127 127 / 0.06);
}

@media (max-width: 640px) {
  .refund-dialog__breakdown {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
