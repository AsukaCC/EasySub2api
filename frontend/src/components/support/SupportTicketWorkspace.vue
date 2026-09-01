<template>
  <div class="page-stack ticket-workspace">
    <div class="ticket-workspace__toolbar card">
      <div class="ticket-workspace__filters">
        <Select v-model="filters.category" :options="categoryOptions" @change="reload" />
        <Select v-model="filters.status" :options="statusOptions" @change="reload" />
        <SearchInput
          v-if="admin"
          v-model="filters.search"
          class="ticket-workspace__search"
          :placeholder="t('tickets.searchPlaceholder')"
          @search="reload"
        />
        <button
          v-if="admin"
          type="button"
          class="ticket-workspace__chip"
          :class="filters.unread && 'ticket-workspace__chip--on'"
          @click="toggleUnread"
        >
          {{ t('tickets.unreadOnly') }}
        </button>
      </div>
      <div class="ticket-workspace__actions">
        <span v-if="summary?.unread" class="ticket-workspace__unread-count">
          {{ t('tickets.unreadBadge', { n: summary.unread }) }}
        </span>
        <button class="btn btn-secondary" :disabled="loading" :title="t('common.refresh')" @click="reload">
          <Icon name="refresh" size="sm" :class="loading && 'ticket-workspace__spin'" />
        </button>
        <button v-if="!admin && canCreate" class="btn btn-primary" @click="openCreate()">
          <Icon name="plus" size="sm" />
          {{ t('tickets.create') }}
        </button>
      </div>
    </div>

    <div v-if="!admin && summary && !summary.feature_enabled && summary.total === 0" class="ticket-workspace__notice card">
      {{ t('tickets.disabled') }}
    </div>

    <div class="ticket-workspace__body card" :class="selectedId && 'ticket-workspace__body--detail'">
      <section class="ticket-workspace__list" :aria-label="t('tickets.list')">
        <div v-if="loading && tickets.length === 0" class="empty-state">
          <p class="empty-state-title">{{ t('common.loading') }}</p>
        </div>
        <div v-else-if="tickets.length === 0" class="empty-state">
          <Icon name="inbox" size="xl" class="empty-state-icon" />
          <p class="empty-state-title">{{ t('tickets.emptyTitle') }}</p>
          <p class="empty-state-description">{{ t('tickets.empty') }}</p>
        </div>
        <button
          v-for="ticket in tickets"
          v-else
          :key="ticket.id"
          type="button"
          class="ticket-item"
          :class="{ 'ticket-item--active': ticket.id === selectedId, 'ticket-item--unread': ticket.unread }"
          @click="selectTicket(ticket.id)"
        >
          <span class="ticket-item__topline">
            <span class="ticket-item__category">{{ categoryLabel(ticket.category) }}</span>
            <span :class="['ticket-status', `ticket-status--${ticket.status.toLowerCase()}`]">{{ statusLabel(ticket.status) }}</span>
          </span>
          <strong class="ticket-item__title">{{ ticket.title }}</strong>
          <span class="ticket-item__meta">
            <span>#{{ ticket.id.slice(0, 8) }}</span>
            <span v-if="admin">{{ ticketOwner(ticket) }}</span>
            <time>{{ formatDate(ticket.updated_at) }}</time>
            <span v-if="ticket.unread" class="ticket-item__dot" :title="t('tickets.unread')"></span>
          </span>
        </button>
        <Pagination
          v-if="pagination.total > pagination.pageSize"
          :page="pagination.page"
          :page-size="pagination.pageSize"
          :total="pagination.total"
          @update:page="changePage"
        />
      </section>

      <section v-if="detail" class="ticket-detail">
        <header class="ticket-detail__header">
          <button type="button" class="ticket-detail__back" :title="t('common.back')" @click="closeDetail">
            <Icon name="arrowLeft" size="sm" />
          </button>
          <div class="ticket-detail__heading">
            <h2>{{ detail.ticket.title }}</h2>
            <p>#{{ detail.ticket.id }} · {{ categoryLabel(detail.ticket.category) }}</p>
          </div>
          <span :class="['ticket-status', `ticket-status--${detail.ticket.status.toLowerCase()}`]">{{ statusLabel(detail.ticket.status) }}</span>
        </header>

        <dl class="ticket-detail__facts">
          <div v-if="admin && detail.user">
            <dt>{{ t('tickets.user') }}</dt>
            <dd>{{ detail.user.username || detail.user.email }}<br />{{ detail.user.email }}</dd>
          </div>
          <div v-if="detail.ticket.api_key_id">
            <dt>{{ t('tickets.apiKey') }}</dt>
            <dd>{{ detail.ticket.api_key_name_snapshot || detail.ticket.api_key_id }}</dd>
          </div>
          <div v-if="detail.ticket.group_id">
            <dt>{{ t('tickets.group') }}</dt>
            <dd>{{ detail.ticket.group_name_snapshot || detail.ticket.group_id }}</dd>
          </div>
          <div v-if="detail.ticket.order_id">
            <dt>{{ t('tickets.order') }}</dt>
            <dd>#{{ detail.ticket.order_id }}</dd>
          </div>
          <div v-if="detail.ticket.approved_principal_amount != null">
            <dt>{{ t('tickets.approvedAmount') }}</dt>
            <dd>RMB {{ detail.ticket.approved_principal_amount.toFixed(2) }}</dd>
          </div>
          <div>
            <dt>{{ t('tickets.createdAt') }}</dt>
            <dd>{{ formatDate(detail.ticket.created_at) }}</dd>
          </div>
        </dl>

        <div class="ticket-timeline" aria-live="polite">
          <article
            v-for="message in detail.messages"
            :key="message.id"
            class="ticket-message"
            :class="{
              'ticket-message--mine': isMine(message),
              'ticket-message--system': message.kind === 'SYSTEM'
            }"
          >
            <div class="ticket-message__meta">
              <strong>{{ authorLabel(message.author_role) }}</strong>
              <time>{{ formatDate(message.created_at) }}</time>
            </div>
            <p>{{ message.body || eventLabel(message.event_type) }}</p>
          </article>
        </div>

        <div v-if="admin && (detail.permissions.can_review_refund || detail.permissions.can_retry_refund)" class="ticket-refund">
          <h3>{{ t('tickets.refundReview') }}</h3>
          <label v-if="detail.permissions.can_review_refund">
            <span>{{ t('tickets.principalAmount') }}</span>
            <input v-model.number="refundAmount" class="input" type="number" min="0.01" step="0.01" />
          </label>
          <div class="ticket-refund__actions">
            <button v-if="detail.permissions.can_review_refund" class="btn btn-primary" :disabled="submitting || !validRefundAmount" @click="reviewRefund('APPROVE')">{{ t('tickets.approve') }}</button>
            <button v-if="detail.permissions.can_review_refund" class="btn btn-danger" :disabled="submitting" @click="reviewRefund('REJECT')">{{ t('tickets.reject') }}</button>
            <button v-if="detail.permissions.can_retry_refund" class="btn btn-secondary" :disabled="submitting" @click="reviewRefund('RETRY')">{{ t('tickets.retry') }}</button>
          </div>
        </div>

        <form v-if="detail.permissions.can_reply" class="ticket-reply" @submit.prevent="sendReply">
          <textarea v-model="reply" class="input" rows="4" maxlength="4000" :placeholder="t('tickets.replyPlaceholder')"></textarea>
          <div class="ticket-reply__aside">
            <span class="ticket-reply__count">{{ reply.length }}/4000</span>
            <button class="btn btn-primary" :disabled="submitting || !reply.trim()" type="submit">
              <Icon name="arrowUp" size="sm" />
              {{ t('tickets.send') }}
            </button>
          </div>
        </form>

        <footer class="ticket-detail__actions">
          <template v-if="admin">
            <button v-if="!isTerminal" class="btn btn-secondary" :disabled="submitting" @click="setStatus('IN_PROGRESS')">{{ t('tickets.markInProgress') }}</button>
            <button v-if="!isTerminal" class="btn btn-secondary" :disabled="submitting" @click="setStatus('RESOLVED')">{{ t('tickets.resolve') }}</button>
            <button v-if="detail.ticket.status !== 'CLOSED'" class="btn btn-secondary" :disabled="submitting" @click="setStatus('CLOSED')">{{ t('tickets.close') }}</button>
          </template>
          <template v-else>
            <button v-if="detail.permissions.can_cancel" class="btn btn-danger" :disabled="submitting" @click="userAction('cancel')">{{ t('tickets.cancel') }}</button>
            <button v-if="detail.permissions.can_close" class="btn btn-secondary" :disabled="submitting" @click="userAction('close')">{{ t('tickets.close') }}</button>
            <button v-if="detail.permissions.can_reopen" class="btn btn-primary" :disabled="submitting" @click="userAction('reopen')">{{ t('tickets.reopen') }}</button>
          </template>
        </footer>
      </section>
      <section v-else class="ticket-workspace__placeholder empty-state">
        <Icon name="clipboard" size="xl" class="empty-state-icon" />
        <p class="empty-state-title">{{ t('tickets.selectHint') }}</p>
      </section>
    </div>

    <BaseDialog :show="showCreate" :title="t('tickets.create')" width="wide" @close="showCreate = false">
      <form id="ticket-create-form" class="ticket-create" @submit.prevent="createTicket">
        <label><span>{{ t('tickets.category') }}</span><Select v-model="draft.category" :options="createCategoryOptions" /></label>
        <label v-if="draft.category === 'ACCOUNT'"><span>{{ t('tickets.subject') }}</span><input v-model="draft.title" class="input" maxlength="120" required /></label>
        <label v-if="draft.category === 'ACCOUNT'"><span>{{ t('tickets.apiKeyOptional') }}</span><Select v-model="draft.apiKeyId" :options="keyOptions" searchable clearable /></label>
        <label v-else><span>{{ t('tickets.order') }}</span><Select v-model="draft.orderId" :options="orderOptions" searchable /></label>
        <label><span>{{ draft.category === 'REFUND' ? t('tickets.refundReason') : t('tickets.descriptionLabel') }}</span><textarea v-model="draft.message" class="input" rows="6" maxlength="4000" required></textarea></label>
      </form>
      <template #footer>
        <button class="btn btn-secondary" @click="showCreate = false">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" form="ticket-create-form" :disabled="submitting || !createValid">{{ t('tickets.submit') }}</button>
      </template>
    </BaseDialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { adminSupportTicketsAPI, supportTicketsAPI } from '@/api/supportTickets'
