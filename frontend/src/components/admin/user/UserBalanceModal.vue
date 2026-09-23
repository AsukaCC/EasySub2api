<template>
  <BaseDialog :show="show" :title="operation === 'add' ? t('admin.users.deposit') : t('admin.users.withdraw')" width="narrow" @close="$emit('close')">
    <form v-if="user" id="balance-form" @submit.prevent="handleBalanceSubmit" class="components-admin-user-user-balance-modal__form">
      <div class="components-admin-user-user-balance-modal__panel">
        <div class="components-admin-user-user-balance-modal__panel-2"><span class="components-admin-user-user-balance-modal__text">{{ user.email.charAt(0).toUpperCase() }}</span></div>
		<div class="components-admin-user-user-balance-modal__panel-3"><p class="components-admin-user-user-balance-modal__description">{{ user.email }}</p><p class="components-admin-user-user-balance-modal__description-2">{{ t('admin.users.currentBalance') }}: {{ formatPoints(user.available_balance ?? 0) }}</p></div>
      </div>
      <div>
        <label for="balance-amount" class="input-label">{{ operation === 'add' ? t('admin.users.depositAmount') : t('admin.users.withdrawAmount') }} ({{ t('common.points') }})</label>
        <div class="components-admin-user-user-balance-modal__panel-4">
          <div class="components-admin-user-user-balance-modal__panel-5">
            <div class="components-admin-user-user-balance-modal__panel-6"><Icon name="points" size="sm" /></div>
            <input id="balance-amount" v-model.number="form.amount" type="number" step="any" min="0" required class="components-admin-user-user-balance-modal__field input" />
          </div>
          <button v-if="operation === 'subtract'" type="button" @click="fillAllBalance" class="components-admin-user-user-balance-modal__action btn btn-secondary">{{ t('admin.users.withdrawAll') }}</button>
        </div>
      </div>
	  <div>
		<label class="input-label">{{ t('admin.users.balanceType') }}</label>
		<Select v-model="form.balanceType" :options="[
		  { value: 'recharge', label: t('admin.users.rechargeBalance') },
		  { value: 'bonus', label: t('admin.users.bonusBalance') }
		]" />
	  </div>
	  <div v-if="operation === 'add' && form.balanceType === 'bonus'">
		<label class="input-label">{{ t('admin.users.bonusValidityDays') }}</label>
		<input v-model.number="form.bonusValidityDays" type="number" min="1" max="3650" class="input" required />
	  </div>
      <div><label class="input-label">{{ t('admin.users.notes') }}</label><textarea v-model="form.notes" rows="3" class="input"></textarea></div>
      <div v-if="isRecharge && tiersLoading" role="status">{{ t('common.loading') }}</div>
      <div v-else-if="isRecharge && tiersFailed" role="alert">
        <p>{{ t('admin.users.rechargeBonusLoadFailed') }}</p>
        <button type="button" class="btn btn-secondary" @click="loadBonusTiers">{{ t('common.retry') }}</button>
      </div>
      <div v-if="form.amount > 0 && (!isRecharge || (!tiersLoading && !tiersFailed))" class="components-admin-user-user-balance-modal__panel-7 recharge-summary">
        <div v-if="isRecharge" class="recharge-summary__row">
          <span class="components-admin-user-user-balance-modal__text-2">
            {{ t('admin.users.rechargeTierBonus') }}:
            <small v-if="bonusPoints > 0" class="recharge-summary__validity">{{ t('admin.users.rechargeTierBonusValidity') }}</small>
          </span>
          <span class="components-admin-user-user-balance-modal__text-3 recharge-summary__amount">{{ formatPoints(bonusPoints) }}</span>
        </div>
        <div class="recharge-summary__row"><span class="components-admin-user-user-balance-modal__text-2">{{ t('admin.users.newBalance') }}:</span><span class="components-admin-user-user-balance-modal__text-3 recharge-summary__amount">{{ formatPoints(calculateNewBalance()) }}</span></div>
      </div>
    </form>
    <template #footer>
      <div class="components-admin-user-user-balance-modal__panel-9">
        <button @click="$emit('close')" class="btn btn-secondary">{{ t('common.cancel') }}</button>
        <button type="submit" form="balance-form" :disabled="submitting || !form.amount || (isRecharge && (tiersLoading || tiersFailed))" class="btn" :class="operation === 'add' ? 'components-admin-user-user-balance-modal__action-2' : 'btn-danger'">{{ submitting ? t('common.saving') : t('common.confirm') }}</button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { extractApiErrorMessage } from '@/utils/apiError'
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import { adminPaymentAPI } from '@/api/admin/payment'
import { rechargeBonusPointsForAmount } from '@/components/payment/rechargeBonus'
import type { RechargeBonusTier } from '@/types/payment'
import type { AdminUser } from '@/types'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatPoints } from '@/utils/format'

