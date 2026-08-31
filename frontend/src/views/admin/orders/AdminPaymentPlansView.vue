<template>
  <AppLayout>
    <div class="views-admin-orders-admin-payment-plans-view__panel">
      <!-- Actions -->
      <div class="views-admin-orders-admin-payment-plans-view__panel-2">
        <button @click="loadPlans" :disabled="plansLoading" class="btn btn-secondary" :title="t('common.refresh')">
          <Icon name="refresh" size="md" :class="plansLoading ? 'views-admin-orders-admin-payment-plans-view__icon' : ''" />
        </button>
        <button @click="openPlanEdit(null)" class="btn btn-primary">{{ t('payment.admin.createPlan') }}</button>
      </div>

      <!-- Plans Table -->
      <DataTable :columns="planColumns" :data="plans" :loading="plansLoading">
        <template #cell-name="{ value, row }">
          <span class="views-admin-orders-admin-payment-plans-view__text" :class="getPlanNameClass(row.group_id)">{{ value }}</span>
        </template>
        <template #cell-group_id="{ value }">
          <span v-if="isGroupMissing(value)" class="views-admin-orders-admin-payment-plans-view__text-2">
            <span class="views-admin-orders-admin-payment-plans-view__text-3">#{{ value }}</span>
            <span class="views-admin-orders-admin-payment-plans-view__text-4 badge badge-danger">{{ t('payment.admin.groupMissing') }}</span>
          </span>
          <GroupBadge
            v-else-if="getGroup(value)"
            :name="getGroup(value)!.name"
            :platform="getGroup(value)!.platform"
            :rate-multiplier="getGroup(value)!.rate_multiplier"
          />
          <span v-else class="views-admin-orders-admin-payment-plans-view__text-5">-</span>
        </template>
        <template #cell-price="{ value, row }">
          <div class="views-admin-orders-admin-payment-plans-view__text-2">
            <span class="views-admin-orders-admin-payment-plans-view__text-6">{{ formatPoints(row.price_points ?? value ?? 0, localeCode) }}</span>
            <span v-if="(row.original_price_points ?? row.original_price ?? 0) > 0" class="views-admin-orders-admin-payment-plans-view__text-8">{{ formatPoints(row.original_price_points ?? row.original_price ?? 0, localeCode) }}</span>
          </div>
        </template>
        <template #cell-validity_days="{ value, row }">
          <span class="views-admin-orders-admin-payment-plans-view__text-2">{{ value }} {{ t('payment.admin.' + (row.validity_unit || 'days')) }}</span>
        </template>
        <template #cell-stock_available="{ row }">
          <span v-if="!row.stock_enabled" class="views-admin-orders-admin-payment-plans-view__text-5">{{ t('payment.admin.unlimitedStock') }}</span>
          <span v-else class="views-admin-orders-admin-payment-plans-view__text-2">
            {{ t('payment.admin.stockSummary', { available: row.stock_available ?? 0, frozen: row.stock_frozen ?? 0, total: row.stock_quantity ?? 0 }) }}
          </span>
        </template>
        <template #cell-for_sale="{ value, row }">
          <button
            type="button"
            :class="[
              'views-admin-orders-admin-payment-plans-view__action-3',
              value ? 'views-admin-orders-admin-payment-plans-view__action-4' : 'views-admin-orders-admin-payment-plans-view__action-5'
            ]"
            @click="toggleForSale(row)"
          >
            <span :class="[
              'views-admin-orders-admin-payment-plans-view__text-10',
              value ? 'toggle-thumb--on' : 'views-admin-orders-admin-payment-plans-view__text-11'
            ]" />
          </button>
        </template>
        <template #cell-actions="{ row }">
          <div class="views-admin-orders-admin-payment-plans-view__panel-3">
            <button @click="openPlanEdit(row)" class="views-admin-orders-admin-payment-plans-view__action">
              <Icon name="edit" size="sm" />
              <span class="views-admin-orders-admin-payment-plans-view__text-9">{{ t('common.edit') }}</span>
            </button>
            <button @click="confirmDeletePlan(row)" class="views-admin-orders-admin-payment-plans-view__action-2">
              <Icon name="trash" size="sm" />
              <span class="views-admin-orders-admin-payment-plans-view__text-9">{{ t('common.delete') }}</span>
            </button>
          </div>
        </template>
      </DataTable>
    </div>

    <!-- Plan Edit Dialog -->
    <PlanEditDialog :show="showPlanDialog" :plan="editingPlan" :groups="groups" @close="showPlanDialog = false" @saved="loadPlans" />

    <ConfirmDialog :show="showDeletePlanDialog" :title="t('payment.admin.deletePlan')" :message="t('payment.admin.deletePlanConfirm')" :confirm-text="t('common.delete')" danger @confirm="handleDeletePlan" @cancel="showDeletePlanDialog = false" />
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractI18nErrorMessage } from '@/utils/apiError'
import adminAPI from '@/api/admin'
import type { SubscriptionPlan } from '@/types/payment'
import type { AdminGroup } from '@/types'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import PlanEditDialog from './PlanEditDialog.vue'
import { formatPoints } from '@/utils/format'
import { platformTextClass } from '@/utils/platformColors'