import { keysAPI } from '@/api/keys'
import { paymentAPI } from '@/api/payment'
import { useAppStore } from '@/stores'
import { extractI18nErrorMessage } from '@/utils/apiError'
import type { ApiKey } from '@/types'
import type { PaymentOrder } from '@/types/payment'
import type { SupportTicket, SupportTicketCategory, SupportTicketDetail, SupportTicketMessage, SupportTicketStatus, SupportTicketSummary } from '@/types/supportTicket'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import Pagination from '@/components/common/Pagination.vue'
import SearchInput from '@/components/common/SearchInput.vue'
import Select from '@/components/common/Select.vue'

const props = defineProps<{ admin?: boolean }>()
const { t, locale } = useI18n()
const route = useRoute()
const router = useRouter()
const appStore = useAppStore()
const admin = computed(() => props.admin === true)
const tickets = ref<SupportTicket[]>([])
const detail = ref<SupportTicketDetail | null>(null)
const summary = ref<SupportTicketSummary | null>(null)
const selectedId = ref('')
const loading = ref(false)
const submitting = ref(false)
const showCreate = ref(false)
const reply = ref('')
const refundAmount = ref<number | null>(null)
const apiKeys = ref<ApiKey[]>([])
const orders = ref<PaymentOrder[]>([])
const filters = reactive<{ category: SupportTicketCategory | ''; status: SupportTicketStatus | ''; search: string; unread: boolean }>({ category: '', status: '', search: '', unread: false })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const draft = reactive<{ category: SupportTicketCategory; title: string; message: string; apiKeyId: string; orderId: string }>({ category: 'ACCOUNT', title: '', message: '', apiKeyId: '', orderId: '' })