const props = defineProps<{ show: boolean, user: AdminUser | null, operation: 'add' | 'subtract' }>()
const emit = defineEmits(['close', 'success']); const { t } = useI18n(); const appStore = useAppStore()

const submitting = ref(false); const form = reactive({ amount: 0, notes: '', balanceType: 'recharge' as 'recharge' | 'bonus', bonusValidityDays: 90 })
const bonusTiers = ref<RechargeBonusTier[]>([])
const tiersLoading = ref(false)
const tiersFailed = ref(false)
const isRecharge = computed(() => props.operation === 'add' && form.balanceType === 'recharge')
const bonusPoints = computed(() => isRecharge.value ? rechargeBonusPointsForAmount(bonusTiers.value, form.amount) : 0)
let tiersRequest = 0

async function loadBonusTiers() {
  const request = ++tiersRequest
  tiersLoading.value = true
  tiersFailed.value = false
  bonusTiers.value = []
  try {
    const { data } = await adminPaymentAPI.getConfig()
    if (request === tiersRequest) bonusTiers.value = data.recharge_bonus_tiers ?? []
  } catch {
    if (request === tiersRequest) tiersFailed.value = true
  } finally {
    if (request === tiersRequest) tiersLoading.value = false
  }
}

watch(() => [props.show, props.operation, props.user?.id] as const, ([show]) => {
  ++tiersRequest
  if (show) {
    form.amount = 0
    form.notes = ''
    form.balanceType = 'recharge'
    form.bonusValidityDays = 90
    if (props.operation === 'add') void loadBonusTiers()
  }
}, { immediate: true })

// 填入全部余额
const fillAllBalance = () => {
  if (props.user) {
		form.amount = form.balanceType === 'bonus' ? (props.user.bonus_balance ?? 0) : (props.user.recharge_balance ?? 0)
  }
}

const calculateNewBalance = () => {
  if (!props.user) return 0
	const current = form.balanceType === 'bonus' ? (props.user.bonus_balance ?? 0) : (props.user.recharge_balance ?? 0)
	const result = props.operation === 'add' ? current + form.amount + bonusPoints.value : current - form.amount
  // 避免浮点数精度问题导致的 -0.00 显示
  return Math.abs(result) < 1e-10 ? 0 : result
}
const handleBalanceSubmit = async () => {
  if (!props.user || submitting.value || (isRecharge.value && (tiersLoading.value || tiersFailed.value))) return
  if (!Number.isFinite(form.amount) || form.amount <= 0) {
    appStore.showError(t('admin.users.amountRequired'))
    return
  }
  // 退款时验证金额不超过实际余额
	const available = form.balanceType === 'bonus' ? (props.user.bonus_balance ?? 0) : (props.user.recharge_balance ?? 0)
	if (props.operation === 'subtract' && form.amount > Math.max(available, 0)) {
    appStore.showError(t('admin.users.insufficientBalance'))
    return
  }
  submitting.value = true
  try {
	await adminAPI.users.updateBalance(props.user.id, form.amount, props.operation, form.notes, form.balanceType, form.bonusValidityDays)
    appStore.showSuccess(t('common.success')); emit('success'); emit('close')
  } catch (e: any) {
    console.error('Failed to update balance:', e)
    appStore.showError(extractApiErrorMessage(e, t('common.error')))
  } finally { submitting.value = false }
}
</script>

<style scoped>
.recharge-summary {
  display: grid;
  gap: 0.5rem;
}

.recharge-summary__row {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.25rem 0.75rem;
}

.recharge-summary__amount {
  max-width: 100%;
  margin-left: auto;
  overflow-wrap: anywhere;
  text-align: right;
}

.recharge-summary__validity {
  display: block;
  font-size: 0.75rem;
}
</style>
