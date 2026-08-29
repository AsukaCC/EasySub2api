<template>
  <BaseDialog :show="show" :title="plan ? t('payment.admin.editPlan') : t('payment.admin.createPlan')" width="wide" @close="emit('close')">
    <form id="plan-form" @submit.prevent="handleSavePlan" class="views-admin-orders-plan-edit-dialog__form">
      <div class="views-admin-orders-plan-edit-dialog__panel">
        <div>
          <label class="input-label">{{ t('payment.admin.planName') }} <span class="views-admin-orders-plan-edit-dialog__text">*</span></label>
          <input v-model="planForm.name" type="text" class="input" required />
        </div>
        <div>
          <label class="input-label">{{ t('payment.admin.group') }} <span class="views-admin-orders-plan-edit-dialog__text">*</span></label>
          <Select v-model="planForm.group_id" :options="groupOptions" :placeholder="t('payment.admin.selectGroup')" class="views-admin-orders-plan-edit-dialog__field">
            <template #selected="{ option }">
              <span v-if="option?.platform" :class="platformTextClass(String(option.platform))">{{ option.label }}</span>
              <span v-else>{{ option?.label || t('payment.admin.selectGroup') }}</span>
            </template>
            <template #option="{ option, selected }">
              <span class="views-admin-orders-plan-edit-dialog__text-2" :class="option.platform ? platformTextClass(String(option.platform)) : ''">{{ option.label }}</span>
              <Icon v-if="selected" name="check" size="sm" class="views-admin-orders-plan-edit-dialog__icon" :stroke-width="2" />
            </template>
          </Select>
        </div>
      </div>

      <!-- Group Info Preview -->
      <div v-if="selectedGroupInfo" class="views-admin-orders-plan-edit-dialog__panel-2">
        <div class="views-admin-orders-plan-edit-dialog__panel-3">
          <GroupBadge :name="selectedGroupInfo.name" :platform="selectedGroupInfo.platform" :rate-multiplier="selectedGroupInfo.rate_multiplier" />
        </div>
        <div class="views-admin-orders-plan-edit-dialog__panel-4">
          <div><span class="views-admin-orders-plan-edit-dialog__text-3">{{ t('payment.admin.dailyLimit') }}:</span> <span class="views-admin-orders-plan-edit-dialog__text-4">{{ groupLimitPoints(selectedGroupInfo.daily_limit_points, selectedGroupInfo.daily_limit_usd) }}</span></div>
          <div><span class="views-admin-orders-plan-edit-dialog__text-3">{{ t('payment.admin.weeklyLimit') }}:</span> <span class="views-admin-orders-plan-edit-dialog__text-4">{{ groupLimitPoints(selectedGroupInfo.weekly_limit_points, selectedGroupInfo.weekly_limit_usd) }}</span></div>
          <div><span class="views-admin-orders-plan-edit-dialog__text-3">{{ t('payment.admin.monthlyLimit') }}:</span> <span class="views-admin-orders-plan-edit-dialog__text-4">{{ groupLimitPoints(selectedGroupInfo.monthly_limit_points, selectedGroupInfo.monthly_limit_usd) }}</span></div>
        </div>
      </div>

      <div><label class="input-label">{{ t('payment.admin.planDescription') }} <span class="views-admin-orders-plan-edit-dialog__text">*</span></label><textarea v-model="planForm.description" rows="2" class="input" required></textarea></div>
      <div class="views-admin-orders-plan-edit-dialog__panel">
        <div>
          <label class="input-label">{{ t('payment.admin.price') }} <span class="views-admin-orders-plan-edit-dialog__text">*</span></label>
          <input v-model.number="planForm.price" type="number" step="0.00000001" min="0.00000001" class="input" required />
          <p class="views-admin-orders-plan-edit-dialog__description">{{ t('payment.admin.planPointsHint') }}</p>
        </div>
        <div><label class="input-label">{{ t('payment.admin.originalPrice') }}</label><input v-model.number="planForm.original_price" type="number" step="0.00000001" min="0" class="input" /></div>
      </div>
      <div class="views-admin-orders-plan-edit-dialog__panel">
        <div><label class="input-label">{{ t('payment.admin.validity') }} <span class="views-admin-orders-plan-edit-dialog__text">*</span></label><input v-model.number="planForm.validity_days" type="number" min="1" class="input" required /></div>
        <div><label class="input-label">{{ t('payment.admin.validityUnit') }} <span class="views-admin-orders-plan-edit-dialog__text">*</span></label><Select v-model="planForm.validity_unit" :options="validityUnitOptions" /></div>
      </div>
      <div><label class="input-label">{{ t('payment.admin.sortOrder') }}</label><input v-model.number="planForm.sort_order" type="number" min="0" class="input" /></div>
      <div>
        <label class="input-label">{{ t('payment.admin.features') }}</label>
        <textarea v-model="planFeaturesText" rows="3" class="input" :placeholder="t('payment.admin.featuresPlaceholder')"></textarea>
        <p class="views-admin-orders-plan-edit-dialog__description-2">{{ t('payment.admin.featuresHint') }}</p>
      </div>
      <div class="views-admin-orders-plan-edit-dialog__panel-5">
        <label class="views-admin-orders-plan-edit-dialog__label">{{ t('payment.admin.forSale') }}</label>
        <button
          type="button"
          :class="[
            'views-admin-orders-plan-edit-dialog__action',
            planForm.for_sale ? 'views-admin-orders-plan-edit-dialog__action-2' : 'views-admin-orders-plan-edit-dialog__action-3'
          ]"
          @click="planForm.for_sale = !planForm.for_sale"
        >
          <span :class="[
            'views-admin-orders-plan-edit-dialog__text-5',
            planForm.for_sale ? 'toggle-thumb--on' : 'views-admin-orders-plan-edit-dialog__text-6'
          ]" />
        </button>
      </div>
    </form>
    <template #footer>
      <div class="views-admin-orders-plan-edit-dialog__panel-6">
        <button type="button" @click="emit('close')" class="btn btn-secondary">{{ t('common.cancel') }}</button>
        <button type="submit" form="plan-form" :disabled="saving" class="btn btn-primary">{{ saving ? t('common.saving') : t('common.save') }}</button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminPaymentAPI } from '@/api/admin/payment'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatPoints } from '@/utils/format'