const categoryOptions = computed(() => [{ value: '', label: t('common.all') }, { value: 'ACCOUNT', label: t('tickets.categories.account') }, { value: 'REFUND', label: t('tickets.categories.refund') }])
const createCategoryOptions = computed(() => [
  ...(summary.value?.can_create_account ? [{ value: 'ACCOUNT', label: t('tickets.categories.account') }] : []),
  ...(summary.value?.can_create_refund ? [{ value: 'REFUND', label: t('tickets.categories.refund') }] : []),
])
const statusOptions = computed(() => [{ value: '', label: t('common.all') }, ...(['PENDING_ADMIN', 'PENDING_USER', 'IN_PROGRESS', 'RESOLVED', 'CLOSED', 'CANCELLED'] as SupportTicketStatus[]).map(value => ({ value, label: statusLabel(value) }))])
const keyOptions = computed(() => [{ value: '', label: t('tickets.noApiKey') }, ...apiKeys.value.map(key => ({ value: key.id, label: key.name }))])
const orderOptions = computed(() => orders.value.map(order => ({ value: order.id, label: `#${order.id} · ${order.out_trade_no} · RMB ${Number(order.pay_amount || 0).toFixed(2)}` })))
const canCreate = computed(() => Boolean(summary.value?.can_create_account || summary.value?.can_create_refund))
const createValid = computed(() => draft.message.trim().length > 0 && (draft.category === 'ACCOUNT' ? draft.title.trim().length > 0 : draft.orderId.length > 0))
const validRefundAmount = computed(() => refundAmount.value != null && Number.isFinite(refundAmount.value) && refundAmount.value > 0 && Math.round(refundAmount.value * 100) === refundAmount.value * 100)
const isTerminal = computed(() => ['RESOLVED', 'CLOSED', 'CANCELLED'].includes(detail.value?.ticket.status || ''))

