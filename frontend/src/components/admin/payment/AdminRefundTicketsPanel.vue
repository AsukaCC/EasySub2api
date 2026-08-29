<template>
  <section class="admin-refund-tickets" aria-labelledby="admin-refund-tickets-heading">
    <div class="admin-refund-tickets__toolbar card">
      <div>
        <h2 id="admin-refund-tickets-heading" class="admin-refund-tickets__title">{{ t('payment.admin.refundTickets.title') }}</h2>
        <p class="admin-refund-tickets__description">{{ t('payment.admin.refundTickets.description') }}</p>
      </div>
      <div class="admin-refund-tickets__filters">
        <Select v-model="statusFilter" :options="statusOptions" @change="handleFilterChange" />
        <button class="btn btn-secondary" :disabled="loading" :title="t('common.refresh')" @click="loadTickets">
          <Icon name="refresh" size="sm" :class="loading ? 'admin-refund-tickets__spin' : ''" />
        </button>
      </div>
    </div>

    <div v-if="loading && tickets.length === 0" class="admin-refund-tickets__empty card">{{ t('common.loading') }}</div>
    <div v-else-if="tickets.length === 0" class="admin-refund-tickets__empty card">{{ t('payment.admin.refundTickets.empty') }}</div>
    <div v-else class="table-container card">
      <table class="admin-refund-tickets__table">
        <thead>
          <tr>
            <th>{{ t('payment.admin.refundTickets.ticketId') }}</th>
            <th>{{ t('payment.orders.orderId') }}</th>
            <th>{{ t('payment.admin.refundTickets.user') }}</th>
            <th>{{ t('payment.admin.refundTickets.comment') }}</th>
            <th>{{ t('payment.orders.status') }}</th>
            <th>{{ t('payment.orders.createdAt') }}</th>
            <th class="admin-refund-tickets__actions-heading">{{ t('common.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="ticket in tickets" :key="ticket.id">
            <td class="admin-refund-tickets__mono">#{{ ticket.id }}</td>
            <td class="admin-refund-tickets__mono">#{{ ticket.order_id }}</td>
            <td class="admin-refund-tickets__mono">#{{ ticket.user_id }}</td>
            <td>
              <p>{{ ticket.comment || '-' }}</p>
              <p v-if="ticket.review_note" class="admin-refund-tickets__note">{{ t('payment.admin.refundTickets.reviewNote') }}: {{ ticket.review_note }}</p>
            </td>
            <td><span :class="['badge', badgeClass(ticket.status)]">{{ statusLabel(ticket.status) }}</span></td>
            <td>{{ formatOrderDateTime(ticket.created_at) }}</td>
            <td class="admin-refund-tickets__actions">
              <button v-if="ticket.status === 'PENDING'" class="btn btn-primary" @click="openReview(ticket)">
                <Icon name="clipboard" size="sm" />
                {{ t('payment.admin.refundTickets.review') }}
              </button>
              <span v-else>-</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <Pagination
      v-if="pagination.total > 0"
      :page="pagination.page"
      :total="pagination.total"
      :page-size="pagination.page_size"
      @update:page="handlePageChange"
      @update:pageSize="handlePageSizeChange"
    />

    <BaseDialog :show="!!reviewTarget" :title="t('payment.admin.refundTickets.reviewTitle')" width="normal" @close="closeReview">
      <div v-if="reviewTarget" class="admin-refund-tickets__review-form">
        <dl class="admin-refund-tickets__summary">
          <div><dt>{{ t('payment.admin.refundTickets.ticketId') }}</dt><dd>#{{ reviewTarget.id }}</dd></div>
          <div><dt>{{ t('payment.orders.orderId') }}</dt><dd>#{{ reviewTarget.order_id }}</dd></div>
          <div><dt>{{ t('payment.admin.refundTickets.user') }}</dt><dd>#{{ reviewTarget.user_id }}</dd></div>
          <div class="admin-refund-tickets__summary-wide"><dt>{{ t('payment.admin.refundTickets.comment') }}</dt><dd>{{ reviewTarget.comment || '-' }}</dd></div>
        </dl>

        <div v-if="reviewOrderLoading" class="admin-refund-tickets__order-state">
          <Icon name="refresh" size="sm" class="admin-refund-tickets__spin" />
          {{ t('payment.admin.refundTickets.loadingOrder') }}
        </div>
        <div v-else-if="reviewOrderError" class="admin-refund-tickets__order-error">
          <Icon name="exclamationCircle" size="sm" />
          <span>{{ reviewOrderError }}</span>
          <button type="button" class="btn btn-secondary" @click="loadReviewOrder">{{ t('payment.admin.refundTickets.retryOrder') }}</button>
        </div>
        <dl v-else-if="reviewOrder" class="admin-refund-tickets__amounts">
          <div><dt>{{ t('payment.admin.refundTickets.originalPrincipal') }}</dt><dd>{{ formatCNY(orderPrincipal, localeCode) }}</dd></div>
          <div><dt>{{ t('payment.admin.refundTickets.refundedPrincipal') }}</dt><dd>{{ formatCNY(refundedPrincipal, localeCode) }}</dd></div>
          <div class="admin-refund-tickets__amounts-total"><dt>{{ t('payment.admin.refundTickets.remainingPrincipal') }}</dt><dd>{{ formatCNY(remainingPrincipal, localeCode) }}</dd></div>
          <div><dt>{{ t('payment.admin.refundTickets.basePoints') }}</dt><dd>{{ formatPoints(reviewOrder.base_points ?? orderPrincipal, localeCode) }}</dd></div>
          <div><dt>{{ t('payment.admin.refundTickets.bonusPoints') }}</dt><dd>{{ formatPoints(reviewOrder.bonus_points ?? 0, localeCode) }}</dd></div>
          <div><dt>{{ t('payment.admin.refundTickets.affiliatePoints') }}</dt><dd>{{ formatPoints(reviewOrder.affiliate_rebate_points ?? 0, localeCode) }}</dd></div>
        </dl>

        <div>
          <label class="input-label">{{ t('payment.admin.refundTickets.decision') }}</label>
          <div class="admin-refund-tickets__segments" role="group" :aria-label="t('payment.admin.refundTickets.decision')">
            <button type="button" :class="{ active: reviewDecision === 'APPROVE' }" @click="reviewDecision = 'APPROVE'">
              <Icon name="check" size="sm" />
              {{ t('payment.admin.refundTickets.approve') }}
            </button>
            <button type="button" :class="{ active: reviewDecision === 'REJECT' }" @click="reviewDecision = 'REJECT'">
              <Icon name="x" size="sm" />
              {{ t('payment.admin.refundTickets.reject') }}
            </button>
          </div>
        </div>

        <div v-if="reviewDecision === 'APPROVE'">
          <label class="input-label" for="approved-principal">{{ t('payment.admin.refundTickets.approvedPrincipal') }}</label>
          <div class="admin-refund-tickets__amount-input">
            <span>RMB</span>
            <input id="approved-principal" v-model.number="approvedPrincipal" class="input" type="number" min="0.01" step="0.01" :max="remainingPrincipal || undefined" required />
          </div>
          <p class="admin-refund-tickets__note">{{ t('payment.admin.refundTickets.approvedPrincipalHint', { amount: formatCNY(remainingPrincipal, localeCode) }) }}</p>
        </div>

        <div>
          <label class="input-label" for="refund-review-note">{{ t('payment.admin.refundTickets.reviewNote') }}</label>
          <textarea id="refund-review-note" v-model="reviewNote" class="input" rows="3" :placeholder="t('payment.admin.refundTickets.reviewNotePlaceholder')" />
        </div>

        <div class="admin-refund-tickets__manual-note">
          <Icon name="infoCircle" size="sm" />
          <span>{{ t('payment.admin.refundTickets.affiliateManual') }}</span>
        </div>
      </div>
      <template #footer>
        <div class="admin-refund-tickets__footer">
          <button class="btn btn-secondary" @click="closeReview">{{ t('common.cancel') }}</button>
          <button
            :class="['btn', reviewDecision === 'REJECT' ? 'btn-danger' : 'btn-primary']"
            :disabled="submitting || reviewOrderLoading || !reviewNote.trim() || (reviewDecision === 'APPROVE' && (!reviewOrder || approvedPrincipal == null || approvedPrincipal <= 0 || approvedPrincipal > remainingPrincipal))"
            @click="submitReview"
          >
            {{ submitting ? t('common.processing') : t('payment.admin.refundTickets.submitReview') }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminPaymentAPI } from '@/api/admin/payment'
import { useAppStore } from '@/stores/app'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatOrderDateTime } from '@/components/payment/orderUtils'
import { formatCNY, formatPoints } from '@/utils/format'
import type { PaymentOrder, RefundTicket } from '@/types/payment'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'

const { t, locale } = useI18n()
const localeCode = computed(() => String(locale.value || ''))
const appStore = useAppStore()

const tickets = ref<RefundTicket[]>([])
const loading = ref(false)
const submitting = ref(false)
const statusFilter = ref('PENDING')
const pagination = reactive({ page: 1, page_size: 20, total: 0 })
const reviewTarget = ref<RefundTicket | null>(null)
const reviewOrder = ref<PaymentOrder | null>(null)
const reviewOrderLoading = ref(false)
const reviewOrderError = ref('')
const reviewDecision = ref<'APPROVE' | 'REJECT'>('APPROVE')
const approvedPrincipal = ref<number | null>(null)
const reviewNote = ref('')

const orderPrincipal = computed(() => Number(
  reviewOrder.value?.principal_amount ?? reviewOrder.value?.base_points ?? 0,
))
const refundedPrincipal = computed(() => Number(reviewOrder.value?.refunded_principal_amount ?? 0))
const remainingPrincipal = computed(() => Math.max(0, orderPrincipal.value - refundedPrincipal.value))

const statusOptions = computed(() => [
  { value: '', label: t('payment.admin.refundTickets.allStatuses') },
  ...['PENDING', 'APPROVED', 'PROCESSING', 'COMPLETED', 'FAILED', 'REJECTED', 'CANCELLED'].map(status => ({
    value: status,
    label: statusLabel(status),
  })),
])

async function loadTickets() {
  loading.value = true
  try {
    const res = await adminPaymentAPI.getRefundTickets({
      status: statusFilter.value || undefined,
      page: pagination.page,
      page_size: pagination.page_size,
    })
    tickets.value = res.data.items || []
    pagination.total = res.data.total || 0
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

function handleFilterChange() { pagination.page = 1; void loadTickets() }
function handlePageChange(page: number) { pagination.page = page; void loadTickets() }
function handlePageSizeChange(size: number) { pagination.page_size = size; pagination.page = 1; void loadTickets() }

function openReview(ticket: RefundTicket) {
  reviewTarget.value = ticket
  reviewDecision.value = 'APPROVE'
  approvedPrincipal.value = null
  reviewNote.value = ''
  reviewOrder.value = null
  reviewOrderError.value = ''
  void loadReviewOrder()
}

async function loadReviewOrder() {
  if (!reviewTarget.value) return
  const ticketId = reviewTarget.value.id
  reviewOrderLoading.value = true
  reviewOrderError.value = ''
  try {
    const res = await adminPaymentAPI.getOrder(reviewTarget.value.order_id)
    if (reviewTarget.value?.id !== ticketId) return
    reviewOrder.value = res.data.order
    approvedPrincipal.value = remainingPrincipal.value > 0 ? remainingPrincipal.value : null
  } catch (err: unknown) {
    if (reviewTarget.value?.id !== ticketId) return
    reviewOrder.value = null
    reviewOrderError.value = extractI18nErrorMessage(err, t, 'payment.errors', t('payment.admin.refundTickets.orderLoadFailed'))
  } finally {
    if (reviewTarget.value?.id === ticketId) reviewOrderLoading.value = false
  }
}

function closeReview() {
  reviewTarget.value = null
  reviewOrder.value = null
  reviewOrderError.value = ''
  approvedPrincipal.value = null
  reviewNote.value = ''
}

async function submitReview() {
  if (!reviewTarget.value || !reviewNote.value.trim()) return
  if (reviewDecision.value === 'APPROVE' && (!reviewOrder.value || approvedPrincipal.value == null || approvedPrincipal.value <= 0 || approvedPrincipal.value > remainingPrincipal.value)) return
  submitting.value = true
  try {
    await adminPaymentAPI.reviewRefundTicket(reviewTarget.value.id, {
      decision: reviewDecision.value,
      approved_principal_amount: reviewDecision.value === 'APPROVE' ? approvedPrincipal.value! : undefined,
      review_note: reviewNote.value.trim(),
      affiliate_action: 'MANUAL',
    })
    appStore.showSuccess(t('payment.admin.refundTickets.reviewed'))
    closeReview()
    await loadTickets()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  } finally {
    submitting.value = false
  }
}

function statusLabel(status: string): string {
  return t(`payment.refundTickets.status.${status.toLowerCase()}`, status)
}

function badgeClass(status: string): string {
  if (status === 'COMPLETED') return 'badge-success'
  if (status === 'REJECTED' || status === 'FAILED') return 'badge-danger'
  if (status === 'CANCELLED') return 'badge-gray'
  if (status === 'PENDING') return 'badge-warning'
  return 'badge-primary'
}

onMounted(() => void loadTickets())
</script>

<style scoped>
.admin-refund-tickets {
  display: grid;
  gap: 1rem;
}

.admin-refund-tickets__toolbar,
.admin-refund-tickets__filters,
.admin-refund-tickets__amount-input,
.admin-refund-tickets__manual-note,
.admin-refund-tickets__footer {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.admin-refund-tickets__toolbar {
  justify-content: space-between;
  padding: 0.875rem 1.25rem;
}

.admin-refund-tickets__title {
  font-size: var(--font-size-base);
  font-weight: 600;
}

.admin-refund-tickets__description,
.admin-refund-tickets__note {
  color: var(--text-secondary);
  font-size: var(--font-size-sm);
}

.admin-refund-tickets__empty {
  padding: 2rem;
  color: var(--text-secondary);
  text-align: center;
}

.admin-refund-tickets__table {
  width: 100%;
  border-collapse: collapse;
}

.admin-refund-tickets__table th,
.admin-refund-tickets__table td {
  padding: 0.75rem;
  border-bottom: 1px solid var(--border-color);
  text-align: left;
  vertical-align: top;
}

.admin-refund-tickets__mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: var(--font-size-sm);
}

.admin-refund-tickets__actions-heading,
.admin-refund-tickets__actions {
  text-align: right !important;
}

.admin-refund-tickets__review-form {
  display: grid;
  gap: 1rem;
}

.admin-refund-tickets__summary {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.admin-refund-tickets__summary dt {
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

.admin-refund-tickets__summary dd {
  margin-top: 0.125rem;
  font-weight: 500;
  overflow-wrap: anywhere;
}

.admin-refund-tickets__summary-wide {
  grid-column: 1 / -1;
}

.admin-refund-tickets__order-state,
.admin-refund-tickets__order-error {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  min-height: 4rem;
  padding: 0.75rem;
  border: 1px solid var(--border-color);
  border-radius: 0.5rem;
  color: var(--text-secondary);
}

.admin-refund-tickets__order-error {
  justify-content: flex-start;
  border-color: rgb(220 38 38 / 0.35);
  background: rgb(239 68 68 / 0.08);
}

.admin-refund-tickets__order-error span {
  flex: 1;
}

.admin-refund-tickets__amounts {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  border: 1px solid var(--border-color);
  border-radius: 0.5rem;
  overflow: hidden;
}

.admin-refund-tickets__amounts > div {
  padding: 0.625rem 0.75rem;
  border-bottom: 1px solid var(--border-color);
}

.admin-refund-tickets__amounts dt {
  color: var(--text-secondary);
  font-size: var(--font-size-xs);
}

.admin-refund-tickets__amounts dd {
  margin-top: 0.125rem;
  font-weight: 600;
}

.admin-refund-tickets__amounts-total {
  background: rgb(127 127 127 / 0.06);
}

.admin-refund-tickets__segments {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin-top: 0.375rem;
  padding: 0.25rem;
  border: 1px solid var(--border-color);
  border-radius: 0.5rem;
}

.admin-refund-tickets__segments button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.375rem;
  min-height: 2.25rem;
  border-radius: 0.375rem;
}

.admin-refund-tickets__segments button.active {
  background: var(--color-primary, #2563eb);
  color: #fff;
}

.admin-refund-tickets__manual-note {
  align-items: flex-start;
  padding: 0.75rem;
  border: 1px solid rgb(217 119 6 / 0.35);
  border-radius: 0.5rem;
  background: rgb(245 158 11 / 0.08);
  font-size: var(--font-size-sm);
}

.admin-refund-tickets__footer {
  justify-content: flex-end;
}

.admin-refund-tickets__spin {
  animation: admin-refund-spin 1s linear infinite;
}

@keyframes admin-refund-spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 640px) {
  .admin-refund-tickets__toolbar {
    align-items: stretch;
    flex-direction: column;
  }

  .admin-refund-tickets__filters > :first-child {
    flex: 1;
  }

  .admin-refund-tickets__amounts {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