import type { SubscriptionPlan } from '@/types/payment'
import type { AdminGroup } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'
import { platformTextClass } from '@/utils/platformColors'

const props = defineProps<{
  show: boolean
  plan: SubscriptionPlan | null
  groups: AdminGroup[]
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const { t, locale } = useI18n()
const localeCode = computed(() => String(locale?.value || ''))
const appStore = useAppStore()

const saving = ref(false)
const planForm = reactive({ name: '', group_id: null as string | null, description: '', price: 0, original_price: 0, validity_days: 30, validity_unit: 'days', sort_order: 0, for_sale: true })
const planFeaturesText = ref('')

const validityUnitOptions = computed(() => [
  { value: 'days', label: t('payment.admin.days') },
  { value: 'weeks', label: t('payment.admin.weeks') },
  { value: 'months', label: t('payment.admin.months') },
])

const groupOptions = computed(() =>
  props.groups
    .filter(g => g.subscription_type === 'subscription')
    .map(g => ({
      value: g.id,
      label: `${g.name} — ${g.platform} (${g.rate_multiplier}x)`,
      platform: g.platform,
    })),
)

const selectedGroupInfo = computed(() => {
  if (!planForm.group_id) return null
  return props.groups.find(g => g.id === planForm.group_id) || null
})

function groupLimitPoints(points: number | null | undefined, legacy: number | null | undefined): string {
  const value = points ?? legacy
  return value == null ? t('payment.admin.unlimited') : formatPoints(value, localeCode.value)
}

// Reset form when dialog opens
watch(() => props.show, (visible) => {
  if (!visible) return
  if (props.plan) {
    Object.assign(planForm, { name: props.plan.name, group_id: props.plan.group_id, description: props.plan.description, price: props.plan.price_points ?? props.plan.price, original_price: props.plan.original_price_points ?? props.plan.original_price ?? 0, validity_days: props.plan.validity_days, validity_unit: props.plan.validity_unit || 'days', sort_order: props.plan.sort_order || 0, for_sale: props.plan.for_sale })
    planFeaturesText.value = (props.plan.features || []).join('\n')
  } else {
    Object.assign(planForm, { name: '', group_id: null, description: '', price: 0, original_price: 0, validity_days: 30, validity_unit: 'days', sort_order: 0, for_sale: true })
    planFeaturesText.value = ''
  }
})

/** Build request payload with snake_case keys matching backend JSON tags */
function buildPlanPayload() {
  const features = planFeaturesText.value.split('\n').map(f => f.trim()).filter(Boolean).join('\n')
  return {
    name: planForm.name,
    group_id: planForm.group_id,
    description: planForm.description,
    price: planForm.price,
    original_price: planForm.original_price || 0,
    validity_days: planForm.validity_days,
    validity_unit: planForm.validity_unit,
    sort_order: planForm.sort_order,
    for_sale: planForm.for_sale,
    features,
  }
}

async function handleSavePlan() {
  if (!planForm.group_id) {
    appStore.showError(t('payment.admin.groupRequired'))
    return
  }
  if (!planForm.price || planForm.price <= 0) {
    appStore.showError(t('payment.admin.priceRequired'))
    return
  }
  if (!planForm.validity_days || planForm.validity_days < 1) {
    appStore.showError(t('payment.admin.validityRequired'))
    return
  }
  saving.value = true
  try {
    const data = buildPlanPayload()
    if (props.plan) { await adminPaymentAPI.updatePlan(props.plan.id, data) }
    else { await adminPaymentAPI.createPlan(data) }
    appStore.showSuccess(t('common.saved'))
    emit('close')
    emit('saved')
  } catch (err: unknown) { appStore.showError(extractApiErrorMessage(err, t('common.error'))) }
  finally { saving.value = false }
}
</script>