function statusLabel(status: string) { return t(`tickets.statuses.${status.toLowerCase()}`, status) }
function categoryLabel(category: string) { return t(`tickets.categories.${category.toLowerCase()}`, category) }
function authorLabel(role: string) { return t(`tickets.authors.${role.toLowerCase()}`, role) }
function eventLabel(event?: string) { return event ? t(`tickets.events.${event.toLowerCase()}`, event) : t('tickets.systemEvent') }
function formatDate(value?: string) { return value ? new Intl.DateTimeFormat(String(locale.value), { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value)) : '-' }
function notifyUpdated() { window.dispatchEvent(new CustomEvent('support-tickets:updated')) }
function ticketOwner(ticket: SupportTicket) { return ticket.username || ticket.email || ticket.user_id.slice(0, 8) }
function isMine(message: SupportTicketMessage) {
  if (message.kind === 'SYSTEM') return false
  return admin.value ? message.author_role === 'ADMIN' : message.author_role === 'USER'
}
function toggleUnread() {
  filters.unread = !filters.unread
  void reload()
}

async function loadSummary() {
  const response = admin.value ? await adminSupportTicketsAPI.summary() : await supportTicketsAPI.summary()
  summary.value = response.data
}

async function reload() {
  loading.value = true
  try {
    const request = admin.value ? adminSupportTicketsAPI.list : supportTicketsAPI.list
    const response = await request({ page: pagination.page, page_size: pagination.pageSize, category: filters.category, status: filters.status, search: filters.search || undefined, unread: filters.unread || undefined })
    tickets.value = response.data.items || []
    pagination.total = response.data.total || 0
    await loadSummary()
  } catch (error) {
    appStore.showError(extractI18nErrorMessage(error, t, 'tickets.errors', t('common.error')))
  } finally { loading.value = false }
}

async function selectTicket(id: string) {
  selectedId.value = id
  if (route.query.ticket !== id) {
    await router.replace({ query: { ...route.query, ticket: id } })
  }
  try {
    const api = admin.value ? adminSupportTicketsAPI : supportTicketsAPI
    detail.value = (await api.detail(id)).data
    refundAmount.value = detail.value.ticket.approved_principal_amount ?? null
    await api.markRead(id)
    const item = tickets.value.find(ticket => ticket.id === id)
    if (item) item.unread = false
    notifyUpdated()
  } catch (error) { appStore.showError(extractI18nErrorMessage(error, t, 'tickets.errors', t('common.error'))) }
}

function closeDetail() { selectedId.value = ''; detail.value = null; void router.replace({ query: { ...route.query, ticket: undefined } }) }
function changePage(page: number) { pagination.page = page; void reload() }

async function loadCreateOptions() {
  const [keys, orderResponse] = await Promise.all([keysAPI.list(1, 100), paymentAPI.getMyOrders({ page: 1, page_size: 100 })])
  apiKeys.value = keys.items || []
  orders.value = (orderResponse.data.items || []).filter(order => ['COMPLETED', 'REFUND_FAILED', 'REFUND_REQUESTED'].includes(order.status))
}