const { t, locale } = useI18n()
const localeCode = computed(() => String(locale?.value || ''))
const appStore = useAppStore()

// ==================== Groups ====================

const groups = ref<AdminGroup[]>([])

async function loadGroups() {
  try {
    groups.value = await adminAPI.groups.getAll()
  } catch { /* ignore */ }
}

function getGroup(id: string): AdminGroup | undefined {
  return groups.value.find(g => g.id === id)
}

function isGroupMissing(id: string): boolean {
  return Boolean(id) && !groups.value.find(g => g.id === id)
}

function getPlanNameClass(groupId: string): string {
  const group = getGroup(groupId)
  return group ? platformTextClass(group.platform) : 'views-admin-orders-admin-payment-plans-view__state'
}


// ==================== Plans ====================

const plansLoading = ref(true)
const plans = ref<SubscriptionPlan[]>([])
const showPlanDialog = ref(false)
const showDeletePlanDialog = ref(false)
const editingPlan = ref<SubscriptionPlan | null>(null)
const deletingPlanId = ref<string | null>(null)

const planColumns = computed((): Column[] => [
  { key: 'id', label: 'ID' },
  { key: 'name', label: t('payment.admin.planName') },
  { key: 'group_id', label: t('payment.admin.group') },
  { key: 'price', label: t('payment.admin.price') },
  { key: 'validity_days', label: t('payment.admin.validity') },
  { key: 'stock_available', label: t('payment.admin.stockQuantity') },
  { key: 'for_sale', label: t('payment.admin.forSale') },
  { key: 'sort_order', label: t('payment.admin.sortOrder') },
  { key: 'actions', label: t('common.actions') },
])

async function loadPlans() {
  plansLoading.value = true
  try {
    const res = await adminPaymentAPI.getPlans()
    // Backend returns features as newline-separated string; parse to array
    plans.value = (res.data || []).map((p: Omit<SubscriptionPlan, 'features'> & { features: string | string[] }) => ({
      ...p,
      features: typeof p.features === 'string'
        ? p.features.split('\n').map((f: string) => f.trim()).filter(Boolean)
        : (p.features || []),
    }))
  }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
  finally { plansLoading.value = false }
}

function openPlanEdit(plan: SubscriptionPlan | null) {
  editingPlan.value = plan
  showPlanDialog.value = true
}


/** Quick toggle for_sale from the list */
async function toggleForSale(plan: SubscriptionPlan) {
  try {
    await adminPaymentAPI.updatePlan(plan.id, { for_sale: !plan.for_sale })
    plan.for_sale = !plan.for_sale
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error')))
  }
}

function confirmDeletePlan(plan: SubscriptionPlan) { deletingPlanId.value = plan.id; showDeletePlanDialog.value = true }
async function handleDeletePlan() {
  if (!deletingPlanId.value) return
  try { await adminPaymentAPI.deletePlan(deletingPlanId.value); appStore.showSuccess(t('common.deleted')); showDeletePlanDialog.value = false; loadPlans() }
  catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
}

// ==================== Lifecycle ====================

onMounted(() => {
  loadGroups()
  loadPlans()
})
</script>