async function openCreate(orderId = '') {
  draft.orderId = orderId
  if (orderId && summary.value?.can_create_refund) draft.category = 'REFUND'
  else draft.category = summary.value?.can_create_account ? 'ACCOUNT' : 'REFUND'
  showCreate.value = true
  try { await loadCreateOptions() } catch { /* selectors remain empty; submit still validates ownership server-side */ }
}

async function createTicket() {
  if (!createValid.value) return
  submitting.value = true
  try {
    const response = await supportTicketsAPI.create({ category: draft.category, title: draft.category === 'ACCOUNT' ? draft.title.trim() : undefined, message: draft.message.trim(), api_key_id: draft.apiKeyId || undefined, order_id: draft.orderId || undefined })
    showCreate.value = false
    draft.title = ''; draft.message = ''; draft.apiKeyId = ''; draft.orderId = ''
    appStore.showSuccess(t('tickets.created'))
    await reload()
    await selectTicket(response.data.ticket.id)
  } catch (error) { appStore.showError(extractI18nErrorMessage(error, t, 'tickets.errors', t('common.error'))) }
  finally { submitting.value = false }
}

async function sendReply() {
  if (!detail.value || !reply.value.trim()) return
  submitting.value = true
  try {
    const api = admin.value ? adminSupportTicketsAPI : supportTicketsAPI
    detail.value = (await api.reply(detail.value.ticket.id, reply.value.trim())).data
    reply.value = ''
    appStore.showSuccess(t('tickets.replied'))
    await reload(); notifyUpdated()
  } catch (error) { appStore.showError(extractI18nErrorMessage(error, t, 'tickets.errors', t('common.error'))) }
  finally { submitting.value = false }
}

async function userAction(action: 'cancel' | 'close' | 'reopen') {
  if (!detail.value) return
  submitting.value = true
  try { detail.value = (await supportTicketsAPI.action(detail.value.ticket.id, action)).data; await reload(); notifyUpdated() }
  catch (error) { appStore.showError(extractI18nErrorMessage(error, t, 'tickets.errors', t('common.error'))) }
  finally { submitting.value = false }
}

async function setStatus(status: SupportTicketStatus) {
  if (!detail.value) return
  submitting.value = true
  try { detail.value = (await adminSupportTicketsAPI.setStatus(detail.value.ticket.id, status)).data; await reload(); notifyUpdated() }
  catch (error) { appStore.showError(extractI18nErrorMessage(error, t, 'tickets.errors', t('common.error'))) }
  finally { submitting.value = false }
}

async function reviewRefund(decision: 'APPROVE' | 'REJECT' | 'RETRY') {
  if (!detail.value) return
  submitting.value = true
  try {
    await adminSupportTicketsAPI.reviewRefund(detail.value.ticket.id, { decision, approved_principal_amount: decision === 'APPROVE' ? refundAmount.value || undefined : undefined })
    await selectTicket(detail.value.ticket.id); await reload(); notifyUpdated()
  } catch (error) { appStore.showError(extractI18nErrorMessage(error, t, 'tickets.errors', t('common.error'))) }
  finally { submitting.value = false }
}

watch(() => route.query.ticket, value => { if (typeof value === 'string' && value !== selectedId.value) void selectTicket(value) })
onMounted(async () => {
  await reload()
  const orderId = typeof route.query.order_id === 'string' ? route.query.order_id : ''
  if (!admin.value && route.query.new === 'refund' && orderId) void openCreate(orderId)
  const ticketId = typeof route.query.ticket === 'string' ? route.query.ticket : ''
  if (ticketId) void selectTicket(ticketId)
})
</script>

<style scoped>
.ticket-workspace {
  min-height: calc(100vh - var(--app-shell-height) - 3.25rem);
}

.ticket-workspace__unread-count {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  min-height: 1.75rem;
  padding: 0.25rem 0.625rem;
  border: 1px solid var(--color-primary-border);
  border-radius: 999px;
  background: var(--glass-tint-brand);
  color: var(--color-text-brand);
  font-size: var(--type-caption-size);
  font-weight: var(--font-weight-medium);
}

.ticket-workspace__toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.875rem 1.25rem;
}

.ticket-workspace__filters,
.ticket-workspace__actions,
.ticket-detail__header,
.ticket-detail__actions,
.ticket-refund__actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.ticket-workspace__filters {
  flex: 1 1 auto;
  min-width: 0;
  flex-wrap: wrap;
}

.ticket-workspace__filters :deep(.app-select) {
  min-width: 10rem;
}

.ticket-workspace__search {
  flex: 1 1 14rem;
  min-width: 12rem;
  max-width: 22rem;
}

.ticket-workspace__actions {
  flex: 0 0 auto;
  margin-left: auto;
}

.ticket-workspace__chip {
  display: inline-flex;
  align-items: center;
  min-height: 2.25rem;
  padding: 0.375rem 0.75rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-md);
  background: var(--glass-layer-inset-bg);
  color: var(--color-text-secondary);
  font-size: var(--type-control-size);
  cursor: pointer;
}

.ticket-workspace__chip--on {
  border-color: var(--color-primary-border);
  background: var(--glass-tint-brand);
  color: var(--color-text-brand);
}

.ticket-workspace__notice {
  padding: 1rem 1.25rem;
  color: var(--color-text-tertiary);
  text-align: center;
}

.ticket-workspace__body {
  display: grid;
  grid-template-columns: minmax(16.875rem, 22.5rem) minmax(0, 1fr);
  flex: 1;
  min-height: 35rem;
  overflow: hidden;
}

.ticket-workspace__list {
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow-y: auto;
  border-right: 1px solid var(--glass-border);
}

.ticket-item {
  display: grid;
  gap: 0.5rem;
  width: 100%;
  min-height: 6.25rem;
  padding: 0.875rem 1rem;
  border: 0;
  border-bottom: 1px solid var(--glass-border);
  background: transparent;
  color: inherit;
  text-align: left;
  cursor: pointer;
  transition: background-color 160ms ease, box-shadow 160ms ease;
}

.ticket-item:hover,
.ticket-item--active {
  background: var(--glass-layer-inset-bg);
  box-shadow: inset 3px 0 var(--color-primary);
}

.ticket-item--unread .ticket-item__title {
  color: var(--color-text-primary);
}

.ticket-item__topline,
.ticket-item__meta,
.ticket-message__meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.ticket-item__topline {
  justify-content: space-between;
}

.ticket-item__category {
  font-size: var(--type-caption-size);
  color: var(--color-text-secondary);
}

.ticket-item__title {
  overflow: hidden;
  font-size: var(--type-control-size);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ticket-item__meta {
  color: var(--color-text-quaternary);
  font-size: var(--type-caption-size);
}

.ticket-item__dot {
  width: 0.4375rem;
  height: 0.4375rem;
  margin-left: auto;
  border-radius: 50%;
  background: var(--color-primary);
  box-shadow: 0 0 0.625rem var(--color-primary);
}

.ticket-status {
  display: inline-flex;
  align-items: center;
  min-height: 1.5rem;
  padding: 0.125rem 0.5rem;
  border: 1px solid var(--glass-border);
  border-radius: 999px;
  background: var(--glass-layer-inset-bg);
  color: var(--color-text-secondary);
  font-size: var(--font-size-2xs);
}

.ticket-status--pending_admin { color: var(--color-text-warning); }
.ticket-status--pending_user { color: var(--color-text-info); }
.ticket-status--in_progress { color: var(--color-text-brand); }
.ticket-status--resolved { color: var(--color-text-success); }
.ticket-status--closed,
.ticket-status--cancelled { color: var(--color-text-quaternary); }

.ticket-detail {
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
}

.ticket-detail__header {
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--glass-border);
}

.ticket-detail__heading {
  min-width: 0;
  flex: 1;
}

.ticket-detail__heading h2 {
  margin: 0;
  font-size: var(--type-dialog-title-size);
}

.ticket-detail__heading p {
  margin: 0.25rem 0 0;
  overflow-wrap: anywhere;
  color: var(--color-text-quaternary);
  font-size: var(--type-caption-size);
}

.ticket-detail__back {
  display: none;
  width: 2rem;
  height: 2rem;
  padding: 0;
  border: 0;
  background: transparent;
  color: inherit;
}

.ticket-detail__facts {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(8.75rem, 1fr));
  gap: 1px;
  margin: 0;
  background: var(--glass-border);
  border-bottom: 1px solid var(--glass-border);
}

.ticket-detail__facts div {
  padding: 0.75rem 1rem;
  background: var(--glass-layer-inset-bg);
}

.ticket-detail__facts dt {
  color: var(--color-text-quaternary);
  font-size: var(--font-size-2xs);
}

.ticket-detail__facts dd {
  margin: 0.25rem 0 0;
  overflow-wrap: anywhere;
  font-size: var(--type-caption-size);
}

.ticket-timeline {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 0.75rem;
  min-height: 13.75rem;
  padding: 1.125rem 1.25rem;
  overflow-y: auto;
}

.ticket-message {
  align-self: flex-start;
  max-width: min(78%, 45rem);
  padding: 0.75rem 0.875rem;
  border: 1px solid var(--glass-border);
  border-radius: var(--radius-lg);
  background: var(--glass-layer-inset-bg);
  -webkit-backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
  backdrop-filter: blur(var(--glass-layer-inset-blur)) saturate(var(--glass-saturate));
}

.ticket-message--mine {
  align-self: flex-end;
  border-color: var(--color-primary-border);
  background: var(--glass-tint-brand);
}

.ticket-message--system {
  align-self: center;
  max-width: 90%;
  color: var(--color-text-tertiary);
}

.ticket-message__meta {
  justify-content: space-between;
  color: var(--color-text-tertiary);
  font-size: var(--font-size-2xs);
}

.ticket-message p {
  margin: 0.5rem 0 0;
  font-size: var(--type-body-size);
  line-height: var(--type-body-line-height);
  overflow-wrap: anywhere;
  white-space: pre-wrap;
}

.ticket-refund,
.ticket-reply {
  padding: 0.875rem 1rem;
  border-top: 1px solid var(--glass-border);
  background: var(--glass-layer-inset-bg);
}

.ticket-refund h3 {
  margin: 0 0 0.625rem;
  font-size: var(--type-card-size);
}

.ticket-refund label {
  display: grid;
  gap: 0.375rem;
  max-width: 16.25rem;
}

.ticket-refund label span,
.ticket-create label > span {
  color: var(--color-text-secondary);
  font-size: var(--type-caption-size);
}

.ticket-refund__actions {
  flex-wrap: wrap;
  margin-top: 0.625rem;
}

.ticket-reply {
  display: flex;
  align-items: flex-end;
  gap: 0.75rem;
}

.ticket-reply textarea {
  flex: 1;
  resize: vertical;
}

.ticket-reply__aside {
  display: grid;
  gap: 0.5rem;
  justify-items: end;
}

.ticket-reply__count {
  color: var(--color-text-quaternary);
  font-size: var(--font-size-2xs);
}

.ticket-detail__actions {
  flex-wrap: wrap;
  justify-content: flex-end;
  padding: 0.75rem 1rem;
  border-top: 1px solid var(--glass-border);
}

.ticket-workspace__placeholder {
  min-height: 100%;
}

.ticket-create {
  display: grid;
  gap: 1rem;
}

.ticket-create label {
  display: grid;
  gap: 0.4375rem;
}

.ticket-create textarea {
  resize: vertical;
}

.ticket-workspace__spin {
  animation: ticket-spin 1s linear infinite;
}

@keyframes ticket-spin {
  to { transform: rotate(360deg); }
}

@media (max-width: 760px) {
  .ticket-workspace {
    min-height: auto;
  }

  .ticket-workspace__body {
    grid-template-columns: 1fr;
    min-height: 32.5rem;
  }

  .ticket-workspace__body--detail .ticket-workspace__list {
    display: none;
  }

  .ticket-workspace__body:not(.ticket-workspace__body--detail) .ticket-workspace__placeholder {
    display: none;
  }

  .ticket-workspace__list {
    border-right: 0;
  }

  .ticket-detail__back {
    display: inline-grid;
    place-items: center;
  }

  .ticket-message {
    max-width: 92%;
  }

  .ticket-reply {
    flex-direction: column;
    align-items: stretch;
  }

  .ticket-reply__aside {
    grid-template-columns: 1fr auto;
    align-items: center;
  }
}
</style>
